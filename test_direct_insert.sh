#!/bin/bash
BASE_URL="http://localhost:8080"

TOKEN=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"direct_$(date +%s)@example.com\",
    \"password\": \"password123\",
    \"name\": \"Test\",
    \"first_name\": \"Test\",
    \"last_name\": \"User\",
    \"phone\": \"+79991234567\",
    \"role\": \"DISTRICT_DOCTOR\",
    \"district_id\": 1
  }" | jq -r '.access_token')

PATIENT_ID=$(curl -s -X POST "$BASE_URL/api/v1/patients" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Test","last_name":"Patient","operation_type":"PHACOEMULSIFICATION","eye":"OD","district_id":1,"diagnosis":"Test"}' | jq -r '.data.id')

# Add custom optional item
CUSTOM=$(curl -s -X POST "$BASE_URL/api/v1/checklists" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"patient_id\": $PATIENT_ID,
    \"name\": \"Кастомный опциональный пункт\",
    \"description\": \"Тест\",
    \"category\": \"Тест\",
    \"is_required\": false
  }")

echo "Custom item created:"
echo "$CUSTOM" | jq '{id: .data.id, name: .data.name, is_required: .data.is_required}'
