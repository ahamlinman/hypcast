package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	// FIXME: Unsafe; for testing purposes only
	CheckOrigin: func(_ *http.Request) bool { return true },
}

type socketHandler struct {
	mu     sync.Mutex
	locked bool
	ws     *websocket.Conn
}

func (h *socketHandler) tryObtainingLock() bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.locked {
		return false
	}

	h.locked = true
	return true
}

func (h *socketHandler) unlock() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.locked = false
	h.ws.Close()
	h.ws = nil
}

func (h *socketHandler) HandleAudioData(buffer []byte, d time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.ws == nil {
		return
	}

	h.ws.WriteJSON(struct {
		Kind     string
		Duration time.Duration
	}{"audio", d})
}

func (h *socketHandler) HandleVideoData(buffer []byte, d time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.ws == nil {
		return
	}

	h.ws.WriteJSON(struct {
		Kind     string
		Duration time.Duration
	}{"video", d})
}

func (h *socketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request(%p): Received", r)

	if !h.tryObtainingLock() {
		log.Printf("Request(%p): Rejected due to existing client", r)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	defer h.unlock()

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
