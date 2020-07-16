package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/ahamlinman/hypcast/internal/gst"
)

func main() {
	if err := gst.Init(); err != nil {
		log.Fatal(err)
	}

	gst.SetSink(gst.SinkTypeRaw, sampleSink("raw"))
	gst.SetSink(gst.SinkTypeVideo, sampleSink("video"))
	gst.SetSink(gst.SinkTypeAudio, sampleSink("audio"))

	gst.Play()

	wait := make(chan os.Signal)
	signal.Notify(wait, os.Interrupt)
	<-wait
}

func sampleSink(name string) gst.Sink {
	return func(buffer []byte, duration time.Duration) {
		log.Printf("%s: %v %v{%d}", name, duration, buffer[:10], len(buffer))
	}
}
