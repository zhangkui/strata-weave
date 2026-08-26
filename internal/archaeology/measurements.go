package archaeology

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type Measurement struct {
	UnitID, Metric, Instrument string
	Value                      float64
	At                         time.Time
	Quality                    string
}
type MetricWindow struct {
	Metric                                           string
	Minimum, Maximum, WarningMinimum, WarningMaximum float64
}
type MetricSummary struct {
	Metric                                 string
	Count                                  int
	Minimum, Maximum, Mean, Median, StdDev float64
	First, Last                            time.Time
}
type ThresholdResult struct {
	Measurement
	Level, Reason string
}

func Summarize(metric string, readings []Measurement) (MetricSummary, error) {
	if len(readings) == 0 {
		return MetricSummary{}, fmt.Errorf("no readings for %s", metric)
	}
	values := make([]float64, 0, len(readings))
	s := MetricSummary{Metric: metric, Count: len(readings), Minimum: readings[0].Value, Maximum: readings[0].Value, First: readings[0].At, Last: readings[0].At}
	for _, r := range readings {
		if r.Metric != metric {
			return s, fmt.Errorf("mixed metric %s", r.Metric)
		}
		if math.IsNaN(r.Value) || math.IsInf(r.Value, 0) {
			return s, fmt.Errorf("non-finite reading")
		}
		values = append(values, r.Value)
		s.Mean += r.Value
		if r.Value < s.Minimum {
			s.Minimum = r.Value
		}
		if r.Value > s.Maximum {
			s.Maximum = r.Value
		}
		if r.At.Before(s.First) {
			s.First = r.At
		}
		if r.At.After(s.Last) {
			s.Last = r.At
		}
	}
	s.Mean /= float64(s.Count)
	sort.Float64s(values)
	mid := s.Count / 2
	s.Median = values[mid]
	if s.Count%2 == 0 {
		s.Median = (values[mid-1] + values[mid]) / 2
	}
	for _, v := range values {
		d := v - s.Mean
		s.StdDev += d * d
	}
	s.StdDev = math.Sqrt(s.StdDev / float64(s.Count))
	return s, nil
}
func EvaluateWindow(w MetricWindow, r Measurement) ThresholdResult {
	out := ThresholdResult{Measurement: r, Level: "normal"}
	if r.Value < w.Minimum || r.Value > w.Maximum {
		out.Level = "critical"
		out.Reason = "outside permitted field range"
		return out
	}
	if r.Value < w.WarningMinimum || r.Value > w.WarningMaximum {
		out.Level = "warning"
		out.Reason = "outside preferred field range"
	}
	return out
}
func RollingMean(readings []Measurement, width int) []Measurement {
	if width <= 0 || len(readings) < width {
		return nil
	}
	ordered := append([]Measurement(nil), readings...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].At.Before(ordered[j].At) })
	out := make([]Measurement, 0, len(ordered)-width+1)
	sum := 0.0
	for i, r := range ordered {
		sum += r.Value
		if i >= width {
			sum -= ordered[i-width].Value
		}
		if i >= width-1 {
			m := r
			m.Value = sum / float64(width)
			m.Instrument = "rolling-mean"
			out = append(out, m)
		}
	}
	return out
}
func DetectDrift(readings []Measurement, minimumSlope float64) (float64, bool) {
	if len(readings) < 3 {
		return 0, false
	}
	ordered := append([]Measurement(nil), readings...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].At.Before(ordered[j].At) })
	base := ordered[0].At
	var sumX, sumY, sumXX, sumXY float64
	for _, r := range ordered {
		x := r.At.Sub(base).Hours()
		sumX += x
		sumY += r.Value
		sumXX += x * x
		sumXY += x * r.Value
	}
	n := float64(len(ordered))
	den := n*sumXX - sumX*sumX
	if den == 0 {
		return 0, false
	}
	slope := (n*sumXY - sumX*sumY) / den
	return slope, math.Abs(slope) >= minimumSlope
}
