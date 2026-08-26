package archaeology

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type Period struct {
	Code       string
	Name       string
	StartYear  int
	EndYear    int
	Confidence float64
}

type PeriodAssignment struct {
	UnitID     string
	PeriodCode string
	Evidence   []string
	AssignedAt time.Time
	Reviewer   string
}

type PeriodCatalogue struct{ entries map[string]Period }

func NewPeriodCatalogue(periods []Period) (*PeriodCatalogue, error) {
	c := &PeriodCatalogue{entries: map[string]Period{}}
	for _, p := range periods {
		if err := validatePeriod(p); err != nil {
			return nil, err
		}
		if _, exists := c.entries[p.Code]; exists {
			return nil, fmt.Errorf("duplicate period code %q", p.Code)
		}
		c.entries[p.Code] = p
	}
	return c, nil
}

func validatePeriod(p Period) error {
	if strings.TrimSpace(p.Code) == "" || strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("period code and name are required")
	}
	if p.StartYear > p.EndYear {
		return fmt.Errorf("period %s has reversed range", p.Code)
	}
	if p.Confidence < 0 || p.Confidence > 1 {
		return fmt.Errorf("period %s confidence outside range", p.Code)
	}
	return nil
}

func (c *PeriodCatalogue) Get(code string) (Period, bool) { p, ok := c.entries[code]; return p, ok }
func (c *PeriodCatalogue) List() []Period {
	out := make([]Period, 0, len(c.entries))
	for _, p := range c.entries {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartYear < out[j].StartYear })
	return out
}

func (c *PeriodCatalogue) Overlaps(code string, from, to int) (Period, bool) {
	p, ok := c.entries[code]
	if !ok {
		return Period{}, false
	}
	return p, p.StartYear <= to && from <= p.EndYear
}

func (c *PeriodCatalogue) CandidatePeriods(from, to int, minimum float64) []Period {
	result := []Period{}
	for _, p := range c.entries {
		if p.Confidence >= minimum && p.StartYear <= to && from <= p.EndYear {
			result = append(result, p)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Confidence == result[j].Confidence {
			return result[i].Code < result[j].Code
		}
		return result[i].Confidence > result[j].Confidence
	})
	return result
}

func ValidateAssignment(a PeriodAssignment, catalogue *PeriodCatalogue) error {
	if strings.TrimSpace(a.UnitID) == "" || strings.TrimSpace(a.PeriodCode) == "" {
		return fmt.Errorf("unit and period are required")
	}
	if _, ok := catalogue.Get(a.PeriodCode); !ok {
		return fmt.Errorf("unknown period %q", a.PeriodCode)
	}
	if len(a.Evidence) == 0 {
		return fmt.Errorf("period assignment requires evidence")
	}
	seen := map[string]bool{}
	for _, item := range a.Evidence {
		item = strings.TrimSpace(item)
		if item == "" {
			return fmt.Errorf("empty period evidence")
		}
		if seen[item] {
			return fmt.Errorf("duplicate period evidence %q", item)
		}
		seen[item] = true
	}
	return nil
}
