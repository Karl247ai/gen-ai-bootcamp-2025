package service

import (
	"context"
	"testing"
	"time"

	"github.com/your-org/lang-portal/internal/errors"
	"github.com/your-org/lang-portal/internal/models"
	"github.com/your-org/lang-portal/internal/repository/sqlite"
	"github.com/your-org/lang-portal/internal/testutil"
)

func TestWordService_CreateWord(t *testing.T) {
	// Setup
	db := testutil.NewTestDB(t)
	testutil.ExecuteSQL(t, db, "../../migrations/0001_init.sql")

	wordRepo := sqlite.NewWordRepository(db)
	wordService := NewWordService(wordRepo)

	// Test case
	word := &models.Word{
		Japanese: "こんにちは",
		Romaji:   "konnichiwa",
		English:  "hello",
		Parts:    "{}",
	}

	// Execute
	err := wordService.CreateWord(word)
	if err != nil {
		t.Fatalf("Failed to create word: %v", err)
	}

	// Verify
	if word.ID == 0 {
		t.Error("Expected word ID to be set after creation")
	}

	// Verify through service
	retrieved, err := wordService.GetWord(word.ID)
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

func TestWordService_ListWords(t *testing.T) {
	// Setup
	db := testutil.NewTestDB(t)
	testutil.ExecuteSQL(t, db, "../../migrations/0001_init.sql")

	wordRepo := sqlite.NewWordRepository(db)
	wordService := NewWordService(wordRepo)

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

	// Test listing with pagination
	page := 1
	pageSize := 2

	retrieved, err := wordService.ListWords(page, pageSize)
	if err != nil {
		t.Fatalf("Failed to list words: %v", err)
	}

	if len(retrieved) != pageSize {
		t.Errorf("Expected %d words, got %d", pageSize, len(retrieved))
	}
}

func TestWordService_CreateWord_Context(t *testing.T) {
	// Setup
	db := testutil.NewTestDB(t)
	wordRepo := sqlite.NewWordRepository(db)
	service := NewWordService(wordRepo)

	tests := []struct {
		name        string
		word        *models.Word
		ctxTimeout  time.Duration
		wantErr     bool
		wantErrCode errors.ErrorCode
	}{
		{
			name: "successful creation",
			word: &models.Word{
				Japanese: "こんにちは",
				Romaji:   "konnichiwa",
				English:  "hello",
			},
			ctxTimeout: time.Second,
			wantErr:    false,
		},
		{
			name: "context timeout",
			word: &models.Word{
				Japanese: "さようなら",
				Romaji:   "sayounara",
				English:  "goodbye",
			},
			ctxTimeout:  time.Nanosecond,
			wantErr:     true,
			wantErrCode: errors.ErrTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.ctxTimeout)
			defer cancel()

			err := service.CreateWord(ctx, tt.word)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.wantErrCode != "" && !errors.IsErrorCode(err, tt.wantErrCode) {
					t.Errorf("Expected error code %v, got %v", tt.wantErrCode, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
} 