package services

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"solback/internal/models"
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

type stubHtmlFetcher struct {
	results map[string]HtmlResult
	errs    map[string]error
}

func (s stubHtmlFetcher) Fetch(ctx context.Context, url string) (HtmlResult, error) {
	if err, ok := s.errs[url]; ok {
		return HtmlResult{URL: url}, err
	}
	if result, ok := s.results[url]; ok {
		return result, nil
	}
	return HtmlResult{URL: url, StatusCode: http.StatusNotFound}, nil
}

type stubOpenAiExtractor struct {
	result OpenAiResult
	err    error
}

func (s stubOpenAiExtractor) ExtractZipLink(ctx context.Context, html string) (OpenAiResult, error) {
	if s.err != nil {
		return OpenAiResult{}, s.err
	}
	return s.result, nil
}

func TestPipelineServiceRefresh(t *testing.T) {
	sources := []models.Source{
		{URL: "https://example.com/ok"},
		{URL: "https://example.com/fail"},
	}
	htmlFetcher := stubHtmlFetcher{
		results: map[string]HtmlResult{
			"https://example.com/ok":   {URL: "https://example.com/ok", StatusCode: http.StatusOK, Body: "<table></table>"},
			"https://example.com/fail": {URL: "https://example.com/fail", StatusCode: http.StatusInternalServerError, Body: "fail"},
		},
	}

	logWriter := &stubLogWriter{}
	service, err := NewPipelineService(
		stubSourceService{sources: sources},
		htmlFetcher,
		stubOpenAiExtractor{result: OpenAiResult{Error: "NO_RESULTS"}},
		logWriter,
	)
	if err != nil {
		t.Fatalf("NewPipelineService: %v", err)
	}

	if err := service.Refresh(context.Background()); err == nil {
		t.Fatalf("Refresh: expected error")
	}

	if len(logWriter.entries) != 3 {
		t.Fatalf("log entries = %d, want 3", len(logWriter.entries))
	}
	if logWriter.entries[0].action != LogActionDataRetrieval {
		t.Fatalf("log action = %q, want %q", logWriter.entries[0].action, LogActionDataRetrieval)
	}
	if logWriter.entries[1].outcome != LogOutcomeSuccess {
		t.Fatalf("first fetch outcome = %q, want %q", logWriter.entries[1].outcome, LogOutcomeSuccess)
	}
	if logWriter.entries[2].outcome != LogOutcomeFail {
		t.Fatalf("second fetch outcome = %q, want %q", logWriter.entries[2].outcome, LogOutcomeFail)
	}
}

func TestPipelineServiceRefreshSourceError(t *testing.T) {
	logWriter := &stubLogWriter{}
	service, err := NewPipelineService(
		stubSourceService{err: errors.New("boom")},
		stubHtmlFetcher{},
		stubOpenAiExtractor{},
		logWriter,
	)
	if err != nil {
		t.Fatalf("NewPipelineService: %v", err)
	}

	if err := service.Refresh(context.Background()); err == nil {
		t.Fatalf("Refresh: expected error")
	}
	if len(logWriter.entries) != 2 {
		t.Fatalf("log entries = %d, want 2", len(logWriter.entries))
	}
	if logWriter.entries[1].outcome != LogOutcomeFail {
		t.Fatalf("log outcome = %q, want %q", logWriter.entries[1].outcome, LogOutcomeFail)
	}
}
