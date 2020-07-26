package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	log.Print("Initializing GStreamer pipeline")
	pipeline, err := gst.NewPipeline(channel)
	if err != nil {
		log.Fatal(err)
	}

	h, err := newSocketHandler()
	if err != nil {
		log.Fatal("unable to create socket handler", err)
	}

	pipeline.SetSink(gst.SinkTypeVideo, h.HandleVideoData)
	pipeline.SetSink(gst.SinkTypeAudio, h.HandleAudioData)

	http.Handle("/control-socket", h)

	log.Print("Starting pipeline")
	pipeline.Start()
	defer func() {
		log.Printf("Result of closing pipeline: %v", pipeline.Close())
	}()

	log.Print("Starting web server")
	server := http.Server{Addr: ":9200"}
	go server.ListenAndServe()

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	<-signalCh

	log.Print("Stopping web server")
	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(stopCtx)
}
