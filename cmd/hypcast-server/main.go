package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/ahamlinman/hypcast/internal/atsc"
	"github.com/ahamlinman/hypcast/internal/atsc/gst"
)

func main() {
	var channelsConf string
	flag.StringVar(&channelsConf, "channels", "/etc/hypcast/channels.conf", "Path to channels.conf")
	flag.Parse()

	log.Print("Reading channels.conf")
	f, err := os.Open(channelsConf)
	if err != nil {
		log.Fatalf("Unable to open channels.conf: %v", err)
	}
	channels, err := atsc.ParseChannelsConf(f)
	f.Close()
	if err != nil {
		log.Fatalf("Unable to read channels.conf: %v", err)
	}

	var channel atsc.Channel
	if flag.NArg() > 0 {
		for _, ch := range channels {
			if ch.Name == flag.Arg(0) {
				channel = ch
				break
			}
		}
	}
	if channel == (atsc.Channel{}) {
		channel = channels[0]
	}
	log.Printf("Watching %v", channel)

	log.Print("Initializing GStreamer")
	if err := gst.Init(channel); err != nil {
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
