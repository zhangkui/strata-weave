package service

import (
    "fmt"
    "strata-weave/internal/model"
    "sync"
    "testing"
    "time"
)

func TestBug001_TelemetryLedgerConcurrentSnapshot(t *testing.T) {
    ledger := NewTelemetryLedger()
    var wg sync.WaitGroup
    for worker := 0; worker < 4; worker++ {
        wg.Add(1)
        go func(worker int) {
            defer wg.Done()
            for n := 0; n < 60; n++ {
                _ = ledger.Record(model.Observation{ID: fmt.Sprintf("o-%d-%d", worker, n), UnitID: "u1", Metric: "elevation", At: time.Now()})
                _ = ledger.Snapshot("u1")
            }
        }(worker)
    }
    wg.Wait()
    if got := len(ledger.Snapshot("u1")); got != 240 { t.Fatalf("telemetry snapshot lost readings: got %d", got) }
}
