#!/bin/bash
set -e

cd "$(dirname "$0")/.."

echo "🗑️  Dropping database..."

# Load DB config from .env
export $(grep -v '^#' .env | xargs)

# Drop and recreate database
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "DROP DATABASE IF EXISTS $DB_NAME;"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE $DB_NAME;"

echo "📦 Running migrations and seeding data..."
go run ./cmd/seed

echo "✅ Database reset and seeded successfully!"
echo ""
echo "Test users:"
echo "  admin@example.com / admin123"
echo "  doctor@example.com / doctor123"
echo "  surgeon@example.com / surgeon123"
echo ""
echo "Test patients with access codes:"
echo "  a1b2c3d4 - Туяра Алексеева"
echo "  e5f6g7h8 - Айаал Степанов"
echo "  i9j0k1l2 - Айыына Павлова"
