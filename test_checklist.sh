#!/bin/bash

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"
TOKEN=""
PATIENT_ID=""
DISTRICT_ID=""

echo -e "${YELLOW}=== Тестирование функциональности чек-листов ===${NC}\n"

# Функция для проверки ответа
check_response() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ $2${NC}"
    else
        echo -e "${RED}✗ $2${NC}"
        exit 1
    fi
}

# 1. Получить публичный список районов
echo -e "\n${YELLOW}1. Получение списка районов (без авторизации)${NC}"
DISTRICT_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/districts")
echo "$DISTRICT_RESPONSE" | jq '.'
DISTRICT_ID=$(echo "$DISTRICT_RESPONSE" | jq -r '.data[0].id')
check_response $? "Получен список районов, district_id=$DISTRICT_ID"

# 2. Регистрация врача
echo -e "\n${YELLOW}2. Регистрация районного врача${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"test_doctor_$(date +%s)@example.com\",
    \"password\": \"password123\",
    \"name\": \"Тестовый Врач\",
    \"first_name\": \"Тестовый\",
    \"last_name\": \"Врач\",
    \"phone\": \"+79991234567\",
    \"role\": \"DISTRICT_DOCTOR\",
    \"district_id\": $DISTRICT_ID
  }")
echo "$REGISTER_RESPONSE" | jq '.'
TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.access_token')
check_response $? "Врач зарегистрирован, получен токен"

# 3. Создание пациента с операцией PHACOEMULSIFICATION
echo -e "\n${YELLOW}3. Создание пациента (Факоэмульсификация)${NC}"
PATIENT_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/patients" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"first_name\": \"Иван\",
    \"last_name\": \"Тестов\",
    \"middle_name\": \"Петрович\",
    \"operation_type\": \"PHACOEMULSIFICATION\",
    \"eye\": \"OD\",
    \"district_id\": $DISTRICT_ID,
    \"diagnosis\": \"Катаракта правого глаза\"
  }")
echo "$PATIENT_RESPONSE" | jq '.'
PATIENT_ID=$(echo "$PATIENT_RESPONSE" | jq -r '.data.id')
PATIENT_STATUS=$(echo "$PATIENT_RESPONSE" | jq -r '.data.status')
check_response $? "Пациент создан, ID=$PATIENT_ID, статус=$PATIENT_STATUS"

# Проверка человекочитаемых полей
STATUS_DISPLAY=$(echo "$PATIENT_RESPONSE" | jq -r '.data.status_display')
OPERATION_DISPLAY=$(echo "$PATIENT_RESPONSE" | jq -r '.data.operation_type_display')
EYE_DISPLAY=$(echo "$PATIENT_RESPONSE" | jq -r '.data.eye_display')
echo -e "  Человекочитаемые поля:"
echo -e "    status_display: $STATUS_DISPLAY"
echo -e "    operation_type_display: $OPERATION_DISPLAY"
echo -e "    eye_display: $EYE_DISPLAY"

# 4. Получение чек-листа пациента
echo -e "\n${YELLOW}4. Получение чек-листа пациента${NC}"
CHECKLIST_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/checklists/patient/$PATIENT_ID" \
  -H "Authorization: Bearer $TOKEN")
echo "$CHECKLIST_RESPONSE" | jq '.'
CHECKLIST_COUNT=$(echo "$CHECKLIST_RESPONSE" | jq '.data | length')
check_response $? "Получен чек-лист, количество пунктов: $CHECKLIST_COUNT"

# Вывод всех пунктов
echo -e "\n  Пункты чек-листа:"
echo "$CHECKLIST_RESPONSE" | jq -r '.data[] | "    [\(.id)] \(.name) - \(.status) (required: \(.is_required))"'

# 5. Получение прогресса чек-листа
echo -e "\n${YELLOW}5. Получение прогресса чек-листа${NC}"
PROGRESS_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/checklists/patient/$PATIENT_ID/progress" \
  -H "Authorization: Bearer $TOKEN")
echo "$PROGRESS_RESPONSE" | jq '.'
check_response $? "Получен прогресс чек-листа"

# 6. Добавление кастомного пункта
echo -e "\n${YELLOW}6. Добавление кастомного пункта в чек-лист${NC}"
CUSTOM_ITEM_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/checklists" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"patient_id\": $PATIENT_ID,
    \"name\": \"Дополнительная консультация офтальмолога\",
    \"description\": \"Повторный осмотр перед операцией\",
    \"category\": \"Консультации\",
    \"is_required\": false,
    \"expires_in_days\": 7
  }")
echo "$CUSTOM_ITEM_RESPONSE" | jq '.'
CUSTOM_ITEM_ID=$(echo "$CUSTOM_ITEM_RESPONSE" | jq -r '.data.id')
check_response $? "Добавлен кастомный пункт, ID=$CUSTOM_ITEM_ID"

# 7. Обновление пункта чек-листа (отметка как выполненного)
echo -e "\n${YELLOW}7. Отметка первого пункта как выполненного${NC}"
FIRST_ITEM_ID=$(echo "$CHECKLIST_RESPONSE" | jq -r '.data[0].id')
UPDATE_RESPONSE=$(curl -s -X PATCH "$BASE_URL/api/v1/checklists/$FIRST_ITEM_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"status\": \"COMPLETED\",
    \"result\": \"Анализ в норме\",
    \"notes\": \"Все показатели в пределах нормы\"
  }")
echo "$UPDATE_RESPONSE" | jq '.'
check_response $? "Пункт $FIRST_ITEM_ID отмечен как выполненный"

# 8. Проверка обновлённого прогресса
echo -e "\n${YELLOW}8. Проверка обновлённого прогресса${NC}"
PROGRESS_RESPONSE2=$(curl -s -X GET "$BASE_URL/api/v1/checklists/patient/$PATIENT_ID/progress" \
  -H "Authorization: Bearer $TOKEN")
echo "$PROGRESS_RESPONSE2" | jq '.'
COMPLETED_COUNT=$(echo "$PROGRESS_RESPONSE2" | jq -r '.data.completed')
REQUIRED_COMPLETED=$(echo "$PROGRESS_RESPONSE2" | jq -r '.data.required_completed')
echo -e "  Выполнено: $COMPLETED_COUNT из $(echo "$PROGRESS_RESPONSE2" | jq -r '.data.total')"
echo -e "  Обязательных выполнено: $REQUIRED_COMPLETED из $(echo "$PROGRESS_RESPONSE2" | jq -r '.data.required')"

# 9. Отметка всех обязательных пунктов как выполненных
echo -e "\n${YELLOW}9. Отметка всех обязательных пунктов как выполненных${NC}"
REQUIRED_ITEMS=$(echo "$CHECKLIST_RESPONSE" | jq -r '.data[] | select(.is_required == true) | .id')
for ITEM_ID in $REQUIRED_ITEMS; do
    echo "  Отметка пункта $ITEM_ID..."
    curl -s -X PATCH "$BASE_URL/api/v1/checklists/$ITEM_ID" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{\"status\": \"COMPLETED\", \"result\": \"Выполнено\"}" > /dev/null
done
check_response $? "Все обязательные пункты отмечены как выполненные"

# 10. Проверка автоматического перехода статуса
echo -e "\n${YELLOW}10. Проверка автоматического перехода статуса пациента${NC}"
sleep 1
PATIENT_STATUS_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/patients/$PATIENT_ID" \
  -H "Authorization: Bearer $TOKEN")
echo "$PATIENT_STATUS_RESPONSE" | jq '.'
NEW_STATUS=$(echo "$PATIENT_STATUS_RESPONSE" | jq -r '.data.status')
NEW_STATUS_DISPLAY=$(echo "$PATIENT_STATUS_RESPONSE" | jq -r '.data.status_display')
echo -e "  Новый статус: $NEW_STATUS ($NEW_STATUS_DISPLAY)"

if [ "$NEW_STATUS" == "PENDING_REVIEW" ]; then
    echo -e "${GREEN}✓ Статус автоматически изменился на PENDING_REVIEW${NC}"
else
    echo -e "${RED}✗ Статус не изменился автоматически (ожидался PENDING_REVIEW, получен $NEW_STATUS)${NC}"
fi

# 11. Тест с другим типом операции (ANTIGLAUCOMA)
echo -e "\n${YELLOW}11. Создание пациента с антиглаукомной операцией${NC}"
PATIENT2_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/patients" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"first_name\": \"Мария\",
    \"last_name\": \"Тестова\",
    \"operation_type\": \"ANTIGLAUCOMA\",
    \"eye\": \"OS\",
    \"district_id\": $DISTRICT_ID,
    \"diagnosis\": \"Глаукома левого глаза\"
  }")
PATIENT2_ID=$(echo "$PATIENT2_RESPONSE" | jq -r '.data.id')
check_response $? "Пациент 2 создан, ID=$PATIENT2_ID"

CHECKLIST2_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/checklists/patient/$PATIENT2_ID" \
  -H "Authorization: Bearer $TOKEN")
CHECKLIST2_COUNT=$(echo "$CHECKLIST2_RESPONSE" | jq '.data | length')
echo -e "  Количество пунктов для ANTIGLAUCOMA: $CHECKLIST2_COUNT"
echo -e "\n  Специфичные пункты для глаукомы:"
echo "$CHECKLIST2_RESPONSE" | jq -r '.data[] | select(.category == "Офтальмология") | "    - \(.name)"'

# 12. Тест с витрэктомией
echo -e "\n${YELLOW}12. Создание пациента с витрэктомией${NC}"
PATIENT3_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/patients" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"first_name\": \"Петр\",
    \"last_name\": \"Тестов\",
    \"operation_type\": \"VITRECTOMY\",
    \"eye\": \"OU\",
    \"district_id\": $DISTRICT_ID,
    \"diagnosis\": \"Отслойка сетчатки обоих глаз\"
  }")
PATIENT3_ID=$(echo "$PATIENT3_RESPONSE" | jq -r '.data.id')
check_response $? "Пациент 3 создан, ID=$PATIENT3_ID"

CHECKLIST3_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/checklists/patient/$PATIENT3_ID" \
  -H "Authorization: Bearer $TOKEN")
CHECKLIST3_COUNT=$(echo "$CHECKLIST3_RESPONSE" | jq '.data | length')
echo -e "  Количество пунктов для VITRECTOMY: $CHECKLIST3_COUNT"
echo -e "\n  Специфичные пункты для витрэктомии:"
echo "$CHECKLIST3_RESPONSE" | jq -r '.data[] | select(.category == "Офтальмология") | "    - \(.name)"'

echo -e "\n${GREEN}=== Все тесты завершены успешно ===${NC}\n"
