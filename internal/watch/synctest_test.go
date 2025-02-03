//go:build goexperiment.synctest

package watch

import (
	"testing"
	"testing/synctest"
)

func TestSyncCancelInactiveHandler(t *testing.T) {
	// The usual case of canceling a watch, where no handler is active at the time
	// of cancellation. Once we cancel, no further handler calls should be made.
	synctest.Run(func() {
		v := NewValue("alice")
		notify := make(chan string, 1)
		w := v.Watch(func(x string) {
			select {
			case notify <- x:
			default:
			}
		})

		assertNextReceive(t, notify, "alice")

		synctest.Wait() // Guarantee that the handler goroutine has exited.
		w.Cancel()

		v.Set("bob")
		synctest.Wait()
		assertSyncBlocked(t, notify)
	})
}

func TestSyncDoubleCancelInactiveHandler(t *testing.T) {
	// A specific test for calling Cancel twice on an inactive handler, and
	// ensuring we don't panic.
	synctest.Run(func() {
		v := NewValue("alice")
		w := v.Watch(func(x string) {})
		synctest.Wait() // Guarantee that the initial handler goroutine has exited.

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
		done := make(chan struct{})
		go func() {
			defer close(done)
			w.Wait()
		}()
		synctest.Wait()
		assertSyncBlocked(t, done)

		// Cancel the watch, and ensure that we are still blocked.
		w.Cancel()
		synctest.Wait()
		assertSyncBlocked(t, done)

		// Allow the handler to finish. At this point, we should become unblocked.
		assertNextReceive(t, notify, "alice")
		assertWatchTerminates(t, w)
	})
}

func assertSyncBlocked[T any](t *testing.T, ch <-chan T) {
	t.Helper()

	select {
	case <-ch:
		t.Fatal("progress was not blocked")
	default:
	}
}
