package workflow

import (
	"fmt"
	"sort"
	"time"
)

type WorkItem struct {
	ID, UnitID, Kind, Assignee string
	Priority                   int
	DueAt                      time.Time
	Status                     string
}
type CrewSchedule struct {
	Day   time.Time
	Items []WorkItem
}
type Conflict struct {
	Left, Right string
	Reason      string
}

func BuildSchedule(day time.Time, items []WorkItem) (CrewSchedule, error) {
	s := CrewSchedule{Day: day.Truncate(24 * time.Hour), Items: append([]WorkItem(nil), items...)}
	for i := range s.Items {
		if s.Items[i].ID == "" || s.Items[i].UnitID == "" {
			return s, fmt.Errorf("work item id and unit required")
		}
		if s.Items[i].Status == "" {
			s.Items[i].Status = "planned"
		}
		if s.Items[i].Priority < 0 {
			s.Items[i].Priority = 0
		}
	}
	sort.SliceStable(s.Items, func(i, j int) bool {
		if s.Items[i].Priority == s.Items[j].Priority {
			return s.Items[i].DueAt.Before(s.Items[j].DueAt)
		}
		return s.Items[i].Priority > s.Items[j].Priority
	})
	return s, nil
}
func DetectConflicts(items []WorkItem) []Conflict {
	byCrew := map[string][]WorkItem{}
	for _, item := range items {
		if item.Assignee != "" {
			byCrew[item.Assignee] = append(byCrew[item.Assignee], item)
		}
	}
	out := []Conflict{}
	for crew, list := range byCrew {
		sort.Slice(list, func(i, j int) bool { return list[i].DueAt.Before(list[j].DueAt) })
		for i := 1; i < len(list); i++ {
			if list[i-1].DueAt.Equal(list[i].DueAt) {
				out = append(out, Conflict{Left: list[i-1].ID, Right: list[i].ID, Reason: "crew has simultaneous work items for " + crew})
			}
		}
	}
	return out
}
func MarkComplete(items []WorkItem, id string) error {
	for i := range items {
		if items[i].ID == id {
			if items[i].Status == "cancelled" {
				return fmt.Errorf("cancelled item cannot complete")
			}
			items[i].Status = "complete"
			return nil
		}
	}
	return fmt.Errorf("work item %s not found", id)
}
func Due(items []WorkItem, now time.Time) []WorkItem {
	out := []WorkItem{}
	for _, item := range items {
		if item.Status != "complete" && item.Status != "cancelled" && !item.DueAt.After(now) {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DueAt.Before(out[j].DueAt) })
	return out
}
