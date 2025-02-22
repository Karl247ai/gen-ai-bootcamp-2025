package test

import (
    "testing"
    "net/http"
    "github.com/stretchr/testify/assert"
    "github.com/karl247ai/lang-portal/internal/models"
)

func TestWordManagementFlow(t *testing.T) {
    tests := []struct {
        name       string
        method     string
        path       string
        body       interface{}
        wantStatus int
        wantBody   string
    }{
        {
            name:       "list_empty_words",
            method:     http.MethodGet,
            path:       "/api/v1/words",
            wantStatus: http.StatusOK,
            wantBody:   "\"data\":[]",
        },
        {
            name:   "create_word",
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