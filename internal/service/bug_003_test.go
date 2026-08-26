package service

import (
    "fmt"
    "sync"
    "testing"
    "time"
)

func TestBug003_DispatchTrackerConcurrentCollection(t *testing.T) {
    tracker := NewDispatchTracker()
    var wg sync.WaitGroup
    for n := 0; n < 8; n++ {
        wg.Add(1)
        go func(n int) { defer wg.Done(); _ = tracker.Transition(fmt.Sprintf("sample-%d", n), "collected", time.Now()) }(n)
    }
    wg.Wait()
    for n := 0; n < 8; n++ { if state, _, ok := tracker.State(fmt.Sprintf("sample-%d", n)); !ok || state != "collected" { t.Fatalf("sample-%d tracking state lost: %q %v", n, state, ok) } }
}
