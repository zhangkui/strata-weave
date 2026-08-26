package regression

import (
    "context"
    "database/sql"
    "fmt"
    "path/filepath"
    "strata-weave/internal/model"
    "strata-weave/internal/service"
    "strata-weave/internal/store"
    "sync"
    "testing"
    "time"
)

func setupBug001(t *testing.T) (*service.Service, *sql.DB) { t.Helper(); db, err := store.Open(filepath.Join(t.TempDir(), "field.db")); if err != nil { t.Fatal(err) }; t.Cleanup(func(){ _ = db.Close() }); app:=service.New(db); if _,err=app.CreateTrench(context.Background(),model.Trench{ID:"t1",Code:"T-1",Site:"Ridge"});err!=nil{t.Fatal(err)}; if _,err=app.CreateUnit(context.Background(),model.Unit{ID:"u1",TrenchID:"t1",Code:"U-1",Phase:1});err!=nil{t.Fatal(err)}; return app,db }
func TestBug001_TelemetryLedgerConcurrentSnapshot(t *testing.T) { app,_:=setupBug001(t); var wg sync.WaitGroup; for w:=0;w<6;w++{wg.Add(1);go func(w int){defer wg.Done();for n:=0;n<30;n++{_ = app.IngestObservation(context.Background(),model.Observation{ID:fmt.Sprintf("o-%d-%d",w,n),UnitID:"u1",Instrument:"total-station",Metric:"elevation",Value:10,At:time.Now(),Quality:"verified"});_ = app.TelemetrySnapshot("u1")}}(w)};wg.Wait() }
