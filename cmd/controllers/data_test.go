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
	group   string
	sumTech bool
	from    string
	to      string
	techIn  string
	sort    string
	limit   string
	deleted int
}

func (s *stubDataService) GetData(ctx context.Context, period string, technology string, groupPeriod string, sumTech bool, from string, to string, techIn string, sort string, limit string) ([]models.AuctionResult, error) {
	s.period = period
	s.tech = technology
	s.group = groupPeriod
	s.sumTech = sumTech
	s.from = from
	s.to = to
	s.techIn = techIn
	s.sort = sort
	s.limit = limit
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

	req := httptest.NewRequest(http.MethodGet, "/data?period=2024-2025&tech=Tech&group_period=year&sum_tech=true&from=2024-01&to=2025-12&tech_in=Solar,Wind&sort=year_desc,month_desc&limit=1", nil)
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
	if service.group != "year" {
		t.Fatalf("group_period = %q, want %q", service.group, "year")
	}
	if !service.sumTech {
		t.Fatalf("sum_tech = %v, want true", service.sumTech)
	}
	if service.from != "2024-01" {
		t.Fatalf("from = %q, want %q", service.from, "2024-01")
	}
	if service.to != "2025-12" {
		t.Fatalf("to = %q, want %q", service.to, "2025-12")
	}
	if service.techIn != "Solar,Wind" {
		t.Fatalf("tech_in = %q, want %q", service.techIn, "Solar,Wind")
	}
	if service.sort != "year_desc,month_desc" {
		t.Fatalf("sort = %q, want %q", service.sort, "year_desc,month_desc")
	}
	if service.limit != "1" {
		t.Fatalf("limit = %q, want %q", service.limit, "1")
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

func TestDataHandlerInvalidMonthRange(t *testing.T) {
	gin.SetMode(gin.TestMode)

	controller, err := NewDataController(&stubDataService{getErr: services.ErrInvalidMonthRange})
	if err != nil {
		t.Fatalf("NewDataController: %v", err)
	}

	router := gin.New()
	if err := controller.RegisterRoutes(router); err != nil {
		t.Fatalf("register data routes: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/data?from=2024-13", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestDataHandlerInvalidSumTech(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &stubDataService{}
	controller, err := NewDataController(service)
	if err != nil {
		t.Fatalf("NewDataController: %v", err)
	}

	router := gin.New()
	if err := controller.RegisterRoutes(router); err != nil {
		t.Fatalf("register data routes: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/data?sum_tech=maybe", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestDataHandlerInvalidLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	controller, err := NewDataController(&stubDataService{getErr: services.ErrInvalidLimit})
	if err != nil {
		t.Fatalf("NewDataController: %v", err)
	}

	router := gin.New()
	if err := controller.RegisterRoutes(router); err != nil {
		t.Fatalf("register data routes: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/data?limit=0", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestDataHandlerInvalidSort(t *testing.T) {
	gin.SetMode(gin.TestMode)

	controller, err := NewDataController(&stubDataService{getErr: services.ErrInvalidSort})
	if err != nil {
		t.Fatalf("NewDataController: %v", err)
	}

	router := gin.New()
	if err := controller.RegisterRoutes(router); err != nil {
		t.Fatalf("register data routes: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/data?sort=region_desc", nil)
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
