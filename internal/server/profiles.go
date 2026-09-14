package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ankit-lilly/nqcli/internal/awsprofile"
	"github.com/ankit-lilly/nqcli/internal/core"
)

type profileInfo struct {
	Profile  string   `json:"profile"`
	Env      string   `json:"env"`
	Profiles []string `json:"profiles"`
}

type profileRequest struct {
	Profile string `json:"profile"`
}

type profileResponse struct {
	Profile string `json:"profile"`
	Env     string `json:"env"`
	Error   string `json:"error,omitempty"`
}

func (s *Server) handleProfiles() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		s.serviceMu.RLock()
		profile := s.profile
		s.serviceMu.RUnlock()

		w.Header().Set("Content-Type", contentTypeJSON)
		_ = json.NewEncoder(w).Encode(profileInfo{
			Profile:  profile,
			Env:      envFromProfile(profile),
			Profiles: awsprofile.Available(),
		})
	}
}

func (s *Server) handleSwitchProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if s.factory == nil {
			writeProfileResponse(w, http.StatusServiceUnavailable, profileResponse{Error: "profile switching not available"})
			return
		}

		defer r.Body.Close()
		r.Body = http.MaxBytesReader(w, r.Body, requestBodyLimit)

		var req profileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeProfileResponse(w, http.StatusBadRequest, profileResponse{Error: "invalid JSON payload"})
			return
		}

		s.serviceMu.RLock()
		same := req.Profile == s.profile
		s.serviceMu.RUnlock()
		if same {
			writeProfileResponse(w, http.StatusOK, profileResponse{Profile: req.Profile, Env: envFromProfile(req.Profile)})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		svc, err := safeCallFactory(s.factory, ctx, req.Profile)
		if err != nil {
			writeProfileResponse(w, http.StatusBadGateway, profileResponse{Error: fmt.Sprintf("failed to switch profile: %v", err)})
			return
		}

		s.serviceMu.Lock()
		s.service = svc
		s.profile = req.Profile
		s.serviceMu.Unlock()

		writeProfileResponse(w, http.StatusOK, profileResponse{Profile: req.Profile, Env: envFromProfile(req.Profile)})
	}
}

func writeProfileResponse(w http.ResponseWriter, status int, resp profileResponse) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

func safeCallFactory(factory ServiceFactory, ctx context.Context, profile string) (svc core.QueryService, err error) {
	defer func() {
		if r := recover(); r != nil {
			svc = nil
			err = fmt.Errorf("credentials error for profile %q: %v (try: aws sso login --profile %s)", profile, r, profile)
		}
	}()
	return factory(ctx, profile)
}

func envFromProfile(profile string) string {
	lower := strings.ToLower(profile)
	switch {
	case lower == "" || strings.Contains(lower, "dev"):
		return "dev"
	case strings.Contains(lower, "qa"):
		return "qa"
	case strings.Contains(lower, "prod"):
		return "prod"
	default:
		return "dev"
	}
}
