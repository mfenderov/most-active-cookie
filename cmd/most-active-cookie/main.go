package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	cookie "github.com/mfenderov/most-active-cookie"
	"github.com/mfenderov/most-active-cookie/src/cli"
)

func main() {
	config := parseAndValidateFlags()
	configureLogging(config.Verbosity)
	cookies := processCookies(config)
	outputResults(cookies)
}

func parseAndValidateFlags() *cli.Config {
	config, err := cli.ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}
	return config
}

func processCookies(config *cli.Config) []string {
	slog.Info("starting cookie processing", "filename", config.Filename, "targetDate", config.TargetDate)

	// Use the library API instead of direct internal imports
	cookies, err := cookie.FindMostActiveCookies(config.Filename, config.TargetDate)
	if err != nil {
		slog.Error("processing failed", "error", err, "filename", config.Filename)
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	slog.Info("cookie processing completed successfully", "cookieCount", len(cookies))
	return cookies
}

func outputResults(cookies []string) {
	if len(cookies) == 0 {
		slog.Debug("no cookies found for target date - exiting quietly")
		os.Exit(0)
	}

	for _, c := range cookies {
		fmt.Println(c)
	}
}

func configureLogging(verbosity int) {
	var level slog.Level
	switch verbosity {
	case 0:
		level = slog.LevelWarn // Default: quiet
	case 1:
		level = slog.LevelInfo // -v: verbose
	default:
		level = slog.LevelDebug // -vv: debug
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)
}
