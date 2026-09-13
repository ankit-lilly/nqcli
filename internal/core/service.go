package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type Service struct {
	backend QueryBackend
}

func NewService(backend QueryBackend) *Service {
	return &Service{backend: backend}
}

func (s *Service) ExecuteQuery(ctx context.Context, query, queryType string, opts QueryOpts) (QueryResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if strings.TrimSpace(query) == "" {
		return QueryResult{}, fmt.Errorf("query content is empty")
	}

	payload, err := s.backend.ExecuteQuery(ctx, query, queryType, opts)
	content := payload.Content
	if content == "" {
		content = payload.Raw
	}
	if err != nil {
		return QueryResult{Content: content, Raw: payload.Raw}, fmt.Errorf("neptune query failed: %w", err)
	}

	processed := ""
	if !opts.SkipFormatting {
		processed = formatOutput(content)
	}
	return QueryResult{
		Content:   content,
		Processed: processed,
		Raw:       payload.Raw,
	}, nil
}

func formatOutput(content string) string {
	var parsed any
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return content
	}

	prettyJSON, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		return content
	}

	return string(prettyJSON)
}
