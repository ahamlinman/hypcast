package gst

// #include "gst.h"
import "C"
import (
	"fmt"
	"sync"
	"time"
	"unsafe"
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

// Sink is a type for functions that receive stream data from a Pipeline.
type Sink func([]byte, time.Duration)

// SinkType represents the data streams available to a Sink.
type SinkType int

const (
	sinkTypeStart SinkType = iota - 1

	// SinkTypeRaw represents the raw MPEG-TS stream.
	SinkTypeRaw
	// SinkTypeVideo represents the H.264-encoded video stream.
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
		sinkRef.global_pipeline_id = C.uint(p.globalID)
		sinkRef.sink_type = C.uint(sinkType)
		p.sinkRefs[sinkType] = sinkRef

		C.hypcast_define_sink(p.gstPipeline, sinkNames[sinkType], sinkRef)
	}

	p.sinks[sinkType] = sink
}

//export hypcastGlobalSink
func hypcastGlobalSink(sinkRef *C.HypcastSinkRef, buf unsafe.Pointer, bufLen C.int, durNs C.int) {
	sinkType := SinkType(sinkRef.sink_type)
	if sinkType <= sinkTypeStart || sinkType >= sinkTypeEnd {
		panic(fmt.Errorf("invalid sink type ID %d", sinkType))
	}

	pipeline := getGlobalPipeline(globalPipelineID(sinkRef.global_pipeline_id))
	if pipeline == nil {
		panic("attempted to sink to nonexistent pipeline")
	}

	sinkFn := pipeline.sinks[sinkType]
	if sinkFn == nil {
		panic("attempted to sink to unregistered sink type")
	}

	var (
		buffer   = C.GoBytes(buf, bufLen)
		duration = time.Duration(durNs)
	)
	sinkFn(buffer, duration)
}

type globalPipelineID C.uint

var (
	globalPipelineLock   sync.RWMutex
	nextGlobalPipelineID globalPipelineID = 0
	globalPipelineMap                     = make(map[globalPipelineID]*Pipeline)
)

func registerGlobalPipeline(p *Pipeline) {
	if p.globalID != 0 {
		panic("attempted to reregister global pipeline")
	}

	globalPipelineLock.Lock()
	defer globalPipelineLock.Unlock()

	nextGlobalPipelineID++
	if _, ok := globalPipelineMap[nextGlobalPipelineID]; ok {
		// If you created a new pipeline every second and never destroyed any of
		// them, even with a 32-bit uint it would take about 136 years to get here.
		// So if we get here we must have seriously corrupted something.
		panic("global pipeline ID collision")
	}

	p.globalID = nextGlobalPipelineID
	globalPipelineMap[p.globalID] = p
}

func unregisterGlobalPipeline(p *Pipeline) {
	globalPipelineLock.Lock()
	defer globalPipelineLock.Unlock()

	delete(globalPipelineMap, p.globalID)
	p.globalID = 0
}

func getGlobalPipeline(id globalPipelineID) *Pipeline {
	globalPipelineLock.RLock()
	defer globalPipelineLock.RUnlock()

	return globalPipelineMap[id]
}
