package test

import (
    "bytes"
    "database/sql"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "github.com/gin-gonic/gin"
    "github.com/karl247ai/lang-portal/internal/models"
    "github.com/karl247ai/lang-portal/internal/repository"
    "github.com/karl247ai/lang-portal/internal/api/handlers"
    _ "github.com/mattn/go-sqlite3"
    "testing"
    "time"
)

// TestResponse represents the standard API response structure
type TestResponse struct {
    Status string      `json:"status"`
    Data   interface{} `json:"data"`
    Error  string      `json:"error,omitempty"`
}

func setupTestDB() *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    if (err != nil) {
        panic(err)
    }

    // Create test tables with additional indices
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS words (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            japanese TEXT NOT NULL,
            romaji TEXT NOT NULL,
            english TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
        
        CREATE INDEX IF NOT EXISTS idx_words_romaji ON words(romaji);
        CREATE INDEX IF NOT EXISTS idx_words_english ON words(english);
    `)
    if err != nil {
        panic(err)
    }

    return db
}

func setupTestRouter() *gin.Engine {
    gin.SetMode(gin.TestMode)
    r := gin.Default()
    
    db := setupTestDB()
    wordRepo := repository.NewWordRepository(db)
    wordHandler := handlers.NewWordHandler(wordRepo)
    
    // Setup routes
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    api := r.Group("/api/v1")
    {
        api.GET("/words", wordHandler.GetWords)
        api.POST("/words", wordHandler.CreateWord)
    }
    
    return r
}

// Update the testCase struct to be more comprehensive
type testCase struct {
    name       string
    method     string
    path       string
    body       interface{}
    headers    map[string]string
    wantStatus int
    wantBody   string
    setupFn    func(*sql.Tx) error
    cleanupFn  func(*sql.Tx) error
    skipCleanup bool
}

// Update performRequest to handle the enhanced testCase
func performRequest(router *gin.Engine, tc testCase) *httptest.ResponseRecorder {
    var reqBody []byte
    if tc.body != nil {
        reqBody, _ = json.Marshal(tc.body)
    }

    w := httptest.NewRecorder()
    req := httptest.NewRequest(tc.method, tc.path, bytes.NewBuffer(reqBody))
    
    if tc.body != nil {
        req.Header.Set("Content-Type", "application/json")
    }
    
    router.ServeHTTP(w, req)
    return w
}

// TransactionWrapper wraps a test in a transaction
func withTransaction(db *sql.DB, test func(*sql.Tx) error) (err error) {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p)
        } else if err != nil {
            tx.Rollback()
        } else {
            err = tx.Commit()
        }
    }()
    
    err = test(tx)
    return err
}

// Add helper for running test with transaction
func runTestWithTx(t *testing.T, db *sql.DB, tc testCase, fn func()) {
    err := withTransaction(db, func(tx *sql.Tx) error {
        if tc.setupFn != nil {
            if err := tc.setupFn(tx); err != nil {
                return err
            }
        }
        fn()
        if tc.cleanupFn != nil && !tc.skipCleanup {
            if err := tc.cleanupFn(tx); err != nil {
                return err
            }
        }
        return nil
    })
    if err != nil {
        t.Fatalf("Transaction failed: %v", err)
    }
}

// Add cleanup helper
func cleanupTestDB(tx *sql.Tx) error {
    queries := []string{
        "DELETE FROM words",
        "DELETE FROM sqlite_sequence WHERE name='words'",
    }
    
    for _, query := range queries {
        if _, err := tx.Exec(query); err != nil {
            return err
        }
    }
    return nil
}

// Add assertion helper
func assertResponse(t *testing.T, w *httptest.ResponseRecorder, tc testCase) {
    t.Helper()
    
    assert.Equal(t, tc.wantStatus, w.Code)
    
    if tc.wantBody != "" {
        assert.Contains(t, w.Body.String(), tc.wantBody)
    }

    var response TestResponse
    if err := json.NewDecoder(w.Body).Decode(&response); err == nil {
        assert.NotEmpty(t, response.Status)
    }
}

// TestOptions represents configuration for test execution
type TestOptions struct {
    SkipCleanup bool
    Timeout     time.Duration
    RetryCount  int
}

// runTestWithOptions runs a test with configurable options
func runTestWithOptions(t *testing.T, db *sql.DB, tc testCase, opts TestOptions, fn func()) {
    t.Helper()
    
    done := make(chan bool)
    var testErr error

    go func() {
        err := withTransaction(db, func(tx *sql.Tx) error {
            if tc.setupFn != nil {
                if err := tc.setupFn(tx); err != nil {
                    return err
                }
            }
            
            fn()
            
            if !opts.SkipCleanup && tc.cleanupFn != nil {
                if err := tc.cleanupFn(tx); err != nil {
                    return err
                }
            }
            return nil
        })
        
        testErr = err
        done <- true
    }()

    select {
    case <-done:
        if testErr != nil {
            t.Fatalf("Test failed: %v", testErr)
        }
    case <-time.After(opts.Timeout):
        t.Fatal("Test timed out")
    }
}

// assertResponseFields verifies specific fields in the response
func assertResponseFields(t *testing.T, w *httptest.ResponseRecorder, fields map[string]interface{}) {
    t.Helper()
    
    var response map[string]interface{}
    if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
        t.Fatalf("Failed to decode response: %v", err)
    }

    for key, want := range fields {
        got, exists := response[key]
        if !exists {
            t.Errorf("Response missing field %q", key)
            continue
        }
        if got != want {
            t.Errorf("Field %q = %v, want %v", key, got, want)
        }
    }
}