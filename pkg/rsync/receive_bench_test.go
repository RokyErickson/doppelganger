package rsync

import (
	"context"
	"sync"
	"testing"
)

func BenchmarkPreemptionCheckOverhead(b *testing.B) {

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func newBenchmarkCallback() func() {

	mutex := &sync.Mutex{}

	return func() {
		mutex.Lock()
		mutex.Unlock()
	}
}

func BenchmarkMonitoringCallbackOverhead(b *testing.B) {

	callback := newBenchmarkCallback()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		callback()
	}
}
