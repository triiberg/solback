package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubRefreshService struct {
	err    error
	called bool
}

func (s *stubRefreshService) Refresh(ctx context.Context) error {
	s.called = true
	return s.err
}

func TestRefreshHandlerSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &stubRefreshService{}
	controller, err := NewRefreshController(service)
	if err != nil {
		t.Fatalf("NewRefreshController: %v", err)
	}

	router := gin.New()
	if err := controller.RegisterRoutes(router); err != nil {
		t.Fatalf("register refresh routes: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/refresh", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if !service.called {
		t.Fatalf("expected refresh to be called")
	}

	var resp RefreshResponse
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Status != "ok" {
		t.Fatalf("status = %q, want %q", resp.Status, "ok")
	}
}

func TestRefreshHandlerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	controller, err := NewRefreshController(&stubRefreshService{err: errors.New("boom")})
	if err != nil {
		t.Fatalf("NewRefreshController: %v", err)
	}

	router := gin.New()
	if err := controller.RegisterRoutes(router); err != nil {
		t.Fatalf("register refresh routes: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/refresh", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
}
