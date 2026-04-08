package e2e

import (
	"fmt"
	"time"
)

// RunnerConfig holds the configuration for the test runner.
type RunnerConfig struct {
	Scenario    string        // Run a specific scenario by name (empty = all).
	Timeout     time.Duration // Max wait for async operations per scenario.
	TrialPrefix string        // Trial alias prefix (e.g. "TST-E2").
	Verbose     bool          // Show detailed output.
	CleanupOnly bool          // Only clean up test data, don't run scenarios.
}

// Run executes the integration test scenarios and returns a report.
func Run(client *SDRClient, cfg RunnerConfig) Report {
	start := time.Now()

	scenarios := AllScenarios(cfg.TrialPrefix)

	pollCfg := PollConfig{
		Timeout:  cfg.Timeout,
		Interval: 5 * time.Second,
	}

	if cfg.CleanupOnly {
		return runCleanupOnly(client, scenarios)
	}

	// Filter to a single scenario if requested.
	if cfg.Scenario != "" {
		var filtered []Scenario
		for _, s := range scenarios {
			if s.Name == cfg.Scenario {
				filtered = append(filtered, s)
			}
		}
		if len(filtered) == 0 {
			fmt.Printf("  Unknown scenario %q. Available:\n", cfg.Scenario)
			for _, s := range scenarios {
				fmt.Printf("    - %s: %s\n", s.Name, s.Description)
			}
			return Report{
				Results:  []Result{{Name: cfg.Scenario, Error: fmt.Errorf("unknown scenario"), Passed: false}},
				Duration: time.Since(start),
			}
		}
		scenarios = filtered
	}

	fmt.Printf("  Running %d scenario(s) against prefix %s\n\n", len(scenarios), cfg.TrialPrefix)

	var results []Result
	for _, scenario := range scenarios {
		result := runScenario(client, scenario, pollCfg, cfg.Verbose)
		results = append(results, result)
	}

	return Report{
		Results:  results,
		Duration: time.Since(start),
	}
}

func runScenario(client *SDRClient, scenario Scenario, pollCfg PollConfig, verbose bool) Result {
	fmt.Printf("  ▶ %s: %s\n", scenario.Name, scenario.Description)

	// Pre-cleanup: remove any leftover data from previous failed runs.
	if scenario.TrialAlias != "" {
		cleanup(client, scenario.TrialAlias, verbose)
	}

	start := time.Now()
	err := scenario.Run(client, pollCfg)
	duration := time.Since(start)

	// Post-cleanup.
	if scenario.TrialAlias != "" {
		cleanup(client, scenario.TrialAlias, verbose)
	}

	if err != nil {
		fmt.Printf("    ✗ FAIL (%s): %s\n\n", formatDuration(duration), err)
		return Result{Name: scenario.Name, Passed: false, Duration: duration, Error: err}
	}

	fmt.Printf("    ✓ PASS (%s)\n\n", formatDuration(duration))
	return Result{Name: scenario.Name, Passed: true, Duration: duration}
}

func runCleanupOnly(client *SDRClient, scenarios []Scenario) Report {
	start := time.Now()
	fmt.Println("  Cleaning up test data...")

	for _, s := range scenarios {
		if s.TrialAlias == "" {
			continue
		}
		fmt.Printf("    cleaning %s...", s.TrialAlias)
		cleanup(client, s.TrialAlias, false)
		fmt.Println(" done")
	}

	fmt.Println("  Cleanup complete.")
	return Report{Duration: time.Since(start)}
}

func cleanup(client *SDRClient, trialAlias string, verbose bool) {
	if verbose {
		fmt.Printf("    [cleanup] deleting %s\n", trialAlias)
	}
	_, _ = client.DeleteAllVersions(trialAlias)
}
