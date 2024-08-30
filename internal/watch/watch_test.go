package watch

import (
	"math/rand/v2"
	"runtime"
	"sync"
	"testing"
	"time"
)

const timeout = 2 * time.Second

func TestValue(t *testing.T) {
	// A stress test meant to be run with the race detector enabled. This test
	// ensures that all access to a Value is synchronized, that handlers run
	// serially, and that handlers are properly notified of the most recent state.

	const (
		nWrites   = 1000
		nWatchers = 50
	)

	v := NewValue(int(0))
	var watches [nWatchers]Watch

	var handlerGroup sync.WaitGroup
	handlerGroup.Add(nWatchers)
	for i := 0; i < nWatchers; i++ {
		var (
			sum      int
			sawFinal bool
		)
		watches[i] = v.Watch(func(x int) {
			// This will quickly make the race detector complain if more than one
			// instance of a handler runs at once.
			sum += x

			if sawFinal && x < nWrites {
				t.Error("read a previous state after the expected final state")
				return
			}
			if !sawFinal && x == nWrites {
				handlerGroup.Done()
				sawFinal = true
			}
		})
	}

	var setGroup sync.WaitGroup
	setGroup.Add(nWrites - 1)
	for i := 1; i <= nWrites-1; i++ {
		// This will quickly make the race detector complain if Set is not properly
		// synchronized.
		go func(i int) {
			defer setGroup.Done()
			v.Set(i)
		}(i)
	}
	setGroup.Wait()

	// Our final Set, which every handler must see at least once.
	v.Set(nWrites)

	done := make(chan struct{})
	go func() {
		defer close(done)
		handlerGroup.Wait()
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		t.Fatalf("reached %v timeout before all watchers saw final state", timeout)
	}

	for _, w := range watches {
		w.Cancel()
		assertWatchTerminates(t, w)
	}
}

func TestGetZeroValue(t *testing.T) {
	// A simple test for getting the zero value of a Value.
	var v Value[any]
	if x := v.Get(); x != nil {
		t.Errorf("zero value of Value contained %v; want nil", x)
	}
}

func TestWatchZeroValue(t *testing.T) {
	var v Value[any]

	notify := make(chan any)
	w := v.Watch(func(x any) { notify <- x })

	select {
	case x := <-notify:
		if x != nil {
			t.Errorf("watch on zero value of Value got %v; want nil", x)
		}
	case <-time.After(timeout):
		t.Fatalf("reached %v timeout before watcher was notified", timeout)
	}

	w.Cancel()
	assertWatchTerminates(t, w)
}

func TestBlockedWatcher(t *testing.T) {
	// A specific test for calling Set while some handlers for a previous Set call
	// are still in progress. We expect that unrelated watchers will continue to
	// receive notifications, and that the blocked watcher will see an additional
	// notification for any state that was set while it was blocked.

	v := NewValue("alice")

	block, notifyBlocked := make(chan struct{}), make(chan string)
	blockedWatcher := v.Watch(func(x string) {
		<-block
		notifyBlocked <- x
	})

	notifyUnblocked := make(chan string)
	unblockedWatcher := v.Watch(func(x string) { notifyUnblocked <- x })

	// Handle the initial notification to both watchers.
	block <- struct{}{}
	assertNextReceive(t, notifyBlocked, "alice")
	assertNextReceive(t, notifyUnblocked, "alice")

	// Notify both watchers. Ensure that the blocked watcher has a running handler
	// for the value "bob" before continuing.
	v.Set("bob")
	block <- struct{}{}

	// Blockage of one watcher should not block the other.
	assertNextReceive(t, notifyUnblocked, "bob")
	v.Set("carol")
	assertNextReceive(t, notifyUnblocked, "carol")
	v.Set("dave")
	assertNextReceive(t, notifyUnblocked, "dave")
	v.Set("eve")
	assertNextReceive(t, notifyUnblocked, "eve")

	// Finish handling the notification that the blocked watcher received for
	// "bob".
	assertNextReceive(t, notifyBlocked, "bob")

	// Ensure that the blocked watcher receives a notification for "eve", which
	// was set while it was blocked.
	close(block)
	assertNextReceive(t, notifyBlocked, "eve")

	// Terminate our watches.
	unblockedWatcher.Cancel()
	assertWatchTerminates(t, unblockedWatcher)
	blockedWatcher.Cancel()
	assertWatchTerminates(t, blockedWatcher)
}

func TestSetFromHandler(t *testing.T) {
	// This is a special case of Set being called while a handler is blocked, as
	// the caller of Set is the handler itself. We don't prevent users from
	// entering a loop of writes and notifications.

	const stopValue = 10

	v := NewValue(int(0))
	done := make(chan struct{})
	w := v.Watch(func(x int) {
		if i := x; i < stopValue {
			v.Set(i + 1)
			v.Set(i + 1)
		} else {
			close(done)
		}
	})

	select {
	case <-done:
	case <-time.After(timeout):
		t.Fatalf("set loop did not complete after %v", timeout)
	}

	if got, want := v.Get(), stopValue; got != want {
		t.Errorf("unexpected value; got %d, want %d", got, want)
	}

	w.Cancel()
	assertWatchTerminates(t, w)
}

func TestGoexitFromHandler(t *testing.T) {
	// A specific test to ensure that terminating the goroutine running the
	// handler does not terminate the watch itself.

	v := NewValue("alice")
	notify := make(chan string)
	w := v.Watch(func(x string) {
		notify <- x
		runtime.Goexit()
	})

	assertNextReceive(t, notify, "alice")
	v.Set("bob")
	assertNextReceive(t, notify, "bob")

	w.Cancel()
	assertWatchTerminates(t, w)
}

func TestCancelInactiveHandler(t *testing.T) {
	// The usual case of canceling a watch, where no handler is active at the time
	// of cancellation. Once we cancel, no further handler calls should be made.

	v := NewValue("alice")
	notify := make(chan string, 1)
	w := v.Watch(func(x string) {
		select {
		case notify <- x:
		default:
		}
	})

	assertNextReceive(t, notify, "alice")
	forceRuntimeProgress() // Try to ensure the handler has fully terminated.

	w.Cancel()
	v.Set("bob")
	assertBlocked(t, notify)
}

func TestDoubleCancelInactiveHandler(t *testing.T) {
	// A specific test for calling Cancel twice on an inactive handler, and
	// ensuring we don't panic.

	v := NewValue("alice")
	w := v.Watch(func(x string) {})
	forceRuntimeProgress() // Try to ensure the initial handler has fully terminated.

	w.Cancel()
	w.Cancel()
	assertWatchTerminates(t, w)
}

func TestCancelBlockedWatcher(t *testing.T) {
	// A specific test for canceling a watch while it is handling a notification.

	v := NewValue("alice")

	block, notify := make(chan struct{}), make(chan string)
	w := v.Watch(func(x string) {
		<-block
		notify <- x
	})

	// Ensure that we have a handler in flight.
	block <- struct{}{}

	// Set some new values. We must schedule another call to the handler following
	// the current execution for the value "alice".
	v.Set("bob")
	v.Set("carol")

	// Cancel the watch while the handler for "alice" is still running. The
	// additional call that we forced to be scheduled above must be canceled.
	w.Cancel()

	// Set another value, to ensure that it doesn't schedule a new handler call
	// either.
	v.Set("dave")

	// Allow the original notification for "alice" to finish, and ensure that no
	// other calls will be made to the handler.
	assertNextReceive(t, notify, "alice")
	assertWatchTerminates(t, w)
}

func TestDoubleCancelFromHandler(t *testing.T) {
	// This is a special case of Cancel being called while a handler is blocked,
	// as the caller of Cancel is the handler itself. We also call Cancel twice,
	// to make sure multi-cancellation works in the active handler case.

	v := NewValue("alice")

	var canceled bool
	watchCh := make(chan Watch)
	w := v.Watch(func(x string) {
		if canceled {
			t.Error("handler called after cancellation")
			return
		}

		v.Set("bob")
		w := <-watchCh
		w.Cancel()
		w.Cancel()
		canceled = true
	})

	watchCh <- w
	assertWatchTerminates(t, w)

	if got, want := v.Get(), "bob"; got != want {
		t.Errorf("unexpected value: got %q, want %q", got, want)
	}
}

func TestWait(t *testing.T) {
	// A specific test to ensure that Wait properly blocks until the watch has
	// terminated.

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
	assertBlocked(t, done)

	// Cancel the watch, and ensure that we are still blocked.
	w.Cancel()
	assertBlocked(t, done)

	// Allow the handler to finish. At this point, we should become unblocked.
	assertNextReceive(t, notify, "alice")
	assertWatchTerminates(t, w)
}

func assertNextReceive[T comparable](t *testing.T, ch chan T, want T) {
	t.Helper()

	select {
	case got := <-ch:
		if got != want {
			t.Fatalf("got %v from channel, want %v", got, want)
		}
	case <-time.After(timeout):
		t.Fatalf("reached %v timeout before watcher was notified", timeout)
	}
}

func assertWatchTerminates(t *testing.T, w Watch) {
	t.Helper()

	done := make(chan struct{})
	go func() {
		defer close(done)
		w.Wait()
	}()

	select {
	case <-done:
	case <-time.After(timeout):
		t.Fatalf("watch not terminated after %v", timeout)
	}
}

func assertBlocked[T any](t *testing.T, ch <-chan T) {
	t.Helper()

	forceRuntimeProgress()
	select {
	case <-ch:
		t.Fatal("progress was not blocked")
	default:
	}
}

// forceRuntimeProgress makes a best-effort attempt to force the Go runtime to
// make progress on all other goroutines in the system, ideally to the point at
// which they will next block if not preempted. It works best if no other
// goroutines are CPU-intensive or change GOMAXPROCS.
func forceRuntimeProgress() {
	gomaxprocs := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(gomaxprocs)
	for range runtime.NumGoroutine() {
		runtime.Gosched()
	}
}

func BenchmarkSet1Watcher(b *testing.B) {
	benchmarkSetWithWatchers(b, 1)
}

func BenchmarkSet10Watchers(b *testing.B) {
	benchmarkSetWithWatchers(b, 10)
}

func BenchmarkSet100Watchers(b *testing.B) {
	benchmarkSetWithWatchers(b, 100)
}

func BenchmarkSet1000Watchers(b *testing.B) {
	benchmarkSetWithWatchers(b, 1000)
}

func benchmarkSetWithWatchers(b *testing.B, nWatchers int) {
	v := NewValue(uint64(0))
	watchers := make([]Watch, nWatchers)
	for i := range watchers {
		var sum uint64
		watchers[i] = v.Watch(func(x uint64) { sum += x })
	}

	b.Cleanup(func() {
		for _, w := range watchers {
			w.Cancel()
			w.Wait()
		}
	})

	b.RunParallel(func(pb *testing.PB) {
		// The choice to set random values is somewhat arbitrary. In practice, the
		// cost of lock contention probably outweighs any strategy for generating
		// these values--even setting a constant every time (unless there were ever
		// an optimization to not trigger watches when the value doesn't change).
		// Having the setters do work that the handlers can't predict feels vaguely
		// more realistic, though, and it's not a huge difference either way since
		// the goal is to compare different watcher implementations (that is, the
		// work just needs to be the same on both sides of the comparison).
		for pb.Next() {
			v.Set(rand.Uint64())
		}
	})
}
