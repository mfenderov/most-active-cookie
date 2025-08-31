package cli

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	// Create a temporary test file for validation
	tmpFile, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	tests := []struct {
		name          string
		args          []string
		expected      *Config
		expectError   bool
		errorContains string
	}{
		{
			name: "valid arguments",
			args: []string{"-f", tmpFile.Name(), "-d", "2018-12-09"},
			expected: &Config{
				Filename:   tmpFile.Name(),
				TargetDate: "2018-12-09",
			},
			expectError: false,
		},
		{
			name:          "missing filename",
			args:          []string{"-d", "2018-12-09"},
			expectError:   true,
			errorContains: "filename is required",
		},
		{
			name:          "missing date",
			args:          []string{"-f", tmpFile.Name()},
			expectError:   true,
			errorContains: "target date is required",
		},
		{
			name:          "non-existent file",
			args:          []string{"-f", "nonexistent.csv", "-d", "2018-12-09"},
			expectError:   true,
			errorContains: "file does not exist",
		},
		{
			name:          "no arguments",
			args:          []string{},
			expectError:   true,
			errorContains: "filename is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flag package state
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			// Save original args and restore them after test
			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			// Set test arguments
			os.Args = append([]string{"test"}, tt.args...)

			config, err := ParseFlags()

			if tt.expectError {
				assert.Error(t, err, "expected error but got none")
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains, "error should contain expected substring")
				}
				return
			}

			assert.NoError(t, err, "unexpected error")
			assert.Equal(t, tt.expected.Filename, config.Filename, "filename mismatch")
			assert.Equal(t, tt.expected.TargetDate, config.TargetDate, "target date mismatch")
		})
	}
}

func TestValidateConfig(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	tests := []struct {
		name          string
		config        *Config
		expectError   bool
		errorContains string
	}{
		{
			name: "valid config",
			config: &Config{
				Filename:   tmpFile.Name(),
				TargetDate: "2018-12-09",
			},
			expectError: false,
		},
		{
			name: "empty filename",
			config: &Config{
				Filename:   "",
				TargetDate: "2018-12-09",
			},
			expectError:   true,
			errorContains: "filename is required",
		},
		{
			name: "empty target date",
			config: &Config{
				Filename:   tmpFile.Name(),
				TargetDate: "",
			},
			expectError:   true,
			errorContains: "target date is required",
		},
		{
			name: "non-existent file",
			config: &Config{
				Filename:   "nonexistent.csv",
				TargetDate: "2018-12-09",
			},
			expectError:   true,
			errorContains: "file does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)

			if tt.expectError {
				assert.Error(t, err, "expected error but got none")
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains, "error should contain expected substring")
				}
				return
			}

			assert.NoError(t, err, "unexpected error")
		})
	}
}
