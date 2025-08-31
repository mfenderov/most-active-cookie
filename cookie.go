// Package cookie provides functionality to analyze cookie log files
// and find the most active cookies for specific dates.
package cookie

import (
	"github.com/mfenderov/most-active-cookie/src/cookie"
	"github.com/mfenderov/most-active-cookie/src/parser"
)

// FindMostActiveCookies analyzes a CSV log file and returns the most active cookie(s)
// for the specified date.
//
// The filename parameter should point to a CSV file with the format:
//
//	cookie,timestamp
//	AtY0laUfhglK3lC7,2018-12-09T14:19:00+00:00
//	SAZuXPGUrfbcn5UA,2018-12-09T10:13:00+00:00
//
// The targetDate parameter should be in YYYY-MM-DD format (UTC timezone).
//
// Returns a sorted slice of cookie names that appeared most frequently on the target date.
// If multiple cookies tie for most active, all are returned.
// Returns an empty slice if no cookies are found for the target date.
//
// Example usage:
//
//	cookies, err := cookie.FindMostActiveCookies("cookie_log.csv", "2018-12-09")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, cookie := range cookies {
//	    fmt.Println(cookie)
//	}
func FindMostActiveCookies(filename, targetDate string) ([]string, error) {
	csvParser := parser.NewCSVParser()
	processor := cookie.NewProcessor(csvParser)
	return processor.FindMostActiveCookies(filename, targetDate)
}
