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

	h, err := newSocketHandler()
	if err != nil {
		log.Fatal("unable to create socket handler", err)
	}

	gst.SetSink(gst.SinkTypeVideo, h.HandleVideoData)
	gst.SetSink(gst.SinkTypeAudio, h.HandleAudioData)

	http.Handle("/control-socket", h)

	log.Print("Starting pipeline")
	gst.Play()

	log.Print("Starting web server")
	http.ListenAndServe(":9200", nil)
}
