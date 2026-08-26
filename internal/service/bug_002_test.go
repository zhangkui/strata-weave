package service

import (
    "sync"
    "testing"
    "time"
)

func TestBug002_ReviewQueueConcurrentSubmission(t *testing.T) {
    queue := NewReviewQueue()
    var wg sync.WaitGroup
    for n := 0; n < 8; n++ {
        wg.Add(1)
        go func() { defer wg.Done(); _ = queue.Enqueue("record-1", time.Now()) }()
    }
    wg.Wait()
    if got := queue.Pending(); len(got) != 1 { t.Fatalf("review queue duplicated record: %v", got) }
}
