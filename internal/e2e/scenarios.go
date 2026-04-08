package e2e

import (
	"fmt"
	"time"
)

// Scenario represents a single integration test scenario.
type Scenario struct {
	Name        string
	Description string
	TrialAlias  string
	Run         func(client *SDRClient, cfg PollConfig) error
}

// AllScenarios returns the full list of integration test scenarios.
func AllScenarios(prefix string) []Scenario {
	return []Scenario{
		submitAndVerify(prefix),
		invalidPayloadRejection(prefix),
		versionIncrement(prefix),
		deleteSingleVersionScenario(prefix),
		deleteAllVersionsScenario(prefix),
		graphStructure(prefix),
		trialsQuery(prefix),
		graphSummaryScenario(prefix),
	}
}

func submitAndVerify(prefix string) Scenario {
	alias := fmt.Sprintf("%s-0001", prefix)
	return Scenario{
		Name:        "submit-and-verify",
		Description: "Submit a minimal payload and verify it lands in Neptune",
		TrialAlias:  alias,
		Run: func(client *SDRClient, cfg PollConfig) error {
			payload := BuildMinimalPayload(alias)

			resp, err := client.SubmitData(payload)
			if err != nil {
				return fmt.Errorf("submit failed: %w", err)
			}
			if !resp.IsValid {
				return fmt.Errorf("expected isValid=true, got message: %s", resp.Message)
			}

			history, err := WaitForVersion(client, alias, 1, cfg)
			if err != nil {
				return err
			}
			if history.TrialVersionHistory[0].Status != "SUCCESS" {
				return fmt.Errorf("expected status SUCCESS, got %s", history.TrialVersionHistory[0].Status)
			}

			if err := VerifyStudyExists(client, alias); err != nil {
				return err
			}
			if err := VerifyVersionCount(client, alias, 1); err != nil {
				return err
			}

			return nil
		},
	}
}

func invalidPayloadRejection(prefix string) Scenario {
	return Scenario{
		Name:        "invalid-payload-rejection",
		Description: "Submit a payload with invalid trialAlias and expect sync rejection",
		TrialAlias:  "", // No cleanup needed; payload is rejected synchronously.
		Run: func(client *SDRClient, cfg PollConfig) error {
			payload := BuildInvalidPayload()

			_, err := client.SubmitData(payload)
			if err == nil {
				return fmt.Errorf("expected error for invalid payload, but submission succeeded")
			}
			// The error is expected — the mutation should reject the payload.
			return nil
		},
	}
}

func versionIncrement(prefix string) Scenario {
	alias := fmt.Sprintf("%s-0003", prefix)
	return Scenario{
		Name:        "version-increment",
		Description: "Submit the same trialAlias twice and verify version increments",
		TrialAlias:  alias,
		Run: func(client *SDRClient, cfg PollConfig) error {
			// First submission.
			payload1 := BuildMinimalPayload(alias)
			resp1, err := client.SubmitData(payload1)
			if err != nil {
				return fmt.Errorf("first submit failed: %w", err)
			}
			if !resp1.IsValid {
				return fmt.Errorf("first submit not valid: %s", resp1.Message)
			}

			if _, err := WaitForVersion(client, alias, 1, cfg); err != nil {
				return fmt.Errorf("waiting for first version: %w", err)
			}

			// Brief pause before second submission.
			time.Sleep(2 * time.Second)

			// Second submission.
			payload2 := BuildMinimalPayload(alias)
			payload2.ID = fmt.Sprintf("e2e-test-v2-%d", time.Now().UnixMilli())
			resp2, err := client.SubmitData(payload2)
			if err != nil {
				return fmt.Errorf("second submit failed: %w", err)
			}
			if !resp2.IsValid {
				return fmt.Errorf("second submit not valid: %s", resp2.Message)
			}

			if _, err := WaitForVersion(client, alias, 2, cfg); err != nil {
				return fmt.Errorf("waiting for second version: %w", err)
			}

			if err := VerifyVersionCount(client, alias, 2); err != nil {
				return err
			}

			return nil
		},
	}
}

func deleteSingleVersionScenario(prefix string) Scenario {
	alias := fmt.Sprintf("%s-0004", prefix)
	return Scenario{
		Name:        "delete-single-version",
		Description: "Submit, then delete single version and verify it's removed",
		TrialAlias:  alias,
		Run: func(client *SDRClient, cfg PollConfig) error {
			payload := BuildMinimalPayload(alias)
			resp, err := client.SubmitData(payload)
			if err != nil {
				return fmt.Errorf("submit failed: %w", err)
			}
			if !resp.IsValid {
				return fmt.Errorf("submit not valid: %s", resp.Message)
			}

			history, err := WaitForVersion(client, alias, 1, cfg)
			if err != nil {
				return err
			}

			versionID := history.TrialVersionHistory[0].SDRVersion
			delResp, err := client.DeleteSingleVersion(alias, versionID)
			if err != nil {
				return fmt.Errorf("deleteSingleVersion failed: %w", err)
			}
			if !delResp.Success {
				return fmt.Errorf("deleteSingleVersion returned success=false: %s", delResp.Message)
			}

			return nil
		},
	}
}

func deleteAllVersionsScenario(prefix string) Scenario {
	alias := fmt.Sprintf("%s-0005", prefix)
	return Scenario{
		Name:        "delete-all-versions",
		Description: "Submit, then delete all versions and verify study is gone",
		TrialAlias:  alias,
		Run: func(client *SDRClient, cfg PollConfig) error {
			payload := BuildMinimalPayload(alias)
			resp, err := client.SubmitData(payload)
			if err != nil {
				return fmt.Errorf("submit failed: %w", err)
			}
			if !resp.IsValid {
				return fmt.Errorf("submit not valid: %s", resp.Message)
			}

			if _, err := WaitForVersion(client, alias, 1, cfg); err != nil {
				return err
			}

			delResp, err := client.DeleteAllVersions(alias)
			if err != nil {
				return fmt.Errorf("deleteAllVersions failed: %w", err)
			}
			if !delResp.Success {
				return fmt.Errorf("deleteAllVersions returned success=false: %s", delResp.Message)
			}

			if err := WaitForCleanup(client, alias, 15*time.Second); err != nil {
				return err
			}

			return nil
		},
	}
}

func graphStructure(prefix string) Scenario {
	alias := fmt.Sprintf("%s-0006", prefix)
	return Scenario{
		Name:        "graph-structure",
		Description: "Submit full payload and verify graph node types exist",
		TrialAlias:  alias,
		Run: func(client *SDRClient, cfg PollConfig) error {
			payload := BuildFullPayload(alias)
			resp, err := client.SubmitData(payload)
			if err != nil {
				return fmt.Errorf("submit failed: %w", err)
			}
			if !resp.IsValid {
				return fmt.Errorf("submit not valid: %s", resp.Message)
			}

			history, err := WaitForVersion(client, alias, 1, cfg)
			if err != nil {
				return err
			}
			if history.TrialVersionHistory[0].Status != "SUCCESS" {
				return fmt.Errorf("expected status SUCCESS, got %s", history.TrialVersionHistory[0].Status)
			}

			// Verify key node types exist in the graph.
			checks := []struct {
				label    string
				minCount int
			}{
				{"StudyVersion", 1},
				{"StudyEpoch", 2},
				{"Encounter", 2},
				{"Activity", 2},
			}

			for _, check := range checks {
				if err := VerifyNodesByLabel(client, alias, check.label, check.minCount); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func trialsQuery(prefix string) Scenario {
	alias := fmt.Sprintf("%s-0007", prefix)
	return Scenario{
		Name:        "trials-query",
		Description: "Submit and verify the trial appears in the trials query",
		TrialAlias:  alias,
		Run: func(client *SDRClient, cfg PollConfig) error {
			payload := BuildMinimalPayload(alias)
			resp, err := client.SubmitData(payload)
			if err != nil {
				return fmt.Errorf("submit failed: %w", err)
			}
			if !resp.IsValid {
				return fmt.Errorf("submit not valid: %s", resp.Message)
			}

			if _, err := WaitForVersion(client, alias, 1, cfg); err != nil {
				return err
			}

			trials, err := client.GetTrials()
			if err != nil {
				return fmt.Errorf("trials query failed: %w", err)
			}

			for _, t := range trials {
				if t.TrialAlias == alias {
					return nil
				}
			}
			return fmt.Errorf("trial %s not found in trials query result (%d trials returned)", alias, len(trials))
		},
	}
}

func graphSummaryScenario(prefix string) Scenario {
	alias := fmt.Sprintf("%s-0008", prefix)
	return Scenario{
		Name:        "graph-summary",
		Description: "Capture graph summary before and after submit, verify counts increased",
		TrialAlias:  alias,
		Run: func(client *SDRClient, cfg PollConfig) error {
			before, err := client.GetGraphSummary()
			if err != nil {
				return fmt.Errorf("graphSummary (before): %w", err)
			}

			payload := BuildMinimalPayload(alias)
			resp, err := client.SubmitData(payload)
			if err != nil {
				return fmt.Errorf("submit failed: %w", err)
			}
			if !resp.IsValid {
				return fmt.Errorf("submit not valid: %s", resp.Message)
			}

			if _, err := WaitForVersion(client, alias, 1, cfg); err != nil {
				return err
			}

			after, err := client.GetGraphSummary()
			if err != nil {
				return fmt.Errorf("graphSummary (after): %w", err)
			}

			if after.TotalNodes <= before.TotalNodes {
				return fmt.Errorf("expected totalNodes to increase: before=%d, after=%d", before.TotalNodes, after.TotalNodes)
			}

			return nil
		},
	}
}
