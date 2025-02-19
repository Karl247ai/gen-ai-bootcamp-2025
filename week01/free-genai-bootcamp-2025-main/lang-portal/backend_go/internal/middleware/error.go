package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/your-org/lang-portal/internal/api/response"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Only handle errors if there are any
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// Handle different types of errors
			switch e := err.Err.(type) {
			case *response.Error:
				// Already formatted error
				c.JSON(e.Status, response.NewErrorResponse(e.Code, e.Message, e.Details))
			default:
				// Unknown error
				c.JSON(http.StatusInternalServerError, response.NewErrorResponse(
					"INTERNAL_ERROR",
					"An unexpected error occurred",
					nil,
				))
			}
		}
	}
} 