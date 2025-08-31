package parser

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/mfenderov/most-active-cookie/src/cookie"

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

	parser := NewCSVParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := createTempCSVFile(t, tt.csvContent)

			var entries []cookie.LogEntry
			err := parser.StreamFile(filename, func(entry cookie.LogEntry) error {
				entries = append(entries, entry)
				return nil
			})

			if tt.expectError {
				assert.Error(t, err, "expected error but got none")
				if tt.errorContains != "" && err != nil {
					assert.Contains(t, err.Error(), tt.errorContains, "error should contain expected substring")
				}
				return
			}

			assert.NoError(t, err, "unexpected error")
			assert.Equal(t, tt.expectedCount, len(entries), "entry count mismatch")
		})
	}
}

func TestCSVParser_StreamFile_NonExistentFile(t *testing.T) {
	parser := NewCSVParser()
	err := parser.StreamFile("nonexistent_file.csv", func(_ cookie.LogEntry) error {
		return nil
	})

	assert.Error(t, err, "expected error for non-existent file")
	assert.Contains(t, err.Error(), "failed to open file", "error should mention file opening failure")
}

func TestCSVParser_StreamFile_ProcessorError(t *testing.T) {
	validCSV := `cookie,timestamp
AtY0laUfhglK3lC7,2018-12-09T14:19:00+00:00`

	parser := NewCSVParser()
	filename := createTempCSVFile(t, validCSV)

	processorError := errors.New("processor failed")
	err := parser.StreamFile(filename, func(_ cookie.LogEntry) error {
		return processorError
	})

	assert.Error(t, err, "expected processor error to propagate")
	assert.Contains(t, err.Error(), "processing error", "error should mention processing failure")
}

func TestCSVParser_ParseLine(t *testing.T) {
	parser := NewCSVParser()

	tests := []struct {
		name           string
		line           string
		expectError    bool
		expectedCookie string
		errorContains  string
	}{
		{
			name:           "valid line",
			line:           "AtY0laUfhglK3lC7,2018-12-09T14:19:00+00:00",
			expectError:    false,
			expectedCookie: "AtY0laUfhglK3lC7",
		},
		{
			name:          "empty cookie",
			line:          ",2018-12-09T14:19:00+00:00",
			expectError:   true,
			errorContains: "empty cookie ID",
		},
		{
			name:          "empty timestamp",
			line:          "AtY0laUfhglK3lC7,",
			expectError:   true,
			errorContains: "empty timestamp",
		},
		{
			name:          "invalid CSV format - too many columns",
			line:          "AtY0laUfhglK3lC7,2018-12-09T14:19:00+00:00,extra",
			expectError:   true,
			errorContains: "invalid CSV format",
		},
		{
			name:          "invalid CSV format - too few columns",
			line:          "AtY0laUfhglK3lC7",
			expectError:   true,
			errorContains: "invalid CSV format",
		},
		{
			name:          "invalid timestamp format",
			line:          "AtY0laUfhglK3lC7,2018-12-09",
			expectError:   true,
			errorContains: "invalid timestamp format",
		},
		{
			name:           "whitespace handling",
			line:           " AtY0laUfhglK3lC7 , 2018-12-09T14:19:00+00:00 ",
			expectError:    false,
			expectedCookie: "AtY0laUfhglK3lC7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := parser.parseLine(tt.line)

			if tt.expectError {
				assert.Error(t, err, "expected error but got none")
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains, "error should contain expected substring")
				}
				return
			}

			assert.NoError(t, err, "unexpected error")
			assert.Equal(t, tt.expectedCookie, entry.Cookie, "cookie mismatch")
		})
	}
}

func TestIsValidHeader(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected bool
	}{
		{
			name:     "valid header",
			header:   "cookie,timestamp",
			expected: true,
		},
		{
			name:     "valid header with case variations",
			header:   "Cookie,Timestamp",
			expected: true,
		},
		{
			name:     "valid header with whitespace",
			header:   " cookie,timestamp ",
			expected: true,
		},
		{
			name:     "invalid header",
			header:   "invalid,header",
			expected: false,
		},
		{
			name:     "empty header",
			header:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidHeader(tt.header)
			assert.Equal(t, tt.expected, result, "header validation result mismatch")
		})
	}
}
