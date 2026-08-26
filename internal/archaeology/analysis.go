package archaeology

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

type Evidence struct {
	Source, Kind       string
	Weight, Confidence float64
	CollectedAt        time.Time
}
type ChronologyCandidate struct {
	Period         Period
	Score, Support float64
	Evidence       []Evidence
}
type ChronologyResult struct {
	UnitID     string
	Candidates []ChronologyCandidate
	Selected   string
	Confidence float64
	Warnings   []string
}
type AnalysisConfig struct {
	MinimumEvidence  float64
	MinimumSelection float64
	AllowUncertain   bool
}

func ValidateEvidence(e Evidence) error {
	if strings.TrimSpace(e.Source) == "" || strings.TrimSpace(e.Kind) == "" {
		return fmt.Errorf("evidence source and kind are required")
	}
	if e.Weight <= 0 || e.Confidence < 0 || e.Confidence > 1 {
		return fmt.Errorf("invalid evidence weight or confidence")
	}
	if e.CollectedAt.IsZero() {
		return fmt.Errorf("evidence date is required")
	}
	return nil
}
func RankCandidates(catalogue *PeriodCatalogue, unit string, from, to int, evidence []Evidence, cfg AnalysisConfig) (ChronologyResult, error) {
	result := ChronologyResult{UnitID: unit}
	if strings.TrimSpace(unit) == "" {
		return result, fmt.Errorf("unit id is required")
	}
	if from > to {
		return result, fmt.Errorf("reversed chronology range")
	}
	if cfg.MinimumEvidence <= 0 {
		cfg.MinimumEvidence = .15
	}
	if cfg.MinimumSelection <= 0 {
		cfg.MinimumSelection = .5
	}
	valid := []Evidence{}
	for _, e := range evidence {
		if err := ValidateEvidence(e); err != nil {
			return result, err
		}
		valid = append(valid, e)
	}
	periods := catalogue.CandidatePeriods(from, to, 0)
	if len(periods) == 0 {
		return result, fmt.Errorf("no periods overlap range")
	}
	for _, period := range periods {
		score := 0.0
		support := 0.0
		for _, ev := range valid {
			factor := ev.Weight * ev.Confidence
			if strings.EqualFold(ev.Kind, "ceramic") && period.StartYear < 0 {
				factor *= 1.1
			}
			if strings.EqualFold(ev.Kind, "radiocarbon") && period.StartYear > 0 {
				factor *= .9
			}
			score += factor
			support += ev.Weight
		}
		if support > 0 {
			score /= support
		}
		if score >= cfg.MinimumEvidence {
			result.Candidates = append(result.Candidates, ChronologyCandidate{Period: period, Score: score, Support: support, Evidence: append([]Evidence(nil), valid...)})
		}
	}
	sort.Slice(result.Candidates, func(i, j int) bool {
		if result.Candidates[i].Score == result.Candidates[j].Score {
			return result.Candidates[i].Period.Code < result.Candidates[j].Period.Code
		}
		return result.Candidates[i].Score > result.Candidates[j].Score
	})
	if len(result.Candidates) == 0 {
		result.Warnings = append(result.Warnings, "evidence does not support any candidate")
	} else {
		best := result.Candidates[0]
		result.Selected = best.Period.Code
		result.Confidence = best.Score
		if best.Score < cfg.MinimumSelection && !cfg.AllowUncertain {
			result.Warnings = append(result.Warnings, "best candidate remains uncertain")
		}
	}
	return result, nil
}
func ExplainChronology(r ChronologyResult) string {
	if len(r.Candidates) == 0 {
		return "no chronology candidate"
	}
	parts := make([]string, 0, len(r.Candidates))
	for _, c := range r.Candidates {
		parts = append(parts, fmt.Sprintf("%s %.2f", c.Period.Code, c.Score))
	}
	return strings.Join(parts, ", ")
}
func MergeChronology(a, b ChronologyResult) ChronologyResult {
	out := a
	if out.UnitID == "" {
		out.UnitID = b.UnitID
	}
	byCode := map[string]ChronologyCandidate{}
	for _, c := range a.Candidates {
		byCode[c.Period.Code] = c
	}
	for _, c := range b.Candidates {
		if old, ok := byCode[c.Period.Code]; ok {
			old.Score = (old.Score + c.Score) / 2
			old.Support += c.Support
			old.Evidence = append(old.Evidence, c.Evidence...)
			byCode[c.Period.Code] = old
		} else {
			byCode[c.Period.Code] = c
		}
	}
	out.Candidates = out.Candidates[:0]
	for _, c := range byCode {
		out.Candidates = append(out.Candidates, c)
	}
	sort.Slice(out.Candidates, func(i, j int) bool { return out.Candidates[i].Score > out.Candidates[j].Score })
	if len(out.Candidates) > 0 {
		out.Selected = out.Candidates[0].Period.Code
		out.Confidence = out.Candidates[0].Score
	}
	out.Warnings = append(append([]string(nil), a.Warnings...), b.Warnings...)
	return out
}

type SequenceNode struct {
	UnitID               string
	Elevation, Thickness float64
	Earlier, Later       []string
	Confidence           float64
}
type SequenceReport struct {
	Nodes         []SequenceNode
	Roots, Leaves []string
	BrokenLinks   []string
	IsValid       bool
}

func AnalyzeSequence(nodes []SequenceNode) (SequenceReport, error) {
	report := SequenceReport{Nodes: append([]SequenceNode(nil), nodes...), IsValid: true}
	byID := map[string]*SequenceNode{}
	for i := range report.Nodes {
		n := &report.Nodes[i]
		if n.UnitID == "" {
			return report, fmt.Errorf("sequence node id missing")
		}
		if _, ok := byID[n.UnitID]; ok {
			return report, fmt.Errorf("duplicate sequence node %s", n.UnitID)
		}
		if n.Thickness <= 0 || math.IsNaN(n.Elevation) {
			return report, fmt.Errorf("invalid geometry for %s", n.UnitID)
		}
		byID[n.UnitID] = n
	}
	inDegree := map[string]int{}
	outDegree := map[string]int{}
	for _, node := range report.Nodes {
		for _, next := range node.Later {
			if _, ok := byID[next]; !ok {
				report.BrokenLinks = append(report.BrokenLinks, node.UnitID+"->"+next)
				report.IsValid = false
				continue
			}
			outDegree[node.UnitID]++
			inDegree[next]++
		}
		for _, prev := range node.Earlier {
			if _, ok := byID[prev]; !ok {
				report.BrokenLinks = append(report.BrokenLinks, prev+"->"+node.UnitID)
				report.IsValid = false
			}
		}
	}
	for _, node := range report.Nodes {
		if inDegree[node.UnitID] == 0 {
			report.Roots = append(report.Roots, node.UnitID)
		}
		if outDegree[node.UnitID] == 0 {
			report.Leaves = append(report.Leaves, node.UnitID)
		}
	}
	sort.Strings(report.Roots)
	sort.Strings(report.Leaves)
	if len(report.BrokenLinks) > 0 {
		sort.Strings(report.BrokenLinks)
	}
	return report, nil
}
func SequenceDepth(report SequenceReport) map[string]int {
	depth := map[string]int{}
	for _, root := range report.Roots {
		depth[root] = 0
	}
	changed := true
	for changed {
		changed = false
		for _, node := range report.Nodes {
			for _, next := range node.Later {
				candidate := depth[node.UnitID] + 1
				if candidate > depth[next] {
					depth[next] = candidate
					changed = true
				}
			}
		}
	}
	return depth
}
func SequenceOutliers(report SequenceReport, tolerance float64) []string {
	out := []string{}
	byID := map[string]SequenceNode{}
	for _, node := range report.Nodes {
		byID[node.UnitID] = node
	}
	for _, node := range report.Nodes {
		for _, next := range node.Later {
			child, ok := byID[next]
			if !ok {
				continue
			}
			expected := node.Elevation - node.Thickness
			if math.Abs(expected-child.Elevation) > tolerance {
				out = append(out, node.UnitID+"->"+next)
			}
		}
	}
	sort.Strings(out)
	return out
}

type DigestEvent struct{ At time.Time }
type DigestRecord struct {
	UnitID, Status string
	Events         []DigestEvent
}
type ReviewDigest struct {
	UnitID                                   string
	RecordCount, Approved, Rejected, Pending int
	Latest                                   time.Time
	Completion                               float64
	Notes                                    []string
}

func DigestRecords(unit string, records []DigestRecord) ReviewDigest {
	d := ReviewDigest{UnitID: unit}
	for _, r := range records {
		if r.UnitID != unit {
			continue
		}
		d.RecordCount++
		switch r.Status {
		case "reviewed":
			d.Approved++
		case "rejected":
			d.Rejected++
		case "submitted":
			d.Pending++
		}
		for _, e := range r.Events {
			if e.At.After(d.Latest) {
				d.Latest = e.At
			}
		}
	}
	if d.RecordCount > 0 {
		d.Completion = float64(d.Approved) / float64(d.RecordCount)
	}
	if d.Pending > 0 {
		d.Notes = append(d.Notes, "records are waiting for review")
	}
	if d.Rejected > 0 {
		d.Notes = append(d.Notes, "rejected records need field follow-up")
	}
	return d
}
func CompareDigests(a, b ReviewDigest) map[string]float64 {
	return map[string]float64{"completion_delta": a.Completion - b.Completion, "approved_delta": float64(a.Approved - b.Approved), "pending_delta": float64(a.Pending - b.Pending)}
}
