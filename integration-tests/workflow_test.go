package integration_test

import (
	"testing"

	"github.com/mfenderov/most-active-cookie/src/cookie"
	"github.com/mfenderov/most-active-cookie/src/parser"

	"github.com/stretchr/testify/assert"
)

// TestWorkflowScenarios tests various end-to-end workflow scenarios
func TestWorkflowScenarios(t *testing.T) {
	tests := []struct {
		name       string
		filename   string
		targetDate string
		expected   []string // empty slice = expect empty, specific values = exact match
	}{
		{
			name:       "sample data workflow",
			filename:   "./test-data/sample_cookie_log.csv",
			targetDate: "2018-12-09",
			expected:   []string{"AtY0laUfhglK3lC7"},
		},
		{
			name:       "massive dataset workflow",
			filename:   "./test-data/large_scale.csv",
			targetDate: "2018-12-15",
			expected:   []string{"sJJeKlQCJ580Nepu"},
		},
		{
			name:       "no matching date workflow",
			filename:   "./test-data/sample_cookie_log.csv",
			targetDate: "2020-01-01",
			expected:   []string{},
		},
		{
			name:       "multiple top cookies workflow",
			filename:   "./test-data/tied_cookies.csv",
			targetDate: "2018-12-09",
			expected:   []string{"CookieA", "CookieB"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create processor with injected parser dependency (following DIP)
			csvParser := parser.NewCSVParser()
			processor := cookie.NewProcessor(csvParser)

			// Process cookies using the processor directly
			cookies, err := processor.FindMostActiveCookies(tt.filename, tt.targetDate)
			assert.NoError(t, err, "Processing should succeed for %s", tt.name)

			assert.Equal(t, tt.expected, cookies, "Results should match expected values for %s", tt.name)
		})
	}
}

// TestErrorHandlingWorkflow tests error scenarios
func TestErrorHandlingWorkflow(t *testing.T) {
	csvParser := parser.NewCSVParser()
	processor := cookie.NewProcessor(csvParser)

	t.Run("NonExistentFile", func(t *testing.T) {
		_, err := processor.FindMostActiveCookies("nonexistent.csv", "2018-12-09")
		assert.Error(t, err, "Should return error for non-existent file")
	})

	t.Run("InvalidDateFormat", func(t *testing.T) {
		_, err := processor.FindMostActiveCookies("./test-data/sample_cookie_log.csv", "invalid-date")
		assert.Error(t, err, "Should return error for invalid date format")
	})
}
