package workflow

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type QualityRule struct {
	Code, Name string
	Required   bool
	Check      func(QualityContext) bool
}
type QualityContext struct {
	RecordID, UnitID, Author, Notes        string
	Attachments, Measurements, Coordinates int
	SubmittedAt                            *time.Time
}
type QualityFinding struct{ Code, Severity, Message string }
type QualityReport struct {
	RecordID    string
	Passed      bool
	Score       float64
	Findings    []QualityFinding
	GeneratedAt time.Time
}

func DefaultRules() []QualityRule {
	return []QualityRule{{Code: "author", Name: "author is present", Required: true, Check: func(c QualityContext) bool { return strings.TrimSpace(c.Author) != "" }}, {Code: "notes", Name: "notes are descriptive", Required: true, Check: func(c QualityContext) bool { return len(strings.TrimSpace(c.Notes)) >= 20 }}, {Code: "attachment", Name: "record has an attachment", Required: false, Check: func(c QualityContext) bool { return c.Attachments > 0 }}, {Code: "measurement", Name: "record has field measurements", Required: false, Check: func(c QualityContext) bool { return c.Measurements > 0 }}, {Code: "coordinates", Name: "record has spatial coordinates", Required: false, Check: func(c QualityContext) bool { return c.Coordinates > 0 }}, {Code: "submitted", Name: "record has submission timestamp", Required: true, Check: func(c QualityContext) bool { return c.SubmittedAt != nil }}}
}
func EvaluateQuality(c QualityContext, rules []QualityRule) QualityReport {
	report := QualityReport{RecordID: c.RecordID, Passed: true, GeneratedAt: time.Now().UTC()}
	if len(rules) == 0 {
		report.Passed = false
		report.Findings = append(report.Findings, QualityFinding{Code: "rules", Severity: "error", Message: "no quality rules configured"})
		return report
	}
	passed := 0
	for _, rule := range rules {
		if rule.Check == nil {
			report.Findings = append(report.Findings, QualityFinding{Code: rule.Code, Severity: "error", Message: "rule has no checker"})
			report.Passed = false
			continue
		}
		if rule.Check(c) {
			passed++
			continue
		}
		severity := "warning"
		if rule.Required {
			severity = "error"
			report.Passed = false
		}
		report.Findings = append(report.Findings, QualityFinding{Code: rule.Code, Severity: severity, Message: rule.Name})
	}
	report.Score = float64(passed) / float64(len(rules))
	return report
}
func SortFindings(report *QualityReport) {
	sort.SliceStable(report.Findings, func(i, j int) bool {
		rank := func(s string) int {
			if s == "error" {
				return 0
			}
			return 1
		}
		return rank(report.Findings[i].Severity) < rank(report.Findings[j].Severity)
	})
}
func RequireQuality(report QualityReport) error {
	if !report.Passed {
		return fmt.Errorf("quality report for %s failed with score %.2f", report.RecordID, report.Score)
	}
	return nil
}
func MergeQuality(a, b QualityReport) QualityReport {
	out := a
	if out.RecordID == "" {
		out.RecordID = b.RecordID
	}
	out.Score = (a.Score + b.Score) / 2
	out.Passed = a.Passed && b.Passed
	out.Findings = append(append([]QualityFinding(nil), a.Findings...), b.Findings...)
	if b.GeneratedAt.After(out.GeneratedAt) {
		out.GeneratedAt = b.GeneratedAt
	}
	SortFindings(&out)
	return out
}
