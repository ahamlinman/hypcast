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
	state      chan watchState[T]
	handle     func(T)
	unregister func(*watch[T])
}

type watchState[T any] struct {
	pending bool
	next    T
	running bool
	closing bool
	done    chan struct{}
}

func newWatch[T any](handle func(T), unregister func(*watch[T])) *watch[T] {
	w := &watch[T]{
		state:      make(chan watchState[T], 1),
		handle:     handle,
		unregister: unregister,
	}
	w.state <- watchState[T]{}
	return w
}

func (w *watch[T]) update(x T) {
	state := <-w.state
	state.pending = true
	state.next = x
	if !state.running {
		state.running = true
		state.done = make(chan struct{})
		go w.run()
	}
	w.state <- state
}

func (w *watch[T]) run() {
	var wg sync.WaitGroup
	for {
		state := <-w.state
		if !state.pending || state.closing {
			close(state.done)
			state.running = false
			w.state <- state
			return
		}

		x := state.next
		state.pending = false
		w.state <- state

		// Insulate the handler from the main loop, e.g. if it calls runtime.Goexit
		// it should not terminate this loop and break the processing of new values.
		wg.Add(1)
		go func() {
			defer wg.Done()
			w.handle(x)
		}()
		wg.Wait()
	}
}

func (w *watch[T]) Cancel() {
	w.unregister(w)
	state := <-w.state
	state.closing = true
	w.state <- state
}

func (w *watch[T]) Wait() {
	state := <-w.state
	if !state.running {
		w.state <- state
		return
	}
	done := state.done
	w.state <- state
	<-done
}
