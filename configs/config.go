package config

import "os"

type R2 struct {
	AccountID  string
	AccessKey  string
	SecretKey  string
	BucketName string
}

type Config struct {
	InstagramClientID     string
	InstagramClientSecret string
	InstagramRedirectURI  string
	TiktokClientKey       string
	TiktokClientSecret    string
	TiktokRedirectURI     string
	GoogleClientID        string
	GoogleClientSecret    string
	GoogleRedirectURI     string
	PostgresURI           string
	DatabaseName          string
	RedisURI              string
	FrontendURL           string
	R2                    R2
	SecretKey             string
	CookieName            string
}

func LoadConfig() *Config {
	return &Config{
		InstagramClientID:     getEnv("INSTAGRAM_CLIENT_ID", ""),
		InstagramClientSecret: getEnv("INSTAGRAM_CLIENT_SECRET", ""),
		InstagramRedirectURI:  getEnv("INSTAGRAM_REDIRECT_URI", ""),
		TiktokClientKey:       getEnv("TIKTOK_CLIENT_KEY", ""),
		TiktokClientSecret:    getEnv("TIKTOK_CLIENT_SECRET", ""),
		TiktokRedirectURI:     getEnv("TIKTOK_REDIRECT_URI", ""),
		GoogleClientID:        getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:    getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURI:     getEnv("GOOGLE_REDIRECT_URI", ""),
		PostgresURI:           getEnv("POSTGRES_URI", ""),
		DatabaseName:          getEnv("DATABASE_NAME", ""),
		RedisURI:              getEnv("REDIS_URI", ""),
		FrontendURL:           getEnv("FRONTEND_URL", "http://localhost:5173"),
		R2: R2{
			AccountID:  getEnv("R2_ACCOUNT_ID", ""),
			AccessKey:  getEnv("R2_ACCESS_KEY", ""),
			SecretKey:  getEnv("R2_SECRET_KEY", ""),
			BucketName: getEnv("R2_BUCKET_NAME", ""),
		},
		SecretKey:  getEnv("SECRET_KEY", ""),
		CookieName: getEnv("COOKIE_NAME", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
