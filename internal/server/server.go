package server

import (
	"github.com/beercut-team/backend-boilerplate/internal/config"
	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/handler"
	"github.com/beercut-team/backend-boilerplate/internal/middleware"
	"github.com/beercut-team/backend-boilerplate/internal/repository"
	"github.com/beercut-team/backend-boilerplate/internal/service"
	"github.com/beercut-team/backend-boilerplate/pkg/storage"
	"github.com/beercut-team/backend-boilerplate/pkg/telegram"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
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

	// --- Repositories ---
	userRepo := repository.NewUserRepository(db)
	districtRepo := repository.NewDistrictRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	patientRepo := repository.NewPatientRepository(db)
	checklistRepo := repository.NewChecklistRepository(db)
	mediaRepo := repository.NewMediaRepository(db)
	iolRepo := repository.NewIOLRepository(db)
	surgeryRepo := repository.NewSurgeryRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	notifRepo := repository.NewNotificationRepository(db)
	telegramRepo := repository.NewTelegramRepository(db)
	telegramTokenRepo := repository.NewTelegramTokenRepository(db)
	syncRepo := repository.NewSyncRepository(db)
	_ = auditRepo

	// --- Storage ---
	var store storage.Storage
	if cfg.StorageMode == "minio" {
		var err error
		store, err = storage.NewMinIOStorage(cfg)
		if err != nil {
			log.Warn().Err(err).Msg("MinIO unavailable, falling back to local storage")
			store = storage.NewLocalStorage(cfg.LocalUploadPath)
		}
	} else {
		store = storage.NewLocalStorage(cfg.LocalUploadPath)
	}

	// --- Telegram Bot (создаём рано, чтобы передать в сервисы) ---
	bot, err := telegram.NewBot(cfg.TelegramBotToken, cfg.BaseURL, patientRepo, telegramRepo, telegramTokenRepo, userRepo)
	if err != nil {
		log.Warn().Err(err).Msg("Telegram bot failed to start")
	}
	if bot != nil {
		bot.Start()
	}

	// --- Services ---
	tokenService := service.NewTokenService(cfg)
	authService := service.NewAuthServiceWithPatient(userRepo, patientRepo, telegramTokenRepo, tokenService)
	districtService := service.NewDistrictService(districtRepo)
	patientService := service.NewPatientService(patientRepo, checklistRepo, notifRepo, bot)
	checklistService := service.NewChecklistService(checklistRepo, patientRepo)
	mediaService := service.NewMediaService(mediaRepo, store)
	iolService := service.NewIOLService(iolRepo)
	surgeryService := service.NewSurgeryService(surgeryRepo, patientRepo, checklistRepo, notifRepo)
	commentService := service.NewCommentService(commentRepo, patientRepo, userRepo, notifRepo)
	notifService := service.NewNotificationService(notifRepo)
	pdfService := service.NewPDFService(patientRepo, checklistRepo)
	syncService := service.NewSyncService(syncRepo)

	// --- Scheduler ---
	scheduler := service.NewSchedulerService(checklistRepo, surgeryRepo, notifRepo, mediaRepo)
	scheduler.Start()

	// --- Handlers ---
	authHandler := handler.NewAuthHandler(authService)
	districtHandler := handler.NewDistrictHandler(districtService)
	patientHandler := handler.NewPatientHandler(patientService)
	checklistHandler := handler.NewChecklistHandler(checklistService)
	mediaHandler := handler.NewMediaHandler(mediaService)
	iolHandler := handler.NewIOLHandler(iolService)
	surgeryHandler := handler.NewSurgeryHandler(surgeryService)
	commentHandler := handler.NewCommentHandler(commentService)
	notifHandler := handler.NewNotificationHandler(notifService)
	printHandler := handler.NewPrintHandler(pdfService)
	syncHandler := handler.NewSyncHandler(syncService)
	adminHandler := handler.NewAdminHandler(authService, db)

	// --- Serve OpenAPI docs ---
	r.StaticFile("/openapi.json", "./openapi.json")
	r.GET("/docs", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(200, scalarHTML)
	})

	// --- Admin panel ---
	r.GET("/admin", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(200, adminHTML)
	})

	// --- Patient pages ---
	r.GET("/patient", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(200, patientPublicHTML)
	})
	r.GET("/patient/login", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(200, patientLoginHTML)
	})
	r.GET("/patient/portal", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(200, patientPortalHTML)
	})

	api := r.Group("/api/v1")
	{
		// Public auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/patient-login", authHandler.PatientLogin)
			auth.POST("/telegram-token-login", authHandler.TelegramTokenLogin)
			auth.POST("/refresh", authHandler.Refresh)
		}

		// Public patient status
		api.GET("/patients/public/:accessCode", patientHandler.GetPublic)

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.Auth(tokenService))
		{
			// Auth
			protected.GET("/auth/me", authHandler.Me)
			protected.POST("/auth/logout", authHandler.Logout)
			protected.GET("/ping", func(c *gin.Context) {
				userID := middleware.GetUserID(c)
				c.JSON(200, gin.H{"message": "pong", "user_id": userID})
			})

			// Districts (ADMIN only for mutations)
			districts := protected.Group("/districts")
			{
				districts.GET("", districtHandler.List)
				districts.GET("/:id", districtHandler.GetByID)
				adminDistricts := districts.Group("")
				adminDistricts.Use(middleware.RequireRole(domain.RoleAdmin))
				{
					adminDistricts.POST("", districtHandler.Create)
					adminDistricts.PATCH("/:id", districtHandler.Update)
					adminDistricts.DELETE("/:id", districtHandler.Delete)
				}
			}

			// Patients
			patients := protected.Group("/patients")
			{
				patients.GET("", patientHandler.List)
				patients.GET("/dashboard", patientHandler.Dashboard)
				patients.GET("/:id", patientHandler.GetByID)
				patients.POST("", middleware.RequireRole(domain.RoleDistrictDoctor, domain.RoleAdmin), patientHandler.Create)
				patients.PATCH("/:id", patientHandler.Update)
				patients.DELETE("/:id", middleware.RequireRole(domain.RoleAdmin), patientHandler.Delete)
				patients.POST("/:id/status", patientHandler.ChangeStatus)
				patients.POST("/:id/regenerate-code", middleware.RequireRole(domain.RoleAdmin), patientHandler.RegenerateAccessCode)
			}

			// Checklists
			checklists := protected.Group("/checklists")
			{
				checklists.GET("/patient/:patientId", checklistHandler.GetByPatient)
				checklists.GET("/patient/:patientId/progress", checklistHandler.GetProgress)
				checklists.PATCH("/:id", checklistHandler.UpdateItem)
				checklists.POST("/:id/review", middleware.RequireRole(domain.RoleSurgeon, domain.RoleAdmin), checklistHandler.ReviewItem)
			}

			// Media
			media := protected.Group("/media")
			{
				media.POST("/upload", mediaHandler.Upload)
				media.GET("/patient/:patientId", mediaHandler.GetByPatient)
				media.GET("/:id/download", mediaHandler.Download)
				media.GET("/:id/download-url", mediaHandler.DownloadURL)
				media.GET("/:id/thumbnail", mediaHandler.Thumbnail)
				media.DELETE("/:id", mediaHandler.Delete)
			}

			// IOL Calculator
			iol := protected.Group("/iol")
			{
				iol.POST("/calculate", iolHandler.Calculate)
				iol.GET("/patient/:patientId/history", iolHandler.History)
			}

			// Surgeries (SURGEON only for creation)
			surgeries := protected.Group("/surgeries")
			{
				surgeries.GET("", surgeryHandler.List)
				surgeries.GET("/:id", surgeryHandler.GetByID)
				surgeries.POST("", middleware.RequireRole(domain.RoleSurgeon, domain.RoleAdmin), surgeryHandler.Schedule)
				surgeries.PATCH("/:id", middleware.RequireRole(domain.RoleSurgeon, domain.RoleAdmin), surgeryHandler.Update)
				surgeries.DELETE("/:id", middleware.RequireRole(domain.RoleSurgeon, domain.RoleAdmin), surgeryHandler.Delete)
			}

			// Comments
			comments := protected.Group("/comments")
			{
				comments.POST("", commentHandler.Create)
				comments.GET("/patient/:patientId", commentHandler.GetByPatient)
				comments.POST("/patient/:patientId/read", commentHandler.MarkAsRead)
			}

			// Notifications
			notifications := protected.Group("/notifications")
			{
				notifications.GET("", notifHandler.List)
				notifications.GET("/unread-count", notifHandler.UnreadCount)
				notifications.POST("/:id/read", notifHandler.MarkAsRead)
				notifications.POST("/read-all", notifHandler.MarkAllAsRead)
			}

			// Print / PDF
			print := protected.Group("/print")
			{
				print.GET("/patient/:patientId/routing-sheet", printHandler.RoutingSheet)
				print.GET("/patient/:patientId/checklist-report", printHandler.ChecklistReport)
			}

			// Sync
			sync := protected.Group("/sync")
			{
				sync.POST("/push", syncHandler.Push)
				sync.GET("/pull", syncHandler.Pull)
			}

			// Admin
			admin := protected.Group("/admin")
			admin.Use(middleware.RequireRole(domain.RoleAdmin))
			{
				admin.GET("/users", adminHandler.ListUsers)
				admin.GET("/stats", adminHandler.Stats)
			}
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
