package store

import (
	"context"
	"database/sql"
	"strata-weave/internal/model"
)

func CreateFind(db *sql.DB, f model.Find) error {
	_, e := db.Exec(`INSERT INTO finds(id,unit_id,catalogue_no,kind,material,condition,reviewed,created_at) VALUES(?,?,?,?,?,?,?,?)`, f.ID, f.UnitID, f.CatalogueNo, f.Kind, f.Material, f.Condition, boolInt(f.Reviewed), stamp(f.CreatedAt))
	return e
}
func GetFind(db *sql.DB, id string) (model.Find, error) {
	var f model.Find
	var s string
	var r int
	e := db.QueryRow(`SELECT id,unit_id,catalogue_no,kind,material,condition,reviewed,created_at FROM finds WHERE id=?`, id).Scan(&f.ID, &f.UnitID, &f.CatalogueNo, &f.Kind, &f.Material, &f.Condition, &r, &s)
	if e == sql.ErrNoRows {
		return f, model.ErrNotFound
	}
	f.Reviewed = r != 0
	f.CreatedAt = parseStamp(s)
	return f, e
}
func ReviewFind(db *sql.DB, id string) error {
	r, e := db.Exec(`UPDATE finds SET reviewed=1 WHERE id=?`, id)
	if e != nil {
		return e
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}
func CreateSample(db *sql.DB, s model.Sample) error {
	_, e := db.Exec(`INSERT INTO samples(id,find_id,label,lab_code,status,collected_at) VALUES(?,?,?,?,?,?)`, s.ID, s.FindID, s.Label, s.LabCode, s.Status, stamp(s.CollectedAt))
	return e
}
func GetSample(db *sql.DB, id string) (model.Sample, error) {
	var s model.Sample
	var at string
	var method, reported sql.NullString
	var age, er sql.NullFloat64
	e := db.QueryRow(`SELECT id,find_id,label,lab_code,status,collected_at,method,age_bp,error_bp,reported_at FROM samples WHERE id=?`, id).Scan(&s.ID, &s.FindID, &s.Label, &s.LabCode, &s.Status, &at, &method, &age, &er, &reported)
	if e == sql.ErrNoRows {
		return s, model.ErrNotFound
	}
	s.CollectedAt = parseStamp(at)
	if method.Valid {
		t := parseStamp(reported.String)
		s.Result = &model.DatingResult{Method: method.String, AgeBP: age.Float64, ErrorBP: er.Float64, ReportedAt: t}
	}
	return s, e
}
func DispatchSample(db *sql.DB, id, lab string) error {
	r, e := db.ExecContext(context.Background(), `UPDATE samples SET status=?,lab_code=? WHERE id=? AND status=?`, model.SampleDispatched, lab, id, model.SampleCollected)
	if e != nil {
		return e
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return model.ErrInvalidState
	}
	return nil
}
func ReportSample(db *sql.DB, id, method string, age, errBP float64, at string) error {
	r, e := db.Exec(`UPDATE samples SET status=?,method=?,age_bp=?,error_bp=?,reported_at=? WHERE id=? AND status=?`, model.SampleReported, method, age, errBP, at, id, model.SampleDispatched)
	if e != nil {
		return e
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return model.ErrInvalidState
	}
	return nil
}

func CreateFieldRecord(db *sql.DB, r model.FieldRecord) error {
	_, e := db.Exec(`INSERT INTO records(id,unit_id,author,notes,status) VALUES(?,?,?,?,?)`, r.ID, r.UnitID, r.Author, r.Notes, r.Status)
	return e
}
func GetFieldRecord(db *sql.DB, id string) (model.FieldRecord, error) {
	var r model.FieldRecord
	var sub, rev sql.NullString
	e := db.QueryRow(`SELECT id,unit_id,author,notes,status,review_note,submitted_at,reviewed_at FROM records WHERE id=?`, id).Scan(&r.ID, &r.UnitID, &r.Author, &r.Notes, &r.Status, &r.ReviewNote, &sub, &rev)
	if e == sql.ErrNoRows {
		return r, model.ErrNotFound
	}
	r.SubmittedAt = scanTime(sub)
	r.ReviewedAt = scanTime(rev)
	return r, e
}
func SubmitRecord(db *sql.DB, id string, at string) error {
	r, e := db.Exec(`UPDATE records SET status=?,submitted_at=? WHERE id=? AND status=?`, model.RecordSubmitted, at, id, model.RecordDraft)
	if e != nil {
		return e
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return model.ErrInvalidState
	}
	return nil
}
func ReviewRecord(db *sql.DB, id, status, note, at string) error {
	r, e := db.Exec(`UPDATE records SET status=?,review_note=?,reviewed_at=? WHERE id=? AND status=?`, status, note, at, id, model.RecordSubmitted)
	if e != nil {
		return e
	}
	nrows, _ := r.RowsAffected()
	if nrows == 0 {
		return model.ErrInvalidState
	}
	return nil
}
func ListRecords(db *sql.DB, f model.RecordFilter) ([]model.FieldRecord, error) {
	q := `SELECT id,unit_id,author,notes,status,review_note,submitted_at,reviewed_at FROM records WHERE 1=1`
	args := []any{}
	if f.UnitID != "" {
		q += ` AND unit_id=?`
		args = append(args, f.UnitID)
	}
	if f.Status != "" {
		q += ` AND status=?`
		args = append(args, f.Status)
	}
	if f.Author != "" {
		q += ` AND author=?`
		args = append(args, f.Author)
	}
	q += ` ORDER BY id LIMIT ?`
	args = append(args, f.NormalizedLimit())
	rows, e := db.Query(q, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []model.FieldRecord{}
	for rows.Next() {
		var r model.FieldRecord
		var sub, rev sql.NullString
		if e = rows.Scan(&r.ID, &r.UnitID, &r.Author, &r.Notes, &r.Status, &r.ReviewNote, &sub, &rev); e != nil {
			return nil, e
		}
		r.SubmittedAt = scanTime(sub)
		r.ReviewedAt = scanTime(rev)
		out = append(out, r)
	}
	return out, rows.Err()
}
