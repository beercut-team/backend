package main

import (
	"fmt"
	"github.com/beercut-team/backend-boilerplate/internal/config"
	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	cfg, _ := config.Load()
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)
	
	db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	
	// Test insert with is_required=false
	item := domain.ChecklistItem{
		PatientID:   999,
		Name:        "Test Item",
		Description: "Test",
		Category:    "Test",
		IsRequired:  false,
		Status:      domain.ChecklistStatusPending,
	}
	
	fmt.Printf("Before insert: IsRequired = %v\n", item.IsRequired)
	
	if err := db.Create(&item).Error; err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("After insert: ID = %d\n", item.ID)
	
	// Read it back
	var readItem domain.ChecklistItem
	db.First(&readItem, item.ID)
	fmt.Printf("Read from DB: IsRequired = %v\n", readItem.IsRequired)
	
	// Clean up
	db.Delete(&item)
}
