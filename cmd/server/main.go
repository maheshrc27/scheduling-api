package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	config "github.com/maheshrc27/postflow/configs"
	"github.com/maheshrc27/postflow/internal/api/handlers"
	"github.com/maheshrc27/postflow/internal/api/middleware"
	job "github.com/maheshrc27/postflow/internal/jobs"
	"github.com/maheshrc27/postflow/internal/queue"
	"github.com/maheshrc27/postflow/internal/repository"
	"github.com/maheshrc27/postflow/internal/service"
	"github.com/robfig/cron"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Failed to load environment variables", err)
	}

	cfg := config.LoadConfig()

	db, err := sql.Open("postgres", cfg.PostgresURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(db)

	if err := db.Ping(); err != nil {
		log.Fatalf("Database is unreachable: %v", err)
	}

	redisConn := asynq.RedisClientOpt{Addr: cfg.RedisURI}
	client := asynq.NewClient(redisConn)
	defer client.Close()

	app := fiber.New(fiber.Config{
		ReadTimeout:  10 * time.Minute,
		WriteTimeout: 10 * time.Minute,
		BodyLimit:    100 * 1024 * 1024, // 100 MB
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Printf("Error: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			return true
		},
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
		MaxAge:           3600,
	}))

	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	socialAccountRepo := repository.NewSocialAccountRepository(db)
	postMediaRepo := repository.NewPostMediaRepository(db)
	selectedAccountRepo := repository.NewSelectedAccountRepository(db)
	mediaAssetRepo := repository.NewMediaAssetRepository(db)
	settingsRepository := repository.NewSettingsRepository(db)
	apiKeyRepository := repository.NewApiKeyRepository(db)

	authService := service.NewAuthService(*cfg, userRepo)
	userService := service.NewUserService(userRepo)
	r2Service := service.NewR2Service(*cfg)
	postService := service.NewPostService(db, postRepo, selectedAccountRepo, mediaAssetRepo, socialAccountRepo, postMediaRepo, *r2Service)
	platformService := service.NewPlatformService(*cfg, socialAccountRepo)
	instagramService := service.NewInstagramService(*cfg, socialAccountRepo, postRepo, postMediaRepo, mediaAssetRepo)
	tiktokService := service.NewTiktokService(*cfg, postRepo, socialAccountRepo, postMediaRepo, mediaAssetRepo)
	youtbeService := service.NewYoutubeService(*cfg, postRepo, socialAccountRepo, postMediaRepo, mediaAssetRepo)
	settingsService := service.NewSettingsService(settingsRepository)
	apiKeyService := service.NewApiKeyService(apiKeyRepository)

	authMiddleware := middleware.NewAuthMiddleware(*cfg, apiKeyService)

	auth := handlers.NewAuthHandler(*cfg, authService)
	app.Get("/login", auth.Login)
	app.Get("/login/callback", auth.LoginCallbackHandler)

	platform := handlers.NewPlatformHandler(platformService, instagramService, tiktokService, youtbeService, *cfg)
	app.Get("/auth/:platform", platform.AddSocialAccount)
	app.Get("/auth/:platform/callback", platform.CallbackHandler)

	api := app.Group("/api")
	api.Use(authMiddleware.AuthMiddleware())

	user := handlers.NewUserHandler(userService)
	api.Get("/user/info", user.GetUserInfo)

	settings := handlers.NewSettingsHandler(settingsService)
	api.Get("/settings/info", settings.GetSettingsInfo)
	api.Post("/settings/update", settings.UpdateSettings)

	apiKeys := handlers.NewApiKeyHandler(apiKeyService)
	api.Post("/api_key/new", apiKeys.CreateApiKey)
	api.Get("/api_key/list", apiKeys.ListKeys)
	api.Post("/api_key/remove", apiKeys.RemoveAPIKey)

	post := handlers.NewPostHandler(postService, client)
	api.Post("/posts/create", post.CreatePost)
	api.Get("/posts", post.ListPosts)
	api.Post("/posts/remove", post.RemovePost)

	// social accounts api routes
	api.Get("/accounts", platform.ListSocialAccounts)
	api.Post("/accounts/remove", platform.DeleteSocialAccount)

	// cron jobs
	refreshTokenJob := job.NewtokenRefreshJob(socialAccountRepo, youtbeService, tiktokService, instagramService)

	//queue
	queueW := queue.NewQueue(postRepo, selectedAccountRepo, mediaAssetRepo, socialAccountRepo, postMediaRepo, youtbeService, tiktokService, instagramService)

	c := cron.New()
	c.AddFunc("@every 00h10m00s", refreshTokenJob.RefreshTokens)
	c.Start()

	go func() {
		server := asynq.NewServer(redisConn, asynq.Config{
			Concurrency: 10,
		})

		mux := asynq.NewServeMux()
		mux.HandleFunc(queue.TaskTypeSchedulePost, queueW.HandleSchedulePostTask)

		log.Println("Starting the Asynq server...")
		if err := server.Run(mux); err != nil {
			log.Fatalf("Could not start Asynq server: %v", err)
		}
	}()

	go func() {
		if err := app.Listen(":3000"); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	log.Println("Server is running on http://localhost:3000")

	gracefulShutdown(app, db)
}

func closeDB(db *sql.DB) {
	fmt.Fprint(os.Stdout, "Closing database connection... ")
	if err := db.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to close database: %v", err)
		return
	}
	fmt.Fprintln(os.Stdout, "Done")
}

func gracefulShutdown(app *fiber.App, db *sql.DB) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	if err := app.Shutdown(); err != nil {
		log.Fatalf("Failed to shut down server: %v", err)
	}

	closeDB(db)
	log.Println("Server shutdown complete.")
}
