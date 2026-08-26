package store

import (
	"database/sql"
	"strata-weave/internal/model"
	"time"
)

func CreateAlert(db *sql.DB, a model.Alert) error {
	createdAt := a.CreatedAt
	if createdAt == nil {
		t := time.Now().UTC()
		createdAt = &t
	}
	_, e := db.Exec(`INSERT INTO alerts(id,unit_id,severity,message,status,created_at) VALUES(?,?,?,?,?,?)`, a.ID, a.UnitID, a.Severity, a.Message, a.Status, stamp(*createdAt))
	return e
}
func ListAlerts(db *sql.DB, f model.AlertFilter) ([]model.Alert, error) {
	q := `SELECT id,unit_id,severity,message,status,created_at,closed_at FROM alerts WHERE 1=1`
	args := []any{}
	if f.UnitID != "" {
		q += ` AND unit_id=?`
		args = append(args, f.UnitID)
	}
	if f.Severity != "" {
		q += ` AND severity=?`
		args = append(args, f.Severity)
	}
	if f.Status != "" {
		q += ` AND status=?`
		args = append(args, f.Status)
	}
	q += ` ORDER BY created_at DESC LIMIT ?`
	args = append(args, f.NormalizedLimit())
	rows, e := db.Query(q, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []model.Alert{}
	for rows.Next() {
		var a model.Alert
		var created, closed string
		if e = rows.Scan(&a.ID, &a.UnitID, &a.Severity, &a.Message, &a.Status, &created, &closed); e != nil {
			return nil, e
		}
		t := parseStamp(created)
		a.CreatedAt = &t
		if closed != "" {
			t = parseStamp(closed)
			a.ClosedAt = &t
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
func CloseAlert(db *sql.DB, id string, at string) error {
	r, e := db.Exec(`UPDATE alerts SET status='closed',closed_at=? WHERE id=? AND status!='closed'`, at, id)
	if e != nil {
		return e
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return model.ErrInvalidState
	}
	return nil
}
