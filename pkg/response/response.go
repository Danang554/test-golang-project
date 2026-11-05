package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"test-golang-project/pkg/apperror"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

func Success(c *gin.Context, statusCode int, message string, data any) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, statusCode int, message string, errors any) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

// FormatValidationError converts raw validator.ValidationErrors into clean, user-friendly messages
func FormatValidationError(err error) any {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		errMap := make(map[string]string)
		for _, fe := range ve {
			field := fe.Field()
			switch fe.Tag() {
			case "required":
				errMap[field] = fmt.Sprintf("%s is required", field)
			case "email":
				errMap[field] = fmt.Sprintf("%s must be a valid email address", field)
			case "min":
				errMap[field] = fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
			case "max":
				errMap[field] = fmt.Sprintf("%s must not exceed %s characters", field, fe.Param())
			default:
				errMap[field] = fmt.Sprintf("%s failed on '%s' validation", field, fe.Tag())
			}
		}
		return errMap
	}

	if errors.Is(err, io.EOF) {
		return "Request body cannot be empty"
	}

	return "Invalid request payload format"
}

// ValidationError responds with a clean 400 Bad Request containing formatted error messages
func ValidationError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Success: false,
		Message: "Validation failed",
		Errors:  FormatValidationError(err),
	})
}

// HandleError is the centralized error handler that inspects any error and responds appropriately
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// 1. Validation or JSON binding errors
	var ve validator.ValidationErrors
	var jsonSyntaxErr *json.SyntaxError
	var jsonUnmarshalErr *json.UnmarshalTypeError
	if errors.As(err, &ve) || errors.Is(err, io.EOF) || errors.As(err, &jsonSyntaxErr) || errors.As(err, &jsonUnmarshalErr) {
		ValidationError(c, err)
		return
	}

	// 2. Structured AppError
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		if appErr.Code >= 500 {
			slog.Error("Internal server error", "error", appErr.RawErr, "path", c.Request.URL.Path)
		}
		c.JSON(appErr.Code, APIResponse{
			Success: false,
			Message: appErr.Message,
			Errors:  appErr.Details,
		})
		return
	}

	// 3. Fallback unhandled error (500)
	slog.Error("Unhandled internal error", "error", err, "path", c.Request.URL.Path)
	c.JSON(http.StatusInternalServerError, APIResponse{
		Success: false,
		Message: "Internal server error",
	})
}
