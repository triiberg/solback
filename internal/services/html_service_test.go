package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHtmlServiceFetchSuccess(t *testing.T) {
	var hits int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	service, err := NewHtmlService(server.Client())
	if err != nil {
		t.Fatalf("NewHtmlService: %v", err)
	}

	result, err := service.Fetch(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if hits != 1 {
		t.Fatalf("hits = %d, want 1", hits)
	}
	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if result.Body != "ok" {
		t.Fatalf("Body = %q, want %q", result.Body, "ok")
	}
}

func TestHtmlServiceFetchNonOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("fail"))
	}))
	defer server.Close()

	service, err := NewHtmlService(server.Client())
	if err != nil {
		t.Fatalf("NewHtmlService: %v", err)
	}

	result, err := service.Fetch(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if result.StatusCode != http.StatusInternalServerError {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusInternalServerError)
	}
	if result.Body != "fail" {
		t.Fatalf("Body = %q, want %q", result.Body, "fail")
	}
}

func TestHtmlServiceFetchEmptyURL(t *testing.T) {
	service, err := NewHtmlService(http.DefaultClient)
	if err != nil {
		t.Fatalf("NewHtmlService: %v", err)
	}

	if _, err := service.Fetch(context.Background(), ""); err == nil {
		t.Fatalf("Fetch empty url: expected error")
	}
}
