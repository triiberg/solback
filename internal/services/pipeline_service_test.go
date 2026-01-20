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

type stubZipDownloader struct {
	result ZipResult
	err    error
}

func (s stubOpenAiExtractor) ExtractZipLink(ctx context.Context, html string, eventID *string) (OpenAiResult, error) {
	if s.err != nil {
		return OpenAiResult{}, s.err
	}
	return s.result, nil
}

func (s stubZipDownloader) Download(ctx context.Context, link string, sourceURL string, eventID *string) (ZipResult, error) {
	if s.err != nil {
		return ZipResult{}, s.err
	}
	return s.result, nil
}

type stubZipProcessor struct {
	payloads []AuctionPayload
	err      error
}

func (s stubZipProcessor) ExtractAuctionPayloads(ctx context.Context, zipBytes []byte) ([]AuctionPayload, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.payloads, nil
}

type stubProcessedFileTracker struct {
	processed map[string]bool
	err       error
	marked    []string
}

func (s *stubProcessedFileTracker) IsProcessed(ctx context.Context, filename string) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	if s.processed == nil {
		return false, nil
	}
	return s.processed[filename], nil
}

func (s *stubProcessedFileTracker) MarkProcessed(ctx context.Context, filename string) error {
	if s.err != nil {
		return s.err
	}
	s.marked = append(s.marked, filename)
	if s.processed == nil {
		s.processed = map[string]bool{}
	}
	s.processed[filename] = true
	return nil
}

type stubAuctionParser struct {
	result AuctionResults
	err    error
}

func (s stubAuctionParser) ParseAuctionResults(ctx context.Context, payload AuctionPayload, eventID *string) (AuctionResults, error) {
	if s.err != nil {
		return s.result, s.err
	}
	return s.result, nil
}

type stubDataStorer struct {
	count int
	err   error
}

func (s *stubDataStorer) StoreAuctionResults(ctx context.Context, results AuctionResults, eventID *string) (int, error) {
	if s.err != nil {
		return 0, s.err
	}
	s.count += len(results.Rows)
	return len(results.Rows), nil
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
	dataStorer := &stubDataStorer{}
	service, err := NewPipelineService(
		stubSourceService{sources: sources},
		htmlFetcher,
		stubOpenAiExtractor{result: OpenAiResult{Link: "https://example.com/file.zip"}},
		stubZipDownloader{result: ZipResult{URL: "https://example.com/file.zip", StatusCode: http.StatusOK, Bytes: []byte("zip")}},
		stubZipProcessor{payloads: []AuctionPayload{{SourceFile: "file.xlsx", Participants: 1, Headers: []string{"Region", "Technology"}, Rows: [][]string{{"Region", "Tech"}}}}},
		&stubProcessedFileTracker{},
		stubAuctionParser{result: AuctionResults{SourceFile: "file.xlsx", Participants: 1, Rows: []AuctionRow{{Year: 2025, Month: 8, Region: "Region", Technology: "Tech", TotalVolumeAuctioned: 1, TotalVolumeSold: 1, WeightedAvgPriceEurPerMwh: 1}}}},
		dataStorer,
		logWriter,
	)
	if err != nil {
		t.Fatalf("NewPipelineService: %v", err)
	}

	if err := service.Refresh(context.Background()); err == nil {
		t.Fatalf("Refresh: expected error")
	}

	if len(logWriter.entries) < 3 {
		t.Fatalf("log entries = %d, want at least 3", len(logWriter.entries))
	}
	for i, entry := range logWriter.entries {
		if entry.eventID == nil || *entry.eventID == "" {
			t.Fatalf("log entry %d missing eventID", i)
		}
	}
	if logWriter.entries[0].action != LogActionDataRetrieval {
		t.Fatalf("log action = %q, want %q", logWriter.entries[0].action, LogActionDataRetrieval)
	}
	if dataStorer.count == 0 {
		t.Fatalf("expected data rows to be stored")
	}
	if logWriter.entries[1].outcome != LogOutcomeSuccess {
		t.Fatalf("first fetch outcome = %q, want %q", logWriter.entries[1].outcome, LogOutcomeSuccess)
	}

	var retrievalFail bool
	for _, entry := range logWriter.entries {
		if entry.action == LogActionDataRetrieval && entry.outcome == LogOutcomeFail {
			retrievalFail = true
			break
		}
	}
	if !retrievalFail {
		t.Fatalf("expected a failed data retrieval log entry")
	}
}

func TestPipelineServiceRefreshSourceError(t *testing.T) {
	logWriter := &stubLogWriter{}
	service, err := NewPipelineService(
		stubSourceService{err: errors.New("boom")},
		stubHtmlFetcher{},
		stubOpenAiExtractor{},
		stubZipDownloader{},
		stubZipProcessor{},
		&stubProcessedFileTracker{},
		stubAuctionParser{},
		&stubDataStorer{},
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

func TestPipelineServiceRefreshStoresPartialCsvResults(t *testing.T) {
	sources := []models.Source{
		{URL: "https://example.com/ok"},
	}
	htmlFetcher := stubHtmlFetcher{
		results: map[string]HtmlResult{
			"https://example.com/ok": {URL: "https://example.com/ok", StatusCode: http.StatusOK, Body: "<table></table>"},
		},
	}

	logWriter := &stubLogWriter{}
	dataStorer := &stubDataStorer{}
	service, err := NewPipelineService(
		stubSourceService{sources: sources},
		htmlFetcher,
		stubOpenAiExtractor{result: OpenAiResult{Link: "https://example.com/file.zip"}},
		stubZipDownloader{result: ZipResult{URL: "https://example.com/file.zip", StatusCode: http.StatusOK, Bytes: []byte("zip")}},
		stubZipProcessor{payloads: []AuctionPayload{{SourceFile: "file.xlsx", Participants: 1, Headers: []string{"Region", "Technology"}, Rows: [][]string{{"Region", "Tech"}}}}},
		&stubProcessedFileTracker{},
		stubAuctionParser{
			result: AuctionResults{SourceFile: "file.xlsx", Participants: 1, Rows: []AuctionRow{{Year: 2025, Month: 8, Region: "Region", Technology: "Tech", TotalVolumeAuctioned: 1, TotalVolumeSold: 1, WeightedAvgPriceEurPerMwh: 1}}},
			err:    errors.New("partial parse failure"),
		},
		dataStorer,
		logWriter,
	)
	if err != nil {
		t.Fatalf("NewPipelineService: %v", err)
	}

	if err := service.Refresh(context.Background()); err == nil {
		t.Fatalf("Refresh: expected error")
	}
	if dataStorer.count == 0 {
		t.Fatalf("expected data rows to be stored")
	}
}

func TestPipelineServiceRefreshSkipsProcessedZip(t *testing.T) {
	sources := []models.Source{
		{URL: "https://example.com/ok"},
	}
	htmlFetcher := stubHtmlFetcher{
		results: map[string]HtmlResult{
			"https://example.com/ok": {URL: "https://example.com/ok", StatusCode: http.StatusOK, Body: "<table></table>"},
		},
	}

	processed := &stubProcessedFileTracker{processed: map[string]bool{"file.zip": true}}
	logWriter := &stubLogWriter{}
	dataStorer := &stubDataStorer{}
	zipDownloader := stubZipDownloader{result: ZipResult{URL: "https://example.com/file.zip", StatusCode: http.StatusOK, Bytes: []byte("zip")}}
	service, err := NewPipelineService(
		stubSourceService{sources: sources},
		htmlFetcher,
		stubOpenAiExtractor{result: OpenAiResult{Link: "https://example.com/file.zip"}},
		zipDownloader,
		stubZipProcessor{payloads: []AuctionPayload{{SourceFile: "file.xlsx", Participants: 1, Headers: []string{"Region", "Technology"}, Rows: [][]string{{"Region", "Tech"}}}}},
		processed,
		stubAuctionParser{result: AuctionResults{SourceFile: "file.xlsx", Participants: 1, Rows: []AuctionRow{{Year: 2025, Month: 8, Region: "Region", Technology: "Tech", TotalVolumeAuctioned: 1, TotalVolumeSold: 1, WeightedAvgPriceEurPerMwh: 1}}}},
		dataStorer,
		logWriter,
	)
	if err != nil {
		t.Fatalf("NewPipelineService: %v", err)
	}

	if err := service.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if dataStorer.count != 0 {
		t.Fatalf("expected no data rows stored")
	}
	if len(processed.marked) != 0 {
		t.Fatalf("expected no new processed marks")
	}
}
