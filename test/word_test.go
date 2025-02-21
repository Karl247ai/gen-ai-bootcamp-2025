package test

import (
    "testing"
    "net/http"
    "github.com/stretchr/testify/assert"
)

func TestWordEndpoints(t *testing.T) {
    tests := []testCase{
        {
            name:       "create_word",
            method:     http.MethodPost,
            path:       "/api/v1/words",
            body: map[string]string{
                "japanese": "猫",
                "romaji":   "neko",
                "english":  "cat",
            },
            wantStatus: http.StatusCreated,
        },
        {
            name:       "get_words",
            method:     http.MethodGet,
            path:       "/api/v1/words",
            wantStatus: http.StatusOK,
        },
    }

    router := setupTestRouter()
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            w := performRequest(router, tt)
            assert.Equal(t, tt.wantStatus, w.Code)
        })
    }
}