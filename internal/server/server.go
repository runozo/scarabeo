// Package server exposes the Scarabeo game over WebSockets and serves the
// static client. All game rules are enforced here, on the server.
package server

import (
	"context"
	"log"
	"net/http"

	"github.com/coder/websocket"
	"github.com/runozo/scarabeo/internal/play"
)

// Server ties the room hub to the HTTP handlers.
type Server struct {
	hub       *Hub
	staticDir string
}

// New creates a server serving static files from staticDir and games on /ws.
// validator is used to check the words played; pass nil to disable word
// checking.
func New(staticDir string, validator play.WordValidator) *Server {
	return &Server{hub: NewHub(validator), staticDir: staticDir}
}

// Handler returns the HTTP handler (static files + /ws).
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWS)
	mux.Handle("/", noCache(http.FileServer(http.Dir(s.staticDir))))
	return mux
}

// noCache forces the browser to revalidate the static client on every load, so
// a stale cached app.js cannot keep an old bug alive after the server changes.
func noCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		h.ServeHTTP(w, r)
	})
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// The page is served by the same server, but this keeps things working
		// when it is opened through a different host/port during development.
		InsecureSkipVerify: true,
	})
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	c := &Client{conn: conn, hub: s.hub, send: make(chan []byte, 32)}
	go c.writePump(ctx)

	c.readPump(ctx)

	cancel()
	_ = conn.Close(websocket.StatusNormalClosure, "bye")
	log.Printf("connessione chiusa (%s)", c.playerID)
}
