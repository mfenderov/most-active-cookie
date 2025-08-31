package cli

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	Filename   string
	TargetDate string
	Verbosity  int // 0=WARN, 1=INFO, 2=DEBUG
}

func ParseFlags() (*Config, error) {
	var config Config

	flag.StringVar(&config.Filename, "f", "", "Cookie log file to process (required)")
	flag.StringVar(&config.TargetDate, "d", "", "Target date in YYYY-MM-DD format (required)")

	var verbose bool
	var veryVerbose bool
	flag.BoolVar(&verbose, "v", false, "Verbose output (INFO level)")
	flag.BoolVar(&veryVerbose, "vv", false, "Very verbose output (DEBUG level)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -f <filename> -d <date> [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nFind the most active cookie(s) for a specific date.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -f cookie_log.csv -d 2018-12-09\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -f cookie_log.csv -d 2018-12-09 -v      # verbose output\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -f cookie_log.csv -d 2018-12-09 -vv     # debug output\n", os.Args[0])
	}

	flag.Parse()

	if veryVerbose {
		config.Verbosity = 2
	} else if verbose {
		config.Verbosity = 1
	} else {
		config.Verbosity = 0
	}

	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func validateConfig(config *Config) error {
	if config.Filename == "" {
		return fmt.Errorf("a filename is required (use -f flag)")
	}

	if config.TargetDate == "" {
		return fmt.Errorf("a target date is required (use -d flag)")
	}

	if _, err := os.Stat(config.Filename); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", config.Filename)
	}

	return nil
}
