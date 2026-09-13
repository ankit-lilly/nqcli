package server

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed webui_dist/*
var explorerDistFS embed.FS

type defaultConnectionConfig struct {
	APIBaseURL  string `json:"apiBaseUrl"`
	QueryEngine string `json:"queryEngine"`
	Profile     string `json:"profile"`
}

func (s *Server) explorerRoutes() {
	distFS, err := fs.Sub(explorerDistFS, "webui_dist")
	if err != nil {
		panic(err)
	}

	s.mux.HandleFunc("/explorer/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/explorer/")
		if path == "" {
			path = "index.html"
		}

		f, err := distFS.(fs.ReadFileFS).ReadFile(path)
		if err != nil {
			// SPA fallback: serve index.html for unknown paths
			f, err = distFS.(fs.ReadFileFS).ReadFile("index.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			path = "index.html"
		}

		ctype := "application/octet-stream"
		switch {
		case strings.HasSuffix(path, ".html"):
			ctype = "text/html; charset=utf-8"
		case strings.HasSuffix(path, ".js"):
			ctype = "application/javascript"
		case strings.HasSuffix(path, ".css"):
			ctype = "text/css"
		case strings.HasSuffix(path, ".json"):
			ctype = "application/json"
		case strings.HasSuffix(path, ".svg"):
			ctype = "image/svg+xml"
		case strings.HasSuffix(path, ".png"):
			ctype = "image/png"
		case strings.HasSuffix(path, ".ico"):
			ctype = "image/x-icon"
		}

		w.Header().Set("Content-Type", ctype)
		w.Write(f)
	})

	s.mux.HandleFunc("/defaultConnection", s.handleDefaultConnection())
	s.mux.HandleFunc("/profiles", s.handleProfiles())
	s.mux.HandleFunc("/profiles/current", s.handleSwitchProfile())
	s.mux.HandleFunc("/gremlin", s.handleGremlinQuery())
	s.mux.HandleFunc("/openCypher", s.handleOpenCypherQuery())
	s.mux.HandleFunc("/schema", s.handleSchema())
	s.mux.HandleFunc("/schema/events", s.handleSchemaEvents())
	s.mux.HandleFunc("/logger", s.handleLogger())
}

func (s *Server) handleDefaultConnection() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		endpoint := scheme + "://" + r.Host
		s.serviceMu.RLock()
		profile := s.profile
		queryEngine := s.queryEngine
		s.serviceMu.RUnlock()

		resp := defaultConnectionConfig{
			APIBaseURL:  endpoint,
			QueryEngine: queryEngine,
			Profile:     profile,
		}

		w.Header().Set("Content-Type", contentTypeJSON)
		json.NewEncoder(w).Encode(resp)
	}
}
