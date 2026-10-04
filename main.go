// Command scarabeo is a brute-force Scarabeo (Italian Scrabble) solver.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/runozo/scarabeo/internal/engine"
	"github.com/runozo/scarabeo/internal/game"
)

// version is overridable at build time with -ldflags "-X main.version=...".
var version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "solve":
		cmdSolve(os.Args[2:])
	case "play":
		cmdPlay(os.Args[2:])
	case "top":
		cmdTop(os.Args[2:])
	case "version", "--version":
		fmt.Println("scarabeo " + version)
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `scarabeo - brute-force Scarabeo (Italian Scrabble) solver

Usage:
  scarabeo solve [flags] <caret>
  scarabeo play  [flags]
  scarabeo top   [flags]
  scarabeo version

Common flags:
  -dict string      path to the word list (default "dicts/italia-1a")
  -rack-size int    number of tiles on the rack (default 8)
  -allow-punct      keep hyphen and apostrophe in words
  -verbose          print dictionary statistics

Run "scarabeo <command> -h" for command specific flags.
`)
}

// commonFlags groups the flags shared by every command.
type commonFlags struct {
	dict       string
	rackSize   int
	allowPunct bool
	verbose    bool
}

func (c *commonFlags) register(fs *flag.FlagSet) {
	fs.StringVar(&c.dict, "dict", "dicts/italia-1a", "path to the word list")
	fs.IntVar(&c.rackSize, "rack-size", 8, "number of tiles on the rack")
	fs.BoolVar(&c.allowPunct, "allow-punct", false, "keep hyphen and apostrophe in words")
	fs.BoolVar(&c.verbose, "verbose", false, "print dictionary statistics")
}

func (c *commonFlags) load() (*engine.Solver, error) {
	dict, err := engine.LoadDictionary(c.dict, engine.LoadOptions{AllowPunct: c.allowPunct})
	if err != nil {
		return nil, err
	}
	if c.verbose {
		fmt.Fprintf(os.Stderr, "dictionary %s: %d words loaded, %d skipped (of %d lines)\n",
			c.dict, dict.Stats.Loaded, dict.Stats.Skipped, dict.Stats.Total)
	}
	return engine.New(dict, c.rackSize), nil
}

func cmdSolve(args []string) {
	fs := flag.NewFlagSet("solve", flag.ExitOnError)
	var common commonFlags
	common.register(fs)
	crossing := fs.String("crossing", "", "crossing pattern; '_' or space is a wildcard (e.g. a_b)")
	asJSON := fs.Bool("json", false, "emit JSON instead of a table")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: scarabeo solve [flags] <caret>")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	caret := strings.Join(fs.Args(), "")
	if caret == "" {
		fs.Usage()
		os.Exit(2)
	}

	solver, err := common.load()
	if err != nil {
		fatal(err)
	}

	words, err := solver.FindWords(caret, *crossing)
	if err != nil {
		fatal(err)
	}

	if *asJSON {
		if words == nil {
			words = []engine.Word{}
		}
		result := struct {
			Caret    string        `json:"caret"`
			Crossing string        `json:"crossing"`
			Count    int           `json:"count"`
			Words    []engine.Word `json:"words"`
		}{Caret: caret, Crossing: *crossing, Count: len(words), Words: words}
		writeJSON(result)
		return
	}

	fmt.Printf("CARET: %s\n", caret)
	if *crossing != "" {
		fmt.Printf("CROSSING: %s\n", *crossing)
	}
	fmt.Printf("WORDS: %d\n\n", len(words))
	printTable(words)
}

func cmdPlay(args []string) {
	fs := flag.NewFlagSet("play", flag.ExitOnError)
	var common commonFlags
	common.register(fs)
	players := fs.Int("players", 2, "number of players")
	seed := fs.Int64("seed", 0, "random seed (0 = time based)")
	maxTurns := fs.Int("max-turns", 0, "stop after N turns (0 = no limit)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: scarabeo play [flags]")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	solver, err := common.load()
	if err != nil {
		fatal(err)
	}

	if *players < 1 {
		fatal(fmt.Errorf("players must be >= 1"))
	}
	names := make([]string, *players)
	for i := range names {
		names[i] = fmt.Sprintf("Player %d", i+1)
	}

	runSeed := *seed
	if runSeed == 0 {
		runSeed = time.Now().UnixNano()
	}

	result := game.Simulate(solver, game.DefaultGlobalCrate, names, game.Config{
		RackSize: common.rackSize,
		Seed:     runSeed,
		MaxTurns: *maxTurns,
	}, os.Stdout)

	fmt.Printf("\nFinal scores after %d turns:\n", result.Turns)
	for _, p := range result.Players {
		fmt.Printf("  %-12s %d\n", p.Name, p.Score)
	}
}

func cmdTop(args []string) {
	fs := flag.NewFlagSet("top", flag.ExitOnError)
	var common commonFlags
	common.register(fs)
	n := fs.Int("n", 20, "number of words to show (<= 0 for all)")
	asJSON := fs.Bool("json", false, "emit JSON instead of a table")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: scarabeo top [flags]")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	solver, err := common.load()
	if err != nil {
		fatal(err)
	}

	words, err := solver.FindWords(game.DefaultGlobalCrate, "")
	if err != nil {
		fatal(err)
	}
	if *n > 0 && len(words) > *n {
		words = words[:*n]
	}

	if *asJSON {
		writeJSON(words)
		return
	}
	printTable(words)
}

func printTable(words []engine.Word) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "WORD\tSCORE")
	for _, word := range words {
		fmt.Fprintf(w, "%s\t%d\n", word.Text, word.Score)
	}
	_ = w.Flush()
}

func writeJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "scarabeo: "+err.Error())
	os.Exit(1)
}
