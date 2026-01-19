package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"solback/internal/models"
	"solback/internal/services"

	"github.com/gin-gonic/gin"
)

type stubDataService struct {
	results []models.AuctionResult
	getErr  error
	delErr  error
	period  string
	tech    string
	deleted int
}

func (s *stubDataService) GetData(ctx context.Context, period string, technology string) ([]models.AuctionResult, error) {
	s.period = period
	s.tech = technology
	if s.getErr != nil {
		return nil, s.getErr
	}
	return s.results, nil
}

func (s *stubDataService) DeleteData(ctx context.Context) (int, error) {
	if s.delErr != nil {
		return 0, s.delErr
	}
	return s.deleted, nil
}

func TestDataHandlerSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &stubDataService{
		results: []models.AuctionResult{{ID: "1", Technology: "Tech"}},
	}
	controller, err := NewDataController(service)
	if err != nil {
		t.Fatalf("NewDataController: %v", err)
	}

	router := gin.New()
	if err := controller.RegisterRoutes(router); err != nil {
		t.Fatalf("register data routes: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/data?period=2024-2025&tech=Tech", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if service.period != "2024-2025" {
		t.Fatalf("period = %q, want %q", service.period, "2024-2025")
	}
	if service.tech != "Tech" {
		t.Fatalf("tech = %q, want %q", service.tech, "Tech")
	}

	var resp []models.AuctionResult
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp) != 1 || resp[0].ID != "1" {
		t.Fatalf("unexpected response: %v", resp)
	}
}

func TestDataHandlerInvalidPeriod(t *testing.T) {
	gin.SetMode(gin.TestMode)

	controller, err := NewDataController(&stubDataService{getErr: services.ErrInvalidPeriod})
	if err != nil {
		t.Fatalf("NewDataController: %v", err)
	}

	router := gin.New()
	if err := controller.RegisterRoutes(router); err != nil {
		t.Fatalf("register data routes: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/data?period=2024", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestDataHandlerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	controller, err := NewDataController(&stubDataService{getErr: errors.New("boom")})
	if err != nil {
		t.Fatalf("NewDataController: %v", err)
	}

	router := gin.New()
	if err := controller.RegisterRoutes(router); err != nil {
		t.Fatalf("register data routes: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/data", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
}

func TestDataDeleteHandlerSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &stubDataService{deleted: 3}
	controller, err := NewDataController(service)
	if err != nil {
		t.Fatalf("NewDataController: %v", err)
	}

	router := gin.New()
	if err := controller.RegisterRoutes(router); err != nil {
		t.Fatalf("register data routes: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/data", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var resp DeleteDataResponse
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Deleted != 3 {
		t.Fatalf("deleted = %d, want %d", resp.Deleted, 3)
	}
}

func TestDataDeleteHandlerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	controller, err := NewDataController(&stubDataService{delErr: errors.New("boom")})
	if err != nil {
		t.Fatalf("NewDataController: %v", err)
	}

	router := gin.New()
	if err := controller.RegisterRoutes(router); err != nil {
		t.Fatalf("register data routes: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/data", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
}
