package gst

// #include "gst.h"
import "C"
import (
	"fmt"
	"sync"
	"time"
	"unsafe"
)

// Signal Handling and the Global Pipeline Map
//
// The signal mechanism for new samples allows us to provide a C function
// pointer to some handler, plus some arbitrary pointer to call the function
// with. The data referenced by that pointer needs to link the global signal
// handler to the correct Pipeline struct in Go so we can call the right Sink
// function.
//
// Since cgo's pointer passing rules prohibit C code from directly holding a
// pointer to the Pipeline struct (Go's garbage collector could, in theory,
// relocate it), we use the classic "big old map" strategy:
// https://medium.com/wallaroo-labs-engineering/adventures-with-cgo-part-1-the-pointering-19506aedf6b.
//
// Pipeline methods take care of managing the Pipeline's presence in the global
// map, as well as any C memory management for the custom data, as necessary.
//
// (Note that we do not have the performance concern mentioned in the blog post.
// Writes to the map are extremely rare compared to reads. Profiling proves the
// impact of the read lock to be negligible.)

// Sink is a type for functions that receive data streams from a Pipeline.
type Sink func([]byte, time.Duration)

type sinkDef struct {
	fn  Sink
	ref *C.HypcastSinkRef
}

// SetSink associates fn with a named appsink element in the pipeline, causing
// it to be continuously called with new samples while the pipeline is running.
//
// SetSink must only be called while the pipeline is stopped. It is possible to
// replace the sink function for an element by calling SetSink again with the
// same name and a new fn. It is not possible to "unset" a sink function for an
// element.
//
// SetSink will panic if name does not correspond to the name of a defined
// appsink, or if fn is nil.
func (p *Pipeline) SetSink(name string, fn Sink) {
	if fn == nil {
		panic("attempted to set nil Sink function")
	}

	if i, ok := p.sinkIdx[name]; ok {
		p.sinks[i].fn = fn
		return
	}

	index := len(p.sinks)

	ref := (*C.HypcastSinkRef)(C.malloc(C.sizeof_HypcastSinkRef))
	ref.pid = p.pid
	ref.index = C.uint(index)

	p.sinkIdx[name] = index
	p.sinks = append(p.sinks, sinkDef{fn: fn, ref: ref})

	var (
		pipelineBin = (*C.GstBin)(unsafe.Pointer(p.gstPipeline))
		nameCString = C.CString(name)
	)
	defer C.free(unsafe.Pointer(nameCString))

	element := C.gst_bin_get_by_name(pipelineBin, nameCString)
	if element == nil {
		panic(fmt.Errorf("unknown sink name %s", name))
	}
	defer C.gst_object_unref(C.gpointer(element))

	C.hypcast_connect_sink(element, ref)
}

// GStreamer calls hypcastSinkSample to pass data from the encoding pipeline
// into Go handler functions. See gst.c for details.
//
//export hypcastSinkSample
func hypcastSinkSample(ref *C.HypcastSinkRef, sample *C.GstSample) C.GstFlowReturn {
	buffer := C.gst_sample_get_buffer(sample)
	if buffer == nil {
		C.gst_sample_unref(sample)
		return C.GST_FLOW_OK
	}

	const offset = 0
	var (
		size = C.gst_buffer_get_size(buffer)
		data = make([]byte, size)
	)
	extracted := C.gst_buffer_extract(buffer, offset, C.gpointer(&data[0]), size)
	data = data[:extracted]

	duration := time.Duration(buffer.duration)

	C.gst_sample_unref(sample) // Invalidates sample and buffer

	var (
		pipeline = getPipeline(ref.pid)
		index    = int(ref.index)
		sink     = pipeline.sinks[index].fn
	)
	sink(data, duration)
	return C.GST_FLOW_OK
}

var (
	pipelineLock sync.RWMutex
	nextPID      C.uint = 0
	pipelines           = make(map[C.uint]*Pipeline)
)

func registerPipeline(p *Pipeline) {
	if p.pid != 0 {
		panic("duplicate pipeline registration")
	}

	pipelineLock.Lock()
	defer pipelineLock.Unlock()

	nextPID++
	if _, ok := pipelines[nextPID]; ok {
		// If you created a new pipeline every second and never destroyed any of
		// them, even with a 32-bit uint it would take about 136 years to get here.
		// So if we get here we must have seriously corrupted something.
		panic("global pipeline ID collision")
	}

	p.pid = nextPID
	pipelines[p.pid] = p
}

func unregisterPipeline(p *Pipeline) {
	pipelineLock.Lock()
	defer pipelineLock.Unlock()

	delete(pipelines, p.pid)
	p.pid = 0
}

func getPipeline(id C.uint) *Pipeline {
	pipelineLock.RLock()
	defer pipelineLock.RUnlock()

	return pipelines[id]
}
