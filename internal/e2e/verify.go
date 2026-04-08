package e2e

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// PollConfig controls polling behavior for async operations.
type PollConfig struct {
	Timeout  time.Duration
	Interval time.Duration
}

func DefaultPollConfig() PollConfig {
	return PollConfig{
		Timeout:  120 * time.Second,
		Interval: 5 * time.Second,
	}
}

// WaitForVersion polls trialHistory until at least expectedCount versions appear
// or the timeout is reached.
func WaitForVersion(client *SDRClient, trialAlias string, expectedCount int, cfg PollConfig) (*TrialVersionHistory, error) {
	deadline := time.Now().Add(cfg.Timeout)

	for time.Now().Before(deadline) {
		history, err := client.GetTrialHistory(trialAlias)
		if err != nil {
			// Trial not found yet is expected during processing.
			if !isNotFoundError(err) {
				return nil, fmt.Errorf("unexpected error polling trialHistory: %w", err)
			}
		} else if len(history.TrialVersionHistory) >= expectedCount {
			return history, nil
		}

		time.Sleep(cfg.Interval)
	}

	return nil, fmt.Errorf("timed out after %s waiting for %d version(s) of %s", cfg.Timeout, expectedCount, trialAlias)
}

// WaitForCleanup polls until a study no longer exists in Neptune.
func WaitForCleanup(client *SDRClient, trialAlias string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	query := fmt.Sprintf(`g.V().has('Study','name','%s').count()`, trialAlias)

	for time.Now().Before(deadline) {
		count, err := gremlinCount(client, query)
		if err != nil {
			return fmt.Errorf("cleanup check failed: %w", err)
		}
		if count == 0 {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("timed out waiting for cleanup of %s", trialAlias)
}

// VerifyStudyExists checks that exactly one Study node exists for the given trialAlias.
func VerifyStudyExists(client *SDRClient, trialAlias string) error {
	query := fmt.Sprintf(`g.V().has('Study','name','%s').count()`, trialAlias)
	count, err := gremlinCount(client, query)
	if err != nil {
		return fmt.Errorf("verifyStudyExists: %w", err)
	}
	if count != 1 {
		return fmt.Errorf("expected 1 Study node for %s, got %d", trialAlias, count)
	}
	return nil
}

// VerifyStudyGone checks that no Study node exists for the given trialAlias.
func VerifyStudyGone(client *SDRClient, trialAlias string) error {
	query := fmt.Sprintf(`g.V().has('Study','name','%s').count()`, trialAlias)
	count, err := gremlinCount(client, query)
	if err != nil {
		return fmt.Errorf("verifyStudyGone: %w", err)
	}
	if count != 0 {
		return fmt.Errorf("expected 0 Study nodes for %s, got %d", trialAlias, count)
	}
	return nil
}

// VerifyVersionCount checks that the expected number of StudyVersion nodes exist.
func VerifyVersionCount(client *SDRClient, trialAlias string, expected int) error {
	query := fmt.Sprintf(`g.V().has('Study','name','%s').out('has_version').hasLabel('StudyVersion').count()`, trialAlias)
	count, err := gremlinCount(client, query)
	if err != nil {
		return fmt.Errorf("verifyVersionCount: %w", err)
	}
	if count != expected {
		return fmt.Errorf("expected %d StudyVersion node(s) for %s, got %d", expected, trialAlias, count)
	}
	return nil
}

// VerifyNodesByLabel checks that at least minCount nodes of the given label
// exist in the subgraph reachable from the study node.
func VerifyNodesByLabel(client *SDRClient, trialAlias string, label string, minCount int) error {
	query := fmt.Sprintf(
		`g.V().has('Study','name','%s').repeat(out()).emit().hasLabel('%s').count()`,
		trialAlias, label,
	)
	count, err := gremlinCount(client, query)
	if err != nil {
		return fmt.Errorf("verifyNodesByLabel(%s): %w", label, err)
	}
	if count < minCount {
		return fmt.Errorf("expected at least %d %s node(s) for %s, got %d", minCount, label, trialAlias, count)
	}
	return nil
}

// gremlinCount executes a Gremlin count query and returns the integer result.
func gremlinCount(client *SDRClient, query string) (int, error) {
	raw, err := client.ExecuteGremlin(query)
	if err != nil {
		return 0, err
	}

	// The response from Neptune count queries can be:
	// - A JSON number: 5
	// - A JSON array: [5]
	// - A JSON string containing a number: "5"
	// - A JSON object with a result field: {"result": [5]}
	var count int

	// Try parsing as a plain number first.
	if err := json.Unmarshal(raw, &count); err == nil {
		return count, nil
	}

	// Try parsing as an array of numbers.
	var arr []int
	if err := json.Unmarshal(raw, &arr); err == nil && len(arr) > 0 {
		return arr[0], nil
	}

	// Try parsing as a nested result object.
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err == nil {
		if result, ok := obj["result"]; ok {
			var resultArr []int
			if err := json.Unmarshal(result, &resultArr); err == nil && len(resultArr) > 0 {
				return resultArr[0], nil
			}
		}
	}

	return 0, fmt.Errorf("unable to parse gremlin count from response: %s", string(raw))
}

func isNotFoundError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "not found") ||
		strings.Contains(msg, "Not Found") ||
		strings.Contains(msg, "does not exist")
}
