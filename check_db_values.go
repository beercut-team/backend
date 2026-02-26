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
	
	var results []map[string]interface{}
	db.Raw(`
		SELECT id, name, is_required 
		FROM checklist_items 
		WHERE patient_id = 5
		ORDER BY id
	`).Scan(&results)
	
	fmt.Println("Checklist items in database:")
	for _, r := range results {
		fmt.Printf("ID: %v, Name: %v, IsRequired: %v\n", r["id"], r["name"], r["is_required"])
	}
}
