package service

import (
    "fmt"
    "strata-weave/internal/model"
    "sync"
    "testing"
)

func TestBug004_AlertLedgerConcurrentCreation(t *testing.T) {
    ledger := NewAlertLedger()
    var wg sync.WaitGroup
    for n := 0; n < 8; n++ {
        wg.Add(1)
        go func(n int) { defer wg.Done(); _ = ledger.Upsert(model.Alert{ID: fmt.Sprintf("a-%d", n), Status: "open"}) }(n)
    }
    wg.Wait()
    if got := len(ledger.Active()); got != 8 { t.Fatalf("alert feed lost alerts: got %d", got) }
}
