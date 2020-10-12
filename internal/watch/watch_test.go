package watch_test

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/ahamlinman/hypcast/internal/watch"
)

const timeout = 2 * time.Second

func TestValue(t *testing.T) {
	// A stress test meant to be run with the race detector enabled. This test
	// ensures that all access to a Value is synchronized, that subscription
	// handlers run serially, and that handlers are properly notified of the most
	// recent state.

	const (
		nSetOperations = 1000
		nSubscribers   = 50
	)

	var v watch.Value

	// Because each handler maintains internal state, the race detector should
	// complain if we run more than one instance of a given handler at a time.
	// Each handler will also let us know once it has seen our final Set at least
	// once, and panic if it sees any other state after that.
	var handlerGroup sync.WaitGroup
	handlerGroup.Add(nSubscribers)
	makeTestHandler := func() func() {
		var (
			max           int
			sawFinalState bool
		)

		return func() {
			current := v.Get().(int)
			if current > max {
				max = current
			}

			if max == nSetOperations && !sawFinalState {
				handlerGroup.Done()
				sawFinalState = true
			} else if sawFinalState && current < nSetOperations {
				t.Fatal("read a previous state after the expected final state")
			}
		}
	}

	var subscriptions [nSubscribers]*watch.Subscription
	for i := 0; i < nSubscribers; i++ {
		subscriptions[i] = v.Subscribe(makeTestHandler())
	}

	// Because we make so many Set calls concurrently, the race detector should
	// complain if access to the Value is not properly synchronized.
	var setGroup sync.WaitGroup
	setGroup.Add(nSetOperations - 1)
	for i := 1; i <= nSetOperations-1; i++ {
		go func(i int) {
			defer setGroup.Done()
			v.Set(i)
		}(i)
	}
	setGroup.Wait()

	// Our final Set, which every handler must see at least once.
	v.Set(nSetOperations)

	done := make(chan struct{})
	go func() {
		defer close(done)
		handlerGroup.Wait()
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		t.Fatalf("reached %v timeout before all subscribers saw final state", timeout)
	}

	for _, s := range subscriptions {
		s.Cancel()
		assertSubscriptionDone(t, s)
	}
}

func TestBlockedSubscriber(t *testing.T) {
	// A specific test for calling Set while some handlers for a previous Set call
	// are still in progress. We expect that unrelated subscribers will continue
	// to receive notifications, and that the blocked subscriber will see an
	// additional notification for any state that was set while it was blocked.

	var v watch.Value
	v.Set("alice")

	var (
		notifyUnblocked = make(chan string)

		block         = make(chan struct{})
		notifyBlocked = make(chan string)
	)

	blockedSub := v.Subscribe(func() {
		saw := v.Get().(string)
		<-block
		notifyBlocked <- saw
	})

	unblockedSub := v.Subscribe(func() {
		notifyUnblocked <- v.Get().(string)
	})

	// Ensure both subscribers are getting values.
	v.Set("bob")
	block <- struct{}{}
	assertNextReceive(t, notifyBlocked, "bob")
	assertNextReceive(t, notifyUnblocked, "bob")

	// Notify both subscribers. Ensure that the blocked subscriber sees the value
	// "carol" before continuing.
	v.Set("carol")
	block <- struct{}{}

	// Blockage of one subscriber should not block the other.
	assertNextReceive(t, notifyUnblocked, "carol")
	v.Set("dave")
	assertNextReceive(t, notifyUnblocked, "dave")
	v.Set("eve")
	assertNextReceive(t, notifyUnblocked, "eve")

	// Finish handling the notification that the blocked subscriber received for
	// "carol".
	assertNextReceive(t, notifyBlocked, "carol")

	// Ensure that the blocked subscriber receives a separate notification to
	// handle "eve", which was set while it was blocked.
	close(block)
	assertNextReceive(t, notifyBlocked, "eve")

	// Cancel our subscriptions.
	unblockedSub.Cancel()
	assertSubscriptionDone(t, unblockedSub)
	blockedSub.Cancel()
	assertSubscriptionDone(t, blockedSub)
}

func TestSetFromHandler(t *testing.T) {
	// This is a special case of Set being called while a handler is blocked, as
	// the caller of Set is the handler itself. We don't prevent users from
	// entering a loop of writes and notifications.

	const nSetOperations = 10
	var v watch.Value

	done := make(chan struct{})
	s := v.Subscribe(func() {
		if i := v.Get().(int); i < nSetOperations {
			v.Set(i + 1)
			v.Set(i + 1)
		} else {
			close(done)
		}
	})

	v.Set(int(1))
	select {
	case <-done:
	case <-time.After(timeout):
		t.Fatalf("set loop did not complete after %v", timeout)
	}

	if got := v.Get().(int); got != nSetOperations {
		t.Errorf("unexpected value; got %d, want %d", got, nSetOperations)
	}

	s.Cancel()
	assertSubscriptionDone(t, s)
}

func TestCancelBlockedSubscriber(t *testing.T) {
	// A specific test for canceling a subscription while it is handling a
	// notification. The requirement is that after Cancel returns, no *new* calls
	// will be made to the handler, regardless of any state updates made while the
	// handler was running.

	var (
		v      watch.Value
		block  = make(chan struct{})
		notify = make(chan string)
	)

	v.Set("alice")

	s := v.Subscribe(func() {
		saw := v.Get().(string)
		<-block
		notify <- saw
	})

	// Set a value, and force our subscriber to block before continuing.
	v.Set("bob")
	block <- struct{}{}

	// Set another value. We must schedule another call to the subscriber
	// following the current blocked execution, since it has already seen the
	// value "bob" and will not see the value "carol" in the current execution.
	v.Set("carol")

	// Cancel the subscription while the handler is still running for "bob". The
	// additional call that we forced to be scheduled above must be canceled.
	s.Cancel()

	// Set another value, to ensure that it doesn't schedule a new handler call
	// either. We use runtime.Gosched to help along any background work that's
	// incorrectly started by the call to Set.
	v.Set("dave")
	runtime.Gosched()

	// Allow the original notification for "bob" to finish, and ensure that no
	// other calls can be made to the handler.
	assertNextReceive(t, notify, "bob")
	assertSubscriptionDone(t, s)
}

func assertNextReceive(t *testing.T, ch chan string, want string) {
	t.Helper()

	select {
	case got := <-ch:
		if got != want {
			t.Fatalf("got %q from channel, want %q", got, want)
		}

	case <-time.After(timeout):
		t.Fatalf("reached %v timeout before subscriber was notified", timeout)
	}
}

func assertSubscriptionDone(t *testing.T, s *watch.Subscription) {
	t.Helper()

	select {
	case <-s.Done():
	case <-time.After(timeout):
		t.Fatalf("subscription routine still running after %v", timeout)
	}
}
