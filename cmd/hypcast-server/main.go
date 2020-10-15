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

	"github.com/ahamlinman/hypcast/internal/api"
	"github.com/ahamlinman/hypcast/internal/atsc"
	"github.com/ahamlinman/hypcast/internal/atsc/tuner"
)

func main() {
	var (
		flagAddr = flag.String(
			"addr", ":9200",
			"Address for the HTTP server to listen on",
		)
		flagChannels = flag.String(
			"channels", "/etc/hypcast/channels.conf",
			"Path to the channels.conf file containing the list of available channels",
		)
	)
	flag.Parse()

	channels, err := readChannelsConf(*flagChannels)
	if err != nil {
		log.Fatalf("Unable to read channels.conf: %v", err)
	}

	tuner := tuner.NewTuner(channels)

	http.Handle("/api/", api.NewHandler(tuner))

	server := http.Server{Addr: *flagAddr}
	go server.ListenAndServe()
	log.Printf("Started Hypcast server on %s", *flagAddr)

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	<-signalCh

	log.Print("Shutting down")
	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(stopCtx)
}

func readChannelsConf(path string) ([]atsc.Channel, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return atsc.ParseChannelsConf(f)
}
