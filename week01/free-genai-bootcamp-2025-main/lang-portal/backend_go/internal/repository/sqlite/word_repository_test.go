package sqlite

import (
	"testing"

	"github.com/your-org/lang-portal/internal/models"
	"github.com/your-org/lang-portal/internal/testutil"
)

func TestWordRepository_Create(t *testing.T) {
	// Setup test database
	db := testutil.NewTestDB(t)
	testutil.ExecuteSQL(t, db, "../../../migrations/0001_init.sql")

	repo := NewWordRepository(db)

	// Test case
	word := &models.Word{
		Japanese: "こんにちは",
		Romaji:   "konnichiwa",
		English:  "hello",
		Parts:    "{}",
	}

	// Execute test
	err := repo.Create(word)
	if err != nil {
		t.Fatalf("Failed to create word: %v", err)
	}

	// Verify results
	if word.ID == 0 {
		t.Error("Expected word ID to be set after creation")
	}

	// Verify we can retrieve the word
	retrieved, err := repo.GetByID(word.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve word: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected to find word but got nil")
	}

	if retrieved.Japanese != word.Japanese {
		t.Errorf("Expected Japanese %q, got %q", word.Japanese, retrieved.Japanese)
	}
}

func TestWordRepository_List(t *testing.T) {
	// Setup test database
	db := testutil.NewTestDB(t)
	testutil.ExecuteSQL(t, db, "../../../migrations/0001_init.sql")

	repo := NewWordRepository(db)

	// Create test data
	words := []*models.Word{
		{Japanese: "一", Romaji: "ichi", English: "one"},
		{Japanese: "二", Romaji: "ni", English: "two"},
		{Japanese: "三", Romaji: "san", English: "three"},
	}

	for _, w := range words {
		if err := repo.Create(w); err != nil {
			t.Fatalf("Failed to create test word: %v", err)
		}
	}

	// Test listing
	retrieved, err := repo.List(0, 10)
	if err != nil {
		t.Fatalf("Failed to list words: %v", err)
	}

	if len(retrieved) != len(words) {
		t.Errorf("Expected %d words, got %d", len(words), len(retrieved))
	}
} 