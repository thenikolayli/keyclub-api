package auth

import "sync"

var loginWaiters sync.Map // attemptID -> chan struct{}

// Registers a login waiter for a specific attempt ID
func RegisterLoginWaiter(attemptID string) (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)
	loginWaiters.Store(attemptID, ch)
	return ch, func() { loginWaiters.Delete(attemptID) }
}

// Notifies a login waiter for a specific attempt ID
func NotifyLoginWaiter(attemptID string) {
	if v, ok := loginWaiters.Load(attemptID); ok {
		select {
		case v.(chan struct{}) <- struct{}{}: // sends a signal to the channel
		default:
		}
	}
}
