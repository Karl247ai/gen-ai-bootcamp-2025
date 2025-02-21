package test

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "bytes"
    "encoding/json"
    "github.com/stretchr/testify/assert"
    "github.com/karl247ai/lang-portal/internal/models"
)

func TestHealthCheck(t *testing.T) {
    router := setupTestRouter()
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/health", nil)
    router.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)
    assert.Contains(t, w.Body.String(), "\"status\":\"ok\"")
}

func TestWordManagement(t *testing.T) {
    tests := []testCase{
        {
            name:       "get_empty_words",
            method:     http.MethodGet,
            path:       "/api/v1/words",
            wantStatus: http.StatusOK,
            wantBody:   "\"data\":[]",
        },
        {
            name:   "create_valid_word",
            method: http.MethodPost,
            path:   "/api/v1/words",
            body: models.Word{
                Japanese: "猫",
                Romaji:   "neko",
                English:  "cat",
            },
            wantStatus: http.StatusCreated,
            wantBody:   "\"japanese\":\"猫\"",
        },
        {
            name:   "create_invalid_word",
            method: http.MethodPost,
            path:   "/api/v1/words",
            body: models.Word{},
            wantStatus: http.StatusBadRequest,
            wantBody:   "\"error\":\"missing required fields\"",
        },
        {
            name:       "get_words_after_create",
            method:     http.MethodGet,
            path:       "/api/v1/words",
            wantStatus: http.StatusOK,
            wantBody:   "\"japanese\":\"猫\"",
        },
    }

    router := setupTestRouter()

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            w := performRequest(router, tt)
            
            assert.Equal(t, tt.wantStatus, w.Code)
            assert.Contains(t, w.Body.String(), tt.wantBody)
        })
    }
}