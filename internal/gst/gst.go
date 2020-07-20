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
	"fmt"
	"text/template"
	"time"
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
	! appsink name=raw

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
	! appsink name=video

	demux.
	! queue leaky=downstream max-size-time=2500000000 max-size-buffers=0 max-size-bytes=0
	! decodebin
	! audioconvert
	! audioresample
	! audio/x-raw,rate=48000
	! opusenc bitrate=128000
	! appsink name=audio
`))

type SinkType int

const (
	sinkTypeStart SinkType = iota - 1

	SinkTypeRaw
	SinkTypeVideo
	SinkTypeAudio

	sinkTypeEnd
)

var modulationMap = map[atsc.Modulation]string{
	atsc.Modulation8VSB:   "8vsb",
	atsc.ModulationQAM64:  "qam-64",
	atsc.ModulationQAM256: "qam-256",
}

var activePipeline *C.GstElement

func Init(channel atsc.Channel) error {
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
		return fmt.Errorf("failed to build pipeline template: %w", err)
	}

	pipelineUnsafe := C.CString(buf.String())
	defer C.free(unsafe.Pointer(pipelineUnsafe))

	var gerror *C.GError
	activePipeline = C.gst_parse_launch(pipelineUnsafe, &gerror)
	if gerror != nil {
		defer C.g_error_free(gerror)
		return fmt.Errorf("failed to initialize pipeline: %s", C.GoString(gerror.message))
	}

	defineSinkType(SinkTypeRaw, "raw")
	defineSinkType(SinkTypeVideo, "video")
	defineSinkType(SinkTypeAudio, "audio")

	return nil
}

func defineSinkType(sinkType SinkType, sinkName string) {
	sinkNameUnsafe := C.CString(sinkName)
	defer C.free(unsafe.Pointer(sinkNameUnsafe))

	C.hyp_define_sink(activePipeline, sinkNameUnsafe, C.int(sinkType))
}

type Sink func([]byte, time.Duration)

var sinks [sinkTypeEnd]Sink

func SetSink(sinkType SinkType, sink Sink) {
	sinks[sinkType] = sink
}

//export hypGoSinkSample
func hypGoSinkSample(cSinkType C.int, cBuffer unsafe.Pointer, cLen C.int, cDuration C.int) {
	sinkType := SinkType(cSinkType)
	if sinkType <= sinkTypeStart || sinkType >= sinkTypeEnd {
		panic(fmt.Errorf("invalid sink type ID %d", sinkType))
	}

	buffer := C.GoBytes(cBuffer, cLen)
	duration := time.Duration(cDuration)

	if sink := sinks[sinkType]; sink != nil {
		sink(buffer, duration)
	}
}

func Play() {
	if activePipeline == nil {
		panic("pipeline not initialized")
	}

	C.gst_element_set_state(activePipeline, C.GST_STATE_PLAYING)
}
