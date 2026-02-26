package main

import (
	"fmt"
	"github.com/beercut-team/backend-boilerplate/internal/config"
	"github.com/beercut-team/backend-boilerplate/internal/domain"
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
	
	var items []domain.ChecklistItem
	db.Where("patient_id = 1").Order("id").Find(&items)
	
	fmt.Println("Checklist items for seeded patient 1:")
	for _, item := range items {
		fmt.Printf("%s - IsRequired: %v\n", item.Name, item.IsRequired)
	}
	
	optionalCount := 0
	for _, item := range items {
		if !item.IsRequired {
			optionalCount++
		}
	}
	fmt.Printf("\nOptional items count: %d (should be 2)\n", optionalCount)
}
