package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// neptuneExecutor defines the capability needed from the Neptune client.
type neptuneExecutor interface {
	ExecuteQueryCtx(ctx context.Context, query string, queryType string) (string, error)
}

// AppService provides query execution against a Neptune database via an
// AppSync GraphQL endpoint.
type AppService struct {
	neptuneClient neptuneExecutor
}

// NewAppService creates an AppService backed by the given Neptune client.
func NewAppService(nc neptuneExecutor) *AppService {
	return &AppService{
		neptuneClient: nc,
	}
}

// Execute reads a query from the given file path (or stdin if empty) and
// executes it against Neptune using a background context. It returns a
// pretty-printed result and the raw JSON response.
func (s *AppService) Execute(queryFilePath string, queryType string) (processedOutput string, rawJSONResponse string, err error) {
	return s.ExecuteCtx(context.Background(), queryFilePath, queryType)
}

// ExecuteCtx reads a query from the given file path (or stdin if empty) and
// executes it against Neptune using the provided context. It returns a
// pretty-printed result and the raw JSON response.
func (s *AppService) ExecuteCtx(ctx context.Context, queryFilePath string, queryType string) (processedOutput string, rawJSONResponse string, err error) {
	query, err := s.readQueryContent(queryFilePath)
	if err != nil {
		return "", "", err
	}

	return s.ExecuteQueryCtx(ctx, query, queryType)
}

// ExecuteQuery runs a raw query string against Neptune and returns the
// pretty-printed result extracted from the GraphQL envelope, along with the
// raw JSON response, using a background context.
func (s *AppService) ExecuteQuery(query string, queryType string) (processedOutput string, rawJSONResponse string, err error) {
	return s.ExecuteQueryCtx(context.Background(), query, queryType)
}

// ExecuteQueryCtx runs a raw query string against Neptune and returns the
// pretty-printed result extracted from the GraphQL envelope, along with the
// raw JSON response.
func (s *AppService) ExecuteQueryCtx(ctx context.Context, query string, queryType string) (processedOutput string, rawJSONResponse string, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if strings.TrimSpace(query) == "" {
		return "", "", fmt.Errorf("query content is empty")
	}

	rawJSONResponse, err = s.neptuneClient.ExecuteQueryCtx(ctx, query, queryType)
	if err != nil {
		return "", rawJSONResponse, fmt.Errorf("neptune query failed: %w", err)
	}

	var responseMap map[string]any
	if err := json.Unmarshal([]byte(rawJSONResponse), &responseMap); err != nil {
		return rawJSONResponse, rawJSONResponse, fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}

	data, ok := responseMap["data"].(map[string]any)

	if !ok || data == nil {
		processedOutput = rawJSONResponse
	} else {
		executeQuery, ok := data["executeQuery"].(string)
		if !ok {
			processedOutput = rawJSONResponse
		} else {
			var innerData any
			if err := json.Unmarshal([]byte(executeQuery), &innerData); err != nil {
				processedOutput = executeQuery
			} else {

				finalOutputData := innerData

				if obj, isObject := innerData.(map[string]any); isObject {
					if finalData, finalDataExists := obj["data"]; finalDataExists {
						finalOutputData = finalData
					}
				}

				prettyJSON, marshalErr := json.MarshalIndent(finalOutputData, "", "  ")
				if marshalErr != nil {
					processedOutput = executeQuery // Fallback
				} else {
					processedOutput = string(prettyJSON)
				}
			}
		}
	}

	return processedOutput, rawJSONResponse, nil
}

func (s *AppService) readQueryContent(queryFilePath string) (string, error) {
	var reader io.Reader

	if queryFilePath != "" {
		file, err := os.Open(queryFilePath)
		if err != nil {
			return "", fmt.Errorf("failed to open query file: %w", err)
		}
		defer file.Close()
		reader = file
	} else {
		fi, err := os.Stdin.Stat()
		if err != nil {
			return "", fmt.Errorf("failed to stat stdin: %w", err)
		}

		if (fi.Mode() & os.ModeCharDevice) == 0 {
			reader = os.Stdin
		} else {
			return "", fmt.Errorf("no query provided. Use 'echo \"query\" | nq-cli' or 'nq-cli <query_file>'")
		}
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read query content: %w", err)
	}

	return string(content), nil
}
