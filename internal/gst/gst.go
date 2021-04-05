// Package gst manages GStreamer pipelines.
//
// This implementation is heavily inspired by https://github.com/pion/rtwatch,
// which uses GStreamer and Pion WebRTC to stream a video file from disk. I
// doubt that I could have figured this out without that project as a reference.
package gst

// #cgo pkg-config: gstreamer-1.0
// #include "gst.h"
import "C"
import (
	"errors"
	"fmt"
	"sync"
	"time"
	"unsafe"
)

func init() {
	C.gst_init(nil, nil)
}

// SinkFunc is a type for functions that receive data from the appsink elements
// of a Pipeline.
type SinkFunc func([]byte, time.Duration)

type sinkDef struct {
	fn  SinkFunc
	ref *C.HypcastSinkRef
}

// Pipeline represents a GStreamer pipeline that can provide sample data to Go
// programs through appsink elements.
type Pipeline struct {
	gstPipeline *C.GstElement

	pid C.uint

	sinkIndexByName map[string]int
	sinks           []sinkDef
}

// NewPipeline creates a GStreamer pipeline based on the syntax used in the
// gst-launch-1.0 utility.
func NewPipeline(description string) (*Pipeline, error) {
	descriptionCString := C.CString(description)
	defer C.free(unsafe.Pointer(descriptionCString))

	var gerror *C.GError
	gstPipeline := C.gst_parse_launch(descriptionCString, &gerror)
	if gerror != nil {
		defer C.g_error_free(gerror)
		return nil, errors.New(C.GoString(gerror.message))
	}

	// gst_parse_launch returns a "floating ref," see here for details:
	// https://developer.gnome.org/gobject/stable/gobject-The-Base-Object-Type.html#floating-ref
	C.gst_object_ref_sink(C.gpointer(gstPipeline))

	p := &Pipeline{
		gstPipeline:     gstPipeline,
		sinkIndexByName: make(map[string]int),
	}
	p.registerForGlobalAccess()
	return p, nil
}

// Start attempts to set the GStreamer pipeline to the PLAYING state, in which
// all elements are processing data and sinks are receiving output.
func (p *Pipeline) Start() error {
	if p.gstPipeline == nil {
		panic("pipeline not initialized")
	}

	result := C.gst_element_set_state(p.gstPipeline, C.GST_STATE_PLAYING)
	if result == C.GST_STATE_CHANGE_FAILURE {
		return errors.New("failed to start pipeline")
	}
	return nil
}

// Stop attempts to set the pipeline to the NULL state, in which no elements are
// processing data and sinks are not receiving any output.
func (p *Pipeline) Stop() error {
	if p.gstPipeline == nil {
		panic("pipeline not initialized")
	}

	result := C.gst_element_set_state(p.gstPipeline, C.GST_STATE_NULL)
	if result == C.GST_STATE_CHANGE_FAILURE {
		return errors.New("failed to stop pipeline")
	}
	return nil
}

// Close stops this pipeline if it is started and releases any resources
// associated with it.
func (p *Pipeline) Close() error {
	p.Stop()
	p.unregisterFromGlobalAccess()

	// The behavior of multiple calls to Close is undefined, however it definitely
	// should not corrupt the C heap with double-free errors. To ensure this:
	// check nil-ness before freeing, and nil after freeing.

	for _, sink := range p.sinks {
		if sink.ref != nil {
			C.free(unsafe.Pointer(sink.ref))
			sink.ref = nil
		}
	}

	if p.gstPipeline != nil {
		C.gst_object_unref(C.gpointer(p.gstPipeline))
		p.gstPipeline = nil
	}

	return nil
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
func (p *Pipeline) SetSink(name string, fn SinkFunc) {
	if fn == nil {
		panic("attempted to set nil Sink function")
	}

	if i, ok := p.sinkIndexByName[name]; ok {
		p.sinks[i].fn = fn
		return
	}

	element := p.getGstElementByName(name)
	if element == nil {
		panic(fmt.Errorf("unknown sink name %s", name))
	}
	defer C.gst_object_unref(C.gpointer(element))

	nextIndex := len(p.sinks)

	ref := (*C.HypcastSinkRef)(C.malloc(C.sizeof_HypcastSinkRef))
	ref.pid = p.pid
	ref.index = C.uint(nextIndex)

	p.sinks = append(p.sinks, sinkDef{fn: fn, ref: ref})
	p.sinkIndexByName[name] = nextIndex

	C.hypcast_connect_sink(element, ref)
}

func (p *Pipeline) getGstElementByName(name string) *C.GstElement {
	nameCString := C.CString(name)
	defer C.free(unsafe.Pointer(nameCString))

	pipelineBin := (*C.GstBin)(unsafe.Pointer(p.gstPipeline))
	return C.gst_bin_get_by_name(pipelineBin, nameCString)
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
		pipeline = getPipelineByPID(ref.pid)
		index    = int(ref.index)
		sinkFn   = pipeline.sinks[index].fn
	)
	sinkFn(data, duration)
	return C.GST_FLOW_OK
}

// cgo pointer passing rules prevent GStreamer from retaining direct references
// to pipelines, so we implement a "big old map" with integer IDs:
// https://medium.com/wallaroo-labs-engineering/adventures-with-cgo-part-1-the-pointering-19506aedf6b.
//
// Note that we don't quite have the performance concerns mentioned in the blog
// post, as writes to this map are extremely rare compared to reads. Profiling
// shows the impact of the read lock to be negligible.
var (
	pipelineLock sync.RWMutex
	nextPID      C.uint = 0
	pipelines           = make(map[C.uint]*Pipeline)
)

func getPipelineByPID(pid C.uint) *Pipeline {
	pipelineLock.RLock()
	defer pipelineLock.RUnlock()

	return pipelines[pid]
}

func (p *Pipeline) registerForGlobalAccess() {
	if p.pid != 0 {
		return
	}

	pipelineLock.Lock()
	defer pipelineLock.Unlock()

	nextPID++
	if _, ok := pipelines[nextPID]; ok {
		// If you created a new pipeline every second and never destroyed any of
		// them, even with a 32-bit uint it would take about 136 years to get here.
		// So if we get here we are probably in a very bad state of some kind.
		panic("global pipeline ID collision")
	}

	p.pid = nextPID
	pipelines[p.pid] = p
}

func (p *Pipeline) unregisterFromGlobalAccess() {
	pipelineLock.Lock()
	defer pipelineLock.Unlock()

	delete(pipelines, p.pid)
	p.pid = 0
}
