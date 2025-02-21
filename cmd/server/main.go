package main

import (
    "log"
    "github.com/gin-gonic/gin"
    _ "github.com/karl247ai/lang-portal/docs" // swagger docs
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
    "net/http"
    "github.com/karl247ai/lang-portal/internal/repository"
    "github.com/karl247ai/lang-portal/internal/api/handlers"
    "github.com/karl247ai/lang-portal/internal/api/middleware"
)

// @title           Language Learning Portal API
// @version         1.0
// @description     API for managing vocabulary and learning progress
// @host           localhost:8080
// @BasePath       /api/v1
func main() {
    // Initialize database
    db, err := repository.NewDB()
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()

    wordRepo := repository.NewWordRepository(db)
    wordHandler := handlers.NewWordHandler(wordRepo)

    r := gin.Default()
    r.Use(middleware.ErrorHandler())
    r.Use(middleware.Logger())
    
    // Add Swagger documentation
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    
    // Health check route
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "ok",
            "message": "Language Portal API is running",
            "database": "connected",
        })
    })

    // Word routes
    v1 := r.Group("/api/v1")
    {
        v1.GET("/words", wordHandler.GetWords)
        v1.POST("/words", wordHandler.CreateWord)
        v1.PUT("/api/v1/words/:id", wordHandler.UpdateWord)
        v1.DELETE("/api/v1/words/:id", wordHandler.DeleteWord)
    }
    
    log.Printf("Server starting on http://localhost:8080")
    if err := r.Run(":8080"); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}