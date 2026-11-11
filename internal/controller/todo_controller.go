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

type TodoController struct {
	todoService service.TodoService
}

func NewTodoController(todoService service.TodoService) *TodoController {
	return &TodoController{todoService: todoService}
}

func parseTodoIDParam(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, apperror.BadRequest("Invalid todo ID parameter")
	}
	return uint(id), nil
}

func getTodoUserID(c *gin.Context) (uint, error) {
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

func (ctrl *TodoController) Create(c *gin.Context) {
	userID, err := getTodoUserID(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req dto.CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	res, err := ctrl.todoService.Create(c.Request.Context(), userID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Todo created successfully", res)
}

func (ctrl *TodoController) GetTodos(c *gin.Context) {
	userID, err := getTodoUserID(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var query dto.TodoQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.HandleError(c, err)
		return
	}

	todos, total, err := ctrl.todoService.GetTodos(c.Request.Context(), userID, &query)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Todos retrieved successfully", todos, query.GetPage(), query.GetLimit(), total)
}

func (ctrl *TodoController) GetByID(c *gin.Context) {
	userID, err := getTodoUserID(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	id, err := parseTodoIDParam(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	res, err := ctrl.todoService.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Todo retrieved successfully", res)
}

func (ctrl *TodoController) Update(c *gin.Context) {
	userID, err := getTodoUserID(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	id, err := parseTodoIDParam(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req dto.UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	res, err := ctrl.todoService.Update(c.Request.Context(), userID, id, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Todo updated successfully", res)
}

func (ctrl *TodoController) ToggleComplete(c *gin.Context) {
	userID, err := getTodoUserID(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	id, err := parseTodoIDParam(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	res, err := ctrl.todoService.ToggleComplete(c.Request.Context(), userID, id)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	msg := "Todo marked as completed"
	if !res.IsCompleted {
		msg = "Todo marked as uncompleted"
	}

	response.Success(c, http.StatusOK, msg, res)
}

func (ctrl *TodoController) Delete(c *gin.Context) {
	userID, err := getTodoUserID(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	id, err := parseTodoIDParam(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	if err := ctrl.todoService.Delete(c.Request.Context(), userID, id); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Todo deleted successfully", nil)
}

func (ctrl *TodoController) GetDeletedTodos(c *gin.Context) {
	userID, err := getTodoUserID(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.HandleError(c, err)
		return
	}

	todos, total, err := ctrl.todoService.GetDeletedTodos(c.Request.Context(), userID, &query)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Deleted todos retrieved successfully", todos, query.GetPage(), query.GetLimit(), total)
}

func (ctrl *TodoController) Restore(c *gin.Context) {
	userID, err := getTodoUserID(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	id, err := parseTodoIDParam(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	res, err := ctrl.todoService.Restore(c.Request.Context(), userID, id)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Todo restored successfully", res)
}

func (ctrl *TodoController) ForceDelete(c *gin.Context) {
	userID, err := getTodoUserID(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	id, err := parseTodoIDParam(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	if err := ctrl.todoService.ForceDelete(c.Request.Context(), userID, id); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Todo permanently deleted successfully", nil)
}
