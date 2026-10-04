// Package server exposes the Scarabeo game over WebSockets and serves the
// static client. All game rules are enforced here, on the server.
package server

import (
	"context"
	"log"
	"net/http"

	"github.com/coder/websocket"
)

// Server ties the room hub to the HTTP handlers.
type Server struct {
	hub       *Hub
	staticDir string
}

// New creates a server serving static files from staticDir and games on /ws.
func New(staticDir string) *Server {
	return &Server{hub: NewHub(), staticDir: staticDir}
}

// Handler returns the HTTP handler (static files + /ws).
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWS)
	mux.Handle("/", http.FileServer(http.Dir(s.staticDir)))
	return mux
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
