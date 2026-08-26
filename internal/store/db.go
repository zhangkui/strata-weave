package store

import (
	"database/sql"
	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err = migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`PRAGMA foreign_keys=ON;
CREATE TABLE IF NOT EXISTS trenches (id TEXT PRIMARY KEY, code TEXT NOT NULL UNIQUE, site TEXT NOT NULL, description TEXT, opened_at TEXT NOT NULL, closed INTEGER NOT NULL DEFAULT 0);
CREATE TABLE IF NOT EXISTS units (id TEXT PRIMARY KEY, trench_id TEXT NOT NULL REFERENCES trenches(id), code TEXT NOT NULL, label TEXT, description TEXT, phase INTEGER NOT NULL, created_at TEXT NOT NULL, UNIQUE(trench_id,code));
CREATE TABLE IF NOT EXISTS relations (id TEXT PRIMARY KEY, earlier_id TEXT NOT NULL REFERENCES units(id), later_id TEXT NOT NULL REFERENCES units(id), note TEXT, created_at TEXT NOT NULL, UNIQUE(earlier_id,later_id));
CREATE TABLE IF NOT EXISTS finds (id TEXT PRIMARY KEY, unit_id TEXT NOT NULL REFERENCES units(id), catalogue_no TEXT NOT NULL UNIQUE, kind TEXT NOT NULL, material TEXT, condition TEXT, reviewed INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS samples (id TEXT PRIMARY KEY, find_id TEXT NOT NULL REFERENCES finds(id), label TEXT NOT NULL, lab_code TEXT, status TEXT NOT NULL, collected_at TEXT NOT NULL, method TEXT, age_bp REAL, error_bp REAL, reported_at TEXT);
CREATE TABLE IF NOT EXISTS records (id TEXT PRIMARY KEY, unit_id TEXT NOT NULL REFERENCES units(id), author TEXT NOT NULL, notes TEXT NOT NULL, status TEXT NOT NULL, review_note TEXT, submitted_at TEXT, reviewed_at TEXT);
CREATE TABLE IF NOT EXISTS observations (id TEXT PRIMARY KEY, unit_id TEXT NOT NULL REFERENCES units(id), instrument TEXT NOT NULL, metric TEXT NOT NULL, value REAL NOT NULL, observed_at TEXT NOT NULL, quality TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS alerts (id TEXT PRIMARY KEY, unit_id TEXT NOT NULL REFERENCES units(id), severity TEXT NOT NULL, message TEXT NOT NULL, status TEXT NOT NULL, created_at TEXT NOT NULL, closed_at TEXT);
CREATE INDEX IF NOT EXISTS idx_units_trench ON units(trench_id);
CREATE INDEX IF NOT EXISTS idx_obs_unit_time ON observations(unit_id,observed_at);
CREATE INDEX IF NOT EXISTS idx_records_status ON records(status);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);`)
	return err
}
