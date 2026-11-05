package routes

import (
	"net/http"

	"test-golang-project/config"
	"test-golang-project/internal/controller"
	"test-golang-project/internal/middleware"
	"test-golang-project/internal/repository"
	"test-golang-project/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Config struct {
	AppConfig        *config.Config
	DB               *gorm.DB
	AuthController   *controller.AuthController
	UserController   *controller.UserController
	RevokedTokenRepo repository.RevokedTokenRepository
}

func SetupRouter(c Config) *gin.Engine {
	r := gin.Default()

	// Global Middlewares
	r.Use(gin.Recovery())

	// Health Check Route
	r.GET("/health", func(ctx *gin.Context) {
		dbStatus := "connected"
		if c.DB == nil {
			dbStatus = "disconnected"
		}
		response.Success(ctx, http.StatusOK, "Service is running healthy", gin.H{
			"app_name": c.AppConfig.AppName,
			"database": dbStatus,
		})
	})

	// API Group
	api := r.Group("/api")
	{
		// Public Auth routes
		auth := api.Group("/auth")
		{
			if c.AuthController != nil {
				auth.POST("/register", c.AuthController.Register)
				auth.POST("/login", c.AuthController.Login)
			}
		}

		// Protected routes (Requires JWT Token)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(c.AppConfig, c.RevokedTokenRepo))
		{
			if c.AuthController != nil {
				protected.POST("/auth/logout", c.AuthController.Logout)
			}
			if c.UserController != nil {
				protected.GET("/profile", c.UserController.GetProfile)
			}
		}
	}

	return r
}
