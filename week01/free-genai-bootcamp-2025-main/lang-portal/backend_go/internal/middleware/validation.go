package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/your-org/lang-portal/internal/api/response"
)

// ValidatePagination validates pagination parameters
func ValidatePagination() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

		if page < 1 {
			c.JSON(http.StatusBadRequest, response.NewErrorResponse(
				"INVALID_PAGE",
				"Page number must be greater than 0",
				nil,
			))
			c.Abort()
			return
		}

		if limit < 1 || limit > 100 {
			c.JSON(http.StatusBadRequest, response.NewErrorResponse(
				"INVALID_LIMIT",
				"Limit must be between 1 and 100",
				nil,
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateID validates ID parameters
func ValidateID(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param(paramName))
		if err != nil || id < 1 {
			c.JSON(http.StatusBadRequest, response.NewErrorResponse(
				"INVALID_ID",
				"Invalid ID format",
				nil,
			))
			c.Abort()
			return
		}

		c.Next()
	}
} 