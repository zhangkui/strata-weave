package store

import (
	"database/sql"
	"time"
)

func stamp(t time.Time) string      { return t.UTC().Format(time.RFC3339Nano) }
func parseStamp(s string) time.Time { t, _ := time.Parse(time.RFC3339Nano, s); return t }
func nullableTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return stamp(*t)
}
func scanTime(s sql.NullString) *time.Time {
	if !s.Valid {
		return nil
	}
	t := parseStamp(s.String)
	return &t
}
func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
