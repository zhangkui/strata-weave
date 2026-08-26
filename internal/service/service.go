package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"strata-weave/internal/model"
	"strata-weave/internal/store"
)

type Service struct {
	db         *sql.DB
	importMu   sync.Mutex
	relationMu sync.Mutex
	ledger     *TelemetryLedger
	reviews    *ReviewQueue
	dispatch   *DispatchTracker
	alerts     *AlertLedger
}

func New(db *sql.DB) *Service {
	return &Service{db: db, ledger: NewTelemetryLedger(), reviews: NewReviewQueue(), dispatch: NewDispatchTracker(), alerts: NewAlertLedger()}
}
func id(prefix string) string { return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano()) }
func now() time.Time          { return time.Now().UTC() }

func (s *Service) CreateTrench(_ context.Context, t model.Trench) (model.Trench, error) {
	if err := model.ValidateTrench(t); err != nil {
		return t, err
	}
	if t.ID == "" {
		t.ID = id("trench")
	}
	if t.OpenedAt.IsZero() {
		t.OpenedAt = now()
	}
	return t, store.CreateTrench(s.db, t)
}
func (s *Service) ListTrenches(_ context.Context) ([]model.Trench, error) {
	return store.ListTrenches(s.db)
}
func (s *Service) CloseTrench(_ context.Context, id string) error { return store.CloseTrench(s.db, id) }
func (s *Service) CreateUnit(_ context.Context, u model.Unit) (model.Unit, error) {
	if err := model.ValidateUnit(u); err != nil {
		return u, err
	}
	if u.ID == "" {
		u.ID = id("unit")
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now()
	}
	if _, e := store.GetTrench(s.db, u.TrenchID); e != nil {
		return u, fmt.Errorf("trench: %w", e)
	}
	return u, store.CreateUnit(s.db, u)
}
func (s *Service) ListUnits(_ context.Context, f model.UnitFilter) ([]model.Unit, error) {
	return store.ListUnits(s.db, f)
}
func (s *Service) AdvanceUnit(_ context.Context, id string, next int) error {
	u, e := store.GetUnit(s.db, id)
	if e != nil {
		return e
	}
	if next < u.Phase || next > u.Phase+1 {
		return model.ErrInvalidState
	}
	return store.UpdateUnitPhase(s.db, id, next)
}

func (s *Service) AddStratigraphicRelation(_ context.Context, r model.Relation) (model.Relation, error) {
	s.relationMu.Lock()
	defer s.relationMu.Unlock()
	if r.EarlierID == r.LaterID {
		return r, model.ErrCycle
	}
	a, e := store.GetUnit(s.db, r.EarlierID)
	if e != nil {
		return r, e
	}
	b, e := store.GetUnit(s.db, r.LaterID)
	if e != nil {
		return r, e
	}
	if a.TrenchID != b.TrenchID {
		return r, model.ErrCrossTrench
	}
	cycle, e := store.HasPath(s.db, r.LaterID, r.EarlierID)
	if e != nil {
		return r, e
	}
	if cycle {
		return r, model.ErrCycle
	}
	if r.ID == "" {
		r.ID = id("rel")
	}
	if r.CreatedAt.IsZero() {
		r.CreatedAt = now()
	}
	return r, store.AddRelation(s.db, r)
}
func (s *Service) Matrix(_ context.Context) ([]model.Relation, error) {
	return store.ListRelations(s.db)
}

func (s *Service) CreateFind(_ context.Context, f model.Find) (model.Find, error) {
	if err := model.ValidateFind(f); err != nil {
		return f, err
	}
	if _, e := store.GetUnit(s.db, f.UnitID); e != nil {
		return f, e
	}
	if f.ID == "" {
		f.ID = id("find")
	}
	if f.CreatedAt.IsZero() {
		f.CreatedAt = now()
	}
	return f, store.CreateFind(s.db, f)
}
func (s *Service) ReviewFind(_ context.Context, id string) error { return store.ReviewFind(s.db, id) }
func (s *Service) CreateSample(_ context.Context, sample model.Sample) (model.Sample, error) {
	if err := model.ValidateSample(sample); err != nil {
		return sample, err
	}
	if _, e := store.GetFind(s.db, sample.FindID); e != nil {
		return sample, e
	}
	if sample.ID == "" {
		sample.ID = id("sample")
	}
	if sample.CollectedAt.IsZero() {
		sample.CollectedAt = now()
	}
	if sample.Status == "" {
		sample.Status = model.SampleCollected
	}
	if err := store.CreateSample(s.db, sample); err != nil {
		return sample, err
	}
	return sample, s.dispatch.Transition(sample.ID, model.SampleCollected, sample.CollectedAt)
}
func (s *Service) DispatchSample(ctx context.Context, id, lab string) error {
	if err := ContextCheckpoint(ctx, "dispatch sample"); err != nil {
		return err
	}
	sample, e := store.GetSample(s.db, id)
	if e != nil {
		return e
	}
	f, e := store.GetFind(s.db, sample.FindID)
	if e != nil {
		return e
	}
	if !f.Reviewed {
		return model.ErrUnreviewedFind
	}
	if lab == "" {
		return model.ErrInvalidInput
	}
	if err := store.DispatchSample(s.db, id, lab); err != nil {
		return err
	}
	return s.dispatch.Transition(id, model.SampleDispatched, now())
}
func (s *Service) ReportSample(ctx context.Context, id, method string, age, errorBP float64) error {
	if err := ContextCheckpoint(ctx, "report sample"); err != nil {
		return err
	}
	if method == "" || age < 0 || errorBP < 0 {
		return model.ErrInvalidInput
	}
	if err := store.ReportSample(s.db, id, method, age, errorBP, now().Format(time.RFC3339Nano)); err != nil {
		return err
	}
	return s.dispatch.Transition(id, model.SampleReported, now())
}

func (s *Service) CreateRecord(_ context.Context, r model.FieldRecord) (model.FieldRecord, error) {
	if r.UnitID == "" || r.Author == "" || r.Notes == "" {
		return r, model.ErrInvalidInput
	}
	if _, e := store.GetUnit(s.db, r.UnitID); e != nil {
		return r, e
	}
	if r.ID == "" {
		r.ID = id("record")
	}
	if r.Status == "" {
		r.Status = model.RecordDraft
	}
	return r, store.CreateFieldRecord(s.db, r)
}
func (s *Service) SubmitRecord(ctx context.Context, id string) error {
	if err := ContextCheckpoint(ctx, "submit record"); err != nil {
		return err
	}
	if err := store.SubmitRecord(s.db, id, now().Format(time.RFC3339Nano)); err != nil {
		return err
	}
	return s.reviews.Enqueue(id, now())
}
func (s *Service) ReviewRecord(ctx context.Context, id string, approved bool, note string) error {
	if err := ContextCheckpoint(ctx, "review record"); err != nil {
		return err
	}
	if strings.TrimSpace(note) == "" {
		return model.ErrInvalidInput
	}
	status := model.RecordRejected
	if approved {
		status = model.RecordReviewed
	}
	if err := s.reviews.Claim(id, "reviewer"); err != nil {
		return err
	}
	if err := store.ReviewRecord(s.db, id, status, note, now().Format(time.RFC3339Nano)); err != nil {
		return err
	}
	return s.reviews.Complete(id)
}
func (s *Service) ListRecords(_ context.Context, f model.RecordFilter) ([]model.FieldRecord, error) {
	return store.ListRecords(s.db, f)
}

func (s *Service) IngestObservation(ctx context.Context, o model.Observation) error {
	if err := ContextCheckpoint(ctx, "ingest observation"); err != nil {
		return err
	}
	if e := model.ValidateObservation(o); e != nil {
		return e
	}
	if _, e := store.GetUnit(s.db, o.UnitID); e != nil {
		return e
	}
	if o.ID == "" {
		o.ID = id("obs")
	}
	if o.Quality == "" {
		o.Quality = "unverified"
	}
	if err := store.InsertObservation(s.db, o); err != nil {
		return err
	}
	return s.ledger.Record(o)
}
func (s *Service) IngestBatch(ctx context.Context, items []model.Observation) (int, []error) {
	s.importMu.Lock()
	defer s.importMu.Unlock()
	accepted := []model.Observation{}
	errs := []error{}
	for _, o := range items {
		select {
		case <-ctx.Done():
			errs = append(errs, ctx.Err())
			return len(accepted), errs
		default:
		}
		if e := model.ValidateObservation(o); e != nil {
			errs = append(errs, e)
			continue
		}
		if o.ID == "" {
			o.ID = id("obs")
		}
		if o.Quality == "" {
			o.Quality = "unverified"
		}
		accepted = append(accepted, o)
	}
	if e := store.InsertObservationsTxContext(ctx, s.db, accepted); e != nil {
		return 0, []error{e}
	}
	for _, item := range accepted {
		if err := s.ledger.Record(item); err != nil {
			return 0, []error{err}
		}
	}
	return len(accepted), errs
}
func (s *Service) ListObservations(_ context.Context, f model.ObservationFilter) ([]model.Observation, error) {
	return store.ListObservations(s.db, f)
}

// TelemetrySnapshot exposes an immutable operational view for field monitors.
func (s *Service) TelemetrySnapshot(unitID string) []model.Observation {
	return s.ledger.Snapshot(unitID)
}

// PendingReviewIDs provides the review desk with outstanding record identifiers.
func (s *Service) PendingReviewIDs() []string { return s.reviews.Pending() }
func (s *Service) CreateAlert(_ context.Context, a model.Alert) (model.Alert, error) {
	if a.UnitID == "" || a.Message == "" || a.Severity == "" {
		return a, model.ErrInvalidInput
	}
	if _, e := store.GetUnit(s.db, a.UnitID); e != nil {
		return a, e
	}
	if a.ID == "" {
		a.ID = id("alert")
	}
	if a.Status == "" {
		a.Status = "open"
	}
	if a.CreatedAt == nil {
		t := now()
		a.CreatedAt = &t
	}
	if err := store.CreateAlert(s.db, a); err != nil {
		return a, err
	}
	return a, s.alerts.Upsert(a)
}
func (s *Service) ListAlerts(_ context.Context, f model.AlertFilter) ([]model.Alert, error) {
	return store.ListAlerts(s.db, f)
}
func (s *Service) CloseAlert(_ context.Context, id string) error {
	if err := store.CloseAlert(s.db, id, now().Format(time.RFC3339Nano)); err != nil {
		return err
	}
	a, err := s.alerts.Get(id)
	if err != nil {
		return err
	}
	if err := s.alerts.Upsert(a); err != nil {
		return err
	}
	a.Status = "closed"
	return s.alerts.Upsert(a)
}

// ActiveRuntimeAlerts exposes the in-process alert feed used by live monitors.
func (s *Service) ActiveRuntimeAlerts() []model.Alert { return s.alerts.Active() }
func (s *Service) Dashboard(_ context.Context) (model.Dashboard, error) {
	var d model.Dashboard
	e := s.db.QueryRow(`SELECT (SELECT count(*) FROM trenches),(SELECT count(*) FROM units WHERE phase<5),(SELECT count(*) FROM records WHERE status='submitted'),(SELECT count(*) FROM alerts WHERE status='open'),(SELECT count(*) FROM samples WHERE status='dispatched')`).Scan(&d.Trenches, &d.OpenUnits, &d.PendingReviews, &d.ActiveAlerts, &d.SamplesInLab)
	return d, e
}
