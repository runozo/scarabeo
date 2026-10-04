// Command webserve runs the Scarabeo WebSocket game server and serves the
// static client from the web/ directory.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/runozo/scarabeo/internal/server"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dir := flag.String("dir", "web", "directory with the static page")
	flag.Parse()

	info, err := os.Stat(*dir)
	if err != nil || !info.IsDir() {
		log.Fatalf("directory %q not found (run from the repository root)", *dir)
	}
	abs, _ := filepath.Abs(*dir)

	srv := server.New(*dir)
	log.Printf("Scarabeo server: http://localhost%s (WebSocket /ws, static %s)", *addr, abs)
	if err := http.ListenAndServe(*addr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
