package workflow

import (
	"fmt"
	"strings"
	"time"
)

type Event struct {
	Name, Actor, Note string
	At                time.Time
}
type Record struct {
	ID, UnitID, Author, Notes, Status string
	Events                            []Event
}
type ReviewPolicy struct {
	RequiredRole      string
	MinimumNoteLength int
	AllowSelfReview   bool
}
type Role string

const (
	RoleRecorder Role = "recorder"
	RoleReviewer Role = "reviewer"
	RoleLead     Role = "lead"
)

func CanTransition(from, to string) bool {
	switch from {
	case "draft":
		return to == "submitted"
	case "submitted":
		return to == "reviewed" || to == "rejected"
	case "rejected":
		return to == "draft"
	}
	return false
}
func Submit(r *Record, actor string, at time.Time) error {
	if r == nil {
		return fmt.Errorf("record is nil")
	}
	if strings.TrimSpace(actor) == "" {
		return fmt.Errorf("submitter required")
	}
	if !CanTransition(r.Status, "submitted") {
		return fmt.Errorf("record %s cannot be submitted from %s", r.ID, r.Status)
	}
	r.Status = "submitted"
	r.Events = append(r.Events, Event{Name: "submitted", Actor: actor, At: at})
	return nil
}
func Review(r *Record, actor string, role Role, approved bool, note string, policy ReviewPolicy, at time.Time) error {
	if r == nil {
		return fmt.Errorf("record is nil")
	}
	if role != Role(policy.RequiredRole) {
		return fmt.Errorf("role %s cannot review", role)
	}
	if !CanTransition(r.Status, "reviewed") && !CanTransition(r.Status, "rejected") {
		return fmt.Errorf("record %s is not awaiting review", r.ID)
	}
	if !policy.AllowSelfReview && actor == r.Author {
		return fmt.Errorf("author cannot review own record")
	}
	if len(strings.TrimSpace(note)) < policy.MinimumNoteLength {
		return fmt.Errorf("review note is too short")
	}
	if approved {
		r.Status = "reviewed"
	} else {
		r.Status = "rejected"
	}
	r.Events = append(r.Events, Event{Name: "reviewed", Actor: actor, Note: note, At: at})
	return nil
}
func Reopen(r *Record, actor string, at time.Time) error {
	if r == nil || r.Status != "rejected" {
		return fmt.Errorf("only rejected records can be reopened")
	}
	r.Status = "draft"
	r.Events = append(r.Events, Event{Name: "reopened", Actor: actor, At: at})
	return nil
}
func LastEvent(r Record) (Event, bool) {
	if len(r.Events) == 0 {
		return Event{}, false
	}
	return r.Events[len(r.Events)-1], true
}
func ValidateHistory(r Record) error {
	if r.Status == "" {
		return fmt.Errorf("record status missing")
	}
	previous := "draft"
	for _, e := range r.Events {
		if e.Name == "submitted" && previous != "draft" {
			return fmt.Errorf("invalid submitted event")
		}
		if e.Name == "reviewed" && previous != "submitted" {
			return fmt.Errorf("invalid review event")
		}
		if e.Name == "reopened" && previous != "rejected" {
			return fmt.Errorf("invalid reopen event")
		}
		switch e.Name {
		case "submitted":
			previous = "submitted"
		case "reviewed":
			previous = "reviewed"
		case "reopened":
			previous = "draft"
		}
	}
	if previous != r.Status && !(r.Status == "rejected" && previous == "reviewed") {
		return fmt.Errorf("status does not match history")
	}
	return nil
}
