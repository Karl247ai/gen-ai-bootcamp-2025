package testutil

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// GenerateTestTraffic generates test traffic for monitoring tests
func GenerateTestTraffic(t *testing.T, router *gin.Engine) {
	// Generate successful requests
	for i := 0; i < 10; i++ {
		PerformRequest(router, "GET", "/api/v1/words", nil)
		PerformRequest(router, "GET", "/api/v1/groups", nil)
		time.Sleep(10 * time.Millisecond)
	}

	// Generate cache activity
	for i := 0; i < 5; i++ {
		PerformRequest(router, "GET", "/api/v1/words/1", nil)
		time.Sleep(10 * time.Millisecond)
	}

	// Generate errors
	PerformRequest(router, "GET", "/api/v1/words/999", nil)
	PerformRequest(router, "GET", "/api/v1/groups/999", nil)

	// Generate slow requests
	for i := 0; i < 3; i++ {
		time.Sleep(100 * time.Millisecond)
		PerformRequest(router, "GET", "/api/v1/words", nil)
	}
}

// PerformRequest performs a test request
func PerformRequest(router *gin.Engine, method, path string, body interface{}) *http.Response {
	req := NewRequest(method, path, body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w.Result()
}

// NewRequest creates a new test request
func NewRequest(method, path string, body interface{}) *http.Request {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, path, reqBody)
	if err != nil {
		panic(err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
} 