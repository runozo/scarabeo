// Command webserve runs the Scarabeo WebSocket game server and serves the
// static client from the web/ directory.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/runozo/scarabeo/internal/engine"
	"github.com/runozo/scarabeo/internal/server"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dir := flag.String("dir", "web", "directory with the static page")
	dictPath := flag.String("dict", "dicts/italia-1a", "path to the word list")
	allowPunct := flag.Bool("allow-punct", false, "keep hyphen and apostrophe in words")
	flag.Parse()

	info, err := os.Stat(*dir)
	if err != nil || !info.IsDir() {
		log.Fatalf("directory %q not found (run from the repository root)", *dir)
	}
	abs, _ := filepath.Abs(*dir)

	dict, err := engine.LoadDictionary(*dictPath, engine.LoadOptions{AllowPunct: *allowPunct})
	if err != nil {
		log.Fatalf("dictionary %q: %v", *dictPath, err)
	}
	log.Printf("dictionary %s: %d words loaded", *dictPath, dict.Stats.Loaded)

	srv := server.New(*dir, dict)
	log.Printf("Scarabeo server: http://localhost%s (WebSocket /ws, static %s)", *addr, abs)
	if err := http.ListenAndServe(*addr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
