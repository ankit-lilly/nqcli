package e2e

import (
	"fmt"
	"strings"
	"time"
)

// Result holds the outcome of a single scenario run.
type Result struct {
	Name     string
	Passed   bool
	Duration time.Duration
	Error    error
	Skipped  bool
}

// Report holds the full test run results.
type Report struct {
	Results  []Result
	Duration time.Duration
}

// PrintReport prints a formatted summary of the test run to stdout.
func PrintReport(report Report) {
	passed, failed, skipped := 0, 0, 0

	fmt.Println()
	fmt.Println(strings.Repeat("─", 70))
	fmt.Println("  E2E Test Results")
	fmt.Println(strings.Repeat("─", 70))

	for _, r := range report.Results {
		switch {
		case r.Skipped:
			skipped++
			fmt.Printf("  SKIP  %-40s\n", r.Name)
		case r.Passed:
			passed++
			fmt.Printf("  PASS  %-40s  %s\n", r.Name, formatDuration(r.Duration))
		default:
			failed++
			fmt.Printf("  FAIL  %-40s  %s\n", r.Name, formatDuration(r.Duration))
			fmt.Printf("        error: %s\n", r.Error)
		}
	}

	fmt.Println(strings.Repeat("─", 70))
	fmt.Printf("  Total: %d  |  Passed: %d  |  Failed: %d  |  Skipped: %d  |  %s\n",
		len(report.Results), passed, failed, skipped, formatDuration(report.Duration))
	fmt.Println(strings.Repeat("─", 70))

	if failed > 0 {
		fmt.Println("\n  RESULT: FAIL")
	} else {
		fmt.Println("\n  RESULT: PASS")
	}
	fmt.Println()
}

// HasFailures returns true if any scenario failed.
func (r Report) HasFailures() bool {
	for _, res := range r.Results {
		if !res.Passed && !res.Skipped {
			return true
		}
	}
	return false
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
