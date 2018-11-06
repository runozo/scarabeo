# scarabeo

This is a simple brute force [Scrabble](https://boardgamegeek.com/boardgame/12747/scarabeo) game solver. For now it's only for italian version but it's very easy to add dicts for other languages.

## Usage

Just pass your caret letters as first parameter then the letter you need to cross with your word as second parameter (may be also empty). It returns a rough list of (word, score) ordered tuples. 

### Quickstart 

#### check how the solver works
```
$ python3 solver.py "asmmquosd" "a"
('asmmquosd', 'a')
CARET LEN: 9
[('squamosa', 51), ('sudammo', 25), ('ammasso', 19), ('squama', 19), ('squamo', 19), ('qua', 15), ('adusa', 11), ('aduso', 11), ('ammusa', 11), ('muda', 11), ('usammo', 11), ('assuma', 10), ('assumo', 10), ('damma', 10), ('domma', 10), ('suda', 10), ('summa', 10), ('adamo', 9), ('assoda', 9), ('mussa', 9), ('sdama', 9), ('sdamo', 9), ('dama', 8), ('damo', 8), ('doma', 8), ('moda', 8), ('musa', 8), ('suasa', 8), ('suaso', 8), ('dosa', 7), ('soda', 7), ('somma', 7), ('susa', 7), ('uosa', 7), ('ada', 6), ('mamo', 6), ('massa', 6), ('masso', 6), ('mossa', 6), ('oda', 6), ('samoa', 6), ('sua', 6), ('usa', 6), ('ad', 5), ('asma', 5), ('au', 5), ('da', 5), ('maso', 5), ('mass', 5), ('samo', 5), ('soma', 5), ('ama', 4), ('amo', 4), ('asso', 4), ('mao', 4), ('mas', 4), ('moa', 4), ('oma', 4), ('ossa', 4), ('ma', 3), ('osa', 3), ('sa', 2)]
```
#### or launch a simulation

```
$ python3 real_game_simulation.py
A game simulation
Turn 0
King Kong crate: ['l', 'n', 'i', 'e', 'z', 's', 'i', 'v'], len: 8, score: 0, words: []
King Kong plays: ('silenzi', 46)
King Kong crate: ['v'], len: 1, score: 46, words: [('silenzi', 46)]
Hulk crate: ['f', 'o', 'c', 'o', 'r', 'r', 'a', 'm'], len: 8, score: 0, words: []
Hulk plays: ('amorfo', 20)
Hulk crate: ['c', 'r'], len: 2, score: 20, words: [('amorfo', 20)]
Global crate: ['a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'b', 'b', 'b', 'b', 'c', 'c', 'c', 'c', 'c', 'c', 'd', 'd', 'd', 'd', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'f', 'f', 'f', 'g', 'g', 'g', 'g', 'h', 'h', 'i', 'i', 'i', 'i', 'i', 'i', 'i', 'i', 'i', 'l', 'l', 'l', 'l', 'l', 'm', 'm', 'm', 'm', 'm', 'n', 'n', 'n', 'n', 'n', 'o', 'o', 'o', 'o', 'o', 'o', 'o', 'o', 'o', 'o', 'p', 'p', 'p', 'p', 'q', 'q', 'r', 'r', 'r', 'r', 'r', 's', 's', 's', 's', 's', 's', 't', 't', 't', 't', 't', 't', 't', 'u', 'u', 'u', 'u', 'v', 'v', 'v', 'z']
Turn 1
King Kong crate: ['v', 'f', 'n', 't', 'e', 'i', 'a', 'm'], len: 8, score: 46, words: [('silenzi', 46)]
King Kong plays: ('mentiva', 42)
King Kong crate: ['f'], len: 1, score: 88, words: [('silenzi', 46), ('mentiva', 42)]
Hulk crate: ['c', 'r', 'i', 'c', 'q', 'm', 'b', 't'], len: 8, score: 20, words: [('amorfo', 20)]
Hulk plays: ('bit', 6)
Hulk crate: ['c', 'r', 'c', 'q', 'm'], len: 5, score: 26, words: [('amorfo', 20), ('bit', 6)]
Global crate: ['a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'b', 'b', 'b', 'c', 'c', 'c', 'c', 'c', 'd', 'd', 'd', 'd', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'e', 'f', 'f', 'g', 'g', 'g', 'g', 'h', 'h', 'i', 'i', 'i', 'i', 'i', 'i', 'i', 'l', 'l', 'l', 'l', 'l', 'm', 'm', 'm', 'n', 'n', 'n', 'n', 'o', 'o', 'o', 'o', 'o', 'o', 'o', 'o', 'o', 'o', 'p', 'p', 'p', 'p', 'q', 'r', 'r', 'r', 'r', 'r', 's', 's', 's', 's', 's', 's', 't', 't', 't', 't', 't', 'u', 'u', 'u', 'u', 'v', 'v', 'v', 'z']
...
...
```


## TODO

Make a RESTful service.

