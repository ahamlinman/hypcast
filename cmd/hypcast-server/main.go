package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ahamlinman/hypcast/internal/gst"
	"github.com/gorilla/websocket"
)

func main() {
	log.Print("Initializing GStreamer")
	if err := gst.Init(); err != nil {
		log.Fatal(err)
	}

	var h socketHandler
	gst.SetSink(gst.SinkTypeAudio, h.HandlePipelineData)

	http.Handle("/hypcast/ws", &h)

	log.Print("Starting pipeline")
	gst.Play()

	log.Print("Starting web server")
	http.ListenAndServe(":9200", nil)
}

var upgrader = websocket.Upgrader{
	// FIXME: Unsafe; for testing purposes only
	CheckOrigin: func(_ *http.Request) bool { return true },
}

type socketHandler struct {
	mu     sync.Mutex
	locked bool
	ws     *websocket.Conn
}

func (h *socketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request(%p): Received", r)

	tryLockingThisClient := func() bool {
		h.mu.Lock()
		defer h.mu.Unlock()

		if h.locked {
			return false
		}

		h.locked = true
		return true
	}

	if !tryLockingThisClient() {
		log.Printf("Request(%p): Rejected due to existing client", r)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	defer func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		h.locked = false
		h.ws.Close()
		h.ws = nil
	}()

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Request(%p): Failed to upgrade connection: %v", r, err)
		return
	}

	h.mu.Lock()
	h.ws = ws
	h.mu.Unlock()

	for err := error(nil); err == nil; _, _, err = ws.ReadMessage() {
	}

	log.Printf("Request(%p): Connection finished", r)
}

func (h *socketHandler) HandlePipelineData(buffer []byte, _ time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.ws != nil {
		h.ws.WriteMessage(websocket.BinaryMessage, buffer)
	}
}
