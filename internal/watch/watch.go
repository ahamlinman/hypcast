// Package watch enables shared state with live updates among multiple parties.
package watch

import "sync"

// Value provides synchronized reads and writes of an arbitrary value, and
// enables watchers to be notified of updates as they are made.
//
// The zero value of a Value is valid and stores the zero value of T.
type Value[T any] struct {
	// Invariant: Every Watch must receive one update call for every value of the
	// Value from the time it is added to the watchers set to the time it is
	// removed.
	//
	// mu protects this invariant, and prevents data races on value.
	mu       sync.RWMutex
	value    T
	watchers map[*watch[T]]struct{}
}

// NewValue creates a Value that stores x.
func NewValue[T any](x T) *Value[T] {
	return &Value[T]{value: x}
}

// Get returns the current value stored in v.
func (v *Value[T]) Get() T {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return v.value
}

// Set stores x in v.
func (v *Value[T]) Set(x T) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.value = x
	for w := range v.watchers {
		w.update(x)
	}
}

// Watch creates a new watch on the value stored in v.
//
// Each active watch executes up to one instance of handler at a time in a new
// goroutine, first with the value stored in v upon creation of the watch, then
// with subsequent values stored in v by calls to Set. If the value stored in v
// changes while a handler execution is in flight, handler will be called once
// more with the latest value stored in v following its current execution.
// Intermediate updates preceding the latest value will be dropped.
//
// Values are not recovered by the garbage collector until all of their
// associated watches have terminated. A watch is terminated after it has been
// canceled by a call to Watch.Cancel, and any pending or in-flight handler
// execution has finished.
func (v *Value[T]) Watch(handler func(x T)) Watch {
	w := newWatch(handler, v.unregisterWatch)
	v.updateAndRegisterWatch(w)
	return w
}

func (v *Value[T]) updateAndRegisterWatch(w *watch[T]) {
	v.mu.Lock()
	defer v.mu.Unlock()

	w.update(v.value)

	if v.watchers == nil {
		v.watchers = make(map[*watch[T]]struct{})
	}
	v.watchers[w] = struct{}{}
}

func (v *Value[T]) unregisterWatch(w *watch[T]) {
	v.mu.Lock()
	defer v.mu.Unlock()

	delete(v.watchers, w)
}

// Watch represents a single watch on a Value. See Value.Watch for details.
type Watch interface {
	// Cancel requests that this watch be terminated as soon as possible,
	// potentially after a pending or in-flight handler execution has finished.
	Cancel()
	// Wait blocks until this watch has terminated following a call to Cancel.
	// After Wait returns, it is guaranteed that no new handler execution will
	// start.
	Wait()
}

// watch is the underlying type-parameterized implementation of Watch created by
// Value. Since the exported methods of a watch don't rely on the type
// parameter, we hide the watch's concrete type via an interface to reduce noise
// for clients.
type watch[T any] struct {
	handler    func(T)
	unregister func(*watch[T])
	pending    chan T
	done       chan struct{}
}

func newWatch[T any](handler func(T), unregister func(*watch[T])) *watch[T] {
	w := &watch[T]{
		handler:    handler,
		unregister: unregister,
		pending:    make(chan T, 1),
		done:       make(chan struct{}),
	}
	go w.run()
	return w
}

func (w *watch[T]) run() {
	var wg sync.WaitGroup
	defer close(w.done)

	for next := range w.pending {
		x := next
		wg.Add(1)
		// Insulate the handler from the main loop, e.g. if it calls runtime.Goexit
		// it should not terminate this loop and break the processing of new values.
		go func() {
			defer wg.Done()
			w.handler(x)
		}()
		wg.Wait()
	}
}

func (w *watch[T]) update(x T) {
	// It's important that this call not block, so we assume w.pending is buffered
	// and drop a pending update to free space if necessary.
	select {
	case <-w.pending:
		w.pending <- x
	case w.pending <- x:
	}
}

func (w *watch[T]) Cancel() {
	w.unregister(w)
	w.clearPending()
	close(w.pending)
}

func (w *watch[T]) clearPending() {
	select {
	case <-w.pending:
	default:
	}
}

func (w *watch[T]) Wait() {
	<-w.done
}
