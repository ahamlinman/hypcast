// Package watch provides primitives to enable shared state with live updates
// among multiple parties.
package watch

import "sync"

// Value provides synchronized reads and writes of an arbitrary interface{}
// value, and enables watchers to be notified of updates as they are made.
//
// The zero value of a Value is valid and has the value nil.
type Value struct {
	// Invariant: Every Watch must receive one update call for every value of the
	// Value from the time it is added to the watchers set to the time it is
	// removed.
	//
	// mu protects this invariant, and prevents data races on value.
	mu       sync.RWMutex
	value    interface{}
	watchers map[*Watch]struct{}
}

// NewValue creates a Value whose value is initially set to x.
func NewValue(x interface{}) *Value {
	return &Value{value: x}
}

// Get returns the current value of v.
func (v *Value) Get() interface{} {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return v.value
}

// Set sets the value of v to x.
func (v *Value) Set(x interface{}) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.value = x

	for w := range v.watchers {
		w.update(x)
	}
}

// Watch creates a new watch on the value of v.
//
// Each active watch executes up to one instance of handler at a time in a new
// goroutine, first with an initial value of v upon creation of the watch, then
// with subsequent values of v as it is updated by calls to Set. If updates are
// made to v while an execution is in flight, handler will be called once more
// with the latest value of v following its current execution. Intermediate
// updates preceding the latest value will be dropped.
//
// Values are not recovered by the garbage collector until all of their
// associated watches have terminated. A watch is terminated after it has been
// canceled by a call to Watch.Cancel, and any pending or in-flight handler
// execution has finished.
func (v *Value) Watch(handler func(x interface{})) *Watch {
	w := newWatch(handler, v.unregisterWatch)
	v.updateAndRegisterWatch(w)
	return w
}

func (v *Value) updateAndRegisterWatch(w *Watch) {
	v.mu.Lock()
	defer v.mu.Unlock()

	w.update(v.value)

	if v.watchers == nil {
		v.watchers = make(map[*Watch]struct{})
	}
	v.watchers[w] = struct{}{}
}

func (v *Value) unregisterWatch(w *Watch) {
	v.mu.Lock()
	defer v.mu.Unlock()

	delete(v.watchers, w)
}

// Watch represents a single watch on a Value. See Value.Watch for details.
type Watch struct {
	handler    func(interface{})
	unregister func(*Watch)

	pending chan interface{}
	done    chan struct{}
}

func newWatch(handler func(interface{}), unregister func(*Watch)) *Watch {
	w := &Watch{
		handler:    handler,
		unregister: unregister,
		pending:    make(chan interface{}, 1),
		done:       make(chan struct{}),
	}
	go w.run()
	return w
}

func (w *Watch) run() {
	var wg sync.WaitGroup
	defer close(w.done)

	for next := range w.pending {
		wg.Add(1)

		// Insulate the handler from the main loop, e.g. if it calls runtime.Goexit
		// it should not terminate this loop and break the processing of new values.
		go func(x interface{}) {
			defer wg.Done()
			w.handler(x)
		}(next)

		wg.Wait()
	}
}

func (w *Watch) update(x interface{}) {
	// It's important that this call not block, so we assume w.pending is buffered
	// and drop a pending update to free space if necessary.
	select {
	case <-w.pending:
		w.pending <- x

	case w.pending <- x:
	}
}

// Cancel requests that this watch be terminated as soon as possible,
// potentially after a pending or in-flight handler execution has finished.
func (w *Watch) Cancel() {
	w.unregister(w)
	w.clearPending()
	close(w.pending)
}

func (w *Watch) clearPending() {
	select {
	case <-w.pending:
	default:
	}
}

// Wait blocks until this watch has terminated following a call to Cancel. After
// Wait returns, it is guaranteed that no new handler execution will start.
func (w *Watch) Wait() {
	<-w.done
}
