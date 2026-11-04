package controller

import (
	"errors"
	"net/http"

	"test-golang-project/internal/entity"
	"test-golang-project/internal/service"
	"test-golang-project/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService service.UserService
}

func NewUserController(userService service.UserService) *UserController {
	return &UserController{userService: userService}
}

func (ctrl *UserController) Register(c *gin.Context) {
	var req entity.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	res, err := ctrl.userService.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			response.Error(c, http.StatusConflict, err.Error(), nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to register user", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "User registered successfully", res)
}

func (ctrl *UserController) Login(c *gin.Context) {
	var req entity.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	res, err := ctrl.userService.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Error(c, http.StatusUnauthorized, err.Error(), nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to login", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Login successful", res)
}

func (ctrl *UserController) GetProfile(c *gin.Context) {
	// userID is set in Gin context by AuthMiddleware
	val, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userID, ok := val.(uint)
	if !ok {
		response.Error(c, http.StatusInternalServerError, "Invalid user ID type in context", nil)
		return
	}

	res, err := ctrl.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.Error(c, http.StatusNotFound, err.Error(), nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get profile", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Profile retrieved successfully", res)
}
