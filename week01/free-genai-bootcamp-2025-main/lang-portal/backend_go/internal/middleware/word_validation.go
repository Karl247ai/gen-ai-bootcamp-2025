package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/your-org/lang-portal/internal/api/response"
	"github.com/your-org/lang-portal/internal/models"
)

func ValidateWordInput() gin.HandlerFunc {
	return func(c *gin.Context) {
		var word models.Word
		if err := c.ShouldBindJSON(&word); err != nil {
			c.JSON(http.StatusBadRequest, response.NewErrorResponse(
				"INVALID_INPUT",
				"Invalid word data format",
				err.Error(),
			))
			c.Abort()
			return
		}

		// Validate required fields
		if word.Japanese == "" {
			c.JSON(http.StatusBadRequest, response.NewErrorResponse(
				"INVALID_INPUT",
				"Japanese text is required",
				nil,
			))
			c.Abort()
			return
		}

		if word.Romaji == "" {
			c.JSON(http.StatusBadRequest, response.NewErrorResponse(
				"INVALID_INPUT",
				"Romaji is required",
				nil,
			))
			c.Abort()
			return
		}

		if word.English == "" {
			c.JSON(http.StatusBadRequest, response.NewErrorResponse(
				"INVALID_INPUT",
				"English translation is required",
				nil,
			))
			c.Abort()
			return
		}

		c.Next()
	}
} 