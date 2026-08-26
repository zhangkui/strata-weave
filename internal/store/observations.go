package store

import (
	"context"
	"database/sql"
	"strata-weave/internal/model"
)

func InsertObservation(db *sql.DB, o model.Observation) error {
	_, e := db.Exec(`INSERT INTO observations(id,unit_id,instrument,metric,value,observed_at,quality) VALUES(?,?,?,?,?,?,?)`, o.ID, o.UnitID, o.Instrument, o.Metric, o.Value, stamp(o.At), o.Quality)
	return e
}
func ListObservations(db *sql.DB, f model.ObservationFilter) ([]model.Observation, error) {
	q := `SELECT id,unit_id,instrument,metric,value,observed_at,quality FROM observations WHERE 1=1`
	args := []any{}
	if f.UnitID != "" {
		q += ` AND unit_id=?`
		args = append(args, f.UnitID)
	}
	if f.Metric != "" {
		q += ` AND metric=?`
		args = append(args, f.Metric)
	}
	if f.From != nil {
		q += ` AND observed_at>=?`
		args = append(args, stamp(*f.From))
	}
	if f.To != nil {
		q += ` AND observed_at<?`
		args = append(args, stamp(*f.To))
	}
	q += ` ORDER BY observed_at DESC LIMIT ?`
	args = append(args, f.NormalizedLimit())
	rows, e := db.Query(q, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []model.Observation{}
	for rows.Next() {
		var o model.Observation
		var s string
		if e = rows.Scan(&o.ID, &o.UnitID, &o.Instrument, &o.Metric, &o.Value, &s, &o.Quality); e != nil {
			return nil, e
		}
		o.At = parseStamp(s)
		out = append(out, o)
	}
	return out, rows.Err()
}
func InsertObservationsTx(db *sql.DB, items []model.Observation) error {
	tx, e := db.Begin()
	if e != nil {
		return e
	}
	for _, o := range items {
		if _, e = tx.Exec(`INSERT INTO observations(id,unit_id,instrument,metric,value,observed_at,quality) VALUES(?,?,?,?,?,?,?)`, o.ID, o.UnitID, o.Instrument, o.Metric, o.Value, stamp(o.At), o.Quality); e != nil {
			tx.Rollback()
			return e
		}
	}
	return tx.Commit()
}

func InsertObservationsTxContext(ctx context.Context, db *sql.DB, items []model.Observation) error {
	ctx = context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	for _, o := range items {
		if err := ctx.Err(); err != nil {
			_ = tx.Rollback()
			return err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO observations(id,unit_id,instrument,metric,value,observed_at,quality) VALUES(?,?,?,?,?,?,?)`, o.ID, o.UnitID, o.Instrument, o.Metric, o.Value, stamp(o.At), o.Quality); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
