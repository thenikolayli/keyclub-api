package sync

import (
	"sync"
	"time"
)

// prevents multiple syncs from happening at the same time
type SyncState struct {
	Mutex         sync.Mutex
	LastUpdated   time.Time
	UpdateTimeout time.Duration
}

func (state *SyncState) ShouldSync() bool {
	state.Mutex.Lock()
	defer state.Mutex.Unlock()
	return time.Since(state.LastUpdated) > state.UpdateTimeout
}
