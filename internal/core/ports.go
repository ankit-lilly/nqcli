package core

import "context"

type QueryOpts struct {
	Serializer       string
	SkipFormatting   bool
	MaxResponseBytes int64
}

type QueryPayload struct {
	Content string
	Raw     string
}

type QueryResult struct {
	Content   string
	Processed string
	Raw       string
}

type QueryService interface {
	ExecuteQuery(ctx context.Context, query, queryType string, opts QueryOpts) (QueryResult, error)
}

type QueryBackend interface {
	ExecuteQuery(ctx context.Context, query, queryType string, opts QueryOpts) (QueryPayload, error)
}
