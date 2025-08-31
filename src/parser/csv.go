package parser

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/mfenderov/most-active-cookie/src/cookie"
)

const (
	expectedColumns = 2
)

type CSVParser struct{}

func NewCSVParser() *CSVParser {
	return &CSVParser{}
}

func (p *CSVParser) StreamFile(filename string, processor cookie.EntryProcessor) error {
	file, err := os.Open(filename) //nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	entriesProcessed := 0

	if scanner.Scan() {
		lineNum++
		header := scanner.Text()
		if !isValidHeader(header) {
			return fmt.Errorf("invalid header format at line %d: expected 'cookie,timestamp', got '%s'", lineNum, header)
		}
	}

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		entry, err := p.parseLine(line)
		if err != nil {
			return fmt.Errorf("error parsing line %d: %w", lineNum, err)
		}

		if err := processor(entry); err != nil {
			if errors.Is(err, cookie.ErrPastTargetDate) {
				break
			}
			return fmt.Errorf("processing error at line %d: %w", lineNum, err)
		}

		entriesProcessed++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file %s: %w", filename, err)
	}

	if entriesProcessed == 0 {
		return fmt.Errorf("no valid entries found in file %s", filename)
	}

	slog.Info("successfully streamed CSV file", "filename", filename, "entriesProcessed", entriesProcessed, "linesProcessed", lineNum)
	return nil
}

func (p *CSVParser) parseLine(line string) (cookie.LogEntry, error) {
	parts := strings.Split(line, ",")
	if len(parts) != expectedColumns {
		return cookie.LogEntry{}, fmt.Errorf("invalid CSV format: expected %d columns, got %d", expectedColumns, len(parts))
	}

	cookieID := strings.TrimSpace(parts[0])
	timestampStr := strings.TrimSpace(parts[1])

	if cookieID == "" {
		return cookie.LogEntry{}, fmt.Errorf("empty cookie ID")
	}

	if timestampStr == "" {
		return cookie.LogEntry{}, fmt.Errorf("empty timestamp")
	}

	if len(timestampStr) < 10 || !strings.Contains(timestampStr, "T") {
		return cookie.LogEntry{}, fmt.Errorf("invalid timestamp format '%s': expected YYYY-MM-DDTHH:mm:ss format", timestampStr)
	}

	return cookie.LogEntry{
		Cookie:    cookieID,
		Timestamp: timestampStr,
	}, nil
}

func isValidHeader(header string) bool {
	expected := "cookie,timestamp"
	return strings.TrimSpace(strings.ToLower(header)) == expected
}
