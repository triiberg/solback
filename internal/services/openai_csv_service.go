package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
)

const (
	csvMaxTokenEstimate  = 8000
	csvMaxRowsPerRequest = 500
	syntheticCellLength  = 8
)

const auctionResultsSchema = `{
  "name": "auction_results",
  "strict": true,
  "schema": {
    "type": "object",
    "properties": {
      "source_file": {
        "type": "string",
        "description": "Original XLSX file name",
        "year": "part of the filename",
        "month": "part of the filename"
      },
      "participants": {
        "type": "integer",
        "description": "Number of participants in the auction"
      },
      "rows": {
        "type": "array",
        "items": {
          "type": "object",
          "properties": {
            "year": {
              "type": "number"
            },
            "month": {
              "type": "number"
            },
            "region": {
              "type": "string"
            },
            "technology": {
              "type": "string"
            },
            "total_volume_auctioned": {
              "type": "number"
            },
            "total_volume_sold": {
              "type": "number"
            },
            "weighted_avg_price_eur_per_mwh": {
              "type": "number"
            }
          },
          "required": [
            "year",
            "month",
            "region",
            "technology",
            "total_volume_auctioned",
            "total_volume_sold",
            "weighted_avg_price_eur_per_mwh"
          ],
          "additionalProperties": false
        }
      }
    },
    "required": ["rows"],
    "additionalProperties": false
  }
}`

type OpenAiCsvService struct {
	apiKey     string
	client     *http.Client
	baseURL    string
	logService LogWriter
}

func NewOpenAiCsvService(apiKey string, logService LogWriter, client *http.Client, baseURL string) (*OpenAiCsvService, error) {
	if apiKey == "" {
		return nil, errors.New("openai api key is empty")
	}
	if logService == nil {
		return nil, errors.New("log service is nil")
	}
	if client == nil {
		client = http.DefaultClient
	}
	if baseURL == "" {
		baseURL = openAiDefaultBaseURL
	}

	return &OpenAiCsvService{
		apiKey:     apiKey,
		client:     client,
		baseURL:    baseURL,
		logService: logService,
	}, nil
}

func (s *OpenAiCsvService) ParseAuctionResults(ctx context.Context, payload AuctionPayload) (AuctionResults, error) {
	if s == nil {
		return AuctionResults{}, errors.New("openai csv service is nil")
	}
	if s.client == nil {
		return AuctionResults{}, errors.New("http client is nil")
	}
	if s.logService == nil {
		return AuctionResults{}, errors.New("log service is nil")
	}
	if s.apiKey == "" {
		return AuctionResults{}, errors.New("openai api key is empty")
	}
	if payload.SourceFile == "" {
		return AuctionResults{}, errors.New("source file is empty")
	}
	if payload.Participants <= 0 {
		return AuctionResults{}, errors.New("participants must be positive")
	}
	if len(payload.Headers) == 0 {
		return AuctionResults{}, errors.New("headers are empty")
	}
	if len(payload.Rows) == 0 {
		return AuctionResults{}, errors.New("rows are empty")
	}

	batches := splitRows(payload.Rows, csvMaxRowsPerRequest)
	estimate := estimateTokens(len(batches[0]), len(payload.Headers))
	ok := estimate <= csvMaxTokenEstimate
	precheckMsg := fmt.Sprintf("rows=%d batches=%d estimate=%d max_tokens=%d max_rows=%d ok=%t", len(payload.Rows), len(batches), estimate, csvMaxTokenEstimate, csvMaxRowsPerRequest, ok)
	if !ok {
		_ = s.logService.CreateLog(ctx, LogActionOpenAICSVParse, LogOutcomeFail, &precheckMsg)
		return AuctionResults{}, errors.New("payload too large for configured token budget")
	}
	_ = s.logService.CreateLog(ctx, LogActionOpenAICSVParse, LogOutcomeSuccess, &precheckMsg)

	combined := AuctionResults{
		SourceFile:   payload.SourceFile,
		Participants: payload.Participants,
	}

	for batchIndex, batchRows := range batches {
		batchPayload := AuctionPayload{
			SourceFile:   payload.SourceFile,
			Participants: payload.Participants,
			Headers:      payload.Headers,
			Rows:         batchRows,
		}

		result, err := s.parseBatch(ctx, batchPayload, batchIndex+1)
		if err != nil {
			return AuctionResults{}, err
		}

		if result.SourceFile == "" {
			result.SourceFile = payload.SourceFile
		}
		if result.Participants == 0 {
			result.Participants = payload.Participants
		}

		if err := validateAuctionResults(result); err != nil {
			msg := fmt.Sprintf("validate openai csv result: %v", err)
			_ = s.logService.CreateLog(ctx, LogActionOpenAICSVParse, LogOutcomeFail, &msg)
			return AuctionResults{}, err
		}

		combined.Rows = append(combined.Rows, result.Rows...)
	}

	if len(combined.Rows) == 0 {
		return AuctionResults{}, errors.New("openai returned empty rows")
	}

	return combined, nil
}

func (s *OpenAiCsvService) parseBatch(ctx context.Context, payload AuctionPayload, batchIndex int) (AuctionResults, error) {
	prompt := buildCsvPrompt(payload, batchIndex)
	for attempt := 1; attempt <= 3; attempt++ {
		result, err := s.callOpenAiStructured(ctx, prompt)
		if err != nil {
			msg := fmt.Sprintf("openai csv parse batch=%d attempt=%d: %v", batchIndex, attempt, err)
			_ = s.logService.CreateLog(ctx, LogActionOpenAICSVParse, LogOutcomeFail, &msg)
			if attempt == 3 {
				return AuctionResults{}, err
			}
			continue
		}

		msg := fmt.Sprintf("openai csv parse batch=%d rows=%d", batchIndex, len(result.Rows))
		_ = s.logService.CreateLog(ctx, LogActionOpenAICSVParse, LogOutcomeSuccess, &msg)
		return result, nil
	}

	return AuctionResults{}, errors.New("openai csv retries exhausted")
}

func (s *OpenAiCsvService) callOpenAiStructured(ctx context.Context, prompt string) (AuctionResults, error) {
	requestBody := openAiStructuredRequest{
		Model:       openAiDefaultModel,
		Temperature: 0,
		Messages: []openAiMessage{
			{Role: "user", Content: prompt},
		},
		ResponseFormat: openAiResponseFormat{
			Type:       "json_schema",
			JSONSchema: json.RawMessage(auctionResultsSchema),
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(requestBody); err != nil {
		return AuctionResults{}, fmt.Errorf("encode request: %w", err)
	}

	endpoint := strings.TrimRight(s.baseURL, "/") + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &buf)
	if err != nil {
		return AuctionResults{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return AuctionResults{}, fmt.Errorf("send request: %w", err)
	}

	body, readErr := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if readErr != nil {
		return AuctionResults{}, fmt.Errorf("read response: %w", readErr)
	}
	if closeErr != nil {
		return AuctionResults{}, fmt.Errorf("close response: %w", closeErr)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return AuctionResults{}, fmt.Errorf("openai status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var response openAiChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return AuctionResults{}, fmt.Errorf("decode response: %w", err)
	}
	if len(response.Choices) == 0 {
		return AuctionResults{}, errors.New("openai response has no choices")
	}

	content := strings.TrimSpace(response.Choices[0].Message.Content)
	if content == "" {
		return AuctionResults{}, errors.New("openai response content is empty")
	}

	result, err := parseAuctionResults(content)
	if err != nil {
		return AuctionResults{}, err
	}

	return result, nil
}

func buildCsvPrompt(payload AuctionPayload, batchIndex int) string {
	request := struct {
		SourceFile   string     `json:"source_file"`
		Participants int        `json:"participants"`
		Headers      []string   `json:"headers"`
		Rows         [][]string `json:"rows"`
		Batch        int        `json:"batch"`
	}{
		SourceFile:   payload.SourceFile,
		Participants: payload.Participants,
		Headers:      payload.Headers,
		Rows:         payload.Rows,
		Batch:        batchIndex,
	}

	payloadJSON, err := json.Marshal(request)
	if err != nil {
		payloadJSON = []byte(`{}`)
	}

	return fmt.Sprintf(`Instructions:
1. Convert the provided rows into the auction_results schema.
2. Map headers to canonical field names.
3. Convert decimal commas to decimal points.
4. Convert "-" or empty cells to null.
5. Coerce numeric values to numbers.
6. Use the source_file to infer year/month when needed.
7. Return only JSON that matches the provided schema.

Payload:
%s`, payloadJSON)
}

func parseAuctionResults(content string) (AuctionResults, error) {
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSpace(trimmed)
		trimmed = strings.TrimPrefix(trimmed, "json")
		if idx := strings.LastIndex(trimmed, "```"); idx != -1 {
			trimmed = trimmed[:idx]
		}
		trimmed = strings.TrimSpace(trimmed)
	}

	var result AuctionResults
	if err := json.Unmarshal([]byte(trimmed), &result); err != nil {
		return AuctionResults{}, fmt.Errorf("parse openai json: %w", err)
	}

	return result, nil
}

func validateAuctionResults(result AuctionResults) error {
	if len(result.Rows) == 0 {
		return errors.New("rows are empty")
	}

	for index, row := range result.Rows {
		if row.Year == 0 {
			return fmt.Errorf("row %d year is empty", index)
		}
		if row.Month == 0 {
			return fmt.Errorf("row %d month is empty", index)
		}
		if math.Trunc(row.Year) != row.Year {
			return fmt.Errorf("row %d year is not an integer", index)
		}
		if math.Trunc(row.Month) != row.Month {
			return fmt.Errorf("row %d month is not an integer", index)
		}
		if strings.TrimSpace(row.Region) == "" {
			return fmt.Errorf("row %d region is empty", index)
		}
		if strings.TrimSpace(row.Technology) == "" {
			return fmt.Errorf("row %d technology is empty", index)
		}
	}

	return nil
}

func estimateTokens(rowCount int, columnCount int) int {
	if rowCount <= 0 || columnCount <= 0 {
		return 0
	}

	chars := rowCount * columnCount * syntheticCellLength
	return chars / 4
}

func splitRows(rows [][]string, maxRows int) [][][]string {
	if maxRows <= 0 || len(rows) == 0 {
		return [][][]string{rows}
	}

	var batches [][][]string
	for start := 0; start < len(rows); start += maxRows {
		end := start + maxRows
		if end > len(rows) {
			end = len(rows)
		}
		batches = append(batches, rows[start:end])
	}
	return batches
}

type openAiStructuredRequest struct {
	Model          string               `json:"model"`
	Messages       []openAiMessage      `json:"messages"`
	Temperature    float32              `json:"temperature"`
	ResponseFormat openAiResponseFormat `json:"response_format"`
}

type openAiResponseFormat struct {
	Type       string          `json:"type"`
	JSONSchema json.RawMessage `json:"json_schema"`
}
