package main

import (
	"log"
	"time"

	"github.com/ahamlinman/hypcast/internal/gst"
)

func main() {
	if err := gst.Init(); err != nil {
		log.Fatal(err)
	}

	gst.SetSink(gst.SinkTypeVideo, func(buffer []byte, duration time.Duration) {
		log.Printf("video: %v %v", duration, buffer[:10])
	})

	log.Print("stay tuned...")
}
