package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"solback/internal/models"

	"github.com/gin-gonic/gin"
)

type stubLogService struct {
	logs    []models.Log
	err     error
	limit   int
	eventID string
	deleted int
}

func (s *stubLogService) GetLogs(ctx context.Context, limit int, eventID string) ([]models.Log, error) {
	s.limit = limit
	s.eventID = eventID
	if s.err != nil {
		return nil, s.err
	}

	return s.logs, nil
}

func (s *stubLogService) TruncateLogs(ctx context.Context) (int, error) {
	if s.err != nil {
		return 0, s.err
	}
	return s.deleted, nil
}

func TestLogsHandlerDefaultLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logs := []models.Log{{ID: "1"}}
	service := &stubLogService{logs: logs}

	controller, err := NewLogsController(service)
	if err != nil {
		t.Fatalf("NewLogsController: %v", err)
	}

	router := gin.New()
	if err := controller.RegisterRoutes(router); err != nil {
		t.Fatalf("register logs routes: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if service.limit != defaultLogsLimit {
		t.Fatalf("limit = %d, want %d", service.limit, defaultLogsLimit)
	}
	if service.eventID != "" {
		t.Fatalf("eventID = %q, want empty", service.eventID)
	}

	var resp []models.Log
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp) != 1 || resp[0].ID != "1" {
		t.Fatalf("unexpected response: %v", resp)
	}
}

func TestLogsHandlerExplicitLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &stubLogService{logs: []models.Log{}}
	controller, err := NewLogsController(service)
	if err != nil {
		t.Fatalf("NewLogsController: %v", err)
	}

	router := gin.New()
	if err := controller.RegisterRoutes(router); err != nil {
		t.Fatalf("register logs routes: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/logs?n=5", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if service.limit != 5 {
		t.Fatalf("limit = %d, want %d", service.limit, 5)
	}
}

func TestLogsHandlerInvalidLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	controller, err := NewLogsController(&stubLogService{})
	if err != nil {
		t.Fatalf("NewLogsController: %v", err)
	}

	router := gin.New()
	if err := controller.RegisterRoutes(router); err != nil {
		t.Fatalf("register logs routes: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/logs?n=invalid", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestLogsHandlerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	controller, err := NewLogsController(&stubLogService{err: errors.New("boom")})
	if err != nil {
		t.Fatalf("NewLogsController: %v", err)
	}

	router := gin.New()
	if err := controller.RegisterRoutes(router); err != nil {
		t.Fatalf("register logs routes: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
}

func TestLogsHandlerEventID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &stubLogService{logs: []models.Log{}}
	controller, err := NewLogsController(service)
	if err != nil {
		t.Fatalf("NewLogsController: %v", err)
	}

	router := gin.New()
	if err := controller.RegisterRoutes(router); err != nil {
		t.Fatalf("register logs routes: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/logs?eventId=abc123", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if service.eventID != "abc123" {
		t.Fatalf("eventID = %q, want %q", service.eventID, "abc123")
	}
}

func TestLogsDeleteHandlerSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &stubLogService{deleted: 4}
	controller, err := NewLogsController(service)
	if err != nil {
		t.Fatalf("NewLogsController: %v", err)
	}

	router := gin.New()
	if err := controller.RegisterRoutes(router); err != nil {
		t.Fatalf("register logs routes: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/logs", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var resp DeleteLogsResponse
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Deleted != 4 {
		t.Fatalf("deleted = %d, want %d", resp.Deleted, 4)
	}
}

func TestLogsDeleteHandlerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	controller, err := NewLogsController(&stubLogService{err: errors.New("boom")})
	if err != nil {
		t.Fatalf("NewLogsController: %v", err)
	}

	router := gin.New()
	if err := controller.RegisterRoutes(router); err != nil {
		t.Fatalf("register logs routes: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/logs", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
}
