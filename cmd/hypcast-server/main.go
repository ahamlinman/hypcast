package main

import (
	"log"
	"net/http"

	"github.com/ahamlinman/hypcast/internal/gst"
)

func main() {
	log.Print("Initializing GStreamer")
	if err := gst.Init(); err != nil {
		log.Fatal(err)
	}

	var h socketHandler
	gst.SetSink(gst.SinkTypeVideo, h.HandleVideoData)
	gst.SetSink(gst.SinkTypeAudio, h.HandleAudioData)

	http.Handle("/hypcast/ws", &h)

	log.Print("Starting pipeline")
	gst.Play()

	log.Print("Starting web server")
	http.ListenAndServe(":9200", nil)
}
