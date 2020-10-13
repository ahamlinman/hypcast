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
	"github.com/ahamlinman/hypcast/internal/oldclient"
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

	// TODO: /old-control-socket is a legacy API; use api.Handler directly once it
	// is fully implemented
	mux := http.NewServeMux()
	mux.Handle("/api/", api.NewHandler(tuner))
	mux.Handle("/old-control-socket", oldclient.TunerControlHandler(tuner))

	log.Printf("Starting Hypcast server on %s", *flagAddr)
	server := http.Server{
		Addr:    *flagAddr,
		Handler: mux,
	}
	go server.ListenAndServe()

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
