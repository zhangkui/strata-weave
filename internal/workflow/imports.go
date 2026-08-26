package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ImportRow struct {
	ID, UnitID, Metric string
	Value              float64
	At                 time.Time
}
type ImportResult struct {
	Accepted, Rejected int
	Errors             []error
}
type Importer struct {
	mu     sync.Mutex
	active bool
}

func (i *Importer) Run(ctx context.Context, rows []ImportRow, validate func(ImportRow) error, write func(context.Context, []ImportRow) error) (ImportResult, error) {
	i.mu.Lock()
	if i.active {
		i.mu.Unlock()
		return ImportResult{}, fmt.Errorf("import already running")
	}
	i.active = true
	i.mu.Unlock()
	defer func() { i.mu.Lock(); i.active = false; i.mu.Unlock() }()
	result := ImportResult{}
	accepted := []ImportRow{}
	for _, row := range rows {
		select {
		case <-ctx.Done():
			result.Errors = append(result.Errors, ctx.Err())
			result.Rejected += len(rows) - result.Accepted
			return result, ctx.Err()
		default:
		}
		if e := validate(row); e != nil {
			result.Rejected++
			result.Errors = append(result.Errors, e)
			continue
		}
		accepted = append(accepted, row)
		result.Accepted++
	}
	if len(accepted) > 0 {
		if e := write(ctx, accepted); e != nil {
			return result, e
		}
	}
	return result, nil
}
func Chunk(rows []ImportRow, size int) [][]ImportRow {
	if size <= 0 {
		return nil
	}
	out := [][]ImportRow{}
	for len(rows) > 0 {
		n := size
		if n > len(rows) {
			n = len(rows)
		}
		out = append(out, append([]ImportRow(nil), rows[:n]...))
		rows = rows[n:]
	}
	return out
}
