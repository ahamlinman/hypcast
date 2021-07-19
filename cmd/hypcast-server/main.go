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

	"github.com/ahamlinman/hypcast/client"
	"github.com/ahamlinman/hypcast/internal/api"
	"github.com/ahamlinman/hypcast/internal/assets"
	"github.com/ahamlinman/hypcast/internal/atsc"
	"github.com/ahamlinman/hypcast/internal/atsc/tuner"
)

var (
	flagAddr          string
	flagChannels      string
	flagAssets        string
	flagVideoPipeline string
)

func init() {
	flag.StringVar(
		&flagAddr, "addr", ":9200",
		"Address for the HTTP server to listen on",
	)
	flag.StringVar(
		&flagChannels, "channels", "/etc/hypcast/channels.conf",
		"Path to the channels.conf file containing the list of available channels",
	)
	flag.StringVar(
		&flagAssets, "assets", "",
		"Path to client assets; overrides any embedded assets",
	)
	flag.StringVar(
		&flagVideoPipeline, "video-pipeline", "default",
		`Video pipeline implementation; use "vaapi" for VA-API hardware acceleration`,
	)
}

func main() {
	flag.Parse()

	log.Printf("Using channels from %s", flagChannels)
	channels, err := readChannelsConf(flagChannels)
	if err != nil {
		log.Fatalf("Unable to read channels.conf: %v", err)
	}

	vp := tuner.ParseVideoPipeline(flagVideoPipeline)
	log.Printf("Using %s video pipeline", vp)
	tuner := tuner.NewTuner(channels, vp)
	http.Handle("/api/", api.NewHandler(tuner))

	if flagAssets != "" {
		log.Printf("Using client assets from %s", flagAssets)
		http.Handle("/", http.FileServer(
			assets.FileSystem{FileSystem: http.Dir(flagAssets)},
		))
	} else if client.Handler != nil {
		log.Print("Using embedded client assets")
		http.Handle("/", client.Handler)
	}

	server := http.Server{Addr: flagAddr}
	go server.ListenAndServe()
	log.Printf("Started Hypcast server on %s", flagAddr)

	signalCh := make(chan os.Signal, 1)
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
