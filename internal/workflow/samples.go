package workflow

import (
	"fmt"
	"strings"
)

type SampleState struct {
	ID, FindID, Status, LabCode string
	ReviewedFind                bool
}
type DispatchRequest struct {
	LabCode                        string
	Courier                        string
	TemperatureMin, TemperatureMax float64
}
type ChainEvent struct{ Status, Actor, Location string }

func ValidateDispatch(s SampleState, r DispatchRequest) error {
	if s.Status != "collected" {
		return fmt.Errorf("sample must be collected, got %s", s.Status)
	}
	if !s.ReviewedFind {
		return fmt.Errorf("parent find has not been reviewed")
	}
	if strings.TrimSpace(r.LabCode) == "" || strings.TrimSpace(r.Courier) == "" {
		return fmt.Errorf("lab and courier are required")
	}
	if r.TemperatureMin >= r.TemperatureMax {
		return fmt.Errorf("invalid temperature range")
	}
	return nil
}
func Dispatch(s *SampleState, r DispatchRequest, actor string) (ChainEvent, error) {
	if e := ValidateDispatch(*s, r); e != nil {
		return ChainEvent{}, e
	}
	s.Status = "dispatched"
	s.LabCode = r.LabCode
	return ChainEvent{Status: s.Status, Actor: actor, Location: r.LabCode}, nil
}
func AcceptAtLab(s *SampleState, actor string) (ChainEvent, error) {
	if s.Status != "dispatched" {
		return ChainEvent{}, fmt.Errorf("sample is not in transit")
	}
	s.Status = "at_lab"
	return ChainEvent{Status: s.Status, Actor: actor, Location: s.LabCode}, nil
}
func AttachResult(s *SampleState, method string, age, errorBP float64, actor string) (ChainEvent, error) {
	if s.Status != "at_lab" {
		return ChainEvent{}, fmt.Errorf("sample has not arrived at lab")
	}
	if strings.TrimSpace(method) == "" || age < 0 || errorBP < 0 {
		return ChainEvent{}, fmt.Errorf("invalid dating result")
	}
	s.Status = "reported"
	return ChainEvent{Status: s.Status, Actor: actor, Location: method}, nil
}
func CanArchive(s SampleState) bool { return s.Status == "reported" }
