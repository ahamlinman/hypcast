package watch

import "sync"

// Value provides synchronized reads and writes of some arbitrary value, and
// notifies subscribers when an updated value is available for them to read.
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

// Set sets the value that will be returned by subsequent Get calls, and
// schedules notifications to all subscribers.
func (v *Value) Set(value interface{}) {
	v.setValue(value)
	v.pingSubscribers()
}

func (v *Value) setValue(value interface{}) {
	v.valueMu.Lock()
	defer v.valueMu.Unlock()
	v.value = value
}

func (v *Value) pingSubscribers() {
	v.subscribersMu.Lock()
	defer v.subscribersMu.Unlock()
	for s := range v.subscribers {
		s.setFlag()
	}
}

// Subscribe sets up a handler to be called whenever the Value may have a new
// value that the subscriber should be aware of.
//
// Each Subscription executes up to one instance of the handler function at a
// time, asynchronously from any calls to Set. If Set is called while the
// handler is running (including from the handler itself), an additional call to
// the handler will be scheduled to ensure that it has a chance to receive the
// latest value. The handler is not guaranteed to be called for every individual
// call to Set, and may see a given value more than once across subsequent
// calls.
//
// Subscriptions are not recovered by the garbage collector until they are
// canceled by a call to Subscription.Cancel.
func (v *Value) Subscribe(handle func(interface{})) *Subscription {
	s := &Subscription{
		value:   v,
		handler: handle,
		flag:    make(chan struct{}, 1),
		done:    make(chan struct{}),
	}

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
	flag    chan struct{} // Must be buffered with size 1
	done    chan struct{} // Unbuffered
}

func (s *Subscription) run() {
	defer close(s.done)
	for range s.flag {
		s.handler(s.value.Get())
	}
}

// Cancel ends a subscription created with Value.Subscribe and cleans up
// resources associated with it. After Cancel returns, no new calls will be made
// to the subscription's handler function. However, an existing asynchronous
// call may still be in progress. Use Done to be notified when all handler calls
// have finished.
func (s *Subscription) Cancel() {
	s.value.unsetSubscription(s)
	close(s.flag)
	s.clearFlag()
}

// Done returns a channel that will be closed after this subscription has been
// canceled and any running handler has finished.
func (s *Subscription) Done() <-chan struct{} {
	return s.done
}

func (s *Subscription) setFlag() {
	select {
	case s.flag <- struct{}{}:
	default:
	}
}

func (s *Subscription) clearFlag() {
	select {
	case <-s.flag:
	default:
	}
}
