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
	"github.com/ahamlinman/hypcast/internal/atsc/tuner"
	"github.com/ahamlinman/hypcast/internal/oldclient"
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

	tuner := tuner.NewTuner(channels)

	http.Handle("/config/channels", oldclient.ChannelListHandler(channels))
	http.Handle("/old-control-socket", oldclient.TunerControlHandler(tuner))

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
