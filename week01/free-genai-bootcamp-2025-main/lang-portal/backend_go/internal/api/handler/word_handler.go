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

type WordHandler struct {
	wordService service.WordService
	metrics     *metrics.Metrics
}

func NewWordHandler(wordService service.WordService, metrics *metrics.Metrics) *WordHandler {
	return &WordHandler{
		wordService: wordService,
		metrics:     metrics,
	}
}

func (h *WordHandler) CreateWord(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())
	timer := h.metrics.NewTimer("handler.word.create")
	defer timer.ObserveDuration()

	var word models.Word
	if err := c.ShouldBindJSON(&word); err != nil {
		log.Error("Failed to bind JSON", "error", err)
		h.metrics.IncCounter("handler.word.create.error")
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(
			errors.ErrInvalidInput.String(),
			"Invalid request body",
			nil,
		))
		return
	}

	if err := h.wordService.CreateWord(c.Request.Context(), &word); err != nil {
		log.Error("Failed to create word", "error", err)
		h.metrics.IncCounter("handler.word.create.error")
		c.JSON(getErrorStatus(err), response.NewErrorResponse(
			getErrorCode(err),
			"Failed to create word",
			nil,
		))
		return
	}

	h.metrics.IncCounter("handler.word.create.success")
	c.JSON(http.StatusCreated, response.NewSuccessResponse(word))
}

func (h *WordHandler) GetWord(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())
	timer := h.metrics.NewTimer("handler.word.get")
	defer timer.ObserveDuration()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Error("Invalid word ID", "error", err)
		h.metrics.IncCounter("handler.word.get.error")
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(
			errors.ErrInvalidInput.String(),
			"Invalid word ID",
			nil,
		))
		return
	}

	word, err := h.wordService.GetWord(c.Request.Context(), id)
	if err != nil {
		log.Error("Failed to get word", "error", err, "id", id)
		h.metrics.IncCounter("handler.word.get.error")
		c.JSON(getErrorStatus(err), response.NewErrorResponse(
			getErrorCode(err),
			"Failed to get word",
			nil,
		))
		return
	}

	h.metrics.IncCounter("handler.word.get.success")
	c.JSON(http.StatusOK, response.NewSuccessResponse(word))
}

func (h *WordHandler) ListWords(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())
	timer := h.metrics.NewTimer("handler.word.list")
	defer timer.ObserveDuration()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	words, err := h.wordService.ListWords(c.Request.Context(), page, pageSize)
	if err != nil {
		log.Error("Failed to list words", "error", err)
		h.metrics.IncCounter("handler.word.list.error")
		c.JSON(getErrorStatus(err), response.NewErrorResponse(
			getErrorCode(err),
			"Failed to list words",
			nil,
		))
		return
	}

	h.metrics.IncCounter("handler.word.list.success")
	c.JSON(http.StatusOK, response.NewSuccessResponse(words))
} 