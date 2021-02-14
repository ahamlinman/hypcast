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
	"unsafe"
)

func init() {
	C.gst_init(nil, nil)
}

// Pipeline represents a GStreamer pipeline that can provide sample data to Go
// programs through appsink elements.
type Pipeline struct {
	gstPipeline *C.GstElement

	pid C.HypcastPID

	sinks    [sinkTypeEnd]Sink
	sinkRefs [sinkTypeEnd]*C.HypcastSinkRef
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

	pipeline := &Pipeline{gstPipeline: gstPipeline}
	registerPipeline(pipeline)
	return pipeline, nil
}

// Start sets the pipeline to the GStreamer PLAYING state.
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

// Stop sets the pipeline to the GStreamer NULL state.
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
	unregisterPipeline(p)

	// The behavior of multiple calls to Close is undefined, however it definitely
	// should not corrupt the C heap with double-free errors. To ensure this:
	// check nil-ness before freeing, and nil after freeing.

	for i, sinkRef := range p.sinkRefs {
		if sinkRef != nil {
			C.free(unsafe.Pointer(sinkRef))
			p.sinkRefs[i] = nil
		}
	}

	if p.gstPipeline != nil {
		C.gst_object_unref(C.gpointer(p.gstPipeline))
		p.gstPipeline = nil
	}

	return nil
}
