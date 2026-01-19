package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAiCsvServiceParseAuctionResults(t *testing.T) {
	payload := AuctionPayload{
		SourceFile:   "20251119_August_2025_83_GLOBAL_Results_detailedresults.xlsx",
		Participants: 34,
		Headers:      []string{"Region", "Technology", "Total Volume Auctionned"},
		Rows: [][]string{
			{"Region1", "Tech1", "1"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var req openAiStructuredRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if req.ResponseFormat.Type != "json_schema" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var schema struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(req.ResponseFormat.JSONSchema, &schema); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if schema.Name != "auction_results" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if len(req.Messages) != 1 || !strings.Contains(req.Messages[0].Content, payload.SourceFile) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resp := openAiChatResponse{
			Choices: []openAiChoice{
				{Message: openAiResponseMessage{Content: `{"source_file":"20251119_August_2025_83_GLOBAL_Results_detailedresults.xlsx","participants":34,"rows":[{"year":2025,"month":8,"region":"Region1","technology":"Tech1","total_volume_auctioned":1,"total_volume_sold":1,"weighted_avg_price_eur_per_mwh":0.3}]}`}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	logWriter := &stubLogWriter{}
	service, err := NewOpenAiCsvService("test-key", logWriter, server.Client(), server.URL)
	if err != nil {
		t.Fatalf("NewOpenAiCsvService: %v", err)
	}

	result, err := service.ParseAuctionResults(context.Background(), payload)
	if err != nil {
		t.Fatalf("ParseAuctionResults: %v", err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(result.Rows))
	}
	if result.Rows[0].Year != 2025 {
		t.Fatalf("year = %v, want 2025", result.Rows[0].Year)
	}
	if len(logWriter.entries) == 0 {
		t.Fatalf("expected log entries")
	}
}
