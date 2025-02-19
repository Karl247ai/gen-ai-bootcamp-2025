package router

import (
	"github.com/gin-gonic/gin"
	"github.com/your-org/lang-portal/internal/api/handler"
	"github.com/your-org/lang-portal/internal/middleware"
)

type Router struct {
	wordHandler  *handler.WordHandler
	groupHandler *handler.GroupHandler
	wordGroupHandler *handler.WordGroupHandler
	healthHandler *handler.HealthHandler
}

func NewRouter(wordHandler *handler.WordHandler, groupHandler *handler.GroupHandler, wordGroupHandler *handler.WordGroupHandler, healthHandler *handler.HealthHandler) *Router {
	return &Router{
		wordHandler:  wordHandler,
		groupHandler: groupHandler,
		wordGroupHandler: wordGroupHandler,
		healthHandler: healthHandler,
	}
}

func (r *Router) Setup(engine *gin.Engine) {
	// Add middleware
	engine.Use(middleware.Logger())
	engine.Use(middleware.CORS())
	engine.Use(middleware.ErrorHandler())

	// API v1 routes
	v1 := engine.Group("/api/v1")
	{
		// Words routes
		words := v1.Group("/words")
		{
			words.POST("", r.wordHandler.Create)
			words.GET("", r.wordHandler.List)
			words.GET("/:id", r.wordHandler.Get)
		}

		// Groups routes
		groups := v1.Group("/groups")
		{
			groups.POST("", r.groupHandler.Create)
			groups.GET("", r.groupHandler.List)
			groups.GET("/:id", r.groupHandler.Get)
		}

		// Word-Group relationship routes
		groups.Group("/:groupId/words").
			POST("/:wordId", r.wordGroupHandler.AddWordToGroup).
			DELETE("/:wordId", r.wordGroupHandler.RemoveWordFromGroup).
			GET("", r.wordGroupHandler.ListGroupWords)
	}

	// Health check
	engine.GET("/health", r.healthHandler.Check)
	engine.GET("/livez", r.healthHandler.LivenessCheck)
	engine.GET("/readyz", r.healthHandler.ReadinessCheck)
} 