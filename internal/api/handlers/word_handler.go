package handlers

import (
    "github.com/gin-gonic/gin"
    "net/http"
    "strconv"
    "github.com/karl247ai/lang-portal/internal/repository"
    "github.com/karl247ai/lang-portal/internal/models"
    "math"
)

// @title           Language Portal API
// @version         1.0
// @description     API for managing language learning vocabulary
// @host           localhost:8080
// @BasePath       /api/v1

type WordHandler struct {
    repo *repository.WordRepository
}

func NewWordHandler(repo *repository.WordRepository) *WordHandler {
    return &WordHandler{repo: repo}
}

// GetWords godoc
// @Summary     Get words list
// @Description Get paginated list of words
// @Tags        words
// @Accept      json
// @Produce     json
// @Param       page  query    int  false  "Page number"
// @Param       limit query    int  false  "Items per page"
// @Success      200  {object}  models.PaginatedResponse
// @Failure     500  {object}  models.ErrorResponse
// @Router      /words [get]
func (h *WordHandler) GetWords(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
    offset := (page - 1) * limit

    words, err := h.repo.GetWords(c.Request.Context(), limit, offset)
    if err != nil {
        c.Error(err)
        return
    }

    totalItems, err := h.repo.GetWordsCount(c.Request.Context())
    if err != nil {
        c.Error(err)
        return
    }

    totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))

    response := models.PaginatedResponse{
        Data: words,
        Pagination: models.PaginationMeta{
            CurrentPage:  page,
            TotalPages:   totalPages,
            TotalItems:   totalItems,
            ItemsPerPage: limit,
        },
    }

    c.JSON(http.StatusOK, response)
}

// CreateWord godoc
// @Summary     Create new word
// @Description Add a new word to the vocabulary
// @Tags        words
// @Accept      json
// @Produce     json
// @Param       word body      models.Word  true  "Word object"
// @Success     201  {object}  models.WordResponse
// @Failure     400  {object}  models.ErrorResponse
// @Failure     500  {object}  models.ErrorResponse
// @Router      /words [post]
func (h *WordHandler) CreateWord(c *gin.Context) {
    var word models.Word
    if err := c.ShouldBindJSON(&word); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Validate required fields
    if word.Japanese == "" || word.Romaji == "" || word.English == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
        return
    }

    if err := h.repo.CreateWord(c.Request.Context(), &word); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"data": word})
}

// UpdateWord godoc
// @Summary     Update word
// @Description Update an existing word
// @Tags        words
// @Accept      json
// @Produce     json
// @Param       id   path      int         true   "Word ID"
// @Param       word body      models.Word true   "Word object"
// @Success     200  {object}  models.WordResponse
// @Failure     400  {object}  models.ErrorResponse
// @Failure     404  {object}  models.ErrorResponse
// @Failure     500  {object}  models.ErrorResponse
// @Router      /words/{id} [put]
func (h *WordHandler) UpdateWord(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid word id"})
        return
    }

    var word models.Word
    if err := c.ShouldBindJSON(&word); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := validator.ValidateWord(&word); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := h.repo.UpdateWord(c.Request.Context(), id, &word); err != nil {
        if err.Error() == "word not found" {
            c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": word})
}

// DeleteWord godoc
// @Summary     Delete word
// @Description Delete an existing word
// @Tags        words
// @Accept      json
// @Produce     json
// @Param       id   path      int  true  "Word ID"
// @Success     204  "No Content"
// @Failure     400  {object}  models.ErrorResponse
// @Failure     404  {object}  models.ErrorResponse
// @Failure     500  {object}  models.ErrorResponse
// @Router      /words/{id} [delete]
func (h *WordHandler) DeleteWord(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid word id"})
        return
    }

    if err := h.repo.DeleteWord(c.Request.Context(), id); err != nil {
        if err.Error() == "word not found" {
            c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.Status(http.StatusNoContent)
}