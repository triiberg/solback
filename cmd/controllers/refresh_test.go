package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type stubRefreshService struct {
	err    error
	called chan struct{}
}

func (s *stubRefreshService) Refresh(ctx context.Context) error {
	if s.called != nil {
		s.called <- struct{}{}
	}
	return s.err
}

func TestRefreshHandlerSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &stubRefreshService{called: make(chan struct{}, 1)}
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

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, recorder.Code)
	}

	select {
	case <-service.called:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected refresh to be called")
	}

	var resp RefreshResponse
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Status != "started" {
		t.Fatalf("status = %q, want %q", resp.Status, "started")
	}
}

func TestRefreshHandlerAsyncError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	controller, err := NewRefreshController(&stubRefreshService{err: context.Canceled})
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

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, recorder.Code)
	}
}
