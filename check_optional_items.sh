#!/bin/bash
BASE_URL="http://localhost:8080"

# Register new user
TOKEN=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"check_$(date +%s)@example.com\",
    \"password\": \"password123\",
    \"name\": \"Test\",
    \"first_name\": \"Test\",
    \"last_name\": \"User\",
    \"phone\": \"+79991234567\",
    \"role\": \"DISTRICT_DOCTOR\",
    \"district_id\": 1
  }" | jq -r '.access_token')

# Create patient
PATIENT_ID=$(curl -s -X POST "$BASE_URL/api/v1/patients" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Test","last_name":"Patient","operation_type":"PHACOEMULSIFICATION","eye":"OD","district_id":1,"diagnosis":"Test"}' | jq -r '.data.id')

echo "Patient ID: $PATIENT_ID"

# Get all checklist items
CHECKLIST=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/v1/checklists/patient/$PATIENT_ID")

echo -e "\nAll items:"
echo "$CHECKLIST" | jq -r '.data[] | "\(.name) | required=\(.is_required)"'

echo -e "\nOptional items (should be 2):"
echo "$CHECKLIST" | jq -r '.data[] | select(.is_required == false) | .name'

echo -e "\nRequired count:"
echo "$CHECKLIST" | jq '[.data[] | select(.is_required == true)] | length'

echo -e "\nOptional count:"
echo "$CHECKLIST" | jq '[.data[] | select(.is_required == false)] | length'
