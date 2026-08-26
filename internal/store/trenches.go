package store

import (
	"database/sql"
	"strata-weave/internal/model"
)

func CreateTrench(db *sql.DB, t model.Trench) error {
	_, e := db.Exec(`INSERT INTO trenches(id,code,site,description,opened_at,closed) VALUES(?,?,?,?,?,?)`, t.ID, t.Code, t.Site, t.Description, stamp(t.OpenedAt), boolInt(t.Closed))
	return e
}
func GetTrench(db *sql.DB, id string) (model.Trench, error) {
	var t model.Trench
	var opened string
	var closed int
	e := db.QueryRow(`SELECT id,code,site,description,opened_at,closed FROM trenches WHERE id=?`, id).Scan(&t.ID, &t.Code, &t.Site, &t.Description, &opened, &closed)
	if e == sql.ErrNoRows {
		return t, model.ErrNotFound
	}
	t.OpenedAt = parseStamp(opened)
	t.Closed = closed != 0
	return t, e
}
func ListTrenches(db *sql.DB) ([]model.Trench, error) {
	rows, e := db.Query(`SELECT id,code,site,description,opened_at,closed FROM trenches ORDER BY code`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []model.Trench{}
	for rows.Next() {
		var t model.Trench
		var s string
		var c int
		if e = rows.Scan(&t.ID, &t.Code, &t.Site, &t.Description, &s, &c); e != nil {
			return nil, e
		}
		t.OpenedAt = parseStamp(s)
		t.Closed = c != 0
		out = append(out, t)
	}
	return out, rows.Err()
}
func CloseTrench(db *sql.DB, id string) error {
	r, e := db.Exec(`UPDATE trenches SET closed=1 WHERE id=?`, id)
	if e != nil {
		return e
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}
