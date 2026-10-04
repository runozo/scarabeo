// Command webserve serves the static Scarabeo board page.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

	log.Printf("Scarabeo board: serving %s on http://localhost%s", abs, *addr)
	if err := http.ListenAndServe(*addr, http.FileServer(http.Dir(*dir))); err != nil {
		log.Fatal(err)
	}
}
