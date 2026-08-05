package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type neptuneExecutor interface {
	ExecuteQueryCtx(ctx context.Context, query string, queryType string) (string, error)
}

type AppService struct {
	neptuneClient neptuneExecutor
}

func NewAppService(nc neptuneExecutor) *AppService {
	return &AppService{
		neptuneClient: nc,
	}
}

func (s *AppService) Execute(queryFilePath string, queryType string) (processedOutput string, rawJSONResponse string, err error) {
	return s.ExecuteCtx(context.Background(), queryFilePath, queryType)
}

func (s *AppService) ExecuteCtx(ctx context.Context, queryFilePath string, queryType string) (processedOutput string, rawJSONResponse string, err error) {
	query, err := s.readQueryContent(queryFilePath)
	if err != nil {
		return "", "", err
	}

	return s.ExecuteQueryCtx(ctx, query, queryType)
}

func (s *AppService) ExecuteQuery(query string, queryType string) (processedOutput string, rawJSONResponse string, err error) {
	return s.ExecuteQueryCtx(context.Background(), query, queryType)
}

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
		// Not a JSON object — try pretty-printing as generic JSON (could be an array)
		var parsed any
		if jsonErr := json.Unmarshal([]byte(rawJSONResponse), &parsed); jsonErr != nil {
			return rawJSONResponse, rawJSONResponse, nil
		}
		prettyJSON, _ := json.MarshalIndent(parsed, "", "  ")
		return string(prettyJSON), rawJSONResponse, nil
	}

	rawData, hasData := responseMap["data"]
	data, isEnvelope := rawData.(map[string]any)

	switch {
	case isEnvelope && data != nil:
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
					processedOutput = executeQuery
				} else {
					processedOutput = string(prettyJSON)
				}
			}
		}
	case hasData && rawData != nil:
		// REST mode: "data" holds the result directly (e.g. an array) — unwrap it
		// so the output matches the AppSync envelope's already-unwrapped shape.
		prettyJSON, marshalErr := json.MarshalIndent(rawData, "", "  ")
		if marshalErr != nil {
			processedOutput = rawJSONResponse
		} else {
			processedOutput = string(prettyJSON)
		}
	default:
		prettyJSON, marshalErr := json.MarshalIndent(responseMap, "", "  ")
		if marshalErr != nil {
			processedOutput = rawJSONResponse
		} else {
			processedOutput = string(prettyJSON)
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
