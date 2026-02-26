#!/bin/bash

BASE_URL="http://localhost:8080"

# Register and login
TOKEN=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"test_auto_$(date +%s)@example.com\",
    \"password\": \"password123\",
    \"name\": \"Test Doctor\",
    \"first_name\": \"Test\",
    \"last_name\": \"Doctor\",
    \"phone\": \"+79991234567\",
    \"role\": \"DISTRICT_DOCTOR\",
    \"district_id\": 1
  }" | jq -r '.access_token')

echo "Token: ${TOKEN:0:20}..."

# Create patient
PATIENT=$(curl -s -X POST "$BASE_URL/api/v1/patients" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"first_name\": \"Иван\",
    \"last_name\": \"Тестов\",
    \"middle_name\": \"Петрович\",
    \"operation_type\": \"PHACOEMULSIFICATION\",
    \"eye\": \"OD\",
    \"district_id\": 1,
    \"diagnosis\": \"Катаракта\"
  }")

PATIENT_ID=$(echo "$PATIENT" | jq -r '.data.id')
echo "Patient ID: $PATIENT_ID"

# Get checklist
CHECKLIST=$(curl -s -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/checklists/patient/$PATIENT_ID")

echo -e "\nChecklist items:"
echo "$CHECKLIST" | jq -r '.data[] | "\(.id) | \(.name) | required=\(.is_required) | \(.status)"'

# Count required items
REQUIRED_COUNT=$(echo "$CHECKLIST" | jq '[.data[] | select(.is_required == true)] | length')
echo -e "\nRequired items count: $REQUIRED_COUNT"

# Get IDs of required items
REQUIRED_IDS=$(echo "$CHECKLIST" | jq -r '[.data[] | select(.is_required == true) | .id] | @sh' | tr -d "'")

echo -e "\nMarking all required items as completed..."
for ID in $REQUIRED_IDS; do
  curl -s -X PATCH "$BASE_URL/api/v1/checklists/$ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"status": "COMPLETED"}' > /dev/null
  echo "  ✓ Item $ID completed"
done

# Check patient status
sleep 1
PATIENT_STATUS=$(curl -s -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/patients/$PATIENT_ID" | jq -r '.data.status')

echo -e "\nFinal patient status: $PATIENT_STATUS"

if [ "$PATIENT_STATUS" = "PENDING_REVIEW" ]; then
  echo "✅ SUCCESS: Auto-transition worked!"
else
  echo "❌ FAILED: Status is $PATIENT_STATUS, expected PENDING_REVIEW"
fi
