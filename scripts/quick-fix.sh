#!/bin/bash
# Быстрая проверка и исправление кодов доступа
# Запустить на сервере: bash quick-fix.sh

set -e

echo "🔍 Проверка контейнера..."
docker ps | grep peak-it-backend || echo "❌ Контейнер не запущен!"

echo ""
echo "📊 Проверка БД..."
docker exec -it postgres psql -U postgres -d peakit -c "SELECT COUNT(*) as total, COUNT(access_code) as with_codes FROM patients;"

echo ""
echo "🔧 Запуск fix-access-codes..."
docker run --rm \
  --network app-network \
  -e "DB_HOST=postgres" \
  -e "DB_PORT=5432" \
  -e "DB_USER=postgres" \
  -e "DB_PASSWORD=${DB_PASSWORD}" \
  -e "DB_NAME=peakit" \
  -e "DB_SSLMODE=disable" \
  peak-it-backend:latest ./fix-access-codes

echo ""
echo "✅ Проверка результата..."
docker exec -it postgres psql -U postgres -d peakit -c "SELECT id, first_name, last_name, access_code FROM patients LIMIT 5;"

echo ""
echo "🎉 Готово! Проверьте админку."
