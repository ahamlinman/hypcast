// Package watch enables shared state with live updates among multiple parties.
package watch

import "sync"

// Value provides synchronized reads and writes of an arbitrary value, and
// enables watchers to be notified of updates as they are made.
//
// The zero value of a Value is valid and stores the zero value of T.
type Value[T any] struct {
	// mu prevents data races on the value, and protects the invariant that every
	// Watch receives one update call for every value of the Value from the time
	// it's added to the watchers set to the time it's removed.
	mu       sync.Mutex
	value    T
	watchers map[*watch[T]]struct{}
}

// NewValue creates a Value that stores x.
func NewValue[T any](x T) *Value[T] {
	return &Value[T]{value: x}
}

// Get returns the current value stored in v.
func (v *Value[T]) Get() T {
	v.mu.Lock()
	defer v.mu.Unlock()

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
// with subsequent values stored in v by calls to [Value.Set]. If the value
// stored in v changes while a handler execution is in flight, handler will be
// called once more with the latest value stored in v following its current
// execution. Intermediate updates preceding the latest value are dropped.
func (v *Value[T]) Watch(handler func(x T)) Watch {
	w := newWatch(handler, v.unregisterWatch)
	v.registerAndUpdateWatch(w)
	return w
}

func (v *Value[T]) registerAndUpdateWatch(w *watch[T]) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.watchers == nil {
		v.watchers = make(map[*watch[T]]struct{})
	}
	v.watchers[w] = struct{}{}
	w.update(v.value)
}

func (v *Value[T]) unregisterWatch(w *watch[T]) {
	v.mu.Lock()
	defer v.mu.Unlock()

	delete(v.watchers, w)
}

// Watch represents a single watch on a Value. See [Value.Watch] for details.
type Watch interface {
	// Cancel requests that this watch terminate as soon as possible, after any
	// pending or in-flight handler execution has finished.
	Cancel()
	// Wait blocks until this watch terminates following a call to Cancel,
	// guaranteeing that no handler is running and no new handlers will execute
	// for this watch.
	Wait()
}

// watch is the underlying type-parameterized implementation of Watch created by
// Value. Since the exported methods of a watch don't rely on the type
// parameter, we hide the watch's concrete type via an interface to reduce noise
// for clients.
type watch[T any] struct {
	handler    func(T)
	unregister func(*watch[T])

	mu      sync.Mutex
	wg      sync.WaitGroup
	next    T
	ok      bool // There is a valid value in next.
	running bool // There is (or will be) a goroutine responsible for handling values.
	cancel  bool // The WaitGroup must be canceled as soon as running == false.
}

func newWatch[T any](handler func(T), unregister func(*watch[T])) *watch[T] {
	w := &watch[T]{
		handler:    handler,
		unregister: unregister,
	}
	w.wg.Add(1)
	return w
}

func (w *watch[T]) update(x T) {
	w.mu.Lock()
	start := !w.running
	w.next, w.ok, w.running = x, true, true
	w.mu.Unlock()
	if start {
		go w.run()
	}
}

func (w *watch[T]) run() {
	var unwind bool
	defer func() {
		if unwind {
			// Only possible if w.running == true, so we must maintain the invariant.
			go w.run()
		}
	}()

	for {
		w.mu.Lock()
		next, cancel := w.next, w.cancel
		stop := !w.ok || cancel
		w.running = !stop
		w.next, w.ok = *new(T), false
		w.mu.Unlock()

		if cancel {
			w.wg.Done()
		}
		if stop {
			return
		}

		unwind = true
		w.handler(next) // May panic or call runtime.Goexit.
		unwind = false
	}
}

func (w *watch[T]) Cancel() {
	w.unregister(w) // After this, we are guaranteed no new w.update calls.
	w.mu.Lock()
	finish := !w.running && !w.cancel
	w.cancel = true
	w.mu.Unlock()
	if finish {
		w.wg.Done()
	}
}

func (w *watch[T]) Wait() {
	w.wg.Wait()
}
