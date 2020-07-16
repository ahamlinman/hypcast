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
	"fmt"
	"time"
	"unsafe"
)

func init() {
	C.gst_init(nil, nil)
}

const pipelineStr = `
	dvbsrc delsys=atsc modulation=8vsb frequency=189028615
	! tsdemux name=demux program-number=3

	demux.
	! queue leaky=downstream max-size-time=2500000000 max-size-buffers=0 max-size-bytes=0
	! decodebin
	! videoconvert
	! videoscale
	! video/x-raw,width=720,height=480
	! vp8enc deadline=1
	! appsink name=video

	demux.
	! queue leaky=downstream max-size-time=2500000000 max-size-buffers=0 max-size-bytes=0
	! decodebin
	! audioconvert
	! audioresample
	! audio/x-raw,rate=48000
	! opusenc bitrate=128000
	! appsink name=audio
`

type SinkType int

const (
	sinkTypeStart SinkType = iota - 1

	SinkTypeVideo
	SinkTypeAudio

	sinkTypeEnd
)

var activePipeline *C.GstElement

func Init() error {
	pipelineUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineUnsafe))

	var gerror *C.GError
	activePipeline = C.gst_parse_launch(pipelineUnsafe, &gerror)
	if gerror != nil {
		defer C.g_error_free(gerror)
		return fmt.Errorf("failed to initialize pipeline: %s", C.GoString(gerror.message))
	}

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
