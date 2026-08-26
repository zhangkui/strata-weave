package regression

import (
    "context"
    "database/sql"
    "path/filepath"
    "testing"
    "time"

    "strata-weave/internal/model"
    "strata-weave/internal/service"
    "strata-weave/internal/store"
)

func setup(t *testing.T) (*service.Service, *sql.DB) {
    t.Helper()
    db, err := store.Open(filepath.Join(t.TempDir(), "field.db"))
    if err != nil { t.Fatal(err) }
    t.Cleanup(func() { _ = db.Close() })
    app := service.New(db)
    ctx := context.Background()
    if _, err = app.CreateTrench(ctx, model.Trench{ID: "t1", Code: "T-1", Site: "Ridge"}); err != nil { t.Fatal(err) }
    for _, unit := range []model.Unit{{ID: "u1", TrenchID: "t1", Code: "U-1", Phase: 1}, {ID: "u2", TrenchID: "t1", Code: "U-2", Phase: 1}} {
        if _, err = app.CreateUnit(ctx, unit); err != nil { t.Fatal(err) }
    }
    return app, db
}

func observation(id string) model.Observation {
    return model.Observation{ID: id, UnitID: "u1", Instrument: "total-station", Metric: "elevation", Value: 10, At: time.Now().UTC(), Quality: "verified"}
}

func reviewedFind(t *testing.T, app *service.Service) model.Find {
    t.Helper()
    ctx := context.Background()
    find, err := app.CreateFind(ctx, model.Find{ID: "f1", UnitID: "u1", CatalogueNo: "CAT-1", Kind: "ceramic"})
    if err != nil { t.Fatal(err) }
    if err = app.ReviewFind(ctx, find.ID); err != nil { t.Fatal(err) }
    return find
}

func TestBug008_CanceledRecordSubmissionStaysDraft(t *testing.T) {
    app, db := setup(t); record, err := app.CreateRecord(context.Background(), model.FieldRecord{ID: "r1", UnitID: "u1", Author: "a", Notes: "record"}); if err != nil { t.Fatal(err) }
    ctx, cancel := context.WithCancel(context.Background()); cancel()
    if err = app.SubmitRecord(ctx, record.ID); err == nil { t.Fatal("cancelled submission succeeded") }
    var status string
    if err = db.QueryRow(`SELECT status FROM records WHERE id=?`, record.ID).Scan(&status); err != nil { t.Fatal(err) }
    if status != model.RecordDraft { t.Fatal("cancelled submission changed the field record") }
}
