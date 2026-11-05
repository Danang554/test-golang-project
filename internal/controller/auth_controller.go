package controller

import (
	"net/http"
	"time"

	"test-golang-project/internal/dto"
	"test-golang-project/internal/service"
	"test-golang-project/pkg/apperror"
	"test-golang-project/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService service.AuthService
}

func NewAuthController(authService service.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

func (ctrl *AuthController) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	res, err := ctrl.authService.Register(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "User registered successfully", res)
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	res, err := ctrl.authService.Login(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Login successful", res)
}

func (ctrl *AuthController) Logout(c *gin.Context) {
	// userID is set in Gin context by AuthMiddleware
	val, exists := c.Get("user_id")
	if !exists {
		response.HandleError(c, apperror.Unauthorized("Unauthorized"))
		return
	}

	userID, ok := val.(uint)
	if !ok {
		response.HandleError(c, apperror.Internal(nil, "Invalid user ID type in context"))
		return
	}

	tokenString, _ := c.Get("token_string")
	tokStr, _ := tokenString.(string)

	expVal, _ := c.Get("token_exp")
	expTime, _ := expVal.(time.Time)

	if err := ctrl.authService.Logout(c.Request.Context(), tokStr, userID, expTime); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Logout successful", nil)
}
