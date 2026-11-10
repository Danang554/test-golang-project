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

type CategoryController struct {
	categoryService service.CategoryService
}

func NewCategoryController(categoryService service.CategoryService) *CategoryController {
	return &CategoryController{categoryService: categoryService}
}

func parseCategoryIDParam(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, apperror.BadRequest("Invalid category ID parameter")
	}
	return uint(id), nil
}

func getUserID(c *gin.Context) (uint, error) {
	val, exists := c.Get("user_id")
	if !exists {
		return 0, apperror.Unauthorized("Unauthorized. Please login first")
	}
	userID, ok := val.(uint)
	if !ok {
		return 0, apperror.Internal(nil, "Invalid user ID type in context")
	}
	return userID, nil
}

func (ctrl *CategoryController) Create(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	res, err := ctrl.categoryService.Create(c.Request.Context(), userID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Category created successfully", res)
}

func (ctrl *CategoryController) GetCategories(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.HandleError(c, err)
		return
	}

	categories, total, err := ctrl.categoryService.GetCategories(c.Request.Context(), userID, &query)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Categories retrieved successfully", categories, query.GetPage(), query.GetLimit(), total)
}

func (ctrl *CategoryController) GetByID(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	id, err := parseCategoryIDParam(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	res, err := ctrl.categoryService.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Category retrieved successfully", res)
}

func (ctrl *CategoryController) Update(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	id, err := parseCategoryIDParam(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	res, err := ctrl.categoryService.Update(c.Request.Context(), userID, id, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Category updated successfully", res)
}

func (ctrl *CategoryController) Delete(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	id, err := parseCategoryIDParam(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	if err := ctrl.categoryService.Delete(c.Request.Context(), userID, id); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Category deleted successfully", nil)
}

func (ctrl *CategoryController) GetDeletedCategories(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.HandleError(c, err)
		return
	}

	categories, total, err := ctrl.categoryService.GetDeletedCategories(c.Request.Context(), userID, &query)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Deleted categories retrieved successfully", categories, query.GetPage(), query.GetLimit(), total)
}

func (ctrl *CategoryController) Restore(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	id, err := parseCategoryIDParam(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	res, err := ctrl.categoryService.Restore(c.Request.Context(), userID, id)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Category restored successfully", res)
}

func (ctrl *CategoryController) ForceDelete(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	id, err := parseCategoryIDParam(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	if err := ctrl.categoryService.ForceDelete(c.Request.Context(), userID, id); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Category permanently deleted successfully", nil)
}
