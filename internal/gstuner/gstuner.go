// Package gstuner supports the management of GStreamer pipelines that process
// live TV signals from DVB devices.
//
// The basic concepts behind this integration are heavily inspired by
// https://github.com/pion/rtwatch.
package gstuner

// #cgo pkg-config: gstreamer-1.0
// #include "gstuner.h"
import "C"

func init() {
	C.hyp_gst_init()
}
