package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/beercut-team/backend-boilerplate/internal/config"
	"github.com/beercut-team/backend-boilerplate/internal/handler"
	"github.com/beercut-team/backend-boilerplate/internal/middleware"
	"github.com/beercut-team/backend-boilerplate/internal/repository"
	"github.com/beercut-team/backend-boilerplate/internal/service"
	"gorm.io/gorm"
)

func NewRouter(cfg *config.Config, db *gorm.DB) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://beercut.tech", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// --- DI wiring ---
	userRepo := repository.NewUserRepository(db)
	tokenService := service.NewTokenService(cfg)
	authService := service.NewAuthService(userRepo, tokenService)
	authHandler := handler.NewAuthHandler(authService)

	// --- Serve OpenAPI docs ---
	r.StaticFile("/openapi.json", "./openapi.json")
	r.GET("/docs", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(200, scalarHTML)
	})

	api := r.Group("/api/v1")
	{
		// Public auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.Auth(tokenService))
		{
			protected.GET("/auth/me", authHandler.Me)
			protected.POST("/auth/logout", authHandler.Logout)
			protected.GET("/ping", func(c *gin.Context) {
				userID := middleware.GetUserID(c)
				c.JSON(200, gin.H{"message": "pong", "user_id": userID})
			})
		}
	}

	return r
}

const scalarHTML = `<!DOCTYPE html>
<html>
<head>
    <title>API Docs</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1"/>
</head>
<body>
    <script id="api-reference" data-url="/openapi.json"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`
