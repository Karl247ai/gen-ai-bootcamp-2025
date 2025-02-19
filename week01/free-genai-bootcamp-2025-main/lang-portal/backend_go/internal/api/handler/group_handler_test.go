package handler

import (
	"net/http"
	"testing"

	"github.com/your-org/lang-portal/internal/api/response"
	"github.com/your-org/lang-portal/internal/models"
	"github.com/your-org/lang-portal/internal/repository/sqlite"
	"github.com/your-org/lang-portal/internal/service"
	"github.com/your-org/lang-portal/internal/testutil"
)

func TestGroupHandler_Create(t *testing.T) {
	// Setup
	db := testutil.NewTestDB(t)
	testutil.ExecuteSQL(t, db, "../../../migrations/0001_init.sql")

	groupRepo := sqlite.NewGroupRepository(db)
	groupService := service.NewGroupService(groupRepo)
	groupHandler := NewGroupHandler(groupService)

	server := testutil.NewAPITestServer()
	server.Engine.POST("/api/v1/groups", groupHandler.Create)

	// Test case
	group := &models.Group{
		Name: "JLPT N5 Vocabulary",
	}

	// Send request
	w := server.SendRequest(t, http.MethodPost, "/api/v1/groups", group)

	// Verify response
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	// Decode response
	var resp response.Response
	testutil.DecodeResponse(t, w, &resp)

	// Verify response data
	createdGroup, ok := resp.Data.(*models.Group)
	if !ok {
		t.Fatal("Expected response data to be a Group")
	}

	if createdGroup.ID == 0 {
		t.Error("Expected created group to have an ID")
	}

	if createdGroup.Name != group.Name {
		t.Errorf("Expected Name %q, got %q", group.Name, createdGroup.Name)
	}
}

func TestGroupHandler_DuplicateName(t *testing.T) {
	// Setup
	db := testutil.NewTestDB(t)
	testutil.ExecuteSQL(t, db, "../../../migrations/0001_init.sql")

	groupRepo := sqlite.NewGroupRepository(db)
	groupService := service.NewGroupService(groupRepo)
	groupHandler := NewGroupHandler(groupService)

	server := testutil.NewAPITestServer()
	server.Engine.POST("/api/v1/groups", groupHandler.Create)

	// Create first group
	group := &models.Group{
		Name: "JLPT N5",
	}

	// Send first request
	w := server.SendRequest(t, http.MethodPost, "/api/v1/groups", group)
	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create first group: %d", w.Code)
	}

	// Try to create duplicate
	w = server.SendRequest(t, http.MethodPost, "/api/v1/groups", group)

	// Verify error response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d for duplicate name, got %d", http.StatusBadRequest, w.Code)
	}
} 