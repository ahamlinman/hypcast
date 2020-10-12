package watch

import (
	"sync"
	"testing"
	"time"
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

	var v Value
	v.Set(int(0))

	// Because each handler maintains internal state, the race detector should
	// complain if we run more than one instance of a given handler at a time.
	// Each handler will also let us know once it has seen our final Set at least
	// once, and panic if it sees any other state after that.
	var handlerGroup sync.WaitGroup
	handlerGroup.Add(nSubscribers)
	makeTestHandler := func() func(interface{}) {
		var (
			max           int
			sawFinalState bool
		)

		return func(x interface{}) {
			current := x.(int)
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

	var subscriptions [nSubscribers]*Subscription
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

	var v Value
	v.Set("alice")

	var (
		notifyUnblocked = make(chan string)

		block         = make(chan struct{})
		notifyBlocked = make(chan string)
	)

	blockedSub := v.Subscribe(func(x interface{}) {
		<-block
		notifyBlocked <- x.(string)
	})

	unblockedSub := v.Subscribe(func(x interface{}) {
		notifyUnblocked <- x.(string)
	})

	// Handle the initial notification to both subscribers.
	block <- struct{}{}
	assertNextReceive(t, notifyBlocked, "alice")
	assertNextReceive(t, notifyUnblocked, "alice")

	// Notify both subscribers. Ensure that the blocked subscriber has a running
	// handler for the value "bob" before continuing.
	v.Set("bob")
	block <- struct{}{}

	// Blockage of one subscriber should not block the other.
	assertNextReceive(t, notifyUnblocked, "bob")
	v.Set("carol")
	assertNextReceive(t, notifyUnblocked, "carol")
	v.Set("dave")
	assertNextReceive(t, notifyUnblocked, "dave")
	v.Set("eve")
	assertNextReceive(t, notifyUnblocked, "eve")

	// Finish handling the notification that the blocked subscriber received for
	// "bob".
	assertNextReceive(t, notifyBlocked, "bob")

	// Ensure that the blocked subscriber receives a notification for "eve", which
	// was set while it was blocked.
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

	const stopValue = 10
	var v Value
	v.Set(int(0))

	done := make(chan struct{})
	s := v.Subscribe(func(x interface{}) {
		if i := x.(int); i < stopValue {
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

	if got, want := v.Get().(int), stopValue; got != want {
		t.Errorf("unexpected value; got %d, want %d", got, want)
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
		v      Value
		block  = make(chan struct{})
		notify = make(chan string)
	)

	v.Set("alice")

	s := v.Subscribe(func(x interface{}) {
		<-block
		notify <- x.(string)
	})

	// Ensure that we have a handler in flight.
	block <- struct{}{}

	// Set some new values. We must schedule another call to the subscriber
	// following the current execution for the value "alice".
	v.Set("bob")
	v.Set("carol")

	// Cancel the subscription while the handler for "alice" is still running. The
	// additional call that we forced to be scheduled above must be canceled.
	s.Cancel()

	// Set another value, to ensure that it doesn't schedule a new handler call
	// either.
	v.Set("dave")

	// Allow the original notification for "alice" to finish, and ensure that no
	// other calls will be made to the handler.
	assertNextReceive(t, notify, "alice")
	assertSubscriptionDone(t, s)
}

func TestCancelFromHandler(t *testing.T) {
	// This is a special case of Cancel being called while a handler is blocked,
	// as the caller of Cancel is the handler itself.

	var v Value
	v.Set("alice")

	var (
		subscriberCanceled bool
		subCh              = make(chan *Subscription)
	)

	s := v.Subscribe(func(x interface{}) {
		if subscriberCanceled {
			t.Fatal("handler called after cancellation")
		}

		v.Set("bob")

		s := <-subCh
		s.Cancel()
		subscriberCanceled = true
	})

	subCh <- s
	assertSubscriptionDone(t, s)

	if got, want := v.Get().(string), "bob"; got != want {
		t.Errorf("unexpected value: got %q, want %q", got, want)
	}
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

func assertSubscriptionDone(t *testing.T, s *Subscription) {
	t.Helper()

	select {
	case <-s.done:
	case <-time.After(timeout):
		t.Fatalf("subscription routine still running after %v", timeout)
	}
}
