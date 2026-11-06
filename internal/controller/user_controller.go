package controller

import (
	"net/http"
	"strconv"

	"test-golang-project/internal/dto"
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

func parseIDParam(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, apperror.BadRequest("Invalid user ID parameter")
	}
	return uint(id), nil
}

func (ctrl *UserController) GetProfile(c *gin.Context) {
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

func (ctrl *UserController) UpdateProfile(c *gin.Context) {
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

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	res, err := ctrl.userService.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Profile updated successfully", res)
}

func (ctrl *UserController) GetUsers(c *gin.Context) {
	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.HandleError(c, err)
		return
	}

	users, total, err := ctrl.userService.GetUsers(c.Request.Context(), &query)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Users retrieved successfully", users, query.GetPage(), query.GetLimit(), total)
}

func (ctrl *UserController) GetUserByID(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	res, err := ctrl.userService.GetUserByID(c.Request.Context(), id)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "User retrieved successfully", res)
}

func (ctrl *UserController) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	res, err := ctrl.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "User created successfully", res)
}

func (ctrl *UserController) UpdateUser(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	res, err := ctrl.userService.UpdateUser(c.Request.Context(), id, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "User updated successfully", res)
}

func (ctrl *UserController) DeleteUser(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	if err := ctrl.userService.DeleteUser(c.Request.Context(), id); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "User deleted successfully", nil)
}

func (ctrl *UserController) GetDeletedUsers(c *gin.Context) {
	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.HandleError(c, err)
		return
	}

	users, total, err := ctrl.userService.GetDeletedUsers(c.Request.Context(), &query)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Deleted users retrieved successfully", users, query.GetPage(), query.GetLimit(), total)
}

func (ctrl *UserController) RestoreUser(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	res, err := ctrl.userService.RestoreUser(c.Request.Context(), id)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "User restored successfully", res)
}

func (ctrl *UserController) ForceDeleteUser(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	if err := ctrl.userService.ForceDeleteUser(c.Request.Context(), id); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "User permanently deleted successfully", nil)
}
