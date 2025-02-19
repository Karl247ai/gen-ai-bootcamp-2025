package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/your-org/lang-portal/internal/api/response"
	"github.com/your-org/lang-portal/internal/models"
)

func TestWordGroupRelationships(t *testing.T) {
	server := setupTestAPI(t)

	// 1. Create a study group
	group := &models.Group{
		Name: "Basic Verbs",
	}
	w := server.SendRequest(t, http.MethodPost, "/api/v1/groups", group)
	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create group: %d", w.Code)
	}

	var groupResp response.Response
	testutil.DecodeResponse(t, w, &groupResp)
	createdGroup, ok := groupResp.Data.(*models.Group)
	if !ok {
		t.Fatal("Expected response data to be a Group")
	}

	// 2. Create words
	words := []*models.Word{
		{Japanese: "食べる", Romaji: "taberu", English: "to eat"},
		{Japanese: "飲む", Romaji: "nomu", English: "to drink"},
		{Japanese: "見る", Romaji: "miru", English: "to see"},
	}

	createdWords := make([]*models.Word, 0, len(words))
	for _, word := range words {
		w := server.SendRequest(t, http.MethodPost, "/api/v1/words", word)
		if w.Code != http.StatusCreated {
			t.Fatalf("Failed to create word %s: %d", word.Japanese, w.Code)
		}

		var wordResp response.Response
		testutil.DecodeResponse(t, w, &wordResp)
		createdWord, ok := wordResp.Data.(*models.Word)
		if !ok {
			t.Fatal("Expected response data to be a Word")
		}
		createdWords = append(createdWords, createdWord)
	}

	// 3. Add words to group
	for _, word := range createdWords {
		w := server.SendRequest(t, http.MethodPost, 
			fmt.Sprintf("/api/v1/groups/%d/words/%d", createdGroup.ID, word.ID), 
			nil)
		if w.Code != http.StatusOK {
			t.Errorf("Failed to add word %d to group %d: %d", 
				word.ID, createdGroup.ID, w.Code)
		}
	}

	// 4. Verify group words
	w = server.SendRequest(t, http.MethodGet, 
		fmt.Sprintf("/api/v1/groups/%d/words", createdGroup.ID), 
		nil)
	if w.Code != http.StatusOK {
		t.Fatalf("Failed to get group words: %d", w.Code)
	}

	var listResp response.PaginatedResponse
	testutil.DecodeResponse(t, w, &listResp)
	groupWords := listResp.Data.([]*models.Word)
	if len(groupWords) != len(words) {
		t.Errorf("Expected %d words in group, got %d", 
			len(words), len(groupWords))
	}

	// 5. Remove a word from group
	w = server.SendRequest(t, http.MethodDelete, 
		fmt.Sprintf("/api/v1/groups/%d/words/%d", 
			createdGroup.ID, createdWords[0].ID), 
		nil)
	if w.Code != http.StatusOK {
		t.Errorf("Failed to remove word from group: %d", w.Code)
	}

	// 6. Verify word removal
	w = server.SendRequest(t, http.MethodGet, 
		fmt.Sprintf("/api/v1/groups/%d/words", createdGroup.ID), 
		nil)
	if w.Code != http.StatusOK {
		t.Fatalf("Failed to get group words after removal: %d", w.Code)
	}

	testutil.DecodeResponse(t, w, &listResp)
	updatedGroupWords := listResp.Data.([]*models.Word)
	if len(updatedGroupWords) != len(words)-1 {
		t.Errorf("Expected %d words in group after removal, got %d", 
			len(words)-1, len(updatedGroupWords))
	}
} 