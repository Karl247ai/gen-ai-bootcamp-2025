package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/your-org/lang-portal/internal/api/response"
	"github.com/your-org/lang-portal/internal/service"
)

type WordGroupHandler struct {
	wordGroupService service.WordGroupService
}

func NewWordGroupHandler(wordGroupService service.WordGroupService) *WordGroupHandler {
	return &WordGroupHandler{
		wordGroupService: wordGroupService,
	}
}

// AddWordToGroup handles adding a word to a group
func (h *WordGroupHandler) AddWordToGroup(c *gin.Context) {
	groupID, err := strconv.Atoi(c.Param("groupId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(
			"INVALID_GROUP_ID",
			"Invalid group ID format",
			nil,
		))
		return
	}

	wordID, err := strconv.Atoi(c.Param("wordId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(
			"INVALID_WORD_ID",
			"Invalid word ID format",
			nil,
		))
		return
	}

	if err := h.wordGroupService.AddWordToGroup(wordID, groupID); err != nil {
		switch err {
		case service.ErrWordNotFound:
			c.JSON(http.StatusNotFound, response.NewErrorResponse(
				"WORD_NOT_FOUND",
				"Word not found",
				nil,
			))
		case service.ErrGroupNotFound:
			c.JSON(http.StatusNotFound, response.NewErrorResponse(
				"GROUP_NOT_FOUND",
				"Group not found",
				nil,
			))
		default:
			c.JSON(http.StatusInternalServerError, response.NewErrorResponse(
				"DATABASE_ERROR",
				"Failed to add word to group",
				nil,
			))
		}
		return
	}

	c.JSON(http.StatusOK, response.NewResponse(nil))
}

// RemoveWordFromGroup handles removing a word from a group
func (h *WordGroupHandler) RemoveWordFromGroup(c *gin.Context) {
	groupID, err := strconv.Atoi(c.Param("groupId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(
			"INVALID_GROUP_ID",
			"Invalid group ID format",
			nil,
		))
		return
	}

	wordID, err := strconv.Atoi(c.Param("wordId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(
			"INVALID_WORD_ID",
			"Invalid word ID format",
			nil,
		))
		return
	}

	if err := h.wordGroupService.RemoveWordFromGroup(wordID, groupID); err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(
			"DATABASE_ERROR",
			"Failed to remove word from group",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, response.NewResponse(nil))
}

// ListGroupWords handles listing all words in a group
func (h *WordGroupHandler) ListGroupWords(c *gin.Context) {
	groupID, err := strconv.Atoi(c.Param("groupId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(
			"INVALID_GROUP_ID",
			"Invalid group ID format",
			nil,
		))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	words, err := h.wordGroupService.GetGroupWords(groupID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(
			"DATABASE_ERROR",
			"Failed to list group words",
			nil,
		))
		return
	}

	totalCount, err := h.wordGroupService.CountGroupWords(groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(
			"DATABASE_ERROR",
			"Failed to count group words",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, response.PaginatedResponse{
		Response: response.NewResponse(words),
		Pagination: response.Pagination{
			CurrentPage:  page,
			ItemsPerPage: pageSize,
			TotalItems:   totalCount,
			TotalPages:   (totalCount + pageSize - 1) / pageSize,
		},
	})
} 