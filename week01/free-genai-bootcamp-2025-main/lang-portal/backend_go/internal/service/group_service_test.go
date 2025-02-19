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

func TestGroupService_CreateGroup(t *testing.T) {
	// Setup
	db := testutil.NewTestDB(t)
	testutil.ExecuteSQL(t, db, "../../migrations/0001_init.sql")

	groupRepo := sqlite.NewGroupRepository(db)
	groupService := NewGroupService(groupRepo)

	// Test case
	group := &models.Group{
		Name: "JLPT N5 Vocabulary",
	}

	// Execute
	err := groupService.CreateGroup(group)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	// Verify
	if group.ID == 0 {
		t.Error("Expected group ID to be set after creation")
	}

	// Verify through service
	retrieved, err := groupService.GetGroup(group.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve group: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected to find group but got nil")
	}

	if retrieved.Name != group.Name {
		t.Errorf("Expected Name %q, got %q", group.Name, retrieved.Name)
	}
}

func TestGroupService_ListGroups(t *testing.T) {
	// Setup
	db := testutil.NewTestDB(t)
	testutil.ExecuteSQL(t, db, "../../migrations/0001_init.sql")

	groupRepo := sqlite.NewGroupRepository(db)
	groupService := NewGroupService(groupRepo)

	// Create test data
	groups := []*models.Group{
		{Name: "JLPT N5"},
		{Name: "JLPT N4"},
		{Name: "Common Phrases"},
	}

	for _, g := range groups {
		if err := groupService.CreateGroup(g); err != nil {
			t.Fatalf("Failed to create test group: %v", err)
		}
	}

	// Test listing with pagination
	page := 1
	pageSize := 2

	retrieved, err := groupService.ListGroups(page, pageSize)
	if err != nil {
		t.Fatalf("Failed to list groups: %v", err)
	}

	if len(retrieved) != pageSize {
		t.Errorf("Expected %d groups, got %d", pageSize, len(retrieved))
	}
}

func TestGroupService_DuplicateGroupName(t *testing.T) {
	// Setup
	db := testutil.NewTestDB(t)
	testutil.ExecuteSQL(t, db, "../../migrations/0001_init.sql")

	groupRepo := sqlite.NewGroupRepository(db)
	groupService := NewGroupService(groupRepo)

	// Create first group
	group1 := &models.Group{
		Name: "JLPT N5",
	}
	if err := groupService.CreateGroup(group1); err != nil {
		t.Fatalf("Failed to create first group: %v", err)
	}

	// Try to create group with same name
	group2 := &models.Group{
		Name: "JLPT N5",
	}
	err := groupService.CreateGroup(group2)
	if err == nil {
		t.Error("Expected error when creating group with duplicate name, got nil")
	}
}

func TestGroupService_AddWordsToGroup(t *testing.T) {
	// Setup
	db := testutil.NewTestDB(t)
	testutil.ExecuteSQL(t, db, "../../migrations/0001_init.sql")

	groupRepo := sqlite.NewGroupRepository(db)
	wordRepo := sqlite.NewWordRepository(db)
	wordGroupRepo := sqlite.NewWordGroupRepository(db)
	service := NewGroupService(groupRepo, wordGroupRepo, db)
	wordService := NewWordService(wordRepo)

	// Create test group
	group := &models.Group{Name: "Test Group"}
	ctx := context.Background()
	if err := service.CreateGroup(ctx, group); err != nil {
		t.Fatalf("Failed to create test group: %v", err)
	}

	// Create test words
	words := []*models.Word{
		{Japanese: "一", Romaji: "ichi", English: "one"},
		{Japanese: "二", Romaji: "ni", English: "two"},
		{Japanese: "三", Romaji: "san", English: "three"},
	}

	wordIDs := make([]int, len(words))
	for i, w := range words {
		if err := wordService.CreateWord(ctx, w); err != nil {
			t.Fatalf("Failed to create test word: %v", err)
		}
		wordIDs[i] = w.ID
	}

	tests := []struct {
		name        string
		groupID     int
		wordIDs     []int
		ctxTimeout  time.Duration
		wantErr     bool
		wantErrCode errors.ErrorCode
	}{
		{
			name:       "successful addition",
			groupID:    group.ID,
			wordIDs:    wordIDs,
			ctxTimeout: time.Second,
			wantErr:    false,
		},
		{
			name:        "context timeout",
			groupID:     group.ID,
			wordIDs:     wordIDs,
			ctxTimeout:  time.Nanosecond,
			wantErr:     true,
			wantErrCode: errors.ErrTimeout,
		},
		{
			name:        "invalid group ID",
			groupID:     0,
			wordIDs:     wordIDs,
			ctxTimeout:  time.Second,
			wantErr:     true,
			wantErrCode: errors.ErrInvalidInput,
		},
		{
			name:        "empty word IDs",
			groupID:     group.ID,
			wordIDs:     []int{},
			ctxTimeout:  time.Second,
			wantErr:     true,
			wantErrCode: errors.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.ctxTimeout)
			defer cancel()

			err := service.AddWordsToGroup(ctx, tt.groupID, tt.wordIDs)

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

			// Verify words were added if no error was expected
			if !tt.wantErr {
				words, err := service.GetGroupWords(ctx, tt.groupID, 1, len(tt.wordIDs))
				if err != nil {
					t.Errorf("Failed to get group words: %v", err)
				}
				if len(words) != len(tt.wordIDs) {
					t.Errorf("Expected %d words in group, got %d", len(tt.wordIDs), len(words))
				}
			}
		})
	}
} 