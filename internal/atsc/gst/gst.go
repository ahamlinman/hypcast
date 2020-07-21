// Package gst supports the management of GStreamer pipelines that process live
// TV signals from DVB devices.
//
// The basic concepts behind this integration are heavily inspired by
// https://github.com/pion/rtwatch.
package gst

// #cgo pkg-config: gstreamer-1.0
// #include "gst.h"
import "C"
import (
	"bytes"
	"errors"
	"fmt"
	"text/template"
	"unsafe"

	"github.com/ahamlinman/hypcast/internal/atsc"
)

func init() {
	C.gst_init(nil, nil)
}

// https://gitlab.freedesktop.org/gstreamer/gstreamer/-/issues/358#note_118032
// TODO: Without drop-allocation the pipeline stalls. I still don't *really*
// understand why.
var pipelineTemplate = template.Must(template.New("").Parse(`
	dvbsrc delsys=atsc modulation={{.Modulation}} frequency={{.Frequency}}
	! tee name=dvbtee
	! identity drop-allocation=true
	! queue leaky=downstream max-size-time=1000000000 max-size-buffers=0 max-size-bytes=0
	! appsink name=raw max-buffers=32 drop=true

	dvbtee.
	! queue leaky=downstream max-size-time=0 max-size-buffers=0 max-size-bytes=0
	! tsdemux name=demux program-number={{.PID}}

	demux.
	! queue leaky=downstream max-size-time=2500000000 max-size-buffers=0 max-size-bytes=0
	! decodebin
	! videoconvert
	! deinterlace
	! videoscale add-borders=true
	! video/x-raw,width=1280,height=720
	! vp8enc cpu-used=8 deadline=1 resize-allowed=true
	! appsink name=video max-buffers=32 drop=true

	demux.
	! queue leaky=downstream max-size-time=2500000000 max-size-buffers=0 max-size-bytes=0
	! decodebin
	! audioconvert
	! audioresample
	! audio/x-raw,rate=48000
	! opusenc bitrate=128000
	! appsink name=audio max-buffers=32 drop=true
`))

// Pipeline represents a GStreamer pipeline that tunes an ATSC tuner card and
// produces streams of data for downstream consumption.
type Pipeline struct {
	gstPipeline *C.GstElement

	globalID globalPipelineID

	sinks    [sinkTypeEnd]Sink
	sinkRefs [sinkTypeEnd]*C.HypSinkRef
}

// NewPipeline creates a new Pipeline that will produce streams for the provided
// Channel when started.
func NewPipeline(channel atsc.Channel) (*Pipeline, error) {
	pipelineString, err := buildPipelineString(channel)
	if err != nil {
		return nil, err
	}

	pipelineUnsafe := C.CString(pipelineString)
	defer C.free(unsafe.Pointer(pipelineUnsafe))

	var gerror *C.GError
	gstPipeline := C.gst_parse_launch(pipelineUnsafe, &gerror)
	if gerror != nil {
		defer C.g_error_free(gerror)
		return nil, fmt.Errorf("failed to initialize pipeline: %s", C.GoString(gerror.message))
	}

	// gst_parse_launch returns a "floating ref," see here for details:
	// https://developer.gnome.org/gobject/stable/gobject-The-Base-Object-Type.html#floating-ref
	C.gst_object_ref_sink(C.gpointer(gstPipeline))

	pipeline := &Pipeline{gstPipeline: gstPipeline}
	registerGlobalPipeline(pipeline)
	return pipeline, nil
}

var modulationMap = map[atsc.Modulation]string{
	atsc.Modulation8VSB:   "8vsb",
	atsc.ModulationQAM64:  "qam-64",
	atsc.ModulationQAM256: "qam-256",
}

func buildPipelineString(channel atsc.Channel) (string, error) {
	var buf bytes.Buffer

	err := pipelineTemplate.Execute(&buf, struct {
		Modulation string
		Frequency  uint
		PID        uint
	}{
		Modulation: modulationMap[channel.Modulation],
		Frequency:  channel.Frequency,
		PID:        channel.ProgramID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to build pipeline template: %w", err)
	}

	return buf.String(), nil
}

// ErrPipelineClosed is returned when invoking methods on a pipeline for which
// Close() has been called.
var ErrPipelineClosed = errors.New("pipeline closed")

// Start sets the pipeline to the GStreamer PLAYING state, in which it will tune
// to a channel and produce streams.
func (p *Pipeline) Start() error {
	if p.gstPipeline == nil {
		return ErrPipelineClosed
	}

	result := C.gst_element_set_state(p.gstPipeline, C.GST_STATE_PLAYING)
	if result == C.GST_STATE_CHANGE_FAILURE {
		return errors.New("failed to change GStreamer pipeline state")
	}
	return nil
}

// Stop sets the pipeline to the GStreamer NULL state, in which it will stop any
// running streams and release the TV tuner device.
func (p *Pipeline) Stop() error {
	if p.gstPipeline == nil {
		return ErrPipelineClosed
	}

	result := C.gst_element_set_state(p.gstPipeline, C.GST_STATE_NULL)
	if result == C.GST_STATE_CHANGE_FAILURE {
		return errors.New("failed to change GStreamer pipeline state")
	}
	return nil
}

// Close stops this pipeline and releases all resources associated with it.
func (p *Pipeline) Close() error {
	p.Stop()
	unregisterGlobalPipeline(p)

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
