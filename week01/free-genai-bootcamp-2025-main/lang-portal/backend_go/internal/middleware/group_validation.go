package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/your-org/lang-portal/internal/api/response"
	"github.com/your-org/lang-portal/internal/models"
)

func ValidateGroupInput() gin.HandlerFunc {
	return func(c *gin.Context) {
		var group models.Group
		if err := c.ShouldBindJSON(&group); err != nil {
			c.JSON(http.StatusBadRequest, response.NewErrorResponse(
				"INVALID_INPUT",
				"Invalid group data format",
				err.Error(),
			))
			c.Abort()
			return
		}

		// Validate required fields
		if group.Name == "" {
			c.JSON(http.StatusBadRequest, response.NewErrorResponse(
				"INVALID_INPUT",
				"Group name is required",
				nil,
			))
			c.Abort()
			return
		}

		// Validate name length
		if len(group.Name) > 100 {
			c.JSON(http.StatusBadRequest, response.NewErrorResponse(
				"INVALID_INPUT",
				"Group name is too long (max 100 characters)",
				nil,
			))
			c.Abort()
			return
		}

		c.Next()
	}
} 