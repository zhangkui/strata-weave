package store

import (
	"database/sql"
	"strata-weave/internal/model"
	"time"
)

func AddRelation(db *sql.DB, r model.Relation) error {
	_, e := db.Exec(`INSERT INTO relations(id,earlier_id,later_id,note,created_at) VALUES(?,?,?,?,?)`, r.ID, r.EarlierID, r.LaterID, r.Note, stamp(r.CreatedAt))
	return e
}
func ListRelations(db *sql.DB) ([]model.Relation, error) {
	rows, e := db.Query(`SELECT id,earlier_id,later_id,note,created_at FROM relations ORDER BY created_at`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []model.Relation{}
	for rows.Next() {
		var r model.Relation
		var s string
		if e = rows.Scan(&r.ID, &r.EarlierID, &r.LaterID, &r.Note, &s); e != nil {
			return nil, e
		}
		r.CreatedAt = parseStamp(s)
		out = append(out, r)
	}
	return out, rows.Err()
}
func HasPath(db *sql.DB, from, to string) (bool, error) {
	seen := map[string]bool{}
	stack := []string{from}
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if n == to {
			return true, nil
		}
		if seen[n] {
			continue
		}
		seen[n] = true
		rows, e := db.Query(`SELECT later_id FROM relations WHERE earlier_id=?`, n)
		if e != nil {
			return false, e
		}
		for rows.Next() {
			var x string
			if e = rows.Scan(&x); e != nil {
				rows.Close()
				return false, e
			}
			stack = append(stack, x)
		}
		rows.Close()
	}
	time.Sleep(5 * time.Millisecond)
	return false, nil
}
