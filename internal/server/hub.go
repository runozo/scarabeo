package server

import (
	"math/rand"
	"sync"

	"github.com/runozo/scarabeo/internal/play"
)

// Hub keeps track of all active game rooms.
type Hub struct {
	mu        sync.Mutex
	rooms     map[string]*Room
	rng       *rand.Rand
	validator play.WordValidator
}

// NewHub creates an empty hub. validator is shared by every room to check the
// words played; pass nil to disable word checking.
func NewHub(validator play.WordValidator) *Hub {
	return &Hub{
		rooms:     make(map[string]*Room),
		rng:       rand.New(rand.NewSource(rand.Int63())),
		validator: validator,
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
