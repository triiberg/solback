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

type stubSourceService struct {
	sources []models.Source
	err     error
}

func (s stubSourceService) GetSources(ctx context.Context) ([]models.Source, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.sources, nil
}

func TestSourcesHandlerSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	comment := "demo"
	sources := []models.Source{
		{ID: "1", URL: "https://example.com", Comment: &comment},
	}

	controller, err := NewSourcesController(stubSourceService{sources: sources})
	if err != nil {
		t.Fatalf("create sources controller: %v", err)
	}

	router := gin.New()
	if err := controller.RegisterRoutes(router); err != nil {
		t.Fatalf("register sources routes: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/sources", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var resp SourcesResponse
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(resp.Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(resp.Sources))
	}

	if resp.Sources[0].ID != "1" {
		t.Fatalf("unexpected source id %q", resp.Sources[0].ID)
	}
}

func TestSourcesHandlerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	controller, err := NewSourcesController(stubSourceService{err: errors.New("boom")})
	if err != nil {
		t.Fatalf("create sources controller: %v", err)
	}

	router := gin.New()
	if err := controller.RegisterRoutes(router); err != nil {
		t.Fatalf("register sources routes: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/sources", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Error == "" {
		t.Fatalf("expected error message")
	}
}
