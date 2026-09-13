package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ankit-lilly/nqcli/internal/core"
	"golang.org/x/sync/errgroup"
)

const (
	schemaStatusEmpty   = "empty"
	schemaStatusRunning = "running"
	schemaStatusReady   = "ready"
	schemaStatusFailed  = "failed"

	schemaSampleSize         = 500
	edgeConnectionSampleSize = 5000
	schemaQueryConcurrency   = 4
	schemaRefreshTimeout     = 5 * time.Minute
	schemaStaleAfter         = 10 * time.Minute
)

type schemaAttribute struct {
	Name     string `json:"name"`
	DataType string `json:"dataType"`
}

type schemaVertex struct {
	Type       string            `json:"type"`
	Attributes []schemaAttribute `json:"attributes"`
	Total      *int              `json:"total,omitempty"`
}

type schemaEdge struct {
	Type       string            `json:"type"`
	Attributes []schemaAttribute `json:"attributes"`
	Total      *int              `json:"total,omitempty"`
}

type schemaEdgeConnection struct {
	SourceVertexType string `json:"sourceVertexType"`
	EdgeType         string `json:"edgeType"`
	TargetVertexType string `json:"targetVertexType"`
}

type schemaSnapshot struct {
	TotalVertices      *int                   `json:"totalVertices,omitempty"`
	Vertices           []schemaVertex         `json:"vertices"`
	TotalEdges         *int                   `json:"totalEdges,omitempty"`
	Edges              []schemaEdge           `json:"edges"`
	EdgeConnections    []schemaEdgeConnection `json:"edgeConnections"`
	Status             string                 `json:"status"`
	Phase              string                 `json:"phase,omitempty"`
	Completed          int                    `json:"completed,omitempty"`
	Total              int                    `json:"total,omitempty"`
	Error              string                 `json:"error,omitempty"`
	LastUpdate         *time.Time             `json:"lastUpdate,omitempty"`
	LastRefreshStarted *time.Time             `json:"-"`
}

type schemaEvent struct {
	Type      string `json:"type"`
	Key       string `json:"key"`
	Status    string `json:"status"`
	Phase     string `json:"phase,omitempty"`
	Completed int    `json:"completed,omitempty"`
	Total     int    `json:"total,omitempty"`
	Error     string `json:"error,omitempty"`
}

type schemaManager struct {
	mu        sync.RWMutex
	snapshots map[string]schemaSnapshot
	running   map[string]chan struct{}
	subs      map[chan schemaEvent]struct{}
	logger    loggerLike
}

type loggerLike interface {
	Info(msg interface{}, keyvals ...interface{})
	Warn(msg interface{}, keyvals ...interface{})
	Error(msg interface{}, keyvals ...interface{})
}

func newSchemaManager(logger loggerLike) *schemaManager {
	return &schemaManager{
		snapshots: make(map[string]schemaSnapshot),
		running:   make(map[string]chan struct{}),
		subs:      make(map[chan schemaEvent]struct{}),
		logger:    logger,
	}
}

func (m *schemaManager) get(key string) schemaSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	snapshot, ok := m.snapshots[key]
	if !ok {
		return emptySchemaSnapshot(schemaStatusEmpty)
	}
	return snapshot
}

func emptySchemaSnapshot(status string) schemaSnapshot {
	return schemaSnapshot{
		Vertices:        []schemaVertex{},
		Edges:           []schemaEdge{},
		EdgeConnections: []schemaEdgeConnection{},
		Status:          status,
	}
}

func (m *schemaManager) start(ctx context.Context, key string, svc core.QueryService, force bool) schemaSnapshot {
	m.mu.Lock()
	existing, ok := m.snapshots[key]
	if _, running := m.running[key]; running {
		m.mu.Unlock()
		if ok {
			return existing
		}
		return emptySchemaSnapshot(schemaStatusRunning)
	}
	if ok && existing.Status == schemaStatusReady && !force {
		m.mu.Unlock()
		return existing
	}

	runningSnapshot := existing
	runningSnapshot.Status = schemaStatusRunning
	runningSnapshot.Phase = "starting"
	runningSnapshot.Error = ""
	now := time.Now()
	runningSnapshot.LastRefreshStarted = &now
	if runningSnapshot.Vertices == nil {
		runningSnapshot.Vertices = []schemaVertex{}
	}
	if runningSnapshot.Edges == nil {
		runningSnapshot.Edges = []schemaEdge{}
	}
	if runningSnapshot.EdgeConnections == nil {
		runningSnapshot.EdgeConnections = []schemaEdgeConnection{}
	}
	m.snapshots[key] = runningSnapshot
	m.running[key] = make(chan struct{})
	m.mu.Unlock()

	m.publish(schemaEvent{Type: "schema.sync.started", Key: key, Status: schemaStatusRunning, Phase: "starting"})

	go m.runRefresh(ctx, key, svc)

	return runningSnapshot
}

func (m *schemaManager) runRefresh(parent context.Context, key string, svc core.QueryService) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), schemaRefreshTimeout)
	defer cancel()

	discovery := schemaDiscovery{
		service: svc,
		onProgress: func(phase string, completed, total int) {
			m.updateProgress(key, phase, completed, total)
		},
	}

	snapshot, err := discovery.discover(ctx)

	m.mu.Lock()
	defer m.mu.Unlock()
	done := m.running[key]
	delete(m.running, key)
	close(done)

	if err != nil {
		existing := m.snapshots[key]
		existing.Status = schemaStatusFailed
		existing.Error = err.Error()
		if existing.Vertices == nil {
			existing.Vertices = []schemaVertex{}
		}
		if existing.Edges == nil {
			existing.Edges = []schemaEdge{}
		}
		if existing.EdgeConnections == nil {
			existing.EdgeConnections = []schemaEdgeConnection{}
		}
		m.snapshots[key] = existing
		m.publishLocked(schemaEvent{Type: "schema.sync.failed", Key: key, Status: schemaStatusFailed, Error: err.Error()})
		if m.logger != nil {
			m.logger.Warn("schema sync failed", "key", key, "error", err)
		}
		return
	}

	now := time.Now()
	snapshot.Status = schemaStatusReady
	snapshot.Phase = "complete"
	snapshot.LastUpdate = &now
	m.snapshots[key] = snapshot
	m.publishLocked(schemaEvent{Type: "schema.sync.completed", Key: key, Status: schemaStatusReady, Phase: "complete"})
}

func (m *schemaManager) wait(ctx context.Context, key string) bool {
	m.mu.RLock()
	done := m.running[key]
	m.mu.RUnlock()
	if done == nil {
		return true
	}

	select {
	case <-done:
		return true
	case <-ctx.Done():
		return false
	}
}

func (m *schemaManager) updateProgress(key, phase string, completed, total int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	snapshot := m.snapshots[key]
	snapshot.Status = schemaStatusRunning
	snapshot.Phase = phase
	snapshot.Completed = completed
	snapshot.Total = total
	m.snapshots[key] = snapshot
	m.publishLocked(schemaEvent{
		Type:      "schema.sync.progress",
		Key:       key,
		Status:    schemaStatusRunning,
		Phase:     phase,
		Completed: completed,
		Total:     total,
	})
}

func (m *schemaManager) subscribe() (chan schemaEvent, func()) {
	ch := make(chan schemaEvent, 16)
	m.mu.Lock()
	m.subs[ch] = struct{}{}
	m.mu.Unlock()

	return ch, func() {
		m.mu.Lock()
		delete(m.subs, ch)
		close(ch)
		m.mu.Unlock()
	}
}

func (m *schemaManager) publish(event schemaEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishLocked(event)
}

func (m *schemaManager) publishLocked(event schemaEvent) {
	for ch := range m.subs {
		select {
		case ch <- event:
		default:
		}
	}
}

type schemaDiscovery struct {
	service    core.QueryService
	onProgress func(phase string, completed, total int)
}

func (d schemaDiscovery) discover(ctx context.Context) (schemaSnapshot, error) {
	d.progress("labels", 0, 2)
	vertexCounts, err := d.groupCounts(ctx, "g.V().label().groupCount()")
	if err != nil {
		return schemaSnapshot{}, fmt.Errorf("discover vertex labels: %w", err)
	}
	d.progress("labels", 1, 2)
	edgeCounts, err := d.groupCounts(ctx, "g.E().label().groupCount()")
	if err != nil {
		return schemaSnapshot{}, fmt.Errorf("discover edge labels: %w", err)
	}
	d.progress("labels", 2, 2)

	vertexLabels := sortedKeys(vertexCounts)
	edgeLabels := sortedKeys(edgeCounts)

	d.progress("vertex-properties", 0, len(vertexLabels))
	vertices, err := discoverByLabel(ctx, vertexLabels, func(ctx context.Context, label string) (schemaVertex, error) {
		attrs, err := d.attributes(ctx, "V", label)
		if err != nil {
			return schemaVertex{}, fmt.Errorf("discover vertex properties for %q: %w", label, err)
		}
		total := vertexCounts[label]
		return schemaVertex{Type: label, Attributes: attrs, Total: &total}, nil
	}, func(completed int) {
		d.progress("vertex-properties", completed, len(vertexLabels))
	})
	if err != nil {
		return schemaSnapshot{}, err
	}

	d.progress("edge-properties", 0, len(edgeLabels))
	edges, err := discoverByLabel(ctx, edgeLabels, func(ctx context.Context, label string) (schemaEdge, error) {
		attrs, err := d.attributes(ctx, "E", label)
		if err != nil {
			return schemaEdge{}, fmt.Errorf("discover edge properties for %q: %w", label, err)
		}
		total := edgeCounts[label]
		return schemaEdge{Type: label, Attributes: attrs, Total: &total}, nil
	}, func(completed int) {
		d.progress("edge-properties", completed, len(edgeLabels))
	})
	if err != nil {
		return schemaSnapshot{}, err
	}

	d.progress("edge-connections", 0, len(edgeLabels))
	connectionsByLabel, err := discoverByLabel(ctx, edgeLabels, func(ctx context.Context, label string) ([]schemaEdgeConnection, error) {
		connections, err := d.edgeConnections(ctx, label)
		if err != nil {
			return nil, fmt.Errorf("discover edge connections for %q: %w", label, err)
		}
		return connections, nil
	}, func(completed int) {
		d.progress("edge-connections", completed, len(edgeLabels))
	})
	if err != nil {
		return schemaSnapshot{}, err
	}

	edgeConnections := make([]schemaEdgeConnection, 0, len(edgeLabels))
	seen := make(map[string]bool)
	for _, connections := range connectionsByLabel {
		for _, connection := range connections {
			key := connection.SourceVertexType + "\x00" + connection.EdgeType + "\x00" + connection.TargetVertexType
			if seen[key] {
				continue
			}
			seen[key] = true
			edgeConnections = append(edgeConnections, connection)
		}
	}

	totalVertices := sumCounts(vertexCounts)
	totalEdges := sumCounts(edgeCounts)
	return schemaSnapshot{
		TotalVertices:   &totalVertices,
		Vertices:        vertices,
		TotalEdges:      &totalEdges,
		Edges:           edges,
		EdgeConnections: edgeConnections,
	}, nil
}

func discoverByLabel[T any](
	ctx context.Context,
	labels []string,
	discover func(context.Context, string) (T, error),
	onProgress func(completed int),
) ([]T, error) {
	results := make([]T, len(labels))
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(schemaQueryConcurrency)
	var completed atomic.Int64

	for i, label := range labels {
		group.Go(func() error {
			result, err := discover(groupCtx, label)
			if err != nil {
				return err
			}
			results[i] = result
			if onProgress != nil {
				onProgress(int(completed.Add(1)))
			}
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

func (d schemaDiscovery) progress(phase string, completed, total int) {
	if d.onProgress != nil {
		d.onProgress(phase, completed, total)
	}
}

func (d schemaDiscovery) groupCounts(ctx context.Context, query string) (map[string]int, error) {
	var raw map[string]any
	if err := d.queryJSON(ctx, query, &raw); err != nil {
		return nil, err
	}
	counts := make(map[string]int, len(raw))
	for key, value := range raw {
		counts[key] = numberToInt(value)
	}
	return counts, nil
}

func (d schemaDiscovery) attributes(ctx context.Context, element, label string) ([]schemaAttribute, error) {
	query := fmt.Sprintf(
		"g.%s().hasLabel(%s).limit(%d).properties().key().dedup().fold()",
		element,
		gremlinString(label),
		schemaSampleSize,
	)
	var names []string
	if err := d.queryJSON(ctx, query, &names); err != nil {
		return nil, err
	}
	sort.Strings(names)
	attrs := make([]schemaAttribute, 0, len(names))
	for _, name := range names {
		if name == "" {
			continue
		}
		attrs = append(attrs, schemaAttribute{Name: name, DataType: "String"})
	}
	return attrs, nil
}

func (d schemaDiscovery) edgeConnections(ctx context.Context, edgeLabel string) ([]schemaEdgeConnection, error) {
	query := fmt.Sprintf(
		"g.E().hasLabel(%s).limit(%d).project('sourceVertexType','targetVertexType').by(outV().label()).by(inV().label()).dedup().fold()",
		gremlinString(edgeLabel),
		edgeConnectionSampleSize,
	)

	var rows []map[string]any
	if err := d.queryJSON(ctx, query, &rows); err != nil {
		return nil, err
	}

	connections := make([]schemaEdgeConnection, 0, len(rows))
	for _, row := range rows {
		source, _ := row["sourceVertexType"].(string)
		target, _ := row["targetVertexType"].(string)
		if source == "" || target == "" {
			continue
		}
		for _, sourceType := range splitCompositeLabel(source) {
			for _, targetType := range splitCompositeLabel(target) {
				connections = append(connections, schemaEdgeConnection{
					SourceVertexType: sourceType,
					EdgeType:         edgeLabel,
					TargetVertexType: targetType,
				})
			}
		}
	}
	return connections, nil
}

func (d schemaDiscovery) queryJSON(ctx context.Context, query string, target any) error {
	result, err := d.service.ExecuteQuery(ctx, query, "gremlin", core.QueryOpts{SkipFormatting: true})
	if err != nil {
		return err
	}
	content := strings.TrimSpace(result.Content)
	if content == "" {
		content = strings.TrimSpace(result.Raw)
	}
	if content == "" {
		return fmt.Errorf("empty response")
	}
	return decodeQueryJSON(content, target)
}

// Neptune represents a traversal's results as a list of traversers. Queries
// such as groupCount() and fold() produce one traverser whose value is itself a
// map or list, so accept both that wire shape and an already-unwrapped value.
func decodeQueryJSON(content string, target any) error {
	directErr := json.Unmarshal([]byte(content), target)
	if directErr == nil {
		return nil
	}

	var traversers []json.RawMessage
	if err := json.Unmarshal([]byte(content), &traversers); err == nil && len(traversers) == 1 {
		if err := json.Unmarshal(traversers[0], target); err == nil {
			return nil
		}
	}

	return fmt.Errorf("decode response %q: %w", content, directErr)
}

func (s *Server) handleSchema() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			key := s.schemaKey()
			snapshot := s.schema.get(key)
			force := r.URL.Query().Get("refresh") == "true"
			wait := r.URL.Query().Get("wait") == "true"
			shouldRefresh := force || snapshot.Status == schemaStatusEmpty || snapshot.Status == schemaStatusFailed || schemaIsStale(snapshot)
			if shouldRefresh {
				started := s.schema.start(r.Context(), key, s.queryService(), force)
				if snapshot.Status == schemaStatusEmpty || force || wait {
					snapshot = started
				}
			}
			if wait && snapshot.Status == schemaStatusRunning && s.schema.wait(r.Context(), key) {
				snapshot = s.schema.get(key)
			}
			w.Header().Set("Content-Type", contentTypeJSON)
			switch snapshot.Status {
			case schemaStatusEmpty, schemaStatusRunning:
				w.WriteHeader(http.StatusAccepted)
			case schemaStatusFailed:
				w.WriteHeader(http.StatusBadGateway)
			}
			_ = json.NewEncoder(w).Encode(snapshot)
		case http.MethodPost:
			key := s.schemaKey()
			snapshot := s.schema.start(r.Context(), key, s.queryService(), true)
			w.Header().Set("Content-Type", contentTypeJSON)
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(snapshot)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func schemaIsStale(snapshot schemaSnapshot) bool {
	if snapshot.Status != schemaStatusReady || snapshot.LastUpdate == nil {
		return false
	}
	return time.Since(*snapshot.LastUpdate) > schemaStaleAfter
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

		ch, unsubscribe := s.schema.subscribe()
		defer unsubscribe()

		_ = writeSchemaSSE(w, schemaEvent{Type: "schema.sync.snapshot", Key: s.schemaKey(), Status: s.schema.get(s.schemaKey()).Status})
		flusher.Flush()

		for {
			select {
			case <-r.Context().Done():
				return
			case event := <-ch:
				if err := writeSchemaSSE(w, event); err != nil {
					return
				}
				flusher.Flush()
			}
		}
	}
}

func writeSchemaSSE(w http.ResponseWriter, event schemaEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "event: schema\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		return err
	}
	return nil
}

func (s *Server) schemaKey() string {
	s.serviceMu.RLock()
	defer s.serviceMu.RUnlock()
	if s.profile == "" {
		return "default"
	}
	return s.profile
}

func sortedKeys(values map[string]int) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sumCounts(values map[string]int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}

func numberToInt(value any) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case json.Number:
		i, _ := v.Int64()
		return int(i)
	default:
		return 0
	}
}

func gremlinString(value string) string {
	return strconv.Quote(value)
}

func splitCompositeLabel(value string) []string {
	parts := strings.Split(value, "::")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	if len(result) == 0 {
		return []string{value}
	}
	return result
}
