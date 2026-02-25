#!/bin/bash

# Reset database script
set -e

echo "🗑️  Dropping database..."
psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS peakit;"

echo "📦 Creating database..."
psql -h localhost -U postgres -c "CREATE DATABASE peakit;"

echo "✅ Database reset complete!"
echo "Run: go run ./cmd/seed to populate with test data"
