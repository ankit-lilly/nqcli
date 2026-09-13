package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	graphschema "github.com/ankit-lilly/nqcli/internal/schema"
)

func (s *Server) handleSchema() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := s.schemaKey()
		switch r.Method {
		case http.MethodGet:
			snapshot := s.schema.Revalidate(r.Context(), key, s.queryService(), r.URL.Query().Get("refresh") == "true")
			if r.URL.Query().Get("wait") == "true" && snapshot.Status == graphschema.StatusRunning && s.schema.Wait(r.Context(), key) {
				snapshot = s.schema.Get(key)
			}
			writeSchemaSnapshot(w, snapshot)
		case http.MethodPost:
			writeSchemaSnapshot(w, s.schema.Revalidate(r.Context(), key, s.queryService(), true))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func writeSchemaSnapshot(w http.ResponseWriter, snapshot graphschema.Snapshot) {
	w.Header().Set("Content-Type", contentTypeJSON)
	switch snapshot.Status {
	case graphschema.StatusEmpty, graphschema.StatusRunning:
		w.WriteHeader(http.StatusAccepted)
	case graphschema.StatusFailed:
		w.WriteHeader(http.StatusBadGateway)
	}
	_ = json.NewEncoder(w).Encode(snapshot)
}

func (s *Server) handleSchemaEvents() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		key := s.schemaKey()
		events, unsubscribe := s.schema.Subscribe()
		defer unsubscribe()
		_ = writeSchemaSSE(w, graphschema.Event{Type: "schema.sync.snapshot", Key: key, Status: s.schema.Get(key).Status})
		flusher.Flush()
		for {
			select {
			case <-r.Context().Done():
				return
			case event, open := <-events:
				if !open || writeSchemaSSE(w, event) != nil {
					return
				}
				flusher.Flush()
			}
		}
	}
}

func writeSchemaSSE(w http.ResponseWriter, event graphschema.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "event: schema\ndata: %s\n\n", data)
	return err
}

func (s *Server) schemaKey() string {
	s.serviceMu.RLock()
	defer s.serviceMu.RUnlock()
	if s.profile == "" {
		return "default"
	}
	return s.profile
}
