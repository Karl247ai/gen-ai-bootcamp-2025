package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// APITestServer represents a test HTTP server
type APITestServer struct {
	Engine *gin.Engine
}

// NewAPITestServer creates a new test server
func NewAPITestServer() *APITestServer {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	return &APITestServer{
		Engine: engine,
	}
}

// SendRequest sends a test HTTP request and returns the response
func (s *APITestServer) SendRequest(t *testing.T, method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	// Create request
	req := httptest.NewRequest(method, path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve request
	s.Engine.ServeHTTP(w, req)

	return w
}

// DecodeResponse decodes the JSON response body
func DecodeResponse(t *testing.T, w *httptest.ResponseRecorder, v interface{}) {
	if err := json.NewDecoder(w.Body).Decode(v); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
} 