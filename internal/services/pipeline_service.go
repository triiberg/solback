package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

type PipelineService struct {
	sourceService SourceProvider
	htmlService   HtmlFetcher
	openAiService OpenAiExtractor
	logService    LogWriter
}

func NewPipelineService(sourceService SourceProvider, htmlService HtmlFetcher, openAiService OpenAiExtractor, logService LogWriter) (*PipelineService, error) {
	if sourceService == nil {
		return nil, errors.New("source service is nil")
	}
	if htmlService == nil {
		return nil, errors.New("html service is nil")
	}
	if openAiService == nil {
		return nil, errors.New("openai service is nil")
	}
	if logService == nil {
		return nil, errors.New("log service is nil")
	}

	return &PipelineService{
		sourceService: sourceService,
		htmlService:   htmlService,
		openAiService: openAiService,
		logService:    logService,
	}, nil
}

func (s *PipelineService) Refresh(ctx context.Context) error {
	if s == nil {
		return errors.New("pipeline service is nil")
	}
	if s.sourceService == nil {
		return errors.New("source service is nil")
	}
	if s.htmlService == nil {
		return errors.New("html service is nil")
	}
	if s.openAiService == nil {
		return errors.New("openai service is nil")
	}
	if s.logService == nil {
		return errors.New("log service is nil")
	}

	startMsg := "pipeline refresh started"
	if err := s.logService.CreateLog(ctx, LogActionDataRetrieval, LogOutcomeSuccess, &startMsg); err != nil {
		return err
	}

	sources, err := s.sourceService.GetSources(ctx)
	if err != nil {
		failMsg := fmt.Sprintf("get sources: %v", err)
		_ = s.logService.CreateLog(ctx, LogActionDataRetrieval, LogOutcomeFail, &failMsg)
		return fmt.Errorf("get sources: %w", err)
	}

	var refreshErr error
	for _, source := range sources {
		if source.URL == "" {
			failMsg := "source url is empty"
			_ = s.logService.CreateLog(ctx, LogActionDataRetrieval, LogOutcomeFail, &failMsg)
			if refreshErr == nil {
				refreshErr = errors.New("source url is empty")
			}
			continue
		}

		result, err := s.htmlService.Fetch(ctx, source.URL)
		if err != nil {
			failMsg := fmt.Sprintf("fetch url=%s: %v", source.URL, err)
			_ = s.logService.CreateLog(ctx, LogActionDataRetrieval, LogOutcomeFail, &failMsg)
			if refreshErr == nil {
				refreshErr = fmt.Errorf("fetch url=%s: %w", source.URL, err)
			}
			continue
		}

		outcome := LogOutcomeSuccess
		if result.StatusCode < http.StatusOK || result.StatusCode >= http.StatusMultipleChoices {
			outcome = LogOutcomeFail
		}

		resultMsg := fmt.Sprintf("url=%s status=%d", source.URL, result.StatusCode)
		if logErr := s.logService.CreateLog(ctx, LogActionDataRetrieval, outcome, &resultMsg); logErr != nil && refreshErr == nil {
			refreshErr = fmt.Errorf("log retrieval result: %w", logErr)
		}

		if outcome == LogOutcomeFail {
			if refreshErr == nil {
				refreshErr = fmt.Errorf("request failed for %s", source.URL)
			}
			continue
		}

		openAiResult, err := s.openAiService.ExtractZipLink(ctx, result.Body)
		if err != nil {
			if refreshErr == nil {
				refreshErr = fmt.Errorf("openai extract: %w", err)
			}
			continue
		}
		if openAiResult.Error != "" && refreshErr == nil {
			refreshErr = fmt.Errorf("openai extract returned error: %s", openAiResult.Error)
		}
	}

	return refreshErr
}
