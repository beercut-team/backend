#!/bin/bash
# Check the actual database schema for checklist_items table
docker exec -it postgres psql -U postgres -d oculus_feldsher -c "\d checklist_items" 2>/dev/null || \
psql -U postgres -d oculus_feldsher -c "\d checklist_items" 2>/dev/null || \
echo "Cannot connect to PostgreSQL. Checking via Go code..."

# Alternative: check via GORM
cat > /tmp/check_schema.go << 'GOEOF'
package main
import (
	"fmt"
	"github.com/beercut-team/backend-boilerplate/internal/config"
	"github.com/beercut-team/backend-boilerplate/pkg/database"
)
func main() {
	cfg := config.Load()
	db, _ := database.NewPostgres(cfg)
	var result []map[string]interface{}
	db.Raw("SELECT column_name, column_default, is_nullable FROM information_schema.columns WHERE table_name = 'checklist_items' AND column_name = 'is_required'").Scan(&result)
	fmt.Printf("%+v\n", result)
}
GOEOF
go run /tmp/check_schema.go
