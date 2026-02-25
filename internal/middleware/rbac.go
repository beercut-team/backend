package middleware

import (
	"net/http"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/gin-gonic/gin"
)

const userRoleKey = "user_role"

func RequireRole(roles ...domain.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetUserRole(c)
		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, domain.ErrorResponse{Error: "недостаточно прав доступа"})
	}
}

func SetUserRole(c *gin.Context, role domain.Role) {
	c.Set(userRoleKey, role)
}

func GetUserRole(c *gin.Context) domain.Role {
	val, exists := c.Get(userRoleKey)
	if !exists {
		return ""
	}
	role, _ := val.(domain.Role)
	return role
}
