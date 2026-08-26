package model

import "time"

type UnitFilter struct {
	TrenchID, Phase string
	Limit           int
	Before          *time.Time
}
type ObservationFilter struct {
	UnitID, Metric string
	From, To       *time.Time
	Limit          int
}
type RecordFilter struct {
	UnitID, Status, Author string
	Limit                  int
}
type AlertFilter struct {
	UnitID, Severity, Status string
	Limit                    int
}

func (f UnitFilter) NormalizedLimit() int {
	if f.Limit <= 0 || f.Limit > 500 {
		return 100
	}
	return f.Limit
}
func (f ObservationFilter) NormalizedLimit() int {
	if f.Limit <= 0 || f.Limit > 1000 {
		return 200
	}
	return f.Limit
}
func (f RecordFilter) NormalizedLimit() int {
	if f.Limit <= 0 || f.Limit > 500 {
		return 100
	}
	return f.Limit
}
func (f AlertFilter) NormalizedLimit() int {
	if f.Limit <= 0 || f.Limit > 500 {
		return 100
	}
	return f.Limit
}
