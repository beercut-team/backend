package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/beercut-team/backend-boilerplate/internal/service"
	"github.com/gin-gonic/gin"
)

// AuditMiddleware логирует все мутации (POST, PUT, PATCH, DELETE)
func AuditMiddleware(auditService service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Логируем только мутации
		method := c.Request.Method
		if method != "POST" && method != "PUT" && method != "PATCH" && method != "DELETE" {
			c.Next()
			return
		}

		// Пропускаем auth endpoints
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/auth") ||
			strings.HasPrefix(c.Request.URL.Path, "/api/public") {
			c.Next()
			return
		}

		// Получаем userID (может быть 0 для неавторизованных запросов)
		userID := GetUserID(c)
		if userID == 0 {
			c.Next()
			return
		}

		// Читаем body для логирования
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		var requestData interface{}
		if len(bodyBytes) > 0 {
			json.Unmarshal(bodyBytes, &requestData)
		}

		// Определяем entity и action из пути
		path := c.Request.URL.Path
		entity := extractEntity(path)
		action := mapMethodToAction(method)

		// Получаем IP
		ip := c.ClientIP()

		// Выполняем запрос
		c.Next()

		// Логируем только успешные операции (2xx)
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			// Пытаемся извлечь entityID из параметров
			entityID := uint(0)
			if idParam := c.Param("id"); idParam != "" {
				// Парсим ID если есть
				var id uint
				if _, err := fmt.Sscanf(idParam, "%d", &id); err == nil {
					entityID = id
				}
			}

			// Асинхронно логируем (не блокируем ответ)
			go func() {
				auditService.LogAction(
					c.Request.Context(),
					userID,
					action,
					entity,
					entityID,
					nil, // oldValue - можно расширить позже
					requestData,
					ip,
				)
			}()
		}
	}
}

func extractEntity(path string) string {
	// Извлекаем entity из пути типа /api/v1/patients/123
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 {
		return parts[2] // patients, districts, etc.
	}
	return "unknown"
}

func mapMethodToAction(method string) string {
	switch method {
	case "POST":
		return "CREATE"
	case "PUT", "PATCH":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	default:
		return method
	}
}
