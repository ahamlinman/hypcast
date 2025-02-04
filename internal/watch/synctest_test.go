//go:build goexperiment.synctest

package watch

import (
	"testing"
	"testing/synctest"
)

func TestSyncCancelInactiveHandler(t *testing.T) {
	// The usual case of canceling a watch, where no handler is active at the time
	// of cancellation.
	synctest.Run(func() {
		v := NewValue("alice")
		notify := make(chan string, 1)
		w := v.Watch(func(x string) {
			select {
			case notify <- x:
			default:
			}
		})

		// Deal with the initial notification. Then, wait for the handler goroutine
		// to exit before canceling the watch.
		assertNextReceive(t, notify, "alice")
		synctest.Wait()
		w.Cancel()

		// Set another value, and ensure that we're not notified even after
		// background goroutines have settled.
		v.Set("bob")
		assertBlockedAfter(synctest.Wait, t, notify)
	})
}

func TestSyncDoubleCancelInactiveHandler(t *testing.T) {
	// A specific test for calling Cancel twice on an inactive handler.
	synctest.Run(func() {
		v := NewValue("alice")
		w := v.Watch(func(x string) {})

		// Wait for the initial handler to exit, then cancel the watch twice.
		// The goal is simply to not panic.
		synctest.Wait()
		w.Cancel()
		w.Cancel()
		assertWatchTerminates(t, w)
	})
}

func TestSyncWait(t *testing.T) {
	// A specific test to ensure that Wait properly blocks until the watch has
	// terminated.
	synctest.Run(func() {
		v := NewValue("alice")

		block, notify := make(chan struct{}), make(chan string)
		w := v.Watch(func(x string) {
			<-block
			notify <- x
		})

		// Ensure that we have a handler in flight.
		block <- struct{}{}

		// Start waiting in the background. We should remain blocked.
		done := makeWaitChannel(w)
		assertBlockedAfter(synctest.Wait, t, done)

		// Cancel the watch, and ensure that Wait is still blocked.
		w.Cancel()
		assertBlockedAfter(synctest.Wait, t, done)

		// Allow the handler to finish. At this point, we should become unblocked.
		assertNextReceive(t, notify, "alice")
		assertWatchTerminates(t, w)
	})
}
