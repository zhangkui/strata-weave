package workflow

import (
	"fmt"
	"sort"
	"time"
)

type UnitReport struct {
	UnitID                                      string
	ObservationCount, FindCount, OpenAlertCount int
	LastObservation                             *time.Time
	Quality                                     string
}
type ReportFilter struct {
	Since         *time.Time
	IncludeClosed bool
}

func BuildReport(units []string, observations map[string][]time.Time, finds map[string]int, alerts map[string][]Alert, filter ReportFilter) []UnitReport {
	out := make([]UnitReport, 0, len(units))
	for _, unit := range units {
		r := UnitReport{UnitID: unit, FindCount: finds[unit], Quality: "ready"}
		for _, at := range observations[unit] {
			if filter.Since != nil && at.Before(*filter.Since) {
				continue
			}
			r.ObservationCount++
			if r.LastObservation == nil || at.After(*r.LastObservation) {
				t := at
				r.LastObservation = &t
			}
		}
		for _, a := range alerts[unit] {
			if !filter.IncludeClosed && a.Status == "closed" {
				continue
			}
			r.OpenAlertCount++
		}
		if r.ObservationCount == 0 {
			r.Quality = "needs-data"
		}
		if r.OpenAlertCount > 0 {
			r.Quality = "attention"
		}
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UnitID < out[j].UnitID })
	return out
}
func ValidateReport(r UnitReport) error {
	if r.UnitID == "" {
		return fmt.Errorf("report unit missing")
	}
	if r.ObservationCount < 0 || r.FindCount < 0 || r.OpenAlertCount < 0 {
		return fmt.Errorf("negative report count")
	}
	return nil
}
