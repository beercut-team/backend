package config

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	AppPort             string `mapstructure:"APP_PORT"`
	DBHost              string `mapstructure:"DB_HOST"`
	DBPort              string `mapstructure:"DB_PORT"`
	DBUser              string `mapstructure:"DB_USER"`
	DBPassword          string `mapstructure:"DB_PASSWORD"`
	DBName              string `mapstructure:"DB_NAME"`
	DBSSLMode           string `mapstructure:"DB_SSLMODE"`
	JWTAccessSecret     string `mapstructure:"JWT_ACCESS_SECRET"`
	JWTRefreshSecret    string `mapstructure:"JWT_REFRESH_SECRET"`
	JWTAccessExpiryMin  int    `mapstructure:"JWT_ACCESS_EXPIRY_MINUTES"`
	JWTRefreshExpiryHrs int    `mapstructure:"JWT_REFRESH_EXPIRY_HOURS"`

	// Redis
	RedisHost     string `mapstructure:"REDIS_HOST"`
	RedisPort     string `mapstructure:"REDIS_PORT"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`

	// MinIO
	MinIOEndpoint  string `mapstructure:"MINIO_ENDPOINT"`
	MinIOAccessKey string `mapstructure:"MINIO_ACCESS_KEY"`
	MinIOSecretKey string `mapstructure:"MINIO_SECRET_KEY"`
	MinIOBucket    string `mapstructure:"MINIO_BUCKET"`
	MinIOUseSSL    bool   `mapstructure:"MINIO_USE_SSL"`

	// Telegram
	TelegramBotToken string `mapstructure:"TELEGRAM_BOT_TOKEN"`

	// Base URL for frontend links (used in Telegram bot, emails, etc.)
	BaseURL string `mapstructure:"BASE_URL"`

	// Storage mode: "minio" or "local"
	StorageMode     string `mapstructure:"STORAGE_MODE"`
	LocalUploadPath string `mapstructure:"LOCAL_UPLOAD_PATH"`
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	viper.AutomaticEnv()

	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("JWT_ACCESS_EXPIRY_MINUTES", 15)
	viper.SetDefault("JWT_REFRESH_EXPIRY_HOURS", 168)
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)
	viper.SetDefault("MINIO_ENDPOINT", "localhost:9000")
	viper.SetDefault("MINIO_ACCESS_KEY", "minioadmin")
	viper.SetDefault("MINIO_SECRET_KEY", "minioadmin")
	viper.SetDefault("MINIO_BUCKET", "oculus")
	viper.SetDefault("MINIO_USE_SSL", false)
	viper.SetDefault("STORAGE_MODE", "local")
	viper.SetDefault("LOCAL_UPLOAD_PATH", "./uploads")
	viper.SetDefault("BASE_URL", "http://localhost:8080")

	cfg := &Config{
		AppPort:             viper.GetString("APP_PORT"),
		DBHost:              viper.GetString("DB_HOST"),
		DBPort:              viper.GetString("DB_PORT"),
		DBUser:              viper.GetString("DB_USER"),
		DBPassword:          viper.GetString("DB_PASSWORD"),
		DBName:              viper.GetString("DB_NAME"),
		DBSSLMode:           viper.GetString("DB_SSLMODE"),
		JWTAccessSecret:     viper.GetString("JWT_ACCESS_SECRET"),
		JWTRefreshSecret:    viper.GetString("JWT_REFRESH_SECRET"),
		JWTAccessExpiryMin:  viper.GetInt("JWT_ACCESS_EXPIRY_MINUTES"),
		JWTRefreshExpiryHrs: viper.GetInt("JWT_REFRESH_EXPIRY_HOURS"),
		RedisHost:           viper.GetString("REDIS_HOST"),
		RedisPort:           viper.GetString("REDIS_PORT"),
		RedisPassword:       viper.GetString("REDIS_PASSWORD"),
		RedisDB:             viper.GetInt("REDIS_DB"),
		MinIOEndpoint:       viper.GetString("MINIO_ENDPOINT"),
		MinIOAccessKey:      viper.GetString("MINIO_ACCESS_KEY"),
		MinIOSecretKey:      viper.GetString("MINIO_SECRET_KEY"),
		MinIOBucket:         viper.GetString("MINIO_BUCKET"),
		MinIOUseSSL:         viper.GetBool("MINIO_USE_SSL"),
		TelegramBotToken:    viper.GetString("TELEGRAM_BOT_TOKEN"),
		BaseURL:             viper.GetString("BASE_URL"),
		StorageMode:         viper.GetString("STORAGE_MODE"),
		LocalUploadPath:     viper.GetString("LOCAL_UPLOAD_PATH"),
	}

	return cfg, nil
}
