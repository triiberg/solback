package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type HtmlResult struct {
	URL        string
	StatusCode int
	Body       string
}

type HtmlService struct {
	client *http.Client
}

func NewHtmlService(client *http.Client) (*HtmlService, error) {
	if client == nil {
		client = http.DefaultClient
	}

	return &HtmlService{client: client}, nil
}

func (s *HtmlService) Fetch(ctx context.Context, url string) (HtmlResult, error) {
	if s == nil {
		return HtmlResult{}, errors.New("html service is nil")
	}
	if s.client == nil {
		return HtmlResult{}, errors.New("http client is nil")
	}
	if url == "" {
		return HtmlResult{}, errors.New("url is empty")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return HtmlResult{URL: url}, fmt.Errorf("build request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return HtmlResult{URL: url}, fmt.Errorf("do request: %w", err)
	}

	body, readErr := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if readErr != nil {
		return HtmlResult{URL: url, StatusCode: resp.StatusCode}, fmt.Errorf("read response: %w", readErr)
	}
	if closeErr != nil {
		return HtmlResult{URL: url, StatusCode: resp.StatusCode, Body: string(body)}, fmt.Errorf("close response: %w", closeErr)
	}

	return HtmlResult{URL: url, StatusCode: resp.StatusCode, Body: string(body)}, nil
}
