package gst

// #include "gst.h"
import "C"
import (
	"fmt"
	"sync"
	"time"
)

// Signal Handling and the Global Pipeline Map
//
// The signal mechanism for new samples allows us to provide a C function
// pointer to some handler, plus some arbitrary pointer to call the function
// with. The data referenced by that pointer needs to link the global signal
// handler to the correct Pipeline struct in Go so we can call the right Sink
// function.
//
// Since cgo's pointer passing rules prohibit C code from directly holding a
// pointer to the Pipeline struct (Go's garbage collector could, in theory,
// relocate it), we use the classic "big old map" strategy:
// https://medium.com/wallaroo-labs-engineering/adventures-with-cgo-part-1-the-pointering-19506aedf6b.
//
// Pipeline methods take care of managing the Pipeline's presence in the global
// map, as well as any C memory management for the custom data, as necessary.
//
// (Note that we do not have the performance concern mentioned in the blog post.
// Writes to the map are extremely rare compared to reads. Profiling proves the
// impact of the read lock to be negligible.)

// Sink is a type for functions that receive data streams from a Pipeline.
type Sink func([]byte, time.Duration)

// SinkType represents the various data streams that can be sent to a Sink.
type SinkType uint

const (
	// SinkTypeRaw represents the raw MPEG-TS stream produced by the tuner,
	// without demuxing or filtering. Program and channel information can be
	// extracted from this stream with an appropriate demuxer and parser.
	SinkTypeRaw SinkType = iota

	// SinkTypeVideo represents the H.264-encoded video stream, using the
	// Constrained Baseline profile to meet WebRTC requirements.
	SinkTypeVideo

	// SinkTypeAudio represents the Opus-encoded audio stream.
	SinkTypeAudio

	sinkTypeEnd
)

// sinkNames matches the appsink element names defined in the pipeline template.
var sinkNames = map[SinkType]*C.char{
	SinkTypeRaw:   C.HYPCAST_SINK_NAME_RAW,
	SinkTypeVideo: C.HYPCAST_SINK_NAME_VIDEO,
	SinkTypeAudio: C.HYPCAST_SINK_NAME_AUDIO,
}

// SetSink sets the Sink function for a particular data stream, as determined by
// the SinkType.
//
// SetSink must only be called while the pipeline is stopped. The behavior of
// setting a sink on a running pipeline is undefined. It is possible to replace
// an existing sink, but not possible to "unset" a sink. SetSink will panic if
// called with a nil sink function.
func (p *Pipeline) SetSink(sinkType SinkType, sink Sink) {
	if sink == nil {
		panic("sink cannot be nil")
	}

	if p.sinkRefs[sinkType] == nil {
		sinkRef := (*C.HypcastSinkRef)(C.malloc(C.sizeof_HypcastSinkRef))
		sinkRef.pid = p.pid
		sinkRef.sink_type = C.HypcastSinkType(sinkType)
		p.sinkRefs[sinkType] = sinkRef

		C.hypcast_define_sink(p.gstPipeline, sinkNames[sinkType], sinkRef)
	}

	p.sinks[sinkType] = sink
}

// GStreamer calls hypcastSinkSample to pass data from the encoding pipeline
// into Go handler functions. See gst.c for details.
//
//export hypcastSinkSample
func hypcastSinkSample(sinkRef *C.HypcastSinkRef, sample *C.GstSample) C.GstFlowReturn {
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

	sink := getSink(sinkRef)
	sink(data, duration)
	return C.GST_FLOW_OK
}

func getSink(sinkRef *C.HypcastSinkRef) Sink {
	pipeline := getPipeline(sinkRef.pid)
	if pipeline == nil {
		panic("attempted to sink to nonexistent pipeline")
	}

	sinkType := SinkType(sinkRef.sink_type)
	if sinkType >= sinkTypeEnd {
		panic(fmt.Errorf("invalid sink type %d", sinkType))
	}

	sink := pipeline.sinks[sinkType]
	if sink == nil {
		panic("attempted to sink to unregistered sink type")
	}

	return sink
}

var (
	pipelineLock sync.RWMutex
	nextPID      C.HypcastPID = 0
	pipelines                 = make(map[C.HypcastPID]*Pipeline)
)

func registerPipeline(p *Pipeline) {
	if p.pid != 0 {
		panic("duplicate pipeline registration")
	}

	pipelineLock.Lock()
	defer pipelineLock.Unlock()

	nextPID++
	if _, ok := pipelines[nextPID]; ok {
		// If you created a new pipeline every second and never destroyed any of
		// them, even with a 32-bit uint it would take about 136 years to get here.
		// So if we get here we must have seriously corrupted something.
		panic("HypcastPID collision")
	}

	p.pid = nextPID
	pipelines[p.pid] = p
}

func unregisterPipeline(p *Pipeline) {
	pipelineLock.Lock()
	defer pipelineLock.Unlock()

	delete(pipelines, p.pid)
	p.pid = 0
}

func getPipeline(id C.HypcastPID) *Pipeline {
	pipelineLock.RLock()
	defer pipelineLock.RUnlock()

	return pipelines[id]
}
