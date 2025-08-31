package parser_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/mfenderov/most-active-cookie/src/cookie"
	"github.com/mfenderov/most-active-cookie/src/parser"

	"github.com/stretchr/testify/assert"
)

// createTempCSVFile creates a temporary CSV file with the given content and returns its path.
// Cleanup is handled automatically using t.Cleanup().
func createTempCSVFile(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "test_*.csv")
	assert.NoError(t, err, "failed to create temp file")

	// Schedule cleanup - will run after test completes
	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err, "failed to write temp file")

	err = tmpFile.Close()
	assert.NoError(t, err, "failed to close temp file")

	return tmpFile.Name()
}

func TestCSVParser_StreamFile(t *testing.T) {
	validCSV := `cookie,timestamp
AtY0laUfhglK3lC7,2018-12-09T14:19:00+00:00
SAZuXPGUrfbcn5UA,2018-12-09T10:13:00+00:00`

	invalidHeaderCSV := `invalid,header
AtY0laUfhglK3lC7,2018-12-09T14:19:00+00:00`

	invalidTimestampCSV := `cookie,timestamp
AtY0laUfhglK3lC7,invalid-timestamp`

	emptyFileCSV := `cookie,timestamp`

	// Edge case: BOM (Byte Order Mark) at start of file
	bomCSV := "\xEF\xBB\xBFcookie,timestamp\nAtY0laUfhglK3lC7,2018-12-09T14:19:00+00:00"

	// Edge case: CRLF line endings (Windows)
	crlfCSV := "cookie,timestamp\r\nAtY0laUfhglK3lC7,2018-12-09T14:19:00+00:00\r\nSAZuXPGUrfbcn5UA,2018-12-09T10:13:00+00:00"

	// Edge case: CR line endings (old Mac)
	crCSV := "cookie,timestamp\rAtY0laUfhglK3lC7,2018-12-09T14:19:00+00:00\rSAZuXPGUrfbcn5UA,2018-12-09T10:13:00+00:00"

	// Edge case: Unicode characters in cookie names
	unicodeCSV := `cookie,timestamp
café🍪μπισκότο,2018-12-09T14:19:00+00:00
тест_куки,2018-12-09T10:13:00+00:00`

	// Edge case: Quoted fields with commas and quotes
	quotedCSV := `cookie,timestamp
"cookie,with,commas",2018-12-09T14:19:00+00:00
"cookie""with""quotes",2018-12-09T10:13:00+00:00`

	// Edge case: Very long cookie name (1000 chars)
	longCookieName := string(make([]byte, 1000))
	for i := range longCookieName {
		longCookieName = longCookieName[:i] + "A" + longCookieName[i+1:]
	}
	longFieldCSV := fmt.Sprintf("cookie,timestamp\n%s,2018-12-09T14:19:00+00:00", longCookieName)

	// Edge case: Malformed UTF-8 (invalid byte sequence)
	malformedUTF8CSV := "cookie,timestamp\n\xFF\xFE\x00invalid,2018-12-09T14:19:00+00:00"

	tests := []struct {
		name          string
		csvContent    string
		expectedCount int
		expectError   bool
		errorContains string
	}{
		{
			name:          "valid CSV file",
			csvContent:    validCSV,
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "invalid header",
			csvContent:    invalidHeaderCSV,
			expectError:   true,
			errorContains: "invalid header format",
		},
		{
			name:          "invalid timestamp",
			csvContent:    invalidTimestampCSV,
			expectError:   true,
			errorContains: "invalid timestamp format",
		},
		{
			name:          "empty file with header only",
			csvContent:    emptyFileCSV,
			expectError:   true,
			errorContains: "no valid entries found",
		},
		{
			name:          "UTF-8 BOM handling",
			csvContent:    bomCSV,
			expectError:   true,
			errorContains: "invalid header format", // Current parser doesn't handle BOM
		},
		{
			name:          "CRLF line endings",
			csvContent:    crlfCSV,
			expectedCount: 2,
			expectError:   false, // Parser handles CRLF correctly
		},
		{
			name:          "CR line endings",
			csvContent:    crCSV,
			expectError:   true,
			errorContains: "invalid header format", // Current parser doesn't handle CR-only
		},
		{
			name:          "Unicode characters in cookie names",
			csvContent:    unicodeCSV,
			expectedCount: 2,
			expectError:   false, // Parser handles Unicode correctly
		},
		{
			name:          "quoted fields with commas and quotes",
			csvContent:    quotedCSV,
			expectError:   true,
			errorContains: "invalid CSV format", // Current parser doesn't handle CSV quoting
		},
		{
			name:          "very long cookie name",
			csvContent:    longFieldCSV,
			expectedCount: 1,
			expectError:   false, // Parser handles long fields correctly
		},
		{
			name:          "malformed UTF-8",
			csvContent:    malformedUTF8CSV,
			expectedCount: 1,
			expectError:   false, // Current parser accepts malformed UTF-8
		},
	}

	csvParser := parser.NewCSVParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := createTempCSVFile(t, tt.csvContent)

			var entries []cookie.LogEntry
			err := csvParser.StreamFile(filename, func(entry cookie.LogEntry) error {
				entries = append(entries, entry)
				return nil
			})

			if tt.expectError {
				assert.Error(t, err, "expected error but got none")
				if tt.errorContains != "" && err != nil {
					assert.ErrorContains(t, err, tt.errorContains, "error should contain expected substring")
				}
				return
			}

			assert.NoError(t, err, "unexpected error")
			assert.Equal(t, tt.expectedCount, len(entries), "entry count mismatch")
		})
	}
}

func TestCSVParser_StreamFile_NonExistentFile(t *testing.T) {
	csvParser := parser.NewCSVParser()
	err := csvParser.StreamFile("nonexistent_file.csv", func(_ cookie.LogEntry) error {
		return nil
	})

	assert.Error(t, err, "expected error for non-existent file")
	assert.ErrorContains(t, err, "failed to open file", "error should mention file opening failure")
}

func TestCSVParser_StreamFile_ProcessorError(t *testing.T) {
	validCSV := `cookie,timestamp
AtY0laUfhglK3lC7,2018-12-09T14:19:00+00:00`

	csvParser := parser.NewCSVParser()
	filename := createTempCSVFile(t, validCSV)

	processorError := errors.New("processor failed")
	err := csvParser.StreamFile(filename, func(_ cookie.LogEntry) error {
		return processorError
	})

	assert.Error(t, err, "expected processor error to propagate")
	assert.ErrorContains(t, err, "processing error", "error should mention processing failure")
}
