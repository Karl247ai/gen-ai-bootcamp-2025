package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/your-org/lang-portal/internal/api/response"
	"github.com/your-org/lang-portal/internal/errors"
	"github.com/your-org/lang-portal/internal/logger"
	"github.com/your-org/lang-portal/internal/metrics"
	"github.com/your-org/lang-portal/internal/models"
	"github.com/your-org/lang-portal/internal/service"
)

type GroupHandler struct {
	groupService service.GroupService
	metrics      *metrics.Metrics
}

func NewGroupHandler(groupService service.GroupService, metrics *metrics.Metrics) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
		metrics:      metrics,
	}
}

func (h *GroupHandler) CreateGroup(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())
	timer := h.metrics.NewTimer("handler.group.create")
	defer timer.ObserveDuration()

	var group models.Group
	if err := c.ShouldBindJSON(&group); err != nil {
		log.Error("Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(
			errors.ErrInvalidInput.String(),
			"Invalid request body",
			nil,
		))
		return
	}

	if err := h.groupService.CreateGroup(c.Request.Context(), &group); err != nil {
		log.Error("Failed to create group", "error", err)
		h.metrics.IncCounter("handler.group.create.error")
		c.JSON(getErrorStatus(err), response.NewErrorResponse(
			getErrorCode(err),
			"Failed to create group",
			nil,
		))
		return
	}

	h.metrics.IncCounter("handler.group.create.success")
	c.JSON(http.StatusCreated, response.NewSuccessResponse(group))
}

func (h *GroupHandler) AddWordsToGroup(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())
	timer := h.metrics.NewTimer("handler.group.add_words")
	defer timer.ObserveDuration()

	groupID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Error("Invalid group ID", "error", err)
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(
			errors.ErrInvalidInput.String(),
			"Invalid group ID",
			nil,
		))
		return
	}

	var request struct {
		WordIDs []int `json:"word_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error("Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(
			errors.ErrInvalidInput.String(),
			"Invalid request body",
			nil,
		))
		return
	}

	if err := h.groupService.AddWordsToGroup(c.Request.Context(), groupID, request.WordIDs); err != nil {
		log.Error("Failed to add words to group", "error", err, "groupID", groupID)
		h.metrics.IncCounter("handler.group.add_words.error")
		c.JSON(getErrorStatus(err), response.NewErrorResponse(
			getErrorCode(err),
			"Failed to add words to group",
			nil,
		))
		return
	}

	h.metrics.IncCounter("handler.group.add_words.success")
	c.JSON(http.StatusOK, response.NewSuccessResponse(nil))
}

func (h *GroupHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse("INVALID_ID", "Invalid group ID", nil))
		return
	}

	group, err := h.groupService.GetGroup(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse("DATABASE_ERROR", "Failed to get group", nil))
		return
	}
	if group == nil {
		c.JSON(http.StatusNotFound, response.NewErrorResponse("NOT_FOUND", "Group not found", nil))
		return
	}

	c.JSON(http.StatusOK, response.NewResponse(group))
}

func (h *GroupHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	groups, err := h.groupService.ListGroups(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse("DATABASE_ERROR", "Failed to list groups", nil))
		return
	}

	resp := response.PaginatedResponse{
		Response: response.NewResponse(groups),
		Pagination: response.Pagination{
			CurrentPage:  page,
			ItemsPerPage: pageSize,
		},
	}

	c.JSON(http.StatusOK, resp)
}

// Helper functions for error handling
func getErrorStatus(err error) int {
	if errors.IsErrorCode(err, errors.ErrInvalidInput) {
		return http.StatusBadRequest
	}
	if errors.IsErrorCode(err, errors.ErrDBNotFound) {
		return http.StatusNotFound
	}
	if errors.IsErrorCode(err, errors.ErrTimeout) {
		return http.StatusGatewayTimeout
	}
	return http.StatusInternalServerError
}

func getErrorCode(err error) string {
	var appErr *errors.AppError
	if errors.As(err, &appErr) {
		return string(appErr.Code)
	}
	return string(errors.ErrInternal)
} 