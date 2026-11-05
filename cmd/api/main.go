package main

import (
	"log/slog"

	"test-golang-project/config"
	"test-golang-project/internal/controller"
	"test-golang-project/internal/repository"
	"test-golang-project/internal/routes"
	"test-golang-project/internal/service"
	"test-golang-project/pkg/database"
	"test-golang-project/pkg/logger"
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
	var authController *controller.AuthController
	var userController *controller.UserController
	var revokedTokenRepo repository.RevokedTokenRepository

	if db != nil {
		userRepo := repository.NewUserRepository(db)
		revokedTokenRepo = repository.NewRevokedTokenRepository(db)
		authService := service.NewAuthService(userRepo, revokedTokenRepo, cfg)
		userService := service.NewUserService(userRepo)

		authController = controller.NewAuthController(authService)
		userController = controller.NewUserController(userService)
	}

	// 5. Setup Router & Routes
	r := routes.SetupRouter(routes.Config{
		AppConfig:        cfg,
		DB:               db,
		AuthController:   authController,
		UserController:   userController,
		RevokedTokenRepo: revokedTokenRepo,
	})

	// 6. Start HTTP Server
	slog.Info("Starting server", "app", cfg.AppName, "port", cfg.AppPort, "env", cfg.AppEnv)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		slog.Error("Server failed to run", "error", err)
	}
}
