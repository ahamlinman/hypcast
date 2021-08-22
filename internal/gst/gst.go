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
	"runtime/cgo"
	"time"
	"unsafe"
)

func init() {
	C.gst_init(nil, nil)
}

// SinkFunc is a type for functions that receive data from the appsink elements
// of a Pipeline.
type SinkFunc func([]byte, time.Duration)

// Pipeline represents a GStreamer pipeline that can provide sample data to Go
// programs through appsink elements.
type Pipeline struct {
	gstPipeline       *C.GstElement
	sinkHandlesByName map[string]cgo.Handle
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

	return &Pipeline{
		gstPipeline:       gstPipeline,
		sinkHandlesByName: make(map[string]cgo.Handle),
	}, nil
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
// associated with it. It is invalid to call any other method of a pipeline
// after it has been closed.
func (p *Pipeline) Close() error {
	p.Stop()

	// The behavior of multiple calls to Close isn't strictly defined, however it
	// probably should not exhibit any form of double-free error, *especially* for
	// things involving the C heap. As such we always check for non-zero-ness
	// before freeing and zero out after freeing.

	for name, handle := range p.sinkHandlesByName {
		handle.Delete()
		delete(p.sinkHandlesByName, name)
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
// SetSink must only be called while the pipeline is stopped, and may only be
// called up to once per appsink over the life of the pipeline. It will panic if
// name does not correspond to the name of a defined appsink, if fn is nil, or
// if SetSink has already been called once for the named appsink.
func (p *Pipeline) SetSink(name string, fn SinkFunc) {
	if fn == nil {
		panic("attempted to set nil sink function")
	}
	if p.sinkHandlesByName[name] > 0 {
		panic("called SetSink more than once for the same appsink")
	}

	handle := cgo.NewHandle(fn)
	p.sinkHandlesByName[name] = handle

	element := p.getGstElementByName(name)
	if element == nil {
		panic(fmt.Errorf("unknown sink name %s", name))
	}
	defer C.gst_object_unref(C.gpointer(element))

	C.hypcast_connect_sink(element, C.uintptr_t(handle))
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
func hypcastSinkSample(sample *C.GstSample, sinkHandle uintptr) C.GstFlowReturn {
	buffer := C.gst_sample_get_buffer(sample)
	if buffer == nil {
		C.gst_sample_unref(sample)
		return C.GST_FLOW_OK
	}

	const offset = 0
	size := C.gst_buffer_get_size(buffer)
	data := make([]byte, size)
	extracted := C.gst_buffer_extract(buffer, offset, C.gpointer(&data[0]), size)

	data = data[:extracted]
	duration := time.Duration(buffer.duration)

	C.gst_sample_unref(sample) // Invalidates sample and buffer

	sinkFn := cgo.Handle(sinkHandle).Value().(SinkFunc)
	sinkFn(data, duration)

	return C.GST_FLOW_OK
}
