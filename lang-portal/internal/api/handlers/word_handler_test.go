package handlers

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "encoding/json"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/karl247ai/lang-portal/internal/models"
    "bytes"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *WordHandler) {
    gin.SetMode(gin.TestMode)
    r := gin.Default()
    
    db := setupTestDB(t)
    repo := repository.NewWordRepository(db)
    handler := NewWordHandler(repo)
    
    return r, handler
}

type ResponseData struct {
    Data       []models.Word `json:"data"`
    Pagination struct {
        Page  int `json:"page"`
        Limit int `json:"limit"`
    } `json:"pagination"`
}

func TestWordHandler_GetWords(t *testing.T) {
    tests := []struct {
        name       string
        url        string
        wantStatus int
        wantCount  int
    }{
        {
            name:       "success_default_pagination",
            url:        "/api/v1/words",
            wantStatus: http.StatusOK,
            wantCount:  1,
        },
        {
            name:       "success_with_pagination",
            url:        "/api/v1/words?page=1&limit=5",
            wantStatus: http.StatusOK,
            wantCount:  1,
        },
        {
            name:       "invalid_page_number",
            url:        "/api/v1/words?page=invalid",
            wantStatus: http.StatusBadRequest,
            wantCount:  0,
        },
    }

    r, handler := setupTestRouter(t)
    r.GET("/api/v1/words", handler.GetWords)

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            w := httptest.NewRecorder()
            req, _ := http.NewRequest("GET", tt.url, nil)
            r.ServeHTTP(w, req)

            assert.Equal(t, tt.wantStatus, w.Code)

            if tt.wantStatus == http.StatusOK {
                var response ResponseData
                err := json.NewDecoder(w.Body).Decode(&response)
                assert.NoError(t, err)
                assert.Len(t, response.Data, tt.wantCount)
                assert.Greater(t, response.Pagination.Limit, 0)
            }
        })
    }
}

func TestWordHandler_CreateWord(t *testing.T) {
    tests := []struct {
        name       string
        payload    map[string]interface{}
        wantStatus int
    }{
        {
            name: "success_create_word",
            payload: map[string]interface{}{
                "japanese": "猫",
                "romaji":   "neko",
                "english":  "cat",
                "parts":    json.RawMessage(`{"type": "noun"}`),
            },
            wantStatus: http.StatusCreated,
        },
        {
            name: "missing_required_fields",
            payload: map[string]interface{}{
                "japanese": "猫",
            },
            wantStatus: http.StatusBadRequest,
        },
    }

    r, handler := setupTestRouter(t)
    r.POST("/api/v1/words", handler.CreateWord)

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            jsonData, _ := json.Marshal(tt.payload)
            w := httptest.NewRecorder()
            req, _ := http.NewRequest("POST", "/api/v1/words", bytes.NewBuffer(jsonData))
            req.Header.Set("Content-Type", "application/json")
            r.ServeHTTP(w, req)

            assert.Equal(t, tt.wantStatus, w.Code)
            
            if tt.wantStatus == http.StatusCreated {
                var response struct {
                    Data models.Word `json:"data"`
                }
                err := json.NewDecoder(w.Body).Decode(&response)
                assert.NoError(t, err)
                assert.NotZero(t, response.Data.ID)
                assert.Equal(t, tt.payload["japanese"], response.Data.Japanese)
            }
        })
    }
}

func TestWordHandler_UpdateWord(t *testing.T) {
    tests := []struct {
        name       string
        wordID     string
        payload    map[string]interface{}
        wantStatus int
    }{
        {
            name:   "success_update_word",
            wordID: "1",
            payload: map[string]interface{}{
                "japanese": "犬",
                "romaji":   "inu",
                "english":  "dog",
            },
            wantStatus: http.StatusOK,
        },
        {
            name:   "invalid_word_id",
            wordID: "999",
            payload: map[string]interface{}{
                "japanese": "犬",
            },
            wantStatus: http.StatusNotFound,
        },
    }

    r, handler := setupTestRouter(t)
    r.PUT("/api/v1/words/:id", handler.UpdateWord)

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            jsonData, _ := json.Marshal(tt.payload)
            w := httptest.NewRecorder()
            req, _ := http.NewRequest("PUT", "/api/v1/words/"+tt.wordID, bytes.NewBuffer(jsonData))
            req.Header.Set("Content-Type", "application/json")
            r.ServeHTTP(w, req)

            assert.Equal(t, tt.wantStatus, w.Code)
        })
    }
}

func TestWordHandler_DeleteWord(t *testing.T) {
    tests := []struct {
        name       string
        wordID     string
        wantStatus int
    }{
        {
            name:       "success_delete_word",
            wordID:     "1",
            wantStatus: http.StatusNoContent,
        },
        {
            name:       "word_not_found",
            wordID:     "999",
            wantStatus: http.StatusNotFound,
        },
        {
            name:       "invalid_id",
            wordID:     "invalid",
            wantStatus: http.StatusBadRequest,
        },
    }

    r, handler := setupTestRouter(t)
    r.DELETE("/api/v1/words/:id", handler.DeleteWord)

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            w := httptest.NewRecorder()
            req, _ := http.NewRequest("DELETE", "/api/v1/words/"+tt.wordID, nil)
            r.ServeHTTP(w, req)

            assert.Equal(t, tt.wantStatus, w.Code)
        })
    }
}