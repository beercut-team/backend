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
	
	// Test batch insert with mixed is_required values
	items := []domain.ChecklistItem{
		{PatientID: 998, Name: "Required Item", IsRequired: true, Status: domain.ChecklistStatusPending},
		{PatientID: 998, Name: "Optional Item", IsRequired: false, Status: domain.ChecklistStatusPending},
		{PatientID: 998, Name: "Another Required", IsRequired: true, Status: domain.ChecklistStatusPending},
	}
	
	fmt.Println("Before batch insert:")
	for i, item := range items {
		fmt.Printf("  %d. %s - IsRequired: %v\n", i+1, item.Name, item.IsRequired)
	}
	
	if err := db.Create(&items).Error; err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Println("\nAfter batch insert:")
	for i, item := range items {
		fmt.Printf("  %d. ID=%d, %s - IsRequired: %v\n", i+1, item.ID, item.Name, item.IsRequired)
	}
	
	// Read back from DB
	var readItems []domain.ChecklistItem
	db.Where("patient_id = 998").Find(&readItems)
	
	fmt.Println("\nRead from DB:")
	for i, item := range readItems {
		fmt.Printf("  %d. %s - IsRequired: %v\n", i+1, item.Name, item.IsRequired)
	}
	
	// Clean up
	db.Where("patient_id = 998").Delete(&domain.ChecklistItem{})
}
