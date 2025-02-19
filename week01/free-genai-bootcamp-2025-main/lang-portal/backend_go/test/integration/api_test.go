package integration

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/your-org/lang-portal/internal/api/handler"
	"github.com/your-org/lang-portal/internal/api/response"
	"github.com/your-org/lang-portal/internal/api/router"
	"github.com/your-org/lang-portal/internal/models"
	"github.com/your-org/lang-portal/internal/repository/sqlite"
	"github.com/your-org/lang-portal/internal/service"
	"github.com/your-org/lang-portal/internal/testutil"
)

func setupTestAPI(t *testing.T) *testutil.APITestServer {
	// Setup database
	db := testutil.NewTestDB(t)
	testutil.ExecuteSQL(t, db, "../../migrations/0001_init.sql")

	// Setup repositories
	wordRepo := sqlite.NewWordRepository(db)
	groupRepo := sqlite.NewGroupRepository(db)

	// Setup services
	wordService := service.NewWordService(wordRepo)
	groupService := service.NewGroupService(groupRepo)

	// Setup handlers
	wordHandler := handler.NewWordHandler(wordService)
	groupHandler := handler.NewGroupHandler(groupService)

	// Setup router
	r := router.NewRouter(wordHandler, groupHandler)
	server := testutil.NewAPITestServer()
	r.Setup(server.Engine)

	return server
}

func TestAPIFlow(t *testing.T) {
	// Setup test environment
	db := testutil.NewTestDB(t)
	testutil.ExecuteSQL(t, db, "../../migrations/0001_init.sql")

	metrics := testutil.NewTestMetrics(t)
	router := setupTestRouter(t, db, metrics)
	server := testutil.NewAPITestServer(t)
	server.Engine = router

	// Create metrics assertion helper
	ma := testutil.NewMetricsAssertion(t, metrics, metrics.Registry())

	t.Run("complete workflow", func(t *testing.T) {
		// 1. Create a group
		group := models.Group{
			Name: "JLPT N5 Words",
		}
		w := server.SendRequest("POST", "/api/v1/groups", group)
		server.AssertResponse(w, http.StatusCreated)
		var groupResp struct {
			Data models.Group `json:"data"`
		}
		server.DecodeResponse(w, &groupResp)
		groupID := groupResp.Data.ID

		// Verify metrics
		ma.AssertCounterValue("handler_group_create_success", 1)
		ma.AssertHistogramCount("handler_group_create", 1)

		// 2. Create multiple words
		words := []models.Word{
			{Japanese: "猫", Romaji: "neko", English: "cat"},
			{Japanese: "犬", Romaji: "inu", English: "dog"},
			{Japanese: "鳥", Romaji: "tori", English: "bird"},
		}

		var wordIDs []int
		for _, word := range words {
			w := server.SendRequest("POST", "/api/v1/words", word)
			server.AssertResponse(w, http.StatusCreated)
			var wordResp struct {
				Data models.Word `json:"data"`
			}
			server.DecodeResponse(w, &wordResp)
			wordIDs = append(wordIDs, wordResp.Data.ID)
		}

		// Verify word creation metrics
		ma.AssertCounterValue("handler_word_create_success", 3)
		ma.AssertHistogramCount("handler_word_create", 3)

		// 3. Add words to group
		addRequest := struct {
			WordIDs []int `json:"word_ids"`
		}{
			WordIDs: wordIDs,
		}
		w = server.SendRequest("POST", "/api/v1/groups/"+strconv.Itoa(groupID)+"/words", addRequest)
		server.AssertResponse(w, http.StatusOK)

		// Verify group update metrics
		ma.AssertCounterValue("handler_group_add_words_success", 1)

		// 4. List group words
		w = server.SendRequest("GET", "/api/v1/groups/"+strconv.Itoa(groupID)+"/words", nil)
		server.AssertResponse(w, http.StatusOK)
		var listResp struct {
			Data []models.Word `json:"data"`
		}
		server.DecodeResponse(w, &listResp)

		if len(listResp.Data) != len(words) {
			t.Errorf("Expected %d words in group, got %d", len(words), len(listResp.Data))
		}

		// 5. Verify performance metrics
		ma.AssertHistogramBounds("handler_group_create", 1.0) // 1 second max
		ma.AssertHistogramBounds("handler_word_create", 1.0)
		ma.AssertHistogramBounds("handler_group_add_words", 1.0)
	})
}

func TestAPIValidation(t *testing.T) {
	// Setup
	server := setupTestAPI(t)
	ma := testutil.NewMetricsAssertion(t, server.Metrics, server.Metrics.Registry())

	tests := []struct {
		name         string
		method       string
		path         string
		body         interface{}
		expectStatus int
	}{
		{
			name:   "empty word",
			method: "POST",
			path:   "/api/v1/words",
			body: models.Word{
				Japanese: "",
				English:  "",
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name:   "empty group name",
			method: "POST",
			path:   "/api/v1/groups",
			body: models.Group{
				Name: "",
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "invalid word ID",
			method:      "GET",
			path:        "/api/v1/words/invalid",
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "invalid pagination",
			method:      "GET",
			path:        "/api/v1/words?page=0&limit=0",
			expectStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := server.SendRequest(tt.method, tt.path, tt.body)
			server.AssertResponse(w, tt.expectStatus)

			// Verify error metrics
			if tt.expectStatus != http.StatusOK {
				ma.AssertCounterValue("handler_error_count", 1)
			}
		})
	}
} 