package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/your-org/lang-portal/internal/api/handler"
	"github.com/your-org/lang-portal/internal/metrics"
	"github.com/your-org/lang-portal/internal/models"
	"github.com/your-org/lang-portal/internal/testutil"
)

func TestGroupIntegration(t *testing.T) {
	// Setup
	db := testutil.NewTestDB(t)
	testutil.ExecuteSQL(t, db, "../../migrations/0001_init.sql")

	// Initialize metrics
	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)

	// Initialize repositories and services
	groupRepo := sqlite.NewGroupRepository(db)
	wordRepo := sqlite.NewWordRepository(db)
	wordGroupRepo := sqlite.NewWordGroupRepository(db)
	groupService := service.NewGroupService(groupRepo, wordGroupRepo, db)
	wordService := service.NewWordService(wordRepo)

	// Initialize handler
	groupHandler := handler.NewGroupHandler(groupService, m)

	// Setup test server
	router := gin.New()
	router.Use(middleware.Recovery())
	router.POST("/groups", groupHandler.CreateGroup)
	router.POST("/groups/:id/words", groupHandler.AddWordsToGroup)

	t.Run("create and add words to group", func(t *testing.T) {
		// Create a group
		group := models.Group{Name: "Test Group"}
		body, _ := json.Marshal(group)
		req := httptest.NewRequest("POST", "/groups", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var response struct {
			Data models.Group `json:"data"`
		}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		groupID := response.Data.ID

		// Create some words
		words := []*models.Word{
			{Japanese: "一", Romaji: "ichi", English: "one"},
			{Japanese: "二", Romaji: "ni", English: "two"},
		}

		wordIDs := make([]int, len(words))
		for i, word := range words {
			if err := wordService.CreateWord(context.Background(), word); err != nil {
				t.Fatalf("Failed to create word: %v", err)
			}
			wordIDs[i] = word.ID
		}

		// Add words to group
		addRequest := struct {
			WordIDs []int `json:"word_ids"`
		}{
			WordIDs: wordIDs,
		}
		body, _ = json.Marshal(addRequest)
		req = httptest.NewRequest("POST", "/groups/"+strconv.Itoa(groupID)+"/words", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Verify metrics
		metrics, err := reg.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		// Check for success counters
		found := false
		for _, m := range metrics {
			if m.GetName() == "handler_group_create_success" {
				found = true
				if m.GetMetric()[0].GetCounter().GetValue() != 1 {
					t.Error("Expected create success counter to be 1")
				}
			}
		}
		if !found {
			t.Error("Create success metric not found")
		}
	})

	t.Run("error handling", func(t *testing.T) {
		tests := []struct {
			name           string
			request        interface{}
			path          string
			expectedCode   int
			expectedError  string
		}{
			{
				name: "invalid group name",
				request: models.Group{
					Name: "",
				},
				path:          "/groups",
				expectedCode:   http.StatusBadRequest,
				expectedError: "INVALID_INPUT",
			},
			{
				name: "invalid word IDs",
				request: struct {
					WordIDs []int `json:"word_ids"`
				}{
					WordIDs: []int{},
				},
				path:          "/groups/1/words",
				expectedCode:   http.StatusBadRequest,
				expectedError: "INVALID_INPUT",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				body, _ := json.Marshal(tt.request)
				req := httptest.NewRequest("POST", tt.path, bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if w.Code != tt.expectedCode {
					t.Errorf("Expected status %d, got %d", tt.expectedCode, w.Code)
				}

				var response struct {
					Error struct {
						Code string `json:"code"`
					} `json:"error"`
				}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if response.Error.Code != tt.expectedError {
					t.Errorf("Expected error code %s, got %s", tt.expectedError, response.Error.Code)
				}
			})
		}
	})
} 