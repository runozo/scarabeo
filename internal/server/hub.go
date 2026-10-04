package server

import (
	"math/rand"
	"sync"
)

// Hub keeps track of all active game rooms.
type Hub struct {
	mu    sync.Mutex
	rooms map[string]*Room
	rng   *rand.Rand
}

// NewHub creates an empty hub.
func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]*Room),
		rng:   rand.New(rand.NewSource(rand.Int63())),
	}
}

// create allocates a new room with a fresh code.
func (h *Hub) create() *Room {
	h.mu.Lock()
	defer h.mu.Unlock()
	code := h.uniqueCodeLocked()
	r := newRoom(code, h)
	h.rooms[code] = r
	return r
}

// get returns an existing room, or nil.
func (h *Hub) get(code string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.rooms[code]
}

// remove deletes a room.
func (h *Hub) remove(code string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.rooms, code)
}

func (h *Hub) uniqueCodeLocked() string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	for {
		b := make([]byte, 4)
		for i := range b {
			b[i] = alphabet[h.rng.Intn(len(alphabet))]
		}
		code := string(b)
		if _, exists := h.rooms[code]; !exists {
			return code
		}
	}
}
