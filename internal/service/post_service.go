package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"mime/multipart"
	"time"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
	"github.com/maheshrc27/postflow/internal/models"
	"github.com/maheshrc27/postflow/internal/repository"
	"github.com/maheshrc27/postflow/internal/transfer"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type PostService interface {
	CreatePost(ctx context.Context, userID int64, pc *transfer.PostCreation, files []*multipart.FileHeader) (int64, time.Duration, error)
	List(ctx context.Context, userID int64) ([]*models.Post, error)
	PostInfo(ctx context.Context, postID, userID int64) (*models.Post, error)
	Remove(ctx context.Context, userID, postID int64) error
}

type postService struct {
	db *sql.DB
	pr repository.PostRepository
	sa repository.SelectedAccountRepository
	ac repository.SocialAccountRepository
	ma repository.MediaAssetRepository
	pm repository.PostMediaRepository
	r2 R2Service
}

func NewPostService(
	db *sql.DB,
	pr repository.PostRepository,
	sa repository.SelectedAccountRepository,
	ma repository.MediaAssetRepository,
	ac repository.SocialAccountRepository,
	pm repository.PostMediaRepository,
	r2 R2Service) PostService {
	return &postService{
		db: db,
		pr: pr,
		sa: sa,
		ac: ac,
		ma: ma,
		pm: pm,
		r2: r2,
	}
}

func (s *postService) CreatePost(ctx context.Context, userID int64, pc *transfer.PostCreation, files []*multipart.FileHeader) (int64, time.Duration, error) {
	var err error
	if pc == nil {
		err = errors.New("nil value for parameter: creation")
		slog.Error(err.Error())
		return 0, 0, err
	}

	if pc.Caption == "" {
		err = errors.New("Caption is empty")
		slog.Info(err.Error())
		return 0, 0, err
	}

	datetime, err := time.Parse("2006-01-02T15:04", pc.ScheduledTime)
	if err != nil {
		slog.Error(err.Error())
		return 0, 0, fmt.Errorf("Unable to parse scheduling time")
	}

	var selectedAccounts []int
	if err := json.Unmarshal([]byte(pc.SelectedAccounts), &selectedAccounts); err != nil {
		slog.Error(err.Error())
		return 0, 0, fmt.Errorf("Invalid format for selected social accounts")
	}

	if files == nil || len(files) == 0 {
		err = errors.New("file is nil or empty")
		slog.Error(err.Error())
		return 0, 0, err
	}

	var postType string
	if len(files) > 1 {
		postType = PostTypeMultiple
	} else {
		postType = PostTypeSingle
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, 0, fmt.Errorf("Failed to start transaction: %w", err)
	}
	defer func() {
		// Rollback if an error occurred
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	post := models.Post{
		UserID:        userID,
		PostType:      postType,
		Caption:       pc.Caption,
		Title:         pc.Title,
		ScheduledTime: datetime,
		Status:        PostStatusScheduled,
	}

	postID, err := s.pr.Create(ctx, tx, &post)
	if err != nil {
		return 0, 0, fmt.Errorf("Error creating post: %w", err)
	}

	for _, accountID := range selectedAccounts {
		isExist, err := s.ac.CheckByUserID(ctx, int64(accountID), userID)
		if err != nil {
			return 0, 0, fmt.Errorf("Error checking social account: %w", err)
		}
		if !isExist {
			return 0, 0, fmt.Errorf("Social account doesn't exist")
		}

		account := models.SelectedAccount{
			PostID:    postID,
			AccountID: int64(accountID),
		}
		if err := s.sa.Create(ctx, tx, &account); err != nil {
			return 0, 0, fmt.Errorf("Error saving selected account: %w", err)
		}
	}

	allowedTypes := map[string]struct{}{
		"mp4":  {},
		"mov":  {},
		"jpeg": {},
	}

	for i, file := range files {
		fileContent, err := file.Open()
		if err != nil {
			return 0, 0, fmt.Errorf("Error opening file: %w", err)
		}
		defer fileContent.Close()

		fileBytes, err := io.ReadAll(fileContent)
		if err != nil {
			return 0, 0, fmt.Errorf("Error reading file content: %w", err)
		}

		fileType, err := filetype.Match(fileBytes)
		if err != nil || fileType == types.Unknown {
			return 0, 0, fmt.Errorf("Unsupported file type")
		}

		if _, ok := allowedTypes[fileType.Extension]; !ok {
			return 0, 0, fmt.Errorf("File type %s is not allowed", fileType.Extension)
		}

		assetID, err := s.saveFile(ctx, tx, userID, fileType.MIME.Value, fileBytes)
		if err != nil {
			return 0, 0, fmt.Errorf("Error uploading file: %w", err)
		}

		pm := models.PostMedia{
			PostID:       postID,
			AssetID:      assetID,
			DisplayOrder: i,
		}
		if err := s.pm.Create(ctx, tx, &pm); err != nil {
			return 0, 0, fmt.Errorf("Error saving media file: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, 0, fmt.Errorf("Failed to commit transaction: %w", err)
	}

	delay := time.Until(datetime)
	if delay < 0 {
		delay = 0
	}

	return postID, delay, nil
}

func (s *postService) saveFile(ctx context.Context, tx *sql.Tx, userID int64, fileType string, file []byte) (int64, error) {
	id, err := gonanoid.New()
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}
	err = s.r2.UploadToR2(ctx, id, file, fileType)
	if err != nil {
		fmt.Println(err.Error())
		return 0, err
	}

	ma := models.MediaAsset{
		UserID:   userID,
		FileName: id,
		FileType: fileType,
		FileURL:  fmt.Sprintf("https://pub-f8f43aa198a449518df6744ec9ce452c.r2.dev/%s", id),
	}

	assetID, err := s.ma.Create(ctx, tx, &ma)
	if err != nil {
		return 0, err
	}

	return assetID, nil
}

func (s *postService) PostInfo(ctx context.Context, postID, userID int64) (*models.Post, error) {
	var err error

	if userID == 0 {
		err = errors.New("User is not valid")
		slog.Info(err.Error())
		return nil, err
	}

	if postID == 0 {
		err = errors.New("post id is not valid")
		slog.Info(err.Error())
		return nil, err
	}

	isValid, err := s.pr.CheckByUserID(ctx, postID, userID)
	if err != nil {
		return nil, err
	}

	if !isValid {
		err = errors.New("Post doesn't exist")
		slog.Info(err.Error())
		return nil, err
	}

	post, err := s.pr.GetByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("Error getting post info")
	}

	return post, nil
}

func (s *postService) List(ctx context.Context, userID int64) ([]*models.Post, error) {
	posts, err := s.pr.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("Error getting API keys")
	}
	return posts, nil
}

func (s *postService) Remove(ctx context.Context, userID, postID int64) error {
	var err error

	if userID == 0 {
		err = errors.New("User is not valid")
		slog.Info(err.Error())
		return err
	}

	if postID == 0 {
		err = errors.New("post_id is not valid")
		slog.Info(err.Error())
		return err
	}

	isValid, err := s.pr.CheckByUserID(ctx, postID, userID)
	if err != nil {
		return err
	}

	if !isValid {
		err = errors.New("Post doesn't exist")
		slog.Info(err.Error())
		return err
	}

	err = s.pr.Remove(ctx, postID)
	if err != nil {
		return fmt.Errorf("Error removing post")
	}

	return nil
}
