package model

import (
	"strings"
	"time"
)

func ValidateTrench(t Trench) error {
	if strings.TrimSpace(t.Code) == "" || strings.TrimSpace(t.Site) == "" {
		return ErrInvalidInput
	}
	return nil
}
func ValidateUnit(u Unit) error {
	if u.TrenchID == "" || strings.TrimSpace(u.Code) == "" || u.Phase < 0 {
		return ErrInvalidInput
	}
	return nil
}
func ValidateFind(f Find) error {
	if f.UnitID == "" || strings.TrimSpace(f.CatalogueNo) == "" || strings.TrimSpace(f.Kind) == "" {
		return ErrInvalidInput
	}
	return nil
}
func ValidateSample(s Sample) error {
	if s.FindID == "" || strings.TrimSpace(s.Label) == "" {
		return ErrInvalidInput
	}
	if s.Result.Method != "" && s.Status == "" {
		return ErrInvalidInput
	}
	if s.Status != "" && s.Status != SampleCollected {
		return ErrInvalidState
	}
	if s.CollectedAt.IsZero() {
		s.CollectedAt = time.Now()
	}
	return nil
}
func ValidateObservation(o Observation) error {
	if o.UnitID == "" || o.Metric == "" || o.Instrument == "" || o.At.IsZero() {
		return ErrInvalidInput
	}
	return nil
}
