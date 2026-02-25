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
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "отсутствует заголовок авторизации"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "неверный заголовок авторизации"})
			return
		}

		userID, role, err := tokenService.ValidateAccessToken(parts[1])
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
