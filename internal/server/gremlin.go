package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"uuid"

	"github.com/ankit-lilly/nqcli/internal/core"
)

const graphsonV3Accept = "application/vnd.gremlin-v3.0+json"

type graphQueryRequest struct {
	Query string `json:"query"`
}

type queryErrorResponse struct {
	Code            string `json:"code"`
	DetailedMessage string `json:"detailedMessage"`
}

func newRequestID() string {
	return uuid.New().String()
}

func (s *Server) handleGremlinQuery() http.HandlerFunc {
	type neptuneStatus struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}

	type neptuneEnvelope struct {
		RequestID string          `json:"requestId"`
		Status    neptuneStatus   `json:"status"`
		Result    json.RawMessage `json:"result"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()
		r.Body = http.MaxBytesReader(w, r.Body, requestBodyLimit)

		var req graphQueryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}
		if req.Query == "" {
			http.Error(w, `{"error":"query is required"}`, http.StatusBadRequest)
			return
		}

		opts := core.QueryOpts{
			Serializer:     graphsonV3Accept,
			SkipFormatting: true,
		}
		result, err := s.queryService().ExecuteQuery(r.Context(), req.Query, "gremlin", opts)

		reqID := newRequestID()

		if err != nil {
			writeQueryError(w, err)
			return
		}

		content := result.Content
		if content == "" {
			content = result.Raw
		}

		// The Lambda returns {"data": <graphson>, "meta": {...}}.
		// Graph Explorer expects result.data to be the GraphSON payload directly.
		if data, ok := unwrapDataField(content); ok {
			content = data
		}
		if !json.Valid([]byte(content)) {
			writeQueryError(w, fmt.Errorf("query returned invalid JSON"))
			return
		}

		resultData, _ := json.Marshal(map[string]json.RawMessage{
			"data": json.RawMessage(content),
		})

		env := neptuneEnvelope{
			RequestID: reqID,
			Status:    neptuneStatus{Message: "", Code: 200},
			Result:    resultData,
		}

		w.Header().Set("Content-Type", contentTypeJSON)
		json.NewEncoder(w).Encode(env)
	}
}

func (s *Server) handleOpenCypherQuery() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()
		r.Body = http.MaxBytesReader(w, r.Body, requestBodyLimit)

		var req graphQueryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}
		if req.Query == "" {
			http.Error(w, `{"error":"query is required"}`, http.StatusBadRequest)
			return
		}

		result, err := s.queryService().ExecuteQuery(
			r.Context(),
			req.Query,
			"openCypher",
			core.QueryOpts{SkipFormatting: true},
		)
		if err != nil {
			writeQueryError(w, err)
			return
		}

		content := result.Content
		if content == "" {
			content = result.Raw
		}
		if data, ok := unwrapDataField(content); ok {
			content = data
		}
		if !json.Valid([]byte(content)) {
			writeQueryError(w, fmt.Errorf("query returned invalid JSON"))
			return
		}

		w.Header().Set("Content-Type", contentTypeJSON)
		_, _ = w.Write([]byte(content))
	}
}

func unwrapDataField(content string) (string, bool) {
	var parsed map[string]json.RawMessage
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return "", false
	}
	data, ok := parsed["data"]
	if !ok {
		return "", false
	}
	return string(data), true
}

func writeQueryError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(http.StatusBadGateway)
	_ = json.NewEncoder(w).Encode(queryErrorResponse{
		Code:            "QueryError",
		DetailedMessage: err.Error(),
	})
}

func (s *Server) handleLogger() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		level := r.Header.Get("level")
		message := r.Header.Get("message")

		if level != "" && message != "" {
			switch level {
			case "error":
				s.logger.Error("[Graph Explorer]", "message", message)
			case "warn":
				s.logger.Warn("[Graph Explorer]", "message", message)
			case "info":
				s.logger.Info("[Graph Explorer]", "message", message)
			default:
				s.logger.Debug("[Graph Explorer]", "level", level, "message", message)
			}
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Log received."))
	}
}
