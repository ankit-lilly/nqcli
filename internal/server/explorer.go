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
	files := http.StripPrefix("/explorer/", http.FileServer(http.FS(distFS)))

	s.mux.HandleFunc("/explorer/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/explorer/")
		if path != "" {
			if _, err := fs.Stat(distFS, path); err != nil {
				path = ""
			}
		}
		if strings.HasPrefix(path, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		request := r.Clone(r.Context())
		request.URL.Path = "/explorer/" + path
		files.ServeHTTP(w, request)
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
