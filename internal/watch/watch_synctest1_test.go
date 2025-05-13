//go:build goexperiment.synctest && !go1.25

package watch

import (
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/assert"
)

func TestCancelInactiveHandlerSynctest1(t *testing.T) {
	// The usual case of canceling a watch, where no handler is active at the time
	// of cancellation.
	synctest.Run(func() {
		notify := make(chan string)
		v := NewValue("alice")
		w := v.Watch(func(x string) { notify <- x })

		// Deal with the initial notification. Then, wait for the handler goroutine
		// to exit before canceling the watch.
		assert.Equal(t, "alice", <-notify)
		synctest.Wait()
		w.Cancel()

		// Set another value, and ensure that we're not notified even after
		// background goroutines have settled.
		v.Set("bob")
		synctest.Wait()
		select {
		case <-notify:
			t.Error("watcher notified after being canceled")
		default:
		}
	})
}

func TestDoubleCancelInactiveHandlerSynctest1(t *testing.T) {
	// A specific test for calling Cancel twice on an inactive handler.
	synctest.Run(func() {
		v := NewValue("alice")
		w := v.Watch(func(x string) {})

		// Wait for the initial handler to exit, then cancel the watch twice.
		// The goal is simply to not panic.
		synctest.Wait()
		w.Cancel()
		w.Cancel()
		w.Wait()
	})
}

func TestWaitSynctest1(t *testing.T) {
	// A specific test to ensure that Wait properly blocks until the watch has
	// terminated.
	synctest.Run(func() {
		notify := make(chan string)
		v := NewValue("alice")
		w := v.Watch(func(x string) { notify <- x })

		// Ensure that we have a handler in flight from the initial notification.
		synctest.Wait()

		// Start waiting in the background. We should remain blocked.
		done := make(chan struct{})
		go func() { defer close(done); w.Wait() }()
		synctest.Wait()
		select {
		case <-done:
			t.Error("watcher finished waiting before cancellation")
			return
		default:
		}

		// Cancel the watch, and ensure that Wait is still blocked.
		w.Cancel()
		synctest.Wait()
		select {
		case <-done:
			t.Error("watcher finished waiting before handler exit")
			return
		default:
		}

		// Allow the handler to finish. At this point, we should become unblocked.
		assert.Equal(t, "alice", <-notify)
		<-done
	})
}
