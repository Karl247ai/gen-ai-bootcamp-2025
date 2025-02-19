package sqlite

import (
	"testing"

	"github.com/your-org/lang-portal/internal/models"
	"github.com/your-org/lang-portal/internal/testutil"
)

func TestGroupRepository_Create(t *testing.T) {
	// Setup test database
	db := testutil.NewTestDB(t)
	testutil.ExecuteSQL(t, db, "../../../migrations/0001_init.sql")

	repo := NewGroupRepository(db)

	// Test case
	group := &models.Group{
		Name: "JLPT N5 Vocabulary",
	}

	// Execute test
	err := repo.Create(group)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	// Verify results
	if group.ID == 0 {
		t.Error("Expected group ID to be set after creation")
	}

	// Verify we can retrieve the group
	retrieved, err := repo.GetByID(group.ID)
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

func TestGroupRepository_List(t *testing.T) {
	// Setup test database
	db := testutil.NewTestDB(t)
	testutil.ExecuteSQL(t, db, "../../../migrations/0001_init.sql")

	repo := NewGroupRepository(db)

	// Create test data
	groups := []*models.Group{
		{Name: "JLPT N5"},
		{Name: "JLPT N4"},
		{Name: "Common Phrases"},
	}

	for _, g := range groups {
		if err := repo.Create(g); err != nil {
			t.Fatalf("Failed to create test group: %v", err)
		}
	}

	// Test listing
	retrieved, err := repo.List(0, 10)
	if err != nil {
		t.Fatalf("Failed to list groups: %v", err)
	}

	if len(retrieved) != len(groups) {
		t.Errorf("Expected %d groups, got %d", len(groups), len(retrieved))
	}
}

func TestGroupRepository_UniqueNameConstraint(t *testing.T) {
	// Setup test database
	db := testutil.NewTestDB(t)
	testutil.ExecuteSQL(t, db, "../../../migrations/0001_init.sql")

	repo := NewGroupRepository(db)

	// Create first group
	group1 := &models.Group{
		Name: "JLPT N5",
	}
	if err := repo.Create(group1); err != nil {
		t.Fatalf("Failed to create first group: %v", err)
	}

	// Try to create group with same name
	group2 := &models.Group{
		Name: "JLPT N5",
	}
	err := repo.Create(group2)
	if err == nil {
		t.Error("Expected error when creating group with duplicate name, got nil")
	}
} 