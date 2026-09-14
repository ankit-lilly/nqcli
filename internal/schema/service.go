// Package schema discovers and caches a property graph schema independently of
// any transport. HTTP and desktop adapters can expose the same service.
package schema

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ankit-lilly/nqcli/internal/core"
	"golang.org/x/sync/errgroup"
)

type Status string

const (
	StatusEmpty   Status = "empty"
	StatusRunning Status = "running"
	StatusReady   Status = "ready"
	StatusFailed  Status = "failed"

	sampleSize           = 500
	connectionSampleSize = 500
	queryConcurrency     = 4
	maxResponseBytes     = 8 << 20
	refreshTimeout       = 5 * time.Minute
	staleAfter           = 10 * time.Minute
)

type Attribute struct {
	Name     string `json:"name"`
	DataType string `json:"dataType"`
}

type Vertex struct {
	Type       string      `json:"type"`
	Attributes []Attribute `json:"attributes"`
	Total      *int        `json:"total,omitempty"`
}

type Edge struct {
	Type       string      `json:"type"`
	Attributes []Attribute `json:"attributes"`
	Total      *int        `json:"total,omitempty"`
}

type EdgeConnection struct {
	SourceVertexType string `json:"sourceVertexType"`
	EdgeType         string `json:"edgeType"`
	TargetVertexType string `json:"targetVertexType"`
}

type Snapshot struct {
	TotalVertices      *int             `json:"totalVertices,omitempty"`
	Vertices           []Vertex         `json:"vertices"`
	TotalEdges         *int             `json:"totalEdges,omitempty"`
	Edges              []Edge           `json:"edges"`
	EdgeConnections    []EdgeConnection `json:"edgeConnections"`
	Status             Status           `json:"status"`
	Phase              string           `json:"phase,omitempty"`
	Completed          int              `json:"completed,omitempty"`
	Total              int              `json:"total,omitempty"`
	Error              string           `json:"error,omitempty"`
	LastUpdate         *time.Time       `json:"lastUpdate,omitempty"`
	LastRefreshStarted *time.Time       `json:"-"`
}

type Event struct {
	Type      string `json:"type"`
	Key       string `json:"key"`
	Status    Status `json:"status"`
	Phase     string `json:"phase,omitempty"`
	Completed int    `json:"completed,omitempty"`
	Total     int    `json:"total,omitempty"`
	Error     string `json:"error,omitempty"`
}

type Logger interface {
	Warn(msg any, keyvals ...any)
}

var ErrCacheMiss = errors.New("schema cache miss")

// Store persists completed snapshots without coupling discovery to a storage transport.
type Store interface {
	Load(key string) (Snapshot, error)
	Save(key string, snapshot Snapshot) error
}

type Option func(*Service)

func WithStore(store Store) Option {
	return func(service *Service) {
		service.store = store
	}
}

// Service coordinates one cached schema snapshot per connection key.
type Service struct {
	mu        sync.RWMutex
	snapshots map[string]Snapshot
	running   map[string]chan struct{}
	subs      map[chan Event]struct{}
	loaded    map[string]bool
	logger    Logger
	store     Store
	ctx       context.Context
	cancel    context.CancelFunc
	workers   sync.WaitGroup
	closed    bool
}

func NewService(logger Logger, options ...Option) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	service := &Service{
		snapshots: make(map[string]Snapshot),
		running:   make(map[string]chan struct{}),
		subs:      make(map[chan Event]struct{}),
		loaded:    make(map[string]bool),
		logger:    logger,
		ctx:       ctx,
		cancel:    cancel,
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func (s *Service) Close() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	s.cancel()
	s.mu.Unlock()
	s.workers.Wait()
}

func EmptySnapshot(status Status) Snapshot {
	return Snapshot{Vertices: []Vertex{}, Edges: []Edge{}, EdgeConnections: []EdgeConnection{}, Status: status}
}

func (s *Service) Get(key string) Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snapshot, ok := s.snapshots[key]
	if !ok {
		return EmptySnapshot(StatusEmpty)
	}
	return snapshot
}

func (s *Service) IsStale(snapshot Snapshot) bool {
	return snapshot.Status == StatusReady && snapshot.LastUpdate != nil && time.Since(*snapshot.LastUpdate) > staleAfter
}

// Revalidate starts a refresh when data is absent, failed, stale, or forced.
// It returns immediately and preserves any cached schema while work continues.
func (s *Service) Revalidate(_ context.Context, key string, query core.QueryService, force bool) Snapshot {
	s.loadCached(key)
	existing := s.Get(key)
	if !force && existing.Status != StatusEmpty && existing.Status != StatusFailed && !s.IsStale(existing) {
		return existing
	}

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return existing
	}
	if _, running := s.running[key]; running {
		snapshot := s.snapshots[key]
		s.mu.Unlock()
		return snapshot
	}
	running := existing
	running.Status = StatusRunning
	running.Phase = "starting"
	running.Error = ""
	running.Completed = 0
	running.Total = 0
	now := time.Now()
	running.LastRefreshStarted = &now
	s.snapshots[key] = running
	s.running[key] = make(chan struct{})
	s.publishLocked(Event{Type: "schema.sync.started", Key: key, Status: StatusRunning, Phase: "starting"})
	s.workers.Go(func() {
		s.runRefresh(key, query)
	})
	s.mu.Unlock()
	return running
}

func (s *Service) loadCached(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.loaded[key] {
		return
	}
	s.loaded[key] = true
	if s.store == nil {
		return
	}
	snapshot, err := s.store.Load(key)
	if err == nil {
		s.snapshots[key] = snapshot
		return
	}
	if !errors.Is(err, ErrCacheMiss) && s.logger != nil {
		s.logger.Warn("failed to load schema cache", "key", key, "error", err)
	}
}

func (s *Service) Wait(ctx context.Context, key string) bool {
	s.mu.RLock()
	done := s.running[key]
	s.mu.RUnlock()
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

func (s *Service) Subscribe() (<-chan Event, func()) {
	ch := make(chan Event, 16)
	s.mu.Lock()
	s.subs[ch] = struct{}{}
	s.mu.Unlock()
	return ch, func() {
		s.mu.Lock()
		if _, exists := s.subs[ch]; exists {
			delete(s.subs, ch)
			close(ch)
		}
		s.mu.Unlock()
	}
}

func (s *Service) runRefresh(key string, query core.QueryService) {
	ctx, cancel := context.WithTimeout(s.ctx, refreshTimeout)
	defer cancel()
	discovery := discovery{query: query, onProgress: func(phase string, completed, total int) {
		s.updateProgress(key, phase, completed, total)
	}}
	snapshot, err := discovery.discover(ctx)

	s.mu.Lock()
	done := s.running[key]
	if err != nil {
		existing := s.snapshots[key]
		existing.Status = StatusFailed
		existing.Error = err.Error()
		s.snapshots[key] = existing
		s.publishLocked(Event{Type: "schema.sync.failed", Key: key, Status: StatusFailed, Error: err.Error()})
		s.mu.Unlock()
		if s.logger != nil {
			s.logger.Warn("schema sync failed", "key", key, "error", err)
		}
		s.finishRefresh(key, done)
		return
	}
	now := time.Now()
	snapshot.Status = StatusReady
	snapshot.Phase = "complete"
	snapshot.LastUpdate = &now
	s.snapshots[key] = snapshot
	s.publishLocked(Event{Type: "schema.sync.completed", Key: key, Status: StatusReady, Phase: "complete"})
	s.mu.Unlock()
	if s.store != nil {
		if err := s.store.Save(key, snapshot); err != nil && s.logger != nil {
			s.logger.Warn("failed to save schema cache", "key", key, "error", err)
		}
	}
	s.finishRefresh(key, done)
}

func (s *Service) finishRefresh(key string, done chan struct{}) {
	if done == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running[key] != done {
		return
	}
	delete(s.running, key)
	close(done)
}

func (s *Service) updateProgress(key, phase string, completed, total int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snapshot := s.snapshots[key]
	snapshot.Status = StatusRunning
	snapshot.Phase = phase
	snapshot.Completed = completed
	snapshot.Total = total
	s.snapshots[key] = snapshot
	s.publishLocked(Event{Type: "schema.sync.progress", Key: key, Status: StatusRunning, Phase: phase, Completed: completed, Total: total})
}

func (s *Service) publish(event Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.publishLocked(event)
}

func (s *Service) publishLocked(event Event) {
	for ch := range s.subs {
		select {
		case ch <- event:
		default:
		}
	}
}

type discovery struct {
	query      core.QueryService
	onProgress func(string, int, int)
}

func (d discovery) discover(ctx context.Context) (Snapshot, error) {
	d.progress("labels", 0, 2)
	var vertexCounts, edgeCounts map[string]int
	labels, labelsCtx := errgroup.WithContext(ctx)
	var labelsCompleted atomic.Int64
	labels.Go(func() error {
		counts, err := d.groupCounts(labelsCtx, "g.V().label().groupCount()")
		if err != nil {
			return fmt.Errorf("discover vertex labels: %w", err)
		}
		vertexCounts = counts
		d.progress("labels", int(labelsCompleted.Add(1)), 2)
		return nil
	})
	labels.Go(func() error {
		counts, err := d.groupCounts(labelsCtx, "g.E().label().groupCount()")
		if err != nil {
			return fmt.Errorf("discover edge labels: %w", err)
		}
		edgeCounts = counts
		d.progress("labels", int(labelsCompleted.Add(1)), 2)
		return nil
	})
	if err := labels.Wait(); err != nil {
		return Snapshot{}, err
	}

	vertexLabels, edgeLabels := sortedKeys(vertexCounts), sortedKeys(edgeCounts)
	vertices := make([]Vertex, len(vertexLabels))
	edges := make([]Edge, len(edgeLabels))
	connectionsByLabel := make([][]EdgeConnection, len(edgeLabels))
	totalDetails := len(vertexLabels) + 2*len(edgeLabels)
	d.progress("details", 0, totalDetails)
	var detailsCompleted atomic.Int64
	details, detailsCtx := errgroup.WithContext(ctx)
	details.SetLimit(queryConcurrency)
	for index := range max(len(vertexLabels), len(edgeLabels)) {
		if index < len(vertexLabels) {
			label := vertexLabels[index]
			details.Go(func() error {
				attributes, err := d.attributes(detailsCtx, "V", label)
				if err != nil {
					return fmt.Errorf("discover vertex properties for %q: %w", label, err)
				}
				total := vertexCounts[label]
				vertices[index] = Vertex{Type: label, Attributes: attributes, Total: &total}
				d.progress("details", int(detailsCompleted.Add(1)), totalDetails)
				return nil
			})
		}
		if index < len(edgeLabels) {
			label := edgeLabels[index]
			details.Go(func() error {
				attributes, err := d.attributes(detailsCtx, "E", label)
				if err != nil {
					return fmt.Errorf("discover edge properties for %q: %w", label, err)
				}
				total := edgeCounts[label]
				edges[index] = Edge{Type: label, Attributes: attributes, Total: &total}
				d.progress("details", int(detailsCompleted.Add(1)), totalDetails)
				return nil
			})
			details.Go(func() error {
				connections, err := d.edgeConnections(detailsCtx, label)
				if err != nil {
					return fmt.Errorf("discover edge connections for %q: %w", label, err)
				}
				connectionsByLabel[index] = connections
				d.progress("details", int(detailsCompleted.Add(1)), totalDetails)
				return nil
			})
		}
	}
	if err := details.Wait(); err != nil {
		return Snapshot{}, err
	}

	connections := make([]EdgeConnection, 0, len(edgeLabels))
	seen := make(map[string]bool)
	for _, group := range connectionsByLabel {
		for _, connection := range group {
			key := connection.SourceVertexType + "\x00" + connection.EdgeType + "\x00" + connection.TargetVertexType
			if !seen[key] {
				seen[key] = true
				connections = append(connections, connection)
			}
		}
	}
	totalVertices, totalEdges := sumCounts(vertexCounts), sumCounts(edgeCounts)
	return Snapshot{TotalVertices: &totalVertices, Vertices: vertices, TotalEdges: &totalEdges, Edges: edges, EdgeConnections: connections}, nil
}

func (d discovery) progress(phase string, completed, total int) {
	if d.onProgress != nil {
		d.onProgress(phase, completed, total)
	}
}

func (d discovery) groupCounts(ctx context.Context, query string) (map[string]int, error) {
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

func (d discovery) attributes(ctx context.Context, element, label string) ([]Attribute, error) {
	query := fmt.Sprintf("g.%s().hasLabel(%s).limit(%d).properties().key().dedup().fold()", element, strconv.Quote(label), sampleSize)
	var names []string
	if err := d.queryJSON(ctx, query, &names); err != nil {
		return nil, err
	}
	sort.Strings(names)
	attributes := make([]Attribute, 0, len(names))
	for _, name := range names {
		if name != "" {
			attributes = append(attributes, Attribute{Name: name, DataType: "String"})
		}
	}
	return attributes, nil
}

func (d discovery) edgeConnections(ctx context.Context, label string) ([]EdgeConnection, error) {
	query := fmt.Sprintf("g.E().hasLabel(%s).limit(%d).project('sourceVertexType','targetVertexType').by(outV().label()).by(inV().label()).dedup().fold()", strconv.Quote(label), connectionSampleSize)
	var rows []map[string]any
	if err := d.queryJSON(ctx, query, &rows); err != nil {
		return nil, err
	}
	connections := make([]EdgeConnection, 0, len(rows))
	for _, row := range rows {
		source, _ := row["sourceVertexType"].(string)
		target, _ := row["targetVertexType"].(string)
		if source == "" || target == "" {
			continue
		}
		for _, sourceType := range splitCompositeLabel(source) {
			for _, targetType := range splitCompositeLabel(target) {
				connections = append(connections, EdgeConnection{SourceVertexType: sourceType, EdgeType: label, TargetVertexType: targetType})
			}
		}
	}
	return connections, nil
}

func (d discovery) queryJSON(ctx context.Context, query string, target any) error {
	result, err := d.query.ExecuteQuery(ctx, query, "gremlin", core.QueryOpts{
		SkipFormatting:   true,
		MaxResponseBytes: maxResponseBytes,
	})
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

func decodeQueryJSON(content string, target any) error {
	return decodeQueryPayload(json.RawMessage(content), target)
}

func decodeQueryPayload(payload json.RawMessage, target any) error {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(payload, &envelope); err == nil {
		data, hasData := envelope["data"]
		_, hasMeta := envelope["meta"]
		if hasData && hasMeta {
			return decodeQueryPayload(data, target)
		}
	}

	directErr := json.Unmarshal(payload, target)
	if directErr == nil {
		return nil
	}
	var traversers []json.RawMessage
	if err := json.Unmarshal(payload, &traversers); err == nil && len(traversers) == 1 {
		if err := json.Unmarshal(traversers[0], target); err == nil {
			return nil
		}
	}
	return fmt.Errorf("decode response %q: %w", string(payload), directErr)
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
	switch value := value.(type) {
	case float64:
		return int(value)
	case int:
		return value
	case int64:
		return int(value)
	case json.Number:
		result, _ := value.Int64()
		return int(result)
	default:
		return 0
	}
}

func splitCompositeLabel(value string) []string {
	parts := strings.Split(value, "::")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			result = append(result, part)
		}
	}
	if len(result) == 0 {
		return []string{value}
	}
	return result
}
