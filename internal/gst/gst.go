// Package gst supports the management of GStreamer pipelines that process live
// TV signals from DVB devices.
//
// The basic concepts behind this integration are heavily inspired by
// https://github.com/pion/rtwatch.
package gst

// #cgo pkg-config: gstreamer-1.0
// #include "gst.h"
import "C"

func init() {
	C.hyp_gst_init()
}
