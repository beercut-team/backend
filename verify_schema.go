package main

import (
	"fmt"
	"github.com/beercut-team/backend-boilerplate/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, _ := config.Load()
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)
	
	db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	
	var result []map[string]interface{}
	db.Raw(`
		SELECT column_name, column_default, is_nullable, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'checklist_items' AND column_name = 'is_required'
	`).Scan(&result)
	
	fmt.Printf("Schema for is_required column:\n%+v\n", result)
}
