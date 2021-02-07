// Package gst manages GStreamer pipelines that process live ATSC TV signals
// from Linux DVB devices.
//
// This implementation is heavily inspired by https://github.com/pion/rtwatch,
// which uses GStreamer and Pion WebRTC to stream a video file from disk. I
// doubt that I could have figured this out without that project as a reference.
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

// Pipeline represents a GStreamer pipeline that tunes an ATSC tuner card and
// produces streams of data for downstream consumption.
type Pipeline struct {
	gstPipeline *C.GstElement

	pid C.HypcastPID

	sinks    [sinkTypeEnd]Sink
	sinkRefs [sinkTypeEnd]*C.HypcastSinkRef
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
		return nil, fmt.Errorf("failed to create pipeline: %s", C.GoString(gerror.message))
	}

	// gst_parse_launch returns a "floating ref," see here for details:
	// https://developer.gnome.org/gobject/stable/gobject-The-Base-Object-Type.html#floating-ref
	C.gst_object_ref_sink(C.gpointer(gstPipeline))

	pipeline := &Pipeline{gstPipeline: gstPipeline}
	registerPipeline(pipeline)
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

// pipelineTemplate is the template for the full GStreamer pipeline meeting our
// requirements.
//
// appsink elements and their names must match up with sink definitions
// elsewhere in this package.
//
// TODO:
// https://gitlab.freedesktop.org/gstreamer/gstreamer/-/issues/358#note_118032
// Without drop-allocation the pipeline stalls. I still don't *really*
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
	! x264enc bitrate=8192 tune=zerolatency speed-preset=ultrafast
	! video/x-h264,profile=constrained-baseline,stream-format=byte-stream
	! appsink name=video max-buffers=32 drop=true

	demux.
	! queue leaky=downstream max-size-time=2500000000 max-size-buffers=0 max-size-bytes=0
	! decodebin
	! audioconvert
	! audioresample
	! audio/x-raw,rate=48000,channels=2
	! opusenc bitrate=128000
	! appsink name=audio max-buffers=32 drop=true
`))

// Start sets the pipeline to the GStreamer PLAYING state, in which it will tune
// to a channel and produce streams.
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

// Stop sets the pipeline to the GStreamer NULL state, in which it will stop any
// running streams and release the TV tuner device.
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
