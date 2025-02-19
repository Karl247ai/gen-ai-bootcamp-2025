package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/your-org/lang-portal/internal/api/response"
	"github.com/your-org/lang-portal/internal/models"
)

func TestStudyWorkflow(t *testing.T) {
	server := setupTestAPI(t)

	// 1. Create multiple groups
	groups := []string{"JLPT N5", "Common Phrases", "Numbers"}
	createdGroups := make([]*models.Group, 0, len(groups))

	for _, name := range groups {
		group := &models.Group{Name: name}
		w := server.SendRequest(t, http.MethodPost, "/api/v1/groups", group)
		if w.Code != http.StatusCreated {
			t.Fatalf("Failed to create group %s: %d", name, w.Code)
		}

		var resp response.Response
		testutil.DecodeResponse(t, w, &resp)
		createdGroup, ok := resp.Data.(*models.Group)
		if !ok {
			t.Fatal("Expected response data to be a Group")
		}
		createdGroups = append(createdGroups, createdGroup)
	}

	// 2. Add words to each group
	numberWords := []*models.Word{
		{Japanese: "一", Romaji: "ichi", English: "one"},
		{Japanese: "二", Romaji: "ni", English: "two"},
		{Japanese: "三", Romaji: "san", English: "three"},
	}

	phraseWords := []*models.Word{
		{Japanese: "おはよう", Romaji: "ohayou", English: "good morning"},
		{Japanese: "こんにちは", Romaji: "konnichiwa", English: "hello"},
	}

	// Add words to respective groups
	for _, word := range numberWords {
		w := server.SendRequest(t, http.MethodPost, "/api/v1/words", word)
		if w.Code != http.StatusCreated {
			t.Fatalf("Failed to create word %s: %d", word.Japanese, w.Code)
		}
	}

	for _, word := range phraseWords {
		w := server.SendRequest(t, http.MethodPost, "/api/v1/words", word)
		if w.Code != http.StatusCreated {
			t.Fatalf("Failed to create word %s: %d", word.Japanese, w.Code)
		}
	}

	// 3. Test group listing and pagination
	w := server.SendRequest(t, http.MethodGet, "/api/v1/groups?page=1&limit=2", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("Failed to list groups: %d", w.Code)
	}

	var listResp response.PaginatedResponse
	testutil.DecodeResponse(t, w, &listResp)
	if len(listResp.Data.([]*models.Group)) != 2 {
		t.Errorf("Expected 2 groups per page, got %d", len(listResp.Data.([]*models.Group)))
	}

	// 4. Test word search and filtering
	searchTests := []struct {
		query    string
		expected int
	}{
		{"?japanese=一", 1},
		{"?english=hello", 1},
		{"?romaji=ichi", 1},
	}

	for _, tt := range searchTests {
		w := server.SendRequest(t, http.MethodGet, "/api/v1/words"+tt.query, nil)
		if w.Code != http.StatusOK {
			t.Errorf("Search failed for query %s: %d", tt.query, w.Code)
			continue
		}

		var resp response.PaginatedResponse
		testutil.DecodeResponse(t, w, &resp)
		words := resp.Data.([]*models.Word)
		if len(words) != tt.expected {
			t.Errorf("Search %s: expected %d results, got %d", tt.query, tt.expected, len(words))
		}
	}
} 