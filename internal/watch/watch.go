// Package watch provides primitives to enable monitoring of live state by
// multiple parties.
package watch

import "sync"

// Value provides synchronized reads and writes of an arbitrary value, and
// continuously provides updates to subscribers as writes are made.
type Value struct {
	valueMu sync.RWMutex
	value   interface{}

	subscribersMu sync.Mutex
	subscribers   map[*Subscription]struct{}
}

// Get returns the value set by the most recent Set.
func (v *Value) Get() interface{} {
	v.valueMu.RLock()
	defer v.valueMu.RUnlock()
	return v.value
}

// Set sets the value of the Value to x, and schedules notifications to
// subscribers to ensure that they eventually receive the new value.
func (v *Value) Set(x interface{}) {
	v.setValue(x)
	v.pingSubscribers()
}

func (v *Value) setValue(x interface{}) {
	v.valueMu.Lock()
	defer v.valueMu.Unlock()
	v.value = x
}

func (v *Value) pingSubscribers() {
	v.subscribersMu.Lock()
	defer v.subscribersMu.Unlock()
	for s := range v.subscribers {
		s.setFlag()
	}
}

// Subscribe sets up a handler function to continuously receive the value of v
// as it is updated, until the associated subscription is canceled.
//
// An initial call to handle will be scheduled when the subscription is first
// created. Subscribers should rely on this call to initialize any state
// associated with their subscription, to avoid losing updates between calls to
// Get and Subscribe.
//
// Each subscription executes up to one instance of handle at a time in a new
// goroutine. Any calls to Set while handle is running will result in handle
// being called once more with the latest value following completion of the
// current call. handle may not receive every value that Set is called with, and
// may see the value from a single call to Set more than once across consecutive
// calls.
//
// Subscriptions are not recovered by the garbage collector until they are
// canceled by a call to Subscription.Cancel and any running handler has
// finished executing. Values are not recovered by the garbage collector until
// all subscriptions have been recovered.
func (v *Value) Subscribe(handle func(x interface{})) *Subscription {
	s := &Subscription{
		value:   v,
		handler: handle,
		flag:    make(chan struct{}, 1),
		done:    make(chan struct{}),
	}

	s.setFlag()
	v.setSubscription(s)
	go s.run()

	return s
}

func (v *Value) setSubscription(s *Subscription) {
	v.subscribersMu.Lock()
	defer v.subscribersMu.Unlock()

	if v.subscribers == nil {
		v.subscribers = make(map[*Subscription]struct{})
	}
	v.subscribers[s] = struct{}{}
}

func (v *Value) unsetSubscription(s *Subscription) {
	v.subscribersMu.Lock()
	defer v.subscribersMu.Unlock()
	delete(v.subscribers, s)
}

// Subscription represents a subscription to the value of a Value. See
// Value.Subscribe for details.
type Subscription struct {
	value   *Value
	handler func(interface{})
	flag    chan struct{} // Buffered with size 1
	done    chan struct{} // Unbuffered; for testing only
}

func (s *Subscription) setFlag() {
	select {
	case s.flag <- struct{}{}:
	default:
	}
}

func (s *Subscription) run() {
	defer close(s.done)
	for range s.flag {
		s.handler(s.value.Get())
	}
}

// Cancel ends a subscription created with Value.Subscribe and enables its
// resources to be released once all calls to the handler have finished. Cancel
// does not wait for a handler call in flight to terminate, and does not
// guarantee that a new call to the handler will not be started after returning.
func (s *Subscription) Cancel() {
	s.value.unsetSubscription(s)
	close(s.flag)
	s.clearFlag()
}

// NOTE: The following interleaving breaks the original invariant (assumed by
// the current code) that no new calls will be made to the subscription handler
// after Cancel returns. While extensive testing has not yet revealed this case
// in practice, it is hypothetically possible as far as I understand.
//
// SET:    setFlag()
// CANCEL: unsetSubscription()
// CANCEL: close(s.flag)
// RUN:    receive from s.flag; able to enter loop body
// CANCEL: s.clearFlag(); reaches default case
// CANCEL: returns; new handler not supposed to run
// RUN:    executes loop body and calls handler
//
// I don't think it's possible to guarantee the invariant I've specified.
// Rather, we must know that run() has completed to guarantee that the handler
// will not be called again.

func (s *Subscription) clearFlag() {
	select {
	case <-s.flag:
	default:
	}
}
