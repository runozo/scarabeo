# scarabeo

A simple brute-force [Scarabeo](https://boardgamegeek.com/boardgame/12747/scarabeo)
(Italian Scrabble) solver written in Go. It ships with an Italian dictionary;
supporting another language is just a matter of pointing `-dict` at a text file
with one word per line.

## Build

```sh
make build          # builds ./bin/scarabeo
```

The Makefile covers cross-compilation, running, testing and benchmarking. Run
`make` (or `make help`) to list every target:

```sh
make build-all              # linux/amd64, darwin/arm64, windows/amd64
make run-solve CARET=asmmquosd CROSSING=a
make run-play PLAYERS=2 SEED=42 MAX_TURNS=5
make run-top TOP_N=10
make test                   # also: test-race, test-verbose, test-cover
make bench BENCHTIME=1s     # also: bench-cpu, bench-mem
make fmt-check vet          # also: fmt, lint, tidy
```

Or build directly with the Go toolchain:

```sh
go build -o scarabeo .
```

## Usage

```
scarabeo solve [flags] <caret>
scarabeo play  [flags]
scarabeo top   [flags]
scarabeo version
```

Flags shared by every command:

| Flag            | Default            | Description                              |
| --------------- | ------------------ | ---------------------------------------- |
| `-dict`         | `dicts/italia-1a`  | Path to the word list                    |
| `-rack-size`    | `8`                | Number of tiles on the rack              |
| `-allow-punct`  | `false`            | Keep hyphen and apostrophe in words      |
| `-verbose`      | `false`            | Print dictionary statistics              |

Flags must come **before** positional arguments (standard Go flag parsing).

### Solve

Pass your rack letters as the positional argument and, optionally, the letters
you must cross using `-crossing`. In a crossing pattern `_` (or a space) is a
wildcard that matches any letter.

```sh
$ ./scarabeo solve -crossing a asmmquosd
CARET: asmmquosd
CROSSING: a
WORDS: 62

WORD      SCORE
squamosa  51
sudammo   25
ammasso   19
squama    19
...
```

Add `-json` for machine-readable output:

```sh
$ ./scarabeo solve -json -crossing a asmmquosd
```

### Top words

Find the most valuable words buildable from the full set of tiles:

```sh
$ ./scarabeo top -n 10
WORD      SCORE
ziqqurat  86
vaghezza  85
...
```

### Game simulation

Play a full game between automated players. `-seed 0` (the default) uses the
current time; pass a fixed seed for a reproducible game.

```sh
$ ./scarabeo play -players 2 -seed 42
Turn 0
Player 1 plays urgevi (25) | total 25 | rack "sz"
Player 2 plays barena (20) | total 20 | rack "se"
Global crate: 111 tiles
...
```

## Scoring

Letter values follow the Italian Scarabeo rules. Length bonuses are applied on
the number of letters drawn from the rack (word length minus the fixed crossing
letters): 6 letters +10, 7 letters +30, 8 letters +50.

## Dictionary

`dicts/italia-1a` is a raw word list: it contains empty lines, NUL bytes,
uppercase entries, punctuation and a few malformed long lines. The loader cleans
it on the fly, keeping only lowercase `a-z` words (`-allow-punct` also keeps `-`
and `'`), rejecting any line that contains invalid characters, and removing
duplicates. `-verbose` reports how many lines were kept and skipped.

## Tests

```sh
go test ./...
go test -bench . ./internal/engine
```
