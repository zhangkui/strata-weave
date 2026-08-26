package store

import (
	"database/sql"
	"strata-weave/internal/model"
)

func CreateUnit(db *sql.DB, u model.Unit) error {
	_, e := db.Exec(`INSERT INTO units(id,trench_id,code,label,description,phase,created_at) VALUES(?,?,?,?,?,?,?)`, u.ID, u.TrenchID, u.Code, u.Label, u.Description, u.Phase, stamp(u.CreatedAt))
	return e
}
func GetUnit(db *sql.DB, id string) (model.Unit, error) {
	var u model.Unit
	var s string
	e := db.QueryRow(`SELECT id,trench_id,code,label,description,phase,created_at FROM units WHERE id=?`, id).Scan(&u.ID, &u.TrenchID, &u.Code, &u.Label, &u.Description, &u.Phase, &s)
	if e == sql.ErrNoRows {
		return u, model.ErrNotFound
	}
	u.CreatedAt = parseStamp(s)
	return u, e
}
func ListUnits(db *sql.DB, f model.UnitFilter) ([]model.Unit, error) {
	q := `SELECT id,trench_id,code,label,description,phase,created_at FROM units WHERE 1=1`
	args := []any{}
	if f.TrenchID != "" {
		q += ` AND trench_id=?`
		args = append(args, f.TrenchID)
	}
	if f.Phase != "" {
		q += ` AND phase=?`
		args = append(args, f.Phase)
	}
	q += ` ORDER BY phase,code LIMIT ?`
	args = append(args, f.NormalizedLimit())
	rows, e := db.Query(q, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []model.Unit{}
	for rows.Next() {
		var u model.Unit
		var s string
		if e = rows.Scan(&u.ID, &u.TrenchID, &u.Code, &u.Label, &u.Description, &u.Phase, &s); e != nil {
			return nil, e
		}
		u.CreatedAt = parseStamp(s)
		out = append(out, u)
	}
	return out, rows.Err()
}
func UpdateUnitPhase(db *sql.DB, id string, phase int) error {
	r, e := db.Exec(`UPDATE units SET phase=? WHERE id=?`, phase, id)
	if e != nil {
		return e
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}
