package workflow

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type Alert struct {
	ID, UnitID, Severity, Message, Status string
	CreatedAt                             time.Time
	Occurrences                           int
}
type AlertGroup struct {
	UnitID, Key string
	Severity    string
	Count       int
	First, Last time.Time
}

func DedupKey(a Alert) string {
	return strings.ToLower(strings.TrimSpace(a.UnitID)) + "|" + strings.ToLower(strings.TrimSpace(a.Message))
}
func MergeAlerts(alerts []Alert) []Alert {
	groups := map[string]Alert{}
	for _, a := range alerts {
		k := DedupKey(a)
		if old, ok := groups[k]; ok {
			old.Occurrences += a.Occurrences + 1
			if a.CreatedAt.Before(old.CreatedAt) {
				old.CreatedAt = a.CreatedAt
			}
			if severityRank(a.Severity) > severityRank(old.Severity) {
				old.Severity = a.Severity
			}
			groups[k] = old
		} else {
			if a.Occurrences == 0 {
				a.Occurrences = 1
			}
			groups[k] = a
		}
	}
	out := make([]Alert, 0, len(groups))
	for _, a := range groups {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}
func severityRank(s string) int {
	switch strings.ToLower(s) {
	case "critical":
		return 3
	case "warning":
		return 2
	case "info":
		return 1
	}
	return 0
}
func Escalate(a Alert, now time.Time, after time.Duration) (Alert, error) {
	if a.Status == "closed" {
		return a, fmt.Errorf("closed alert cannot escalate")
	}
	if now.Sub(a.CreatedAt) < after {
		return a, nil
	}
	if severityRank(a.Severity) < 3 {
		a.Severity = "critical"
	}
	return a, nil
}
