package desktop

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ankit-lilly/nqcli/internal/awsprofile"
	"github.com/ankit-lilly/nqcli/internal/core"
	graphschema "github.com/ankit-lilly/nqcli/internal/schema"
	"github.com/ankit-lilly/nqcli/internal/schema/filecache"
)

type ServiceFactory func(ctx context.Context, profile string) (core.QueryService, error)

type QueryRequest struct {
	Query      string `json:"query"`
	Type       string `json:"type"`
	Serializer string `json:"serializer"`
}

type QueryResponse struct {
	Processed string `json:"processed"`
	Raw       string `json:"raw,omitempty"`
	Error     string `json:"error,omitempty"`
}

type GraphResponse struct {
	Warning  string         `json:"warning,omitempty"`
	Elements []GraphElement `json:"elements"`
	JSON     string         `json:"json,omitempty"`
	Error    string         `json:"error,omitempty"`
}

type ProfileInfo struct {
	Profile  string   `json:"profile"`
	Env      string   `json:"env"`
	Profiles []string `json:"profiles"`
}

type ProfileRequest struct {
	Profile string `json:"profile"`
}

type ProfileResponse struct {
	Profile string `json:"profile"`
	Env     string `json:"env"`
	Error   string `json:"error,omitempty"`
}

type SchemaRequest struct {
	Refresh bool `json:"refresh"`
}

type DesktopService struct {
	svc       core.QueryService
	ctx       context.Context
	profile   string
	factory   ServiceFactory
	mu        sync.RWMutex
	switchMu  sync.Mutex
	cancel    context.CancelFunc
	lifecycle context.Context
	schema    *graphschema.Service
}

func NewDesktopService(svc core.QueryService, profile string, factory ServiceFactory) *DesktopService {
	ctx, cancel := context.WithCancel(context.Background())
	return &DesktopService{
		svc: svc, profile: profile, factory: factory, ctx: ctx, cancel: cancel,
		lifecycle: context.Background(), schema: graphschema.NewService(nil, graphschema.WithStore(filecache.Default())),
	}
}

//wails:ignore
func (d *DesktopService) Startup(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cancel()
	d.lifecycle = ctx
	d.ctx, d.cancel = context.WithCancel(ctx)
	return nil
}

func (d *DesktopService) GetProfile() ProfileInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return ProfileInfo{
		Profile:  d.profile,
		Env:      envFromProfile(d.profile),
		Profiles: awsprofile.Available(),
	}
}

//wails:ignore
func (d *DesktopService) Shutdown() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cancel()
	d.schema.Close()
	return nil
}

// Cancel currently active requests without discarding the application lifecycle context.
func (d *DesktopService) CancelQueries() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cancel()
	d.ctx, d.cancel = context.WithCancel(d.lifecycle)
}

func (d *DesktopService) SwitchProfile(req ProfileRequest) ProfileResponse {
	d.switchMu.Lock()
	defer d.switchMu.Unlock()
	d.mu.RLock()
	same := req.Profile == d.profile
	lifecycle := d.lifecycle
	d.mu.RUnlock()
	if same {
		return ProfileResponse{Profile: req.Profile, Env: envFromProfile(req.Profile)}
	}
	if d.factory == nil {
		return ProfileResponse{Error: "profile switching not available"}
	}
	ctx, cancel := context.WithTimeout(lifecycle, 30*time.Second)
	defer cancel()
	svc, err := safeCallFactory(d.factory, ctx, req.Profile)
	if err != nil {
		return ProfileResponse{Error: fmt.Sprintf("failed to switch profile: %v", err)}
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cancel()
	d.ctx, d.cancel = context.WithCancel(d.lifecycle)
	d.svc, d.profile = svc, req.Profile
	return ProfileResponse{Profile: req.Profile, Env: envFromProfile(req.Profile)}
}

func (d *DesktopService) ExecuteQuery(req QueryRequest) QueryResponse {
	d.mu.RLock()
	svc := d.svc
	ctx := d.ctx
	d.mu.RUnlock()

	if req.Type == "" {
		req.Type = "gremlin"
	}
	opts := core.QueryOpts{Serializer: req.Serializer, MaxResponseBytes: 8 << 20}
	result, err := svc.ExecuteQuery(ctx, req.Query, req.Type, opts)
	resp := QueryResponse{
		Processed: result.Processed,
	}
	if err != nil {
		resp.Error = err.Error()
	}
	return resp
}

func (d *DesktopService) ExecuteGraphQuery(req QueryRequest) GraphResponse {
	d.mu.RLock()
	svc := d.svc
	ctx := d.ctx
	d.mu.RUnlock()

	if req.Type == "" {
		req.Type = "gremlin"
	}
	opts := core.QueryOpts{MaxResponseBytes: 8 << 20}
	result, err := svc.ExecuteQuery(ctx, req.Query, req.Type, opts)
	if err != nil {
		return GraphResponse{Error: err.Error()}
	}

	content := result.Content
	if content == "" {
		content = result.Raw
	}
	elements, parseErr := ParseGraphSON(content)
	if parseErr != nil {
		return GraphResponse{Error: fmt.Sprintf("graph parse error: %v", parseErr)}
	}
	elements, warning := boundGraph(elements)
	jsonResult := result.Processed
	if jsonResult == "" {
		jsonResult = content
	}
	return GraphResponse{Elements: elements, JSON: jsonResult, Warning: warning}
}

// GetSchema returns cached schema data immediately and revalidates it in the
// background when absent, stale, or explicitly refreshed.
func (d *DesktopService) GetSchema(req SchemaRequest) graphschema.Snapshot {
	d.mu.RLock()
	svc, ctx, key := d.svc, d.ctx, d.profile
	d.mu.RUnlock()
	if key == "" {
		key = "default"
	}
	return d.schema.Revalidate(ctx, key, svc, req.Refresh)
}

type VertexPropsRequest struct {
	ID string `json:"id"`
}

type VertexPropsResponse struct {
	Properties map[string]any `json:"properties"`
	Error      string         `json:"error,omitempty"`
}

func (d *DesktopService) GetVertexProperties(req VertexPropsRequest) VertexPropsResponse {
	d.mu.RLock()
	svc := d.svc
	ctx := d.ctx
	d.mu.RUnlock()

	query := fmt.Sprintf("g.V(%s).valueMap(true)", gremlinString(req.ID))
	opts := core.QueryOpts{SkipFormatting: true, MaxResponseBytes: 8 << 20}
	result, err := svc.ExecuteQuery(ctx, query, "gremlin", opts)
	if err != nil {
		return VertexPropsResponse{Error: err.Error()}
	}

	content := result.Content
	if content == "" {
		content = result.Raw
	}

	props, err := parseValueMap(content)
	if err != nil {
		return VertexPropsResponse{Error: err.Error()}
	}
	return VertexPropsResponse{Properties: props}
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
	switch {
	case profile == "" || contains(profile, "dev"):
		return "dev"
	case contains(profile, "qa"):
		return "qa"
	case contains(profile, "prod"):
		return "prod"
	default:
		return "dev"
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		match := true
		for j := range substr {
			c := s[i+j]
			sc := substr[j]
			if c >= 'A' && c <= 'Z' {
				c += 32
			}
			if sc >= 'A' && sc <= 'Z' {
				sc += 32
			}
			if c != sc {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func boundGraph(elements []GraphElement) ([]GraphElement, string) {
	return boundGraphWithExternal(elements, nil)
}

func boundGraphWithExternal(elements []GraphElement, externalNodes map[string]bool) ([]GraphElement, string) {
	const maxNodes, maxEdges = 500, 1000
	nodes := map[string]bool{}
	result := make([]GraphElement, 0, min(len(elements), maxNodes+maxEdges))
	for _, e := range elements {
		if e.Group == "nodes" && len(nodes) < maxNodes {
			id, _ := e.Data["id"].(string)
			nodes[id] = true
			result = append(result, e)
		}
	}
	edges := 0
	for _, e := range elements {
		if e.Group != "edges" || edges >= maxEdges {
			continue
		}
		from, _ := e.Data["source"].(string)
		to, _ := e.Data["target"].(string)
		if (nodes[from] || externalNodes[from]) && (nodes[to] || externalNodes[to]) {
			result = append(result, e)
			edges++
		}
	}
	warning := ""
	if len(result) != len(elements) {
		warning = "Graph preview limited to 500 nodes and 1,000 connected edges. Narrow the query to explore more."
	}
	if len(elements) == 0 {
		warning = "No graph elements returned. Use JSON for counts and projected results."
	}
	return result, warning
}
