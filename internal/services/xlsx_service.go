package services

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

const (
	aggregatedTitleMarker  = "aggregated auction results"
	participantsLabelMatch = "number of participants"
)

type XlsxService struct{}

func NewXlsxService() (*XlsxService, error) {
	return &XlsxService{}, nil
}

func (s *XlsxService) ExtractAuctionPayloads(ctx context.Context, zipBytes []byte) ([]AuctionPayload, error) {
	if s == nil {
		return nil, errors.New("xlsx service is nil")
	}
	if len(zipBytes) == 0 {
		return nil, errors.New("zip bytes are empty")
	}
	_ = ctx

	reader := bytes.NewReader(zipBytes)
	zipReader, err := zip.NewReader(reader, int64(len(zipBytes)))
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}

	var payloads []AuctionPayload
	for _, file := range zipReader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		if strings.HasPrefix(file.Name, "__MACOSX") {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(file.Name), ".xlsx") {
			continue
		}

		payload, err := parseXlsxFile(file)
		if err != nil {
			return nil, err
		}
		payloads = append(payloads, payload)
	}

	if len(payloads) == 0 {
		return nil, errors.New("no xlsx files found in zip")
	}

	return payloads, nil
}

func parseXlsxFile(file *zip.File) (AuctionPayload, error) {
	if file == nil {
		return AuctionPayload{}, errors.New("xlsx file is nil")
	}

	reader, err := file.Open()
	if err != nil {
		return AuctionPayload{}, fmt.Errorf("open xlsx file: %w", err)
	}

	content, readErr := io.ReadAll(reader)
	closeErr := reader.Close()
	if readErr != nil {
		return AuctionPayload{}, fmt.Errorf("read xlsx file: %w", readErr)
	}
	if closeErr != nil {
		return AuctionPayload{}, fmt.Errorf("close xlsx file: %w", closeErr)
	}

	workbook, err := excelize.OpenReader(bytes.NewReader(content))
	if err != nil {
		return AuctionPayload{}, fmt.Errorf("open workbook: %w", err)
	}

	sheetName, rows, err := selectSheetRows(workbook)
	if err != nil {
		closeErr := workbook.Close()
		if closeErr != nil {
			return AuctionPayload{}, fmt.Errorf("close workbook: %w", closeErr)
		}
		return AuctionPayload{}, err
	}
	_ = sheetName

	participants, err := extractParticipants(rows)
	if err != nil {
		closeErr := workbook.Close()
		if closeErr != nil {
			return AuctionPayload{}, fmt.Errorf("close workbook: %w", closeErr)
		}
		return AuctionPayload{}, err
	}

	headerIndex, headerRow, regionIndex, techIndex, err := findHeaderRow(rows)
	if err != nil {
		closeErr := workbook.Close()
		if closeErr != nil {
			return AuctionPayload{}, fmt.Errorf("close workbook: %w", closeErr)
		}
		return AuctionPayload{}, err
	}

	dataRows := extractDataRows(rows, headerIndex+1, len(headerRow), regionIndex, techIndex)
	if len(dataRows) == 0 {
		closeErr := workbook.Close()
		if closeErr != nil {
			return AuctionPayload{}, fmt.Errorf("close workbook: %w", closeErr)
		}
		return AuctionPayload{}, errors.New("no data rows found after header")
	}

	if closeErr := workbook.Close(); closeErr != nil {
		return AuctionPayload{}, fmt.Errorf("close workbook: %w", closeErr)
	}

	return AuctionPayload{
		SourceFile:   file.Name,
		Participants: participants,
		Headers:      headerRow,
		Rows:         dataRows,
	}, nil
}

func selectSheetRows(workbook *excelize.File) (string, [][]string, error) {
	sheets := workbook.GetSheetList()
	if len(sheets) == 0 {
		return "", nil, errors.New("workbook has no sheets")
	}

	firstRows, err := workbook.GetRows(sheets[0])
	if err != nil {
		return "", nil, fmt.Errorf("get rows for %s: %w", sheets[0], err)
	}
	if containsAggregatedTitle(firstRows) {
		return sheets[0], firstRows, nil
	}

	for _, sheet := range sheets[1:] {
		rows, err := workbook.GetRows(sheet)
		if err != nil {
			return "", nil, fmt.Errorf("get rows for %s: %w", sheet, err)
		}
		if containsAggregatedTitle(rows) {
			return sheet, rows, nil
		}
	}

	return sheets[0], firstRows, nil
}

func containsAggregatedTitle(rows [][]string) bool {
	for _, row := range rows {
		for _, cell := range row {
			if strings.Contains(strings.ToLower(cell), aggregatedTitleMarker) {
				return true
			}
		}
	}
	return false
}

func extractParticipants(rows [][]string) (int, error) {
	for _, row := range rows {
		for i, cell := range row {
			if strings.Contains(strings.ToLower(cell), participantsLabelMatch) {
				if i+1 >= len(row) {
					return 0, errors.New("participants value is missing")
				}
				value := strings.TrimSpace(row[i+1])
				if value == "" {
					return 0, errors.New("participants value is empty")
				}
				participants, err := parseInt(value)
				if err != nil {
					return 0, err
				}
				return participants, nil
			}
		}
	}

	return 0, errors.New("participants row not found")
}

func findHeaderRow(rows [][]string) (int, []string, int, int, error) {
	for index, row := range rows {
		regionIndex := -1
		techIndex := -1
		for i, cell := range row {
			cellLower := strings.ToLower(cell)
			if regionIndex == -1 && strings.Contains(cellLower, "region") {
				regionIndex = i
			}
			if techIndex == -1 && strings.Contains(cellLower, "technology") {
				techIndex = i
			}
		}
		if regionIndex != -1 && techIndex != -1 {
			return index, row, regionIndex, techIndex, nil
		}
	}

	return 0, nil, -1, -1, errors.New("header row not found")
}

func extractDataRows(rows [][]string, startIndex int, headerLen int, regionIndex int, techIndex int) [][]string {
	if startIndex < 0 || startIndex >= len(rows) {
		return [][]string{}
	}

	var data [][]string
	started := false
	for _, row := range rows[startIndex:] {
		normalized := normalizeRow(row, headerLen)
		if rowIsEmpty(normalized) {
			if started {
				break
			}
			continue
		}

		if regionIndex >= len(normalized) || techIndex >= len(normalized) {
			if started {
				break
			}
			continue
		}

		region := strings.TrimSpace(normalized[regionIndex])
		technology := strings.TrimSpace(normalized[techIndex])
		if region == "" || technology == "" {
			if started {
				break
			}
			continue
		}

		data = append(data, normalized)
		started = true
	}

	return data
}

func normalizeRow(row []string, length int) []string {
	if len(row) >= length {
		return row
	}
	normalized := make([]string, length)
	copy(normalized, row)
	return normalized
}

func rowIsEmpty(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func parseInt(value string) (int, error) {
	cleaned := strings.ReplaceAll(value, " ", "")
	parsed, err := strconv.Atoi(cleaned)
	if err != nil {
		return 0, fmt.Errorf("parse int %q: %w", value, err)
	}
	return parsed, nil
}
