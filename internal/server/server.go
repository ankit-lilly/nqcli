package server

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"time"

	"github.com/ankit-lilly/nqcli/internal/app"

	"github.com/charmbracelet/log"
)

const (
	defaultAddr       = ":8080"
	requestBodyLimit  = 1 << 20 // 1 MiB
	defaultQueryType  = "gremlin"
	contentTypeJSON   = "application/json"
	contentTypeHTML   = "text/html; charset=utf-8"
	serverReadTimeout = 10 * time.Second
)

//go:embed templates/*.html
var templateFS embed.FS

var indexTemplate = template.Must(template.ParseFS(templateFS, "templates/index.html"))

type Server struct {
	app    *app.AppService
	logger *log.Logger
	mux    *http.ServeMux
}

func New(appService *app.AppService, logger *log.Logger) *Server {
	s := &Server{
		app:    appService,
		logger: logger,
		mux:    http.NewServeMux(),
	}

	s.routes()

	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/", s.handleIndex())
	s.mux.HandleFunc("/healthz", s.handleHealthz())
	s.mux.HandleFunc("/queries", s.handleExecuteQuery())
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) Start(ctx context.Context, addr string) error {
	if addr == "" {
		addr = defaultAddr
	}

	server := &http.Server{
		Addr:        addr,
		Handler:     s,
		ReadTimeout: serverReadTimeout,
	}

	errCh := make(chan error, 1)

	go func() {
		s.logger.Info("HTTP server starting", "addr", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.logger.Info("HTTP server shutting down")
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func (s *Server) handleHealthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
}

func (s *Server) handleExecuteQuery() http.HandlerFunc {
	type queryRequest struct {
		Type  string `json:"type"`
		Query string `json:"query"`
	}

	type queryResponse struct {
		Type         string `json:"type"`
		Processed    string `json:"processed"`
		RawResponse  string `json:"rawResponse"`
		ErrorMessage string `json:"error,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()
		r.Body = http.MaxBytesReader(w, r.Body, requestBodyLimit)

		var req queryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}

		queryType := req.Type
		if queryType == "" {
			queryType = defaultQueryType
		}

		processed, raw, err := s.app.ExecuteQuery(req.Query, queryType)
		resp := queryResponse{
			Type:        queryType,
			Processed:   processed,
			RawResponse: raw,
		}

		status := http.StatusOK
		if err != nil {
			resp.ErrorMessage = err.Error()
			status = http.StatusBadRequest
		}

		w.Header().Set("Content-Type", contentTypeJSON)
		w.WriteHeader(status)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			s.logger.Error("failed to write JSON response", "error", err)
		}
	}
}

func (s *Server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", contentTypeHTML)
		if err := indexTemplate.Execute(w, nil); err != nil {
			s.logger.Error("failed to render template", "error", err)
		}
	}
}
