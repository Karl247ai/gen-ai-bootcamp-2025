package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-org/lang-portal/internal/api/handler"
	"github.com/your-org/lang-portal/internal/errors"
	"github.com/your-org/lang-portal/internal/models"
	"github.com/your-org/lang-portal/internal/testutil"
)

func TestErrorHandling(t *testing.T) {
	// Setup test environment
	db := testutil.NewTestDB(t)
	testutil.ExecuteSQL(t, db, "../../migrations/0001_init.sql")

	// Initialize metrics and handlers
	metrics := testutil.NewTestMetrics(t)
	router := setupTestRouter(t, db, metrics)

	tests := []struct {
		name         string
		method       string
		path         string
		body         interface{}
		expectCode   int
		expectError  errors.ErrorCode
		setupContext func(context.Context) context.Context
	}{
		{
			name:   "invalid word input",
			method: "POST",
			path:   "/api/v1/words",
			body: models.Word{
				Japanese: "", // Empty required field
				English:  "",
			},
			expectCode:  http.StatusBadRequest,
			expectError: errors.ErrInvalidInput,
		},
		{
			name:   "duplicate group name",
			method: "POST",
			path:   "/api/v1/groups",
			body: models.Group{
				Name: "Test Group",
			},
			expectCode:  http.StatusBadRequest,
			expectError: errors.ErrDBDuplicate,
			setupContext: func(ctx context.Context) context.Context {
				// Create a group first to trigger duplicate error
				group := models.Group{Name: "Test Group"}
				body, _ := json.Marshal(group)
				req := httptest.NewRequest("POST", "/api/v1/groups", bytes.NewBuffer(body))
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				return ctx
			},
		},
		{
			name:        "invalid group ID",
			method:      "GET",
			path:        "/api/v1/groups/invalid",
			expectCode:  http.StatusBadRequest,
			expectError: errors.ErrInvalidInput,
		},
		{
			name:        "request timeout",
			method:      "GET",
			path:        "/api/v1/words",
			expectCode:  http.StatusGatewayTimeout,
			expectError: errors.ErrTimeout,
			setupContext: func(ctx context.Context) context.Context {
				ctx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
				time.Sleep(1 * time.Millisecond) // Ensure timeout
				defer cancel()
				return ctx
			},
		},
		{
			name:   "malformed JSON",
			method: "POST",
			path:   "/api/v1/words",
			body:   `{"invalid": json`,
			expectCode:  http.StatusBadRequest,
			expectError: errors.ErrInvalidInput,
		},
		{
			name:   "transaction rollback",
			method: "POST",
			path:   "/api/v1/groups/1/words",
			body: struct {
				WordIDs []int `json:"word_ids"`
			}{
				WordIDs: []int{999999}, // Non-existent word ID
			},
			expectCode:  http.StatusNotFound,
			expectError: errors.ErrDBNotFound,
		},
		{
			name:   "concurrent modification",
			method: "POST",
			path:   "/api/v1/groups",
			body: models.Group{
				Name: "Concurrent Test",
			},
			setupContext: func(ctx context.Context) context.Context {
				// Simulate concurrent modification
				go func() {
					group := models.Group{Name: "Concurrent Test"}
					body, _ := json.Marshal(group)
					req := httptest.NewRequest("POST", "/api/v1/groups", bytes.NewBuffer(body))
					router.ServeHTTP(httptest.NewRecorder(), req)
				}()
				time.Sleep(10 * time.Millisecond)
				return ctx
			},
			expectCode:  http.StatusConflict,
			expectError: errors.ErrDBDuplicate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.setupContext != nil {
				ctx = tt.setupContext(ctx)
			}

			var req *http.Request
			if tt.body != nil {
				body, _ := json.Marshal(tt.body)
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("Expected status code %d, got %d", tt.expectCode, w.Code)
			}

			var resp struct {
				Error struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}

			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if string(tt.expectError) != resp.Error.Code {
				t.Errorf("Expected error code %s, got %s", tt.expectError, resp.Error.Code)
			}

			if resp.Error.Message == "" {
				t.Error("Expected error message to be non-empty")
			}
		})
	}
}

func setupTestRouter(t *testing.T, db *sql.DB, metrics *metrics.Metrics) *gin.Engine {
	router := gin.New()
	router.Use(middleware.RequestContext())
	router.Use(middleware.Recovery())
	setupHandlers(router, db, metrics)
	return router
} 