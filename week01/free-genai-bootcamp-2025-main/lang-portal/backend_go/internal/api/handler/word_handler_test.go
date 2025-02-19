package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/your-org/lang-portal/internal/api/response"
	"github.com/your-org/lang-portal/internal/metrics"
	"github.com/your-org/lang-portal/internal/models"
	"github.com/your-org/lang-portal/internal/repository/sqlite"
	"github.com/your-org/lang-portal/internal/service"
	"github.com/your-org/lang-portal/internal/testutil"
)

func TestWordHandler(t *testing.T) {
	// Setup
	db := testutil.NewTestDB(t)
	testutil.ExecuteSQL(t, db, "../../../migrations/0001_init.sql")

	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)

	wordRepo := sqlite.NewWordRepository(db)
	wordService := service.NewWordService(wordRepo)
	handler := NewWordHandler(wordService, m)

	router := gin.New()
	router.POST("/words", handler.CreateWord)
	router.GET("/words/:id", handler.GetWord)
	router.GET("/words", handler.ListWords)

	// Create test data
	testWords := []*models.Word{
		{Japanese: "猫", Romaji: "neko", English: "cat"},
		{Japanese: "犬", Romaji: "inu", English: "dog"},
		{Japanese: "鳥", Romaji: "tori", English: "bird"},
	}

	var createdWords []*models.Word
	for _, w := range testWords {
		body, _ := json.Marshal(w)
		req := httptest.NewRequest("POST", "/words", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp struct {
			Data models.Word `json:"data"`
		}
		json.NewDecoder(w.Body).Decode(&resp)
		createdWords = append(createdWords, &resp.Data)
	}

	t.Run("get word", func(t *testing.T) {
		tests := []struct {
			name         string
			wordID      string
			expectCode  int
			expectError bool
		}{
			{
				name:         "valid word",
				wordID:      strconv.Itoa(createdWords[0].ID),
				expectCode:  http.StatusOK,
				expectError: false,
			},
			{
				name:         "invalid word ID",
				wordID:      "invalid",
				expectCode:  http.StatusBadRequest,
				expectError: true,
			},
			{
				name:         "non-existent word",
				wordID:      "999999",
				expectCode:  http.StatusNotFound,
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest("GET", "/words/"+tt.wordID, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if w.Code != tt.expectCode {
					t.Errorf("Expected status %d, got %d", tt.expectCode, w.Code)
				}

				if tt.expectError {
					var resp struct {
						Error struct {
							Code string `json:"code"`
						} `json:"error"`
					}
					if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
						t.Fatalf("Failed to decode error response: %v", err)
					}
					if resp.Error.Code == "" {
						t.Error("Expected error code in response")
					}
				} else {
					var resp struct {
						Data models.Word `json:"data"`
					}
					if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
						t.Fatalf("Failed to decode success response: %v", err)
					}
					if resp.Data.ID == 0 {
						t.Error("Expected word data in response")
					}
				}
			})
		}
	})

	t.Run("list words", func(t *testing.T) {
		tests := []struct {
			name        string
			query       string
			expectCount int
		}{
			{
				name:        "default pagination",
				query:       "",
				expectCount: 3,
			},
			{
				name:        "custom page size",
				query:       "?limit=2",
				expectCount: 2,
			},
			{
				name:        "second page",
				query:       "?page=2&limit=2",
				expectCount: 1,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest("GET", "/words"+tt.query, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
				}

				var resp struct {
					Data []models.Word `json:"data"`
				}
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if len(resp.Data) != tt.expectCount {
					t.Errorf("Expected %d words, got %d", tt.expectCount, len(resp.Data))
				}
			})
		}

		// Verify metrics for list operation
		metrics, err := reg.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		found := false
		for _, m := range metrics {
			if m.GetName() == "handler_word_list_success" {
				found = true
				if m.GetMetric()[0].GetCounter().GetValue() < 3 {
					t.Error("Expected list success counter to be at least 3")
				}
			}
		}
		if !found {
			t.Error("List success metric not found")
		}
	})
}

func TestWordHandler_Create(t *testing.T) {
	// Setup
	db := testutil.NewTestDB(t)
	testutil.ExecuteSQL(t, db, "../../../migrations/0001_init.sql")

	wordRepo := sqlite.NewWordRepository(db)
	wordService := service.NewWordService(wordRepo)
	wordHandler := NewWordHandler(wordService)

	server := testutil.NewAPITestServer()
	server.Engine.POST("/api/v1/words", wordHandler.Create)

	// Test case
	word := &models.Word{
		Japanese: "こんにちは",
		Romaji:   "konnichiwa",
		English:  "hello",
		Parts:    "{}",
	}

	// Send request
	w := server.SendRequest(t, http.MethodPost, "/api/v1/words", word)

	// Verify response
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	// Decode response
	var resp response.Response
	testutil.DecodeResponse(t, w, &resp)

	// Verify response data
	createdWord, ok := resp.Data.(*models.Word)
	if !ok {
		t.Fatal("Expected response data to be a Word")
	}

	if createdWord.ID == 0 {
		t.Error("Expected created word to have an ID")
	}

	if createdWord.Japanese != word.Japanese {
		t.Errorf("Expected Japanese %q, got %q", word.Japanese, createdWord.Japanese)
	}
}

func TestWordHandler_List(t *testing.T) {
	// Setup
	db := testutil.NewTestDB(t)
	testutil.ExecuteSQL(t, db, "../../../migrations/0001_init.sql")

	wordRepo := sqlite.NewWordRepository(db)
	wordService := service.NewWordService(wordRepo)
	wordHandler := NewWordHandler(wordService)

	server := testutil.NewAPITestServer()
	server.Engine.GET("/api/v1/words", wordHandler.List)

	// Create test data
	words := []*models.Word{
		{Japanese: "一", Romaji: "ichi", English: "one"},
		{Japanese: "二", Romaji: "ni", English: "two"},
		{Japanese: "三", Romaji: "san", English: "three"},
	}

	for _, w := range words {
		if err := wordService.CreateWord(w); err != nil {
			t.Fatalf("Failed to create test word: %v", err)
		}
	}

	// Send request
	w := server.SendRequest(t, http.MethodGet, "/api/v1/words?page=1&limit=2", nil)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Decode response
	var resp response.PaginatedResponse
	testutil.DecodeResponse(t, w, &resp)

	// Verify pagination
	if resp.Pagination.CurrentPage != 1 {
		t.Errorf("Expected current page 1, got %d", resp.Pagination.CurrentPage)
	}

	if resp.Pagination.ItemsPerPage != 2 {
		t.Errorf("Expected items per page 2, got %d", resp.Pagination.ItemsPerPage)
	}

	// Verify response data
	wordList, ok := resp.Data.([]*models.Word)
	if !ok {
		t.Fatal("Expected response data to be a slice of Words")
	}

	if len(wordList) != 2 {
		t.Errorf("Expected 2 words, got %d", len(wordList))
	}
} 