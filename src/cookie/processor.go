package cookie

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

type LogEntry struct {
	Cookie    string
	Timestamp string
}

type EntryProcessor func(entry LogEntry) error

var ErrPastTargetDate = errors.New("past the target date")

type FileParser interface {
	StreamFile(filename string, processor EntryProcessor) error
}

type Processor struct {
	parser FileParser
}

func NewProcessor(parser FileParser) *Processor {
	return &Processor{
		parser: parser,
	}
}

func (p *Processor) FindMostActiveCookies(filename, targetDate string) ([]string, error) {
	if filename == "" {
		return nil, fmt.Errorf("filename cannot be empty")
	}
	err := validateDate(targetDate)
	if err != nil {
		return []string{}, fmt.Errorf("invalid target date: %w", err)
	}

	cookieCounts := make(map[string]int)
	err = p.parser.StreamFile(filename, processLogEntry(targetDate, cookieCounts))
	if err != nil && !errors.Is(err, ErrPastTargetDate) {
		return nil, fmt.Errorf("failed to stream file: %w", err)
	}

	if len(cookieCounts) == 0 {
		return []string{}, nil
	}

	var mostActiveCookies []string
	maxCount := 0

	for cookie, count := range cookieCounts {
		if count > maxCount {
			maxCount = count
			mostActiveCookies = []string{cookie}
		} else if count == maxCount {
			mostActiveCookies = append(mostActiveCookies, cookie)
		}
	}

	sort.Strings(mostActiveCookies)

	return mostActiveCookies, nil
}

func validateDate(targetDate string) error {
	if targetDate == "" {
		return fmt.Errorf("the target date cannot be empty")
	}

	if _, err := time.Parse("2006-01-02", targetDate); err != nil {
		return fmt.Errorf("invalid target date: expected YYYY-MM-DD, got '%s'", targetDate)
	}
	return nil
}

func processLogEntry(targetDate string, cookieCounts map[string]int) func(entry LogEntry) error {
	return func(entry LogEntry) error {
		timestamp := entry.Timestamp
		if len(timestamp) < 10 {
			return fmt.Errorf("timestamp too short: %s", timestamp)
		}

		entryDate := timestamp[:10]

		if entryDate > targetDate {
			return ErrPastTargetDate
		}

		if entryDate == targetDate {
			cookieCounts[entry.Cookie]++
		}

		return nil
	}
}
