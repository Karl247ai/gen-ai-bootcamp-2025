package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/your-org/lang-portal/internal/api/response"
)

func TestValidatePagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		wantStatus int
	}{
		{
			name:       "Valid pagination",
			query:      "?page=1&limit=10",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid page",
			query:      "?page=0&limit=10",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid limit",
			query:      "?page=1&limit=0",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Limit too high",
			query:      "?page=1&limit=101",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(ValidatePagination())
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test"+tt.query, nil)
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestValidateID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramName  string
		paramValue string
		wantStatus int
	}{
		{
			name:       "Valid ID",
			paramName:  "id",
			paramValue: "1",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid ID format",
			paramName:  "id",
			paramValue: "abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Negative ID",
			paramName:  "id",
			paramValue: "-1",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(ValidateID(tt.paramName))
			router.GET("/test/:"+tt.paramName, func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test/"+tt.paramValue, nil)
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}

			if w.Code == http.StatusBadRequest {
				var resp response.Response
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if resp.Error == nil {
					t.Error("Expected error response")
				}
			}
		})
	}
} 