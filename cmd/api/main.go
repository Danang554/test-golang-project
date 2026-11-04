package main

import (
	"log/slog"
	"net/http"

	"test-golang-project/config"
	"test-golang-project/internal/controller"
	"test-golang-project/internal/middleware"
	"test-golang-project/internal/repository"
	"test-golang-project/internal/service"
	"test-golang-project/pkg/database"
	"test-golang-project/pkg/logger"
	"test-golang-project/pkg/response"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load application configuration (.env)
	cfg := config.LoadConfig()

	// 2. Initialize logger (configured via .env)
	logger.InitLogger(cfg)

	// 3. Initialize database connection (MySQL)
	db, err := database.NewMySQLDatabase(cfg)
	if err != nil {
		slog.Warn("Could not connect to MySQL", "error", err)
		slog.Info("Server will continue running for health checks. Start MySQL to use DB endpoints.")
	}

	// 4. Initialize layers (Dependency Injection)
	var userController *controller.UserController
	if db != nil {
		userRepo := repository.NewUserRepository(db)
		userService := service.NewUserService(userRepo, cfg)
		userController = controller.NewUserController(userService)
	}

	// 5. Initialize Gin router
	r := gin.Default()

	// 6. Global Middlewares
	r.Use(gin.Recovery())

	// 7. Base Routes
	r.GET("/health", func(c *gin.Context) {
		dbStatus := "connected"
		if db == nil {
			dbStatus = "disconnected"
		}
		response.Success(c, http.StatusOK, "Service is running healthy", gin.H{
			"app_name": cfg.AppName,
			"database": dbStatus,
		})
	})

	// 8. API Routes
	api := r.Group("/api/v1")
	{
		// Public Auth routes
		auth := api.Group("/auth")
		{
			if userController != nil {
				auth.POST("/register", userController.Register)
				auth.POST("/login", userController.Login)
			}
		}

		// Protected routes (Require JWT)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			if userController != nil {
				protected.GET("/profile", userController.GetProfile)
			}
		}
	}

	// 9. Start HTTP Server
	slog.Info("Starting server", "app", cfg.AppName, "port", cfg.AppPort, "env", cfg.AppEnv)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		slog.Error("Server failed to run", "error", err)
	}
}
