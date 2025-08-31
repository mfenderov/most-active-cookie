package cookie

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcessor_FindMostActiveCookies(t *testing.T) {
	tests := []struct {
		name           string
		entries        []LogEntry
		targetDate     string
		expectedResult []string
		expectError    bool
		errorContains  string
	}{
		{
			name: "single most active cookie",
			entries: []LogEntry{
				{Cookie: "A", Timestamp: "2018-12-09T14:19:00+00:00"},
				{Cookie: "A", Timestamp: "2018-12-09T06:19:00+00:00"},
				{Cookie: "B", Timestamp: "2018-12-09T10:13:00+00:00"},
			},
			targetDate:     "2018-12-09",
			expectedResult: []string{"A"},
		},
		{
			name: "multiple most active cookies (tie)",
			entries: []LogEntry{
				{Cookie: "A", Timestamp: "2018-12-09T14:19:00+00:00"},
				{Cookie: "B", Timestamp: "2018-12-09T10:13:00+00:00"},
				{Cookie: "C", Timestamp: "2018-12-09T07:25:00+00:00"},
			},
			targetDate:     "2018-12-09",
			expectedResult: []string{"A", "B", "C"},
		},
		{
			name: "no entries for target date",
			entries: []LogEntry{
				{Cookie: "A", Timestamp: "2018-12-08T14:19:00+00:00"},
				{Cookie: "B", Timestamp: "2018-12-07T10:13:00+00:00"},
			},
			targetDate:     "2018-12-09",
			expectedResult: []string{},
		},
		{
			name: "entries from multiple dates, filter correctly",
			entries: []LogEntry{
				{Cookie: "A", Timestamp: "2018-12-09T14:19:00+00:00"},
				{Cookie: "A", Timestamp: "2018-12-09T06:19:00+00:00"},
				{Cookie: "B", Timestamp: "2018-12-08T22:03:00+00:00"},
				{Cookie: "B", Timestamp: "2018-12-08T21:30:00+00:00"},
				{Cookie: "C", Timestamp: "2018-12-09T07:25:00+00:00"},
			},
			targetDate:     "2018-12-09",
			expectedResult: []string{"A"},
		},
		{
			name: "timezone handling",
			entries: []LogEntry{
				{Cookie: "A", Timestamp: "2018-12-09T14:19:00+00:00"},
				{Cookie: "B", Timestamp: "2018-12-09T09:19:00-05:00"},
			},
			targetDate:     "2018-12-09",
			expectedResult: []string{"A", "B"},
		},
		{
			name:          "invalid target date format",
			targetDate:    "invalid-date",
			expectError:   true,
			errorContains: "invalid target date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockParser := NewMockFileParser(t)
			if !tt.expectError || tt.errorContains != "invalid target date" {
				mockParser.EXPECT().StreamFile("test.csv", mock.AnythingOfType("EntryProcessor")).Run(func(_ string, processor EntryProcessor) {
					for _, entry := range tt.entries {
						processor(entry)
					}
				}).Return(nil)
			}
			processor := NewProcessor(mockParser)

			cookies, err := processor.FindMostActiveCookies("test.csv", tt.targetDate)

			if tt.expectError {
				assert.Error(t, err, "expected error but got none")
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains, "error should contain expected substring")
				}
				return
			}

			assert.NoError(t, err, "unexpected error")
			assert.Equal(t, tt.expectedResult, cookies, "result mismatch")
		})
	}
}

func TestProcessor_FindMostActiveCookies_ParserError(t *testing.T) {
	mockParser := NewMockFileParser(t)
	mockParser.EXPECT().StreamFile("test.csv", mock.AnythingOfType("EntryProcessor")).Return(errors.New("parser error"))
	processor := NewProcessor(mockParser)

	_, err := processor.FindMostActiveCookies("test.csv", "2018-12-09")

	assert.Error(t, err, "expected error from parser")
	assert.Contains(t, err.Error(), "failed to stream file", "error should mention streaming failure")
}
