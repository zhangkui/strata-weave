package model

import "time"

type Trench struct {
	ID, Code, Site, Description string
	OpenedAt                    time.Time
	Closed                      bool
}
type Unit struct {
	ID, TrenchID, Code, Label, Description string
	Phase                                  int
	CreatedAt                              time.Time
}
type Relation struct {
	ID, EarlierID, LaterID, Note string
	CreatedAt                    time.Time
}
type Find struct {
	ID, UnitID, CatalogueNo, Kind, Material string
	Condition                               string
	Reviewed                                bool
	CreatedAt                               time.Time
}
type Sample struct {
	ID, FindID, Label, LabCode, Status string
	CollectedAt                        time.Time
	Result                             *DatingResult
}
type DatingResult struct {
	Method     string
	AgeBP      float64
	ErrorBP    float64
	ReportedAt time.Time
}
type FieldRecord struct {
	ID, UnitID, Author, Notes, Status, ReviewNote string
	SubmittedAt, ReviewedAt                       *time.Time
}
type Observation struct {
	ID, UnitID, Instrument, Metric string
	Value                          float64
	At                             time.Time
	Quality                        string
}
type Alert struct {
	ID, UnitID, Severity, Message, Status string
	CreatedAt, ClosedAt                   *time.Time
}
type Dashboard struct {
	Trenches       int `json:"trenches"`
	OpenUnits      int `json:"open_units"`
	PendingReviews int `json:"pending_reviews"`
	ActiveAlerts   int `json:"active_alerts"`
	SamplesInLab   int `json:"samples_in_lab"`
}

const (
	RecordDraft      = "draft"
	RecordSubmitted  = "submitted"
	RecordReviewed   = "reviewed"
	RecordRejected   = "rejected"
	SampleCollected  = "collected"
	SampleDispatched = "dispatched"
	SampleReported   = "reported"
)
