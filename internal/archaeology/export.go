package archaeology

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

func WriteMeasurements(w io.Writer, rows []Measurement) error {
	c := csv.NewWriter(w)
	if e := c.Write([]string{"unit_id", "metric", "instrument", "value", "observed_at", "quality"}); e != nil {
		return e
	}
	for _, r := range rows {
		if e := c.Write([]string{r.UnitID, r.Metric, r.Instrument, strconv.FormatFloat(r.Value, 'f', 6, 64), r.At.Format("2006-01-02T15:04:05Z07:00"), r.Quality}); e != nil {
			return e
		}
	}
	c.Flush()
	return c.Error()
}
func ReadMeasurements(r io.Reader) ([]Measurement, error) {
	c := csv.NewReader(r)
	header, e := c.Read()
	if e != nil {
		return nil, e
	}
	if len(header) != 6 || strings.ToLower(header[0]) != "unit_id" {
		return nil, fmt.Errorf("unexpected measurement header")
	}
	out := []Measurement{}
	for {
		row, e := c.Read()
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, e
		}
		if len(row) != 6 {
			return nil, fmt.Errorf("measurement row has %d columns", len(row))
		}
		value, e := strconv.ParseFloat(row[3], 64)
		if e != nil {
			return nil, e
		}
		at, e := timeParse(row[4])
		if e != nil {
			return nil, e
		}
		out = append(out, Measurement{UnitID: row[0], Metric: row[1], Instrument: row[2], Value: value, At: at, Quality: row[5]})
	}
	return out, nil
}
func timeParse(value string) (time.Time, error) { return time.Parse(time.RFC3339, value) }
