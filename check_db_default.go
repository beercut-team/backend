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
		SELECT 
			column_name, 
			column_default, 
			is_nullable,
			data_type
		FROM information_schema.columns 
		WHERE table_name = 'checklist_items' 
		AND column_name IN ('is_required', 'status')
		ORDER BY column_name
	`).Scan(&result)
	
	fmt.Println("Database schema for checklist_items:")
	for _, r := range result {
		fmt.Printf("\nColumn: %v\n", r["column_name"])
		fmt.Printf("  Type: %v\n", r["data_type"])
		fmt.Printf("  Nullable: %v\n", r["is_nullable"])
		fmt.Printf("  Default: %v\n", r["column_default"])
	}
}
