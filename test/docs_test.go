package test

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "github.com/stretchr/testify/assert"
)

func TestAPIDocumentation(t *testing.T) {
    tests := []struct {
        name       string
        path       string
        method     string
        wantStatus int
        wantBody   string
    }{
        {
            name:       "swagger_ui",
            path:       "/swagger/index.html",
            method:     http.MethodGet,
            wantStatus: http.StatusOK,
            wantBody:   "Swagger UI",
        },
        {
            name:       "api_docs_json",
            path:       "/swagger/doc.json",
            method:     http.MethodGet,
            wantStatus: http.StatusOK,
            wantBody:   "\"swagger\":\"2.0\"",
        },
        {
            name:       "health_check",
            path:       "/health",
            method:     http.MethodGet,
            wantStatus: http.StatusOK,
            wantBody:   "\"status\":\"ok\"",
        },
    }

    router := setupTestRouter()

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            w := httptest.NewRecorder()
            req := httptest.NewRequest(tt.method, tt.path, nil)
            router.ServeHTTP(w, req)

            assert.Equal(t, tt.wantStatus, w.Code)
            assert.Contains(t, w.Body.String(), tt.wantBody)
        })
    }
}