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
	"github.com/maheshrc27/scheduling-api/internal/models"
	"github.com/maheshrc27/scheduling-api/internal/repository"
	"github.com/maheshrc27/scheduling-api/internal/transfer"
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
	// Validate input parameters
	if pc == nil {
		err := errors.New("post creation data is nil")
		slog.Error(err.Error())
		return 0, 0, err
	}
	if pc.Caption == "" {
		err := errors.New("caption cannot be empty")
		slog.Info(err.Error())
		return 0, 0, err
	}

	// Parse scheduled time
	scheduledTime, err := time.Parse("2006-01-02T15:04", pc.ScheduledTime)
	if err != nil {
		err = fmt.Errorf("invalid scheduled time format: %w", err)
		slog.Error(err.Error())
		return 0, 0, err
	}

	// Parse selected accounts
	var selectedAccounts []int
	if err := json.Unmarshal([]byte(pc.SelectedAccounts), &selectedAccounts); err != nil {
		err = fmt.Errorf("invalid selected accounts format: %w", err)
		slog.Error(err.Error())
		return 0, 0, err
	}
	if len(selectedAccounts) == 0 {
		err := errors.New("no social accounts selected")
		slog.Error(err.Error())
		return 0, 0, err
	}

	// Validate files
	if len(files) == 0 {
		err := errors.New("no files provided for the post")
		slog.Error(err.Error())
		return 0, 0, err
	}

	postType := PostTypeSingle
	if len(files) > 1 {
		postType = PostTypeMultiple
	}

	// Begin database transaction
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	// Create post
	post := models.Post{
		UserID:        userID,
		PostType:      postType,
		Caption:       pc.Caption,
		Title:         pc.Title,
		ScheduledTime: scheduledTime,
		Status:        PostStatusScheduled,
	}

	postID, err := s.pr.Create(ctx, tx, &post)
	if err != nil {
		return 0, 0, fmt.Errorf("error creating post: %w", err)
	}

	// Validate and save selected accounts
	if err := s.saveSelectedAccounts(ctx, tx, userID, postID, selectedAccounts); err != nil {
		return 0, 0, fmt.Errorf("error processing selected accounts: %w", err)
	}

	// Process and save files
	if err := s.processFiles(ctx, tx, userID, postID, files); err != nil {
		return 0, 0, fmt.Errorf("error processing files: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return 0, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	delay := time.Until(scheduledTime)
	if delay < 0 {
		delay = 0
	}

	return postID, delay, nil
}

func (s *postService) saveSelectedAccounts(ctx context.Context, tx *sql.Tx, userID, postID int64, accounts []int) error {
	for _, accountID := range accounts {
		exists, err := s.ac.CheckByUserID(ctx, int64(accountID), userID)
		if err != nil {
			return fmt.Errorf("error checking social account %d: %w", accountID, err)
		}
		if !exists {
			return fmt.Errorf("social account %d does not exist", accountID)
		}

		account := models.SelectedAccount{
			PostID:    postID,
			AccountID: int64(accountID),
		}
		if err := s.sa.Create(ctx, tx, &account); err != nil {
			return fmt.Errorf("error saving selected account %d: %w", accountID, err)
		}
	}
	return nil
}

func (s *postService) processFiles(ctx context.Context, tx *sql.Tx, userID, postID int64, files []*multipart.FileHeader) error {
	allowedTypes := map[string]struct{}{
		"mp4": {}, "mov": {}, "jpeg": {}, "png": {}, "jpg": {},
	}

	for i, file := range files {
		fileContent, err := file.Open()
		if err != nil {
			return fmt.Errorf("error opening file: %w", err)
		}
		defer fileContent.Close()

		fileBytes, err := io.ReadAll(fileContent)
		if err != nil {
			return fmt.Errorf("error reading file content: %w", err)
		}

		fileType, err := filetype.Match(fileBytes)
		if err != nil || fileType == types.Unknown {
			return fmt.Errorf("unsupported file type: %w", err)
		}
		if _, ok := allowedTypes[fileType.Extension]; !ok {
			return fmt.Errorf("file type %s is not allowed", fileType.Extension)
		}

		assetID, err := s.saveFile(ctx, tx, userID, fileType.MIME.Value, fileBytes)
		if err != nil {
			return fmt.Errorf("error uploading file: %w", err)
		}

		postMedia := models.PostMedia{
			PostID:       postID,
			AssetID:      assetID,
			DisplayOrder: i,
		}
		if err := s.pm.Create(ctx, tx, &postMedia); err != nil {
			return fmt.Errorf("error saving media file: %w", err)
		}
	}
	return nil
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
