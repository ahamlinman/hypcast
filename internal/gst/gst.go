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

type Bin int

const (
	binStart Bin = iota - 1
	BinVideo
	BinAudio
	binEnd
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

	defineReceiver("video", BinVideo)
	defineReceiver("audio", BinAudio)

	return nil
}

func defineReceiver(binName string, binID Bin) {
	binNameUnsafe := C.CString(binName)
	defer C.free(unsafe.Pointer(binNameUnsafe))

	C.hyp_define_receiver(activePipeline, binNameUnsafe, C.int(binID))
}

type Receiver func([]byte, time.Duration)

var receivers [binEnd]Receiver

func SetReceiver(bin Bin, receiver Receiver) {
	receivers[bin] = receiver
}

//export hypGoReceiveSample
func hypGoReceiveSample(cBin C.int, cBuffer unsafe.Pointer, cLen C.int, cDuration C.int) {
	bin := Bin(cBin)
	if bin <= binStart || bin >= binEnd {
		panic("bin ID outside of range")
	}

	buffer := C.GoBytes(cBuffer, cLen)
	duration := time.Duration(cDuration)

	if receiver := receivers[bin]; receiver != nil {
		receiver(buffer, duration)
	}
}
