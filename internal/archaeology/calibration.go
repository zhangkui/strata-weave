package archaeology

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

type DatingMeasurement struct {
	ID, SampleID, Method string
	AgeBP, ErrorBP       float64
	Lab, Analyst         string
	MeasuredAt           time.Time
	Quality              string
}
type CalibrationCurvePoint struct{ AgeBP, CalendarYear float64 }
type CalibratedRange struct {
	SampleID         string
	FromYear, ToYear float64
	Probability      float64
	Method           string
	Warnings         []string
}
type CalibrationCurve struct {
	Name   string
	Points []CalibrationCurvePoint
}

func ValidateDatingMeasurement(m DatingMeasurement) error {
	if strings.TrimSpace(m.ID) == "" || strings.TrimSpace(m.SampleID) == "" || strings.TrimSpace(m.Method) == "" {
		return fmt.Errorf("dating measurement identity is incomplete")
	}
	if m.AgeBP < 0 || m.ErrorBP < 0 || m.ErrorBP > m.AgeBP+1 {
		return fmt.Errorf("dating measurement range is invalid")
	}
	if strings.TrimSpace(m.Lab) == "" || strings.TrimSpace(m.Analyst) == "" {
		return fmt.Errorf("laboratory and analyst are required")
	}
	if m.MeasuredAt.IsZero() {
		return fmt.Errorf("measurement time is required")
	}
	return nil
}
func ValidateCurve(c CalibrationCurve) error {
	if strings.TrimSpace(c.Name) == "" || len(c.Points) < 2 {
		return fmt.Errorf("calibration curve needs a name and points")
	}
	last := c.Points[0]
	for _, p := range c.Points[1:] {
		if p.AgeBP <= last.AgeBP {
			return fmt.Errorf("curve ages must increase")
		}
		if p.CalendarYear >= last.CalendarYear {
			return fmt.Errorf("curve calendar years must decrease")
		}
		last = p
	}
	return nil
}
func Interpolate(c CalibrationCurve, age float64) (float64, error) {
	if e := ValidateCurve(c); e != nil {
		return 0, e
	}
	if age < c.Points[0].AgeBP || age > c.Points[len(c.Points)-1].AgeBP {
		return 0, fmt.Errorf("age %.2f outside curve", age)
	}
	for i := 1; i < len(c.Points); i++ {
		left, right := c.Points[i-1], c.Points[i]
		if age <= right.AgeBP {
			ratio := (age - left.AgeBP) / (right.AgeBP - left.AgeBP)
			return left.CalendarYear + ratio*(right.CalendarYear-left.CalendarYear), nil
		}
	}
	return 0, fmt.Errorf("age interpolation failed")
}
func Calibrate(m DatingMeasurement, c CalibrationCurve) (CalibratedRange, error) {
	if e := ValidateDatingMeasurement(m); e != nil {
		return CalibratedRange{}, e
	}
	center, e := Interpolate(c, m.AgeBP)
	if e != nil {
		return CalibratedRange{}, e
	}
	low := m.AgeBP - m.ErrorBP
	high := m.AgeBP + m.ErrorBP
	from, e := Interpolate(c, low)
	if e != nil {
		from = center
	}
	to, e := Interpolate(c, high)
	if e != nil {
		to = center
	}
	if from > to {
		from, to = to, from
	}
	out := CalibratedRange{SampleID: m.SampleID, FromYear: from, ToYear: to, Probability: .68, Method: c.Name}
	if m.Quality != "accepted" {
		out.Warnings = append(out.Warnings, "laboratory quality is not accepted")
	}
	if m.ErrorBP > 200 {
		out.Warnings = append(out.Warnings, "wide laboratory error range")
	}
	return out, nil
}
func CombineRanges(ranges []CalibratedRange) (CalibratedRange, error) {
	if len(ranges) == 0 {
		return CalibratedRange{}, fmt.Errorf("no calibrated ranges")
	}
	out := ranges[0]
	for _, r := range ranges[1:] {
		if r.SampleID != out.SampleID {
			return CalibratedRange{}, fmt.Errorf("ranges belong to different samples")
		}
		out.FromYear = math.Max(out.FromYear, r.FromYear)
		out.ToYear = math.Min(out.ToYear, r.ToYear)
		out.Probability *= r.Probability
		out.Warnings = append(out.Warnings, r.Warnings...)
	}
	if out.FromYear > out.ToYear {
		out.Warnings = append(out.Warnings, "calibrated ranges do not overlap")
	}
	return out, nil
}
func SortMeasurements(rows []DatingMeasurement) []DatingMeasurement {
	out := append([]DatingMeasurement(nil), rows...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].SampleID == out[j].SampleID {
			return out[i].MeasuredAt.Before(out[j].MeasuredAt)
		}
		return out[i].SampleID < out[j].SampleID
	})
	return out
}
func MeasurementSummary(rows []DatingMeasurement) map[string]float64 {
	summary := map[string]float64{}
	counts := map[string]int{}
	for _, r := range rows {
		summary[r.Method] += r.AgeBP
		counts[r.Method]++
	}
	for method, count := range counts {
		summary[method] /= float64(count)
	}
	return summary
}
func AssessReplicates(rows []DatingMeasurement, tolerance float64) []string {
	bySample := map[string][]DatingMeasurement{}
	for _, r := range rows {
		bySample[r.SampleID] = append(bySample[r.SampleID], r)
	}
	out := []string{}
	for sample, items := range bySample {
		if len(items) < 2 {
			continue
		}
		sort.Slice(items, func(i, j int) bool { return items[i].AgeBP < items[j].AgeBP })
		if items[len(items)-1].AgeBP-items[0].AgeBP > tolerance {
			out = append(out, sample)
		}
	}
	sort.Strings(out)
	return out
}
func AgeLabel(year float64) string {
	if year < 0 {
		return fmt.Sprintf("%.0f BCE", -year)
	}
	return fmt.Sprintf("%.0f CE", year)
}
