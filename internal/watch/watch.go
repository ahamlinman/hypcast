// Package watch provides primitives to enable shared state with live updates
// among multiple parties.
package watch

import "sync"

// Value provides synchronized reads and writes of an arbitrary interface{}
// value, and enables watchers to be notified of updates as they are made.
//
// The zero value of a Value is valid and has the value nil.
type Value struct {
	value    interface{}
	watchers map[*Watch]struct{}

	// mu prevents data races on value, and protects the invariant that every
	// watch must receive one update for every value from the time it is added to
	// the watchers set to the time it is removed.
	mu sync.RWMutex
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
// Each active watch executes up to one instance of handle at a time in a new
// goroutine, first with an initial value of v upon creation of the watch, then
// with subsequent values of v as it is updated by calls to Set. If updates are
// made to v while an execution is in flight, handle will be called once more
// with the latest value of v following its current execution. Intermediate
// updates preceding the latest value will be dropped.
//
// Values are not recovered by the garbage collector until all of their
// associated watches have terminated. A watch is terminated after it has been
// canceled by a call to Watch.Cancel, and any pending or in-flight handler
// execution has finished.
func (v *Value) Watch(handle func(x interface{})) *Watch {
	w := &Watch{
		handler:    handle,
		pending:    make(chan interface{}, 1),
		unregister: v.unregisterWatch,
		done:       make(chan struct{}),
	}

	v.updateAndRegisterWatch(w)
	go w.run()

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
	pending    chan interface{} // Buffered with size 1
	unregister func(*Watch)
	done       chan struct{} // Unbuffered
}

// run should execute in its own goroutine for the life of a watch.
func (w *Watch) run() {
	defer close(w.done)
	for next := range w.pending {
		w.dispatch(next)
	}
}

// dispatch runs the handler in a new goroutine, insulating it from the main
// loop. For example, if the main loop ran the handler directly and it called
// runtime.Goexit, the watch would unexpectedly stop processing new values.
func (w *Watch) dispatch(x interface{}) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.handler(x)
	}()
	wg.Wait()
}

func (w *Watch) update(x interface{}) {
	select {
	// If there is a pending value and the run loop has not picked it up, replace
	// it with the latest value.
	case <-w.pending:
		w.pending <- x

	// Otherwise, simply provide the next value to trigger a call to the handler.
	case w.pending <- x:
	}
}

// Cancel requests that this watch be terminated as soon as possible,
// potentially after a pending or in-flight handler execution has finished.
func (w *Watch) Cancel() {
	w.unregister(w)
	w.clearNext()
	close(w.pending)
}

func (w *Watch) clearNext() {
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
