package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/your-org/lang-portal/internal/api/response"
)

// Recovery middleware handles panics and returns a 500 error
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the stack trace
				stack := debug.Stack()
				fmt.Printf("Recovery from panic: %v\nStack: %s\n", err, stack)

				// Return error response
				c.JSON(http.StatusInternalServerError, response.NewErrorResponse(
					"INTERNAL_SERVER_ERROR",
					"An unexpected error occurred",
					nil,
				))
				c.Abort()
			}
		}()
		c.Next()
	}
} 