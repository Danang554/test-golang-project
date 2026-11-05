package controller

import (
	"net/http"

	"test-golang-project/internal/service"
	"test-golang-project/pkg/apperror"
	"test-golang-project/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService service.UserService
}

func NewUserController(userService service.UserService) *UserController {
	return &UserController{userService: userService}
}

func (ctrl *UserController) GetProfile(c *gin.Context) {
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

	res, err := ctrl.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Profile retrieved successfully", res)
}
