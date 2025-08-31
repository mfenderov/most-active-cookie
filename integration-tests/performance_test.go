package integration_test

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/mfenderov/most-active-cookie/src/cookie"
	"github.com/mfenderov/most-active-cookie/src/parser"

	"github.com/stretchr/testify/assert"
)

// TestPermanentMassiveDataset validates performance with the committed 100K entry dataset
func TestPermanentMassiveDataset(t *testing.T) {
	filename := "./test-data/large_scale.csv"

	// Measure initial memory
	var memBefore runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	// Process the permanent massive dataset using streaming
	t.Log("Processing permanent 100K entry dataset...")
	start := time.Now()

	// Create processor with CSV parser
	csvParser := parser.NewCSVParser()
	processor := cookie.NewProcessor(csvParser)

	cookies, err := processor.FindMostActiveCookies(filename, "2018-12-15")

	elapsed := time.Since(start)

	// Measure final memory
	var memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memAfter)

	// Validate results
	assert.NoError(t, err, "Processing permanent dataset should succeed")
	assert.NotEmpty(t, cookies, "Should return results from 100K entry dataset")

	// Performance metrics
	memUsedMB := float64(memAfter.TotalAlloc-memBefore.TotalAlloc) / 1024 / 1024
	entriesPerSec := 100000.0 / elapsed.Seconds()

	t.Logf("✓ Results: %v", cookies)
	t.Logf("✓ Performance: %.0f entries/sec", entriesPerSec)
	t.Logf("✓ Duration: %v", elapsed)
	t.Logf("✓ Memory: %.2f MB", memUsedMB)

	// Validate streaming performance expectations
	assert.LessOrEqual(t, elapsed, 100*time.Millisecond, "Processing 100K entries should take under 100ms with streaming")
	assert.LessOrEqual(t, memUsedMB, 15.0, "Memory usage should not exceed 15MB limit for streaming")
	assert.GreaterOrEqual(t, entriesPerSec, 1000000.0, "Processing rate should exceed 1M entries/sec with streaming")

}

// TestMassiveScaleProcessing demonstrates streaming can handle production-scale datasets
func TestMassiveScaleProcessing(t *testing.T) {
	scales := []struct {
		name        string
		entries     int
		maxDuration time.Duration
		maxMemoryMB float64
	}{
		{"Small", 10000, 50 * time.Millisecond, 10},
		{"Medium", 100000, 100 * time.Millisecond, 15},
		{"Large", 1000000, 500 * time.Millisecond, 100},
		{"Massive", 5000000, 2 * time.Second, 400},
	}

	for _, scale := range scales {
		t.Run(scale.name, func(t *testing.T) {
			filename := fmt.Sprintf("perf_test_%d.csv", scale.entries)
			defer os.Remove(filename)

			t.Logf("Generating %d entries...", scale.entries)
			generateOptimizedPerfData(filename, scale.entries)

			// Measure initial memory
			var memBefore runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&memBefore)

			// Process the massive dataset using streaming
			t.Logf("Processing %d entries with streaming...", scale.entries)
			start := time.Now()

			// Create processor with CSV parser
			csvParser := parser.NewCSVParser()
			processor := cookie.NewProcessor(csvParser)

			cookies, err := processor.FindMostActiveCookies(filename, "2018-12-15")

			elapsed := time.Since(start)

			// Measure final memory
			var memAfter runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&memAfter)

			// Validate results
			assert.NoError(t, err, "Processing should succeed")
			assert.NotEmpty(t, cookies, "Should return results for %d entries", scale.entries)

			// Performance validation
			memUsedMB := float64(memAfter.TotalAlloc-memBefore.TotalAlloc) / 1024 / 1024
			entriesPerSec := float64(scale.entries) / elapsed.Seconds()

			t.Logf("Results: %v", cookies)
			t.Logf("Performance: %.2f entries/sec", entriesPerSec)
			t.Logf("Duration: %v (max: %v)", elapsed, scale.maxDuration)
			t.Logf("Memory: %.2f MB (max: %.2f MB)", memUsedMB, scale.maxMemoryMB)

			// Assert performance requirements
			assert.LessOrEqual(t, elapsed, scale.maxDuration, "Processing took too long")
			assert.LessOrEqual(t, memUsedMB, scale.maxMemoryMB, "Memory usage exceeded limit")

			// Validate high throughput (streaming should be fast)
			minThroughput := 1000000.0 // 1M+ entries/sec expected with streaming
			assert.GreaterOrEqual(t, entriesPerSec, minThroughput, "Processing rate too slow (expected >%.0f/sec)", minThroughput)

		})
	}
}

// BenchmarkStreamingThroughput benchmarks streaming approach throughput
func BenchmarkStreamingThroughput(b *testing.B) {
	filename := "./test-data/large_scale.csv"
	targetDate := "2018-12-15"

	// Create processor with CSV parser
	csvParser := parser.NewCSVParser()
	processor := cookie.NewProcessor(csvParser)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := processor.FindMostActiveCookies(filename, targetDate)
		assert.NoError(b, err, "Benchmark iteration should succeed")
	}
}

// generateOptimizedPerfData creates deterministic test data with minimal allocations
func generateOptimizedPerfData(filename string, entries int) {
	file, err := os.Create(filename)
	if err != nil {
		panic(fmt.Sprintf("Failed to create performance test file: %v", err))
	}
	defer file.Close()

	// Use buffered writer for performance
	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Write header
	writer.WriteString("cookie,timestamp\n")

	// Pre-calculate pattern data to avoid repeated formatting
	cookies := []string{"TopCookie", "SecondCookie", "ThirdCookie"}
	baseTime := time.Date(2018, 12, 15, 0, 0, 0, 0, time.UTC)

	for i := 0; i < entries; i++ {
		// Simple deterministic pattern - no complex modulo math
		cookie := cookies[i%3]

		// Distribute entries throughout the day
		timestamp := baseTime.Add(time.Second * time.Duration(i%(24*60*60)))

		// Write with minimal allocations
		writer.WriteString(cookie)
		writer.WriteByte(',')
		writer.WriteString(timestamp.Format("2006-01-02T15:04:05Z07:00"))
		writer.WriteByte('\n')
	}
}
