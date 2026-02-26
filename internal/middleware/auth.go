package middleware

import (
	"net/http"
	"strings"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/service"
	"github.com/gin-gonic/gin"
)

const userIDKey = "user_id"

func Auth(tokenService service.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// Try Authorization header first
		header := c.GetHeader("Authorization")
		if header != "" {
			parts := strings.SplitN(header, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				token = parts[1]
			}
		}

		// Fall back to query parameter if no header
		if token == "" {
			token = c.Query("token")
		}

		// No token found in either location
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "отсутствует токен авторизации"})
			return
		}

		userID, role, err := tokenService.ValidateAccessToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "недействительный или просроченный токен"})
			return
		}

		c.Set(userIDKey, userID)
		SetUserRole(c, role)
		c.Next()
	}
}

func GetUserID(c *gin.Context) uint {
	id, _ := c.Get(userIDKey)
	userID, _ := id.(uint)
	return userID
}
