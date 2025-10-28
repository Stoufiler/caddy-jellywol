package server_state

import (
	"sync"
	"testing"
)

func TestServerState(t *testing.T) {
	t.Run("initial state is not waking up", func(t *testing.T) {
		state := &ServerState{}
		if state.IsWakingUp() {
			t.Error("initial state should not be waking up")
		}
	})

	t.Run("can start and stop waking up", func(t *testing.T) {
		state := &ServerState{}

		if !state.StartWakingUp() {
			t.Error("should be able to start waking up from initial state")
		}

		if !state.IsWakingUp() {
			t.Error("state should be waking up after starting")
		}

		if state.StartWakingUp() {
			t.Error("should not be able to start waking up again")
		}

		state.DoneWakingUp()

		if state.IsWakingUp() {
			t.Error("state should not be waking up after done")
		}
	})

	t.Run("is safe for concurrent use", func(t *testing.T) {
		state := &ServerState{}
		var wg sync.WaitGroup

		// Simulate multiple concurrent requests trying to start the wake-up process.
		startCount := 0
		startCountMutex := &sync.Mutex{}
		numConcurrentRequests := 100

		wg.Add(numConcurrentRequests)
		for i := 0; i < numConcurrentRequests; i++ {
			go func() {
				defer wg.Done()
				if state.StartWakingUp() {
					startCountMutex.Lock()
					startCount++
					startCountMutex.Unlock()
				}
			}()
		}

		wg.Wait()

		if startCount != 1 {
			t.Errorf("expected exactly one routine to start waking up, but got %d", startCount)
		}
	})
}
