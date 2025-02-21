package test

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "github.com/gin-gonic/gin"
    "github.com/karl247ai/lang-portal/internal/models"
    "github.com/karl247ai/lang-portal/internal/repository"
    "github.com/karl247ai/lang-portal/internal/api/handlers"
    _ "github.com/mattn/go-sqlite3"
)

func setupTestDB() *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        panic(err)
    }

    // Create test tables
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS words (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            japanese TEXT NOT NULL,
            romaji TEXT NOT NULL,
            english TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
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

type testCase struct {
    name       string
    method     string
    path       string
    body       interface{}
    wantStatus int
    wantBody   string
}

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