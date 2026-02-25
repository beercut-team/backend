package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/service"
)

const userIDKey = "user_id"

func Auth(tokenService service.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "missing authorization header"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "invalid authorization header"})
			return
		}

		userID, err := tokenService.ValidateAccessToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "invalid or expired token"})
			return
		}

		c.Set(userIDKey, userID)
		c.Next()
	}
}

func GetUserID(c *gin.Context) uint {
	id, _ := c.Get(userIDKey)
	userID, _ := id.(uint)
	return userID
}
