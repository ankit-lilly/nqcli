package e2e

import (
	"encoding/json"
	"fmt"

	neptune "github.com/ankit-lilly/nqcli/internal/gq"
)

// SDRClient wraps the gq.Client to provide SDR-specific GraphQL operations.
type SDRClient struct {
	gql     *neptune.Client
	verbose bool
}

func NewSDRClient(gql *neptune.Client, verbose bool) *SDRClient {
	return &SDRClient{gql: gql, verbose: verbose}
}

func (c *SDRClient) SubmitData(payload SdrPayload) (*SubmitResponse, error) {
	const mutation = `mutation SubmitData($input: SdrPayload!) {
		submitData(input: $input) {
			message
			isValid
		}
	}`

	vars := map[string]any{"input": payload}
	result, err := c.execute(mutation, vars)
	if err != nil {
		return nil, fmt.Errorf("submitData: %w", err)
	}

	data, err := extractField(result, "submitData")
	if err != nil {
		return nil, fmt.Errorf("submitData response: %w", err)
	}

	var resp SubmitResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("submitData unmarshal: %w", err)
	}
	return &resp, nil
}

func (c *SDRClient) GetTrialHistory(trialAlias string) (*TrialVersionHistory, error) {
	const query = `query TrialHistory($trialAlias: String!) {
		trialHistory(trialAlias: $trialAlias) {
			status
			trialAlias
			therapeuticAreas
			sponsor
			initialExportDate
			studyPhase
			studyType
			dsLatestVersion
			sdrLatestVersion
			author
			collaborators
			trialVersionHistory {
				studyId
				trialAlias
				dsVersion
				dsVersionTimestamp
				sdrVersion
				sdrIngestionTimestamp
				status
				notificationStatus
			}
		}
	}`

	vars := map[string]any{"trialAlias": trialAlias}
	result, err := c.execute(query, vars)
	if err != nil {
		return nil, fmt.Errorf("trialHistory: %w", err)
	}

	data, err := extractField(result, "trialHistory")
	if err != nil {
		return nil, fmt.Errorf("trialHistory response: %w", err)
	}

	var resp TrialVersionHistory
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("trialHistory unmarshal: %w", err)
	}
	return &resp, nil
}

func (c *SDRClient) GetTrials() ([]Trial, error) {
	const query = `query { trials {
		trialAlias dsVersion dsVersionTimestamp sdrVersion
		sdrIngestionTimestamp status therapeuticAreas studyPhase studyType
	}}`

	result, err := c.execute(query, nil)
	if err != nil {
		return nil, fmt.Errorf("trials: %w", err)
	}

	data, err := extractField(result, "trials")
	if err != nil {
		return nil, fmt.Errorf("trials response: %w", err)
	}

	var resp []Trial
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("trials unmarshal: %w", err)
	}
	return resp, nil
}

func (c *SDRClient) GetGraphSummary() (*GraphSummary, error) {
	const query = `query { graphSummary {
		totalNodes totalEdges totalNodeTypes
		nodes { label totalCount }
	}}`

	result, err := c.execute(query, nil)
	if err != nil {
		return nil, fmt.Errorf("graphSummary: %w", err)
	}

	data, err := extractField(result, "graphSummary")
	if err != nil {
		return nil, fmt.Errorf("graphSummary response: %w", err)
	}

	var resp GraphSummary
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("graphSummary unmarshal: %w", err)
	}
	return &resp, nil
}

func (c *SDRClient) ExecuteGremlin(query string) (json.RawMessage, error) {
	raw, err := c.gql.ExecuteQuery(query, "gremlin")
	if err != nil {
		return nil, fmt.Errorf("gremlin: %w", err)
	}

	// The response is: {"data":{"executeQuery":"<escaped JSON>"}}
	var envelope struct {
		Data struct {
			ExecuteQuery string `json:"executeQuery"`
		} `json:"data"`
		Errors []graphQLError `json:"errors"`
	}
	if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
		return nil, fmt.Errorf("gremlin unmarshal envelope: %w", err)
	}
	if len(envelope.Errors) > 0 {
		return nil, fmt.Errorf("gremlin errors: %s", envelope.Errors[0].Message)
	}

	return json.RawMessage(envelope.Data.ExecuteQuery), nil
}

func (c *SDRClient) DeleteAllVersions(trialAlias string) (*DeleteResponse, error) {
	const mutation = `mutation DeleteAll($trialAlias: String!) {
		deleteAllVersionsByTrialAlias(trialAlias: $trialAlias) {
			success
			message
		}
	}`

	vars := map[string]any{"trialAlias": trialAlias}
	result, err := c.execute(mutation, vars)
	if err != nil {
		return nil, fmt.Errorf("deleteAll: %w", err)
	}

	data, err := extractField(result, "deleteAllVersionsByTrialAlias")
	if err != nil {
		return nil, fmt.Errorf("deleteAll response: %w", err)
	}

	var resp DeleteResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("deleteAll unmarshal: %w", err)
	}
	return &resp, nil
}

func (c *SDRClient) DeleteSingleVersion(trialAlias string, versionID string) (*DeleteResponse, error) {
	const mutation = `mutation DeleteVersion($trialAlias: String!, $versionIdentifier: ID!) {
		deleteSingleVersion(trialAlias: $trialAlias, versionIdentifier: $versionIdentifier) {
			success
			message
		}
	}`

	vars := map[string]any{"trialAlias": trialAlias, "versionIdentifier": versionID}
	result, err := c.execute(mutation, vars)
	if err != nil {
		return nil, fmt.Errorf("deleteVersion: %w", err)
	}

	data, err := extractField(result, "deleteSingleVersion")
	if err != nil {
		return nil, fmt.Errorf("deleteVersion response: %w", err)
	}

	var resp DeleteResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("deleteVersion unmarshal: %w", err)
	}
	return &resp, nil
}

// execute sends a GraphQL operation and returns the parsed data map.
func (c *SDRClient) execute(query string, variables any) (map[string]json.RawMessage, error) {
	raw, err := c.gql.ExecuteGraphQL(query, variables)
	if err != nil {
		return nil, err
	}

	if c.verbose {
		fmt.Printf("  [graphql] response: %s\n", raw)
	}

	var envelope struct {
		Data   map[string]json.RawMessage `json:"data"`
		Errors []graphQLError             `json:"errors"`
	}
	if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(envelope.Errors) > 0 {
		return nil, fmt.Errorf("graphql error: %s", envelope.Errors[0].Message)
	}

	return envelope.Data, nil
}

type graphQLError struct {
	Message string `json:"message"`
}

// extractField extracts a single field from the data map.
func extractField(data map[string]json.RawMessage, field string) (json.RawMessage, error) {
	raw, ok := data[field]
	if !ok {
		return nil, fmt.Errorf("field %q not found in response", field)
	}
	return raw, nil
}
