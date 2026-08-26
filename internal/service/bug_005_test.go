package service

import (
    "fmt"
    "strata-weave/internal/model"
    "sync"
    "testing"
    "time"
)

func TestBug005_BatchAndSingleTelemetryShareLedger(t *testing.T) {
    ledger := NewTelemetryLedger()
    var wg sync.WaitGroup
    for n := 0; n < 8; n++ {
        wg.Add(1)
        go func(n int) { defer wg.Done(); _ = ledger.Record(model.Observation{ID: fmt.Sprintf("single-%d", n), UnitID: "u1", Metric: "elevation", At: time.Now()}); _ = ledger.Snapshot("u1") }(n)
    }
    wg.Wait()
    if got := len(ledger.Snapshot("u1")); got != 8 { t.Fatalf("shared telemetry ledger lost readings: got %d", got) }
}
