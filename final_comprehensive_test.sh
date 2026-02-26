#!/bin/bash
set -e

BASE_URL="http://localhost:8080"
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}=== Комплексное тестирование чек-листов ===${NC}\n"

# 1. Register and login
echo "1. Регистрация врача..."
TOKEN=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"final_test_$(date +%s)@example.com\",
    \"password\": \"password123\",
    \"name\": \"Test Doctor\",
    \"first_name\": \"Test\",
    \"last_name\": \"Doctor\",
    \"phone\": \"+79991234567\",
    \"role\": \"DISTRICT_DOCTOR\",
    \"district_id\": 1
  }" | jq -r '.access_token')
echo -e "${GREEN}✓ Врач зарегистрирован${NC}"

# 2. Create patient
echo -e "\n2. Создание пациента..."
PATIENT=$(curl -s -X POST "$BASE_URL/api/v1/patients" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Иван","last_name":"Тестов","middle_name":"Петрович","operation_type":"PHACOEMULSIFICATION","eye":"OD","district_id":1,"diagnosis":"Катаракта"}')
PATIENT_ID=$(echo "$PATIENT" | jq -r '.data.id')
PATIENT_STATUS=$(echo "$PATIENT" | jq -r '.data.status')
echo -e "${GREEN}✓ Пациент создан (ID: $PATIENT_ID, статус: $PATIENT_STATUS)${NC}"

# 3. Get checklist
echo -e "\n3. Получение чек-листа..."
CHECKLIST=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/v1/checklists/patient/$PATIENT_ID")
TOTAL_COUNT=$(echo "$CHECKLIST" | jq '.data | length')
REQUIRED_COUNT=$(echo "$CHECKLIST" | jq '[.data[] | select(.is_required == true)] | length')
OPTIONAL_COUNT=$(echo "$CHECKLIST" | jq '[.data[] | select(.is_required == false)] | length')
echo -e "${GREEN}✓ Чек-лист получен: $TOTAL_COUNT пунктов ($REQUIRED_COUNT обязательных, $OPTIONAL_COUNT опциональных)${NC}"

# 4. Add custom optional item
echo -e "\n4. Добавление кастомного опционального пункта..."
CUSTOM=$(curl -s -X POST "$BASE_URL/api/v1/checklists" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"patient_id\":$PATIENT_ID,\"name\":\"Дополнительная консультация\",\"description\":\"Тест\",\"category\":\"Консультации\",\"is_required\":false}")
CUSTOM_ID=$(echo "$CUSTOM" | jq -r '.data.id')
CUSTOM_REQUIRED=$(echo "$CUSTOM" | jq -r '.data.is_required')
echo -e "${GREEN}✓ Кастомный пункт добавлен (ID: $CUSTOM_ID, required: $CUSTOM_REQUIRED)${NC}"

# 5. Mark all required items as completed
echo -e "\n5. Отметка всех обязательных пунктов как выполненных..."
REQUIRED_IDS=$(echo "$CHECKLIST" | jq -r '[.data[] | select(.is_required == true) | .id] | @sh' | tr -d "'")
for ID in $REQUIRED_IDS; do
  curl -s -X PATCH "$BASE_URL/api/v1/checklists/$ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"status":"COMPLETED"}' > /dev/null
done
echo -e "${GREEN}✓ Все обязательные пункты отмечены${NC}"

# 6. Check auto-transition
echo -e "\n6. Проверка автоперехода статуса..."
sleep 1
FINAL_STATUS=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/v1/patients/$PATIENT_ID" | jq -r '.data.status')
if [ "$FINAL_STATUS" = "PENDING_REVIEW" ]; then
  echo -e "${GREEN}✅ УСПЕХ: Автопереход сработал (статус: $FINAL_STATUS)${NC}"
else
  echo -e "${RED}❌ ОШИБКА: Статус не изменился (ожидался PENDING_REVIEW, получен $FINAL_STATUS)${NC}"
  exit 1
fi

# 7. Check progress
echo -e "\n7. Проверка прогресса..."
PROGRESS=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/v1/checklists/patient/$PATIENT_ID/progress")
echo "$PROGRESS" | jq '{total, completed, required, required_completed, percentage}'

echo -e "\n${GREEN}=== Все тесты пройдены успешно ===${NC}"
