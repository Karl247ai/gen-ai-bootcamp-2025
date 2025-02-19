package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/your-org/lang-portal/internal/middleware"
)

type APITestServer struct {
	Engine *gin.Engine
	T      testing.TB
}

func NewAPITestServer(t testing.TB) *APITestServer {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middleware.RequestContext())
	engine.Use(middleware.Recovery())

	return &APITestServer{
		Engine: engine,
		T:      t,
	}
}

func (s *APITestServer) SendRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			s.T.Fatalf("Failed to marshal request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Engine.ServeHTTP(w, req)
	return w
}

func (s *APITestServer) AssertResponse(w *httptest.ResponseRecorder, expectedStatus int) {
	if w.Code != expectedStatus {
		s.T.Errorf("Expected status code %d, got %d", expectedStatus, w.Code)
	}
}

func (s *APITestServer) DecodeResponse(w *httptest.ResponseRecorder, v interface{}) {
	if err := json.NewDecoder(w.Body).Decode(v); err != nil {
		s.T.Fatalf("Failed to decode response: %v", err)
	}
} 