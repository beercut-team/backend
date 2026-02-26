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
	
	// Check patient 6 (created via API)
	var apiItems []domain.ChecklistItem
	db.Where("patient_id = 6").Order("id").Find(&apiItems)
	
	fmt.Println("Patient 6 (created via API):")
	optionalCount := 0
	for _, item := range apiItems {
		if !item.IsRequired {
			fmt.Printf("  OPTIONAL: %s\n", item.Name)
			optionalCount++
		}
	}
	fmt.Printf("Total optional: %d\n", optionalCount)
	
	// Check patient 1 (seeded)
	var seedItems []domain.ChecklistItem
	db.Where("patient_id = 1").Order("id").Find(&seedItems)
	
	fmt.Println("\nPatient 1 (seeded):")
	optionalCount = 0
	for _, item := range seedItems {
		if !item.IsRequired {
			fmt.Printf("  OPTIONAL: %s\n", item.Name)
			optionalCount++
		}
	}
	fmt.Printf("Total optional: %d\n", optionalCount)
	
	// Check raw SQL
	var rawResults []map[string]interface{}
	db.Raw("SELECT id, name, is_required FROM checklist_items WHERE patient_id = 6 AND is_required = false").Scan(&rawResults)
	fmt.Printf("\nRaw SQL query for patient 6 (is_required = false): %d rows\n", len(rawResults))
	
	db.Raw("SELECT id, name, is_required FROM checklist_items WHERE patient_id = 6 AND is_required IS NULL").Scan(&rawResults)
	fmt.Printf("Raw SQL query for patient 6 (is_required IS NULL): %d rows\n", len(rawResults))
}
