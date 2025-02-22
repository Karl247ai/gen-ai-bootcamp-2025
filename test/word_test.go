package test

import (
    "testing"
    "net/http"
    "database/sql"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/karl247ai/lang-portal/internal/models"
)

// TestWordSuite runs all word-related tests
func TestWordSuite(t *testing.T) {
    t.Parallel()
    
    // Common setup
    db := setupTestDB()
    defer db.Close()
    router := setupTestRouter()

    // Run test groups in parallel
    t.Run("group=crud", func(t *testing.T) {
        t.Parallel()
        testWordCRUD(t, db, router)
    })

    t.Run("group=validation", func(t *testing.T) {
        t.Parallel()
        testWordValidation(t, db, router)
    })
}

// testWordCRUD tests CRUD operations
func testWordCRUD(t *testing.T, db *sql.DB, router *gin.Engine) {
    tests := []testCase{
        {
            name:       "list_empty",
            method:     http.MethodGet,
            path:       "/api/v1/words",
            wantStatus: http.StatusOK,
            wantBody:   `{"data":[],"pagination":{"current_page":1}}`,
        },
        {
            name:   "create_success",
            method: http.MethodPost,
            path:   "/api/v1/words",
            body: models.Word{
                Japanese: "猫",
                Romaji:   "neko",
                English:  "cat",
            },
            wantStatus: http.StatusCreated,
            wantBody:   `"japanese":"猫"`,
        },
    }

    for _, tt := range tests {
        tt := tt // capture range variable
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            runTestWithTx(t, db, tt, func() {
                w := performRequest(router, tt)
                assertResponse(t, w, tt)
            })
        })
    }
}

// testWordValidation tests input validation
func testWordValidation(t *testing.T, db *sql.DB, router *gin.Engine) {
    tests := []testCase{
        {
            name:   "validate_empty_japanese",
            method: http.MethodPost,
            path:   "/api/v1/words",
            body: models.Word{
                Japanese: "",
                Romaji:   "neko",
                English:  "cat",
            },
            wantStatus: http.StatusBadRequest,
            wantBody:   `"error":"japanese is required"`,
        },
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            runTestWithTx(t, db, tt, func() {
                w := performRequest(router, tt)
                assertResponse(t, w, tt)
            })
        })
    }
}

// TestWordManagement tests the complete word management lifecycle
func TestWordManagement(t *testing.T) {
    // Enable parallel testing
    t.Parallel()

    tests := []testCase{
        {
            name:       "list_empty_words",
            method:     http.MethodGet,
            path:       "/api/v1/words",
            wantStatus: http.StatusOK,
            wantBody:   `{"data":[],"pagination":{"current_page":1}}`,
        },
        {
            name:   "create_valid_word",
            method: http.MethodPost,
            path:   "/api/v1/words",
            body: models.Word{
                Japanese: "猫",
                Romaji:   "neko",
                English:  "cat",
            },
            wantStatus: http.StatusCreated,
            setupFn:    nil,
            cleanupFn:  cleanupTestDB,
        },
        {
            name:   "get_word_by_id",
            method: http.MethodGet,
            path:   "/api/v1/words/1",
            wantStatus: http.StatusOK,
            setupFn: func(tx *sql.Tx) error {
                return LoadTestWords(tx)
            },
        },
    }

    db := setupTestDB()
    defer db.Close()
    
    router := setupTestRouter()

    for _, tt := range tests {
        tt := tt // Capture range variable
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            runTestWithTx(t, db, tt, func() {
                w := performRequest(router, tt)
                assertResponse(t, w, tt)
            })
        })
    }
}