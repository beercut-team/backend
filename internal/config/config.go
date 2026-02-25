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
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	viper.AutomaticEnv()

	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("JWT_ACCESS_EXPIRY_MINUTES", 15)
	viper.SetDefault("JWT_REFRESH_EXPIRY_HOURS", 168)

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
	}

	return cfg, nil
}
