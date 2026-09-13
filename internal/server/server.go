package server

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"io/fs"
	"net/http"
	"sync"
	"time"

	"github.com/ankit-lilly/nqcli/internal/core"
	graphschema "github.com/ankit-lilly/nqcli/internal/schema"
	"github.com/charmbracelet/log"
)

const (
	defaultAddr       = ":8080"
	requestBodyLimit  = 1 << 20
	defaultQueryType  = "gremlin"
	contentTypeJSON   = "application/json"
	contentTypeHTML   = "text/html; charset=utf-8"
	serverReadTimeout = 15 * time.Second
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed assets/*
var assetsFS embed.FS

var (
	pageTemplates   = template.Must(template.ParseFS(templateFS, "templates/*.html"))
	assetFileSystem = mustSubFS(assetsFS, "assets")
)

type Server struct {
	serviceMu   sync.RWMutex
	service     core.QueryService
	factory     ServiceFactory
	profile     string
	queryEngine string
	logger      *log.Logger
	mux         *http.ServeMux
	schema      *graphschema.Service
}

type ServiceFactory func(context.Context, string) (core.QueryService, error)

type Options struct {
	Profile        string
	QueryEngine    string
	ServiceFactory ServiceFactory
}

func New(service core.QueryService, logger *log.Logger) *Server {
	return NewWithOptions(service, logger, Options{})
}

func NewWithOptions(service core.QueryService, logger *log.Logger, opts Options) *Server {
	queryEngine := opts.QueryEngine
	if queryEngine == "" {
		queryEngine = defaultQueryType
	}
	s := &Server{
		service:     service,
		factory:     opts.ServiceFactory,
		profile:     opts.Profile,
		queryEngine: queryEngine,
		logger:      logger,
		mux:         http.NewServeMux(),
	}
	s.schema = graphschema.NewService(logger)

	s.routes()

	return s
}

func (s *Server) queryService() core.QueryService {
	s.serviceMu.RLock()
	defer s.serviceMu.RUnlock()
	return s.service
}

func (s *Server) routes() {
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/explorer/", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})
	s.mux.Handle("/minimal/assets/", http.StripPrefix("/minimal/assets/", http.FileServer(http.FS(assetFileSystem))))
	s.mux.HandleFunc("/minimal", s.handleMinimalIndex())
	s.mux.HandleFunc("/minimal/", s.handleMinimalIndex())
	s.mux.HandleFunc("/healthz", s.handleHealthz())
	s.mux.HandleFunc("/queries", s.handleExecuteQuery())

	s.explorerRoutes()
}

// ServeHTTP implements [http.Handler].
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// Start begins listening on addr and blocks until ctx is cancelled, at which
// point the server gracefully shuts down.
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
		Type       string `json:"type"`
		Query      string `json:"query"`
		Serializer string `json:"serializer,omitempty"`
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

		opts := core.QueryOpts{Serializer: req.Serializer}
		result, err := s.queryService().ExecuteQuery(r.Context(), req.Query, queryType, opts)

		resp := queryResponse{
			Type:        queryType,
			Processed:   result.Processed,
			RawResponse: result.Raw,
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

func (s *Server) handleMinimalIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", contentTypeHTML)
		if err := pageTemplates.ExecuteTemplate(w, "index", nil); err != nil {
			s.logger.Error("failed to render template", "error", err)
		}
	}
}

func mustSubFS(fsys embed.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}

	return sub
}
