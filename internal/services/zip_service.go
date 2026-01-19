package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type ZipService struct {
	client     *http.Client
	logService LogWriter
}

func NewZipService(logService LogWriter, client *http.Client) (*ZipService, error) {
	if logService == nil {
		return nil, errors.New("log service is nil")
	}
	if client == nil {
		client = http.DefaultClient
	}

	return &ZipService{
		client:     client,
		logService: logService,
	}, nil
}

func (s *ZipService) Download(ctx context.Context, link string, sourceURL string, eventID *string) (ZipResult, error) {
	if s == nil {
		return ZipResult{}, errors.New("zip service is nil")
	}
	if s.client == nil {
		return ZipResult{}, errors.New("http client is nil")
	}
	if s.logService == nil {
		return ZipResult{}, errors.New("log service is nil")
	}
	if link == "" {
		return ZipResult{}, errors.New("zip link is empty")
	}

	zipURL, err := resolveZipURL(link, sourceURL)
	if err != nil {
		failMsg := fmt.Sprintf("resolve zip url: %v", err)
		_ = s.logService.CreateLog(ctx, eventID, LogActionZipDownload, LogOutcomeFail, &failMsg)
		return ZipResult{}, err
	}

	if !strings.HasSuffix(strings.ToLower(zipURL), ".zip") {
		failMsg := fmt.Sprintf("zip url does not end with .zip: %s", zipURL)
		_ = s.logService.CreateLog(ctx, eventID, LogActionZipDownload, LogOutcomeFail, &failMsg)
		return ZipResult{}, errors.New("zip url must end with .zip")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, zipURL, nil)
	if err != nil {
		failMsg := fmt.Sprintf("build zip request: %v", err)
		_ = s.logService.CreateLog(ctx, eventID, LogActionZipDownload, LogOutcomeFail, &failMsg)
		return ZipResult{}, fmt.Errorf("build zip request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		failMsg := fmt.Sprintf("download zip: %v", err)
		_ = s.logService.CreateLog(ctx, eventID, LogActionZipDownload, LogOutcomeFail, &failMsg)
		return ZipResult{}, fmt.Errorf("download zip: %w", err)
	}

	body, readErr := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if readErr != nil {
		failMsg := fmt.Sprintf("read zip response: %v", readErr)
		_ = s.logService.CreateLog(ctx, eventID, LogActionZipDownload, LogOutcomeFail, &failMsg)
		return ZipResult{}, fmt.Errorf("read zip response: %w", readErr)
	}
	if closeErr != nil {
		failMsg := fmt.Sprintf("close zip response: %v", closeErr)
		_ = s.logService.CreateLog(ctx, eventID, LogActionZipDownload, LogOutcomeFail, &failMsg)
		return ZipResult{}, fmt.Errorf("close zip response: %w", closeErr)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		failMsg := fmt.Sprintf("zip download status=%d url=%s", resp.StatusCode, zipURL)
		_ = s.logService.CreateLog(ctx, eventID, LogActionZipDownload, LogOutcomeFail, &failMsg)
		return ZipResult{URL: zipURL, StatusCode: resp.StatusCode}, fmt.Errorf("zip download failed with status %d", resp.StatusCode)
	}

	successMsg := fmt.Sprintf("zip download status=%d url=%s bytes=%d", resp.StatusCode, zipURL, len(body))
	_ = s.logService.CreateLog(ctx, eventID, LogActionZipDownload, LogOutcomeSuccess, &successMsg)

	return ZipResult{URL: zipURL, StatusCode: resp.StatusCode, Bytes: body}, nil
}

func resolveZipURL(link string, sourceURL string) (string, error) {
	parsed, err := url.Parse(link)
	if err != nil {
		return "", fmt.Errorf("parse link: %w", err)
	}
	if parsed.IsAbs() {
		return parsed.String(), nil
	}
	if sourceURL == "" {
		return "", errors.New("source url is empty")
	}

	base, err := url.Parse(sourceURL)
	if err != nil {
		return "", fmt.Errorf("parse source url: %w", err)
	}

	resolved := base.ResolveReference(parsed)
	return resolved.String(), nil
}
