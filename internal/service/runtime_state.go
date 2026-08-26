package service

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"strata-weave/internal/model"
)

// TelemetryLedger keeps the latest field readings available to threshold and
// dashboard workflows without exposing mutable slices to callers.
type TelemetryLedger struct {
	mu     sync.RWMutex
	items  map[string][]model.Observation
	closed map[string]bool
}

func NewTelemetryLedger() *TelemetryLedger {
	return &TelemetryLedger{items: map[string][]model.Observation{}, closed: map[string]bool{}}
}

func (l *TelemetryLedger) Record(o model.Observation) error {
	if o.UnitID == "" || o.Metric == "" {
		return model.ErrInvalidInput
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed[o.UnitID] {
		return fmt.Errorf("unit %s telemetry is closed", o.UnitID)
	}
	l.items[o.UnitID] = append(l.items[o.UnitID], o)
	return nil
}

func (l *TelemetryLedger) Snapshot(unitID string) []model.Observation {
	l.mu.RLock()
	defer l.mu.RUnlock()
	items := append([]model.Observation(nil), l.items[unitID]...)
	sort.SliceStable(items, func(i, j int) bool { return items[i].At.Before(items[j].At) })
	return items
}

func (l *TelemetryLedger) Close(unitID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.items[unitID]; !ok {
		return model.ErrNotFound
	}
	l.closed[unitID] = true
	return nil
}

func (l *TelemetryLedger) Reopen(unitID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.items[unitID]; !ok {
		return model.ErrNotFound
	}
	l.closed[unitID] = false
	return nil
}

type ReviewQueue struct {
	mu      sync.Mutex
	queued  map[string]time.Time
	claimed map[string]string
}

func NewReviewQueue() *ReviewQueue {
	return &ReviewQueue{queued: map[string]time.Time{}, claimed: map[string]string{}}
}

func (q *ReviewQueue) Enqueue(recordID string, at time.Time) error {
	if recordID == "" {
		return model.ErrInvalidInput
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if _, ok := q.claimed[recordID]; ok {
		return model.ErrInvalidState
	}
	q.queued[recordID] = at
	return nil
}

func (q *ReviewQueue) Claim(recordID, reviewer string) error {
	if reviewer == "" {
		return model.ErrInvalidInput
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if _, ok := q.queued[recordID]; !ok {
		return model.ErrNotFound
	}
	delete(q.queued, recordID)
	q.claimed[recordID] = reviewer
	return nil
}

func (q *ReviewQueue) Complete(recordID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if _, ok := q.claimed[recordID]; !ok {
		return model.ErrNotFound
	}
	delete(q.claimed, recordID)
	return nil
}

func (q *ReviewQueue) Pending() []string {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]string, 0, len(q.queued)+len(q.claimed))
	for id := range q.queued {
		out = append(out, id)
	}
	for id := range q.claimed {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

type DispatchTracker struct {
	mu      sync.RWMutex
	states  map[string]string
	updated map[string]time.Time
}

func NewDispatchTracker() *DispatchTracker {
	return &DispatchTracker{states: map[string]string{}, updated: map[string]time.Time{}}
}

func (t *DispatchTracker) Transition(sampleID, next string, at time.Time) error {
	if sampleID == "" || next == "" {
		return model.ErrInvalidInput
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	current := t.states[sampleID]
	if current != "" && !validDispatchTransition(current, next) {
		return model.ErrInvalidState
	}
	t.states[sampleID] = next
	t.updated[sampleID] = at
	return nil
}

func (t *DispatchTracker) State(sampleID string) (string, time.Time, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	state, ok := t.states[sampleID]
	return state, t.updated[sampleID], ok
}

func validDispatchTransition(from, to string) bool {
	switch from {
	case "collected":
		return to == "dispatched"
	case "dispatched":
		return to == "at_lab"
	case "at_lab":
		return to == "reported"
	case "reported":
		return to == "archived"
	}
	return false
}

func RunWithContext(ctx context.Context, operation string, work func() error) error {
	if ctx == nil {
		return model.ErrInvalidInput
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("%s canceled before start: %w", operation, err)
	}
	done := make(chan error, 1)
	go func() { done <- work() }()
	select {
	case <-ctx.Done():
		return fmt.Errorf("%s canceled while running: %w", operation, ctx.Err())
	case err := <-done:
		return err
	}
}

func ContextCheckpoint(ctx context.Context, operation string) error {
	if ctx == nil {
		return fmt.Errorf("%s has nil context", operation)
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("%s canceled: %w", operation, err)
	}
	return nil
}

type AlertLedger struct {
	mu     sync.RWMutex
	alerts map[string]model.Alert
	order  []string
}

func NewAlertLedger() *AlertLedger { return &AlertLedger{alerts: map[string]model.Alert{}} }

func (l *AlertLedger) Upsert(a model.Alert) error {
	if a.ID == "" {
		return model.ErrInvalidInput
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, exists := l.alerts[a.ID]; exists {
		return nil
	}
	l.order = append(l.order, a.ID)
	l.alerts[a.ID] = a
	return nil
}

func (l *AlertLedger) Get(id string) (model.Alert, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	a, ok := l.alerts[id]
	if !ok {
		return model.Alert{}, model.ErrNotFound
	}
	return a, nil
}

func (l *AlertLedger) Active() []model.Alert {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]model.Alert, 0, len(l.order))
	for _, id := range l.order {
		if a, ok := l.alerts[id]; ok && a.Status != "closed" {
			out = append(out, a)
		}
	}
	return out
}
