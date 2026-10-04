/* Scarabeo — tavolo di gioco (versione italiana, 17x17)
 * Tileset SVG originale, layout ricostruito dalla plancia Editrice Giochi.
 * Regole: prima mossa sul centro, collegamento, linea continua, +100 SCARABEO,
 * parole giocate bloccate, turni tra 2-4 giocatori.
 */
(() => {
  "use strict";

  const NS = "http://www.w3.org/2000/svg";
  const SIZE = 17;
  const CENTER = 8;
  const RACK_SIZE = 8;

  // --- Layout della plancia ---------------------------------------------
  // N normale · L doppia lettera (2L) · T tripla lettera (3L)
  // D doppia parola (2P) · W tripla parola (3P) · C centro (scarabeo)
  const LAYOUT = [
    "WNNNLNNNWNNNLNNNW",
    "NDNNNNTNNNTNNNNDN",
    "NNDNNNNLNLNNNNDNN",
    "NNNDNNNNLNNNNDNNN",
    "LNNNDNNNNNNNDNNNL",
    "NNNNNDNNNNNDNNNNN",
    "NTNNNNTNNNTNNNNTN",
    "NNLNNNNLNLNNNNLNN",
    "NNNLNNNNCNNNNLNNN",
    "NNLNNNNLNLNNNNLNN",
    "NTNNNNTNNNTNNNNTN",
    "NNNNNDNNNNNDNNNNN",
    "LNNNDNNNNNNNDNNNL",
    "NNNDNNNNLNNNNDNNN",
    "NNDNNNNLNLNNNNDNN",
    "NDNNNNTNNNTNNNNDN",
    "WNNNLNNNWNNNLNNNW",
  ];

  const CELL_LABEL = { L: "2L", T: "3L", D: "2P", W: "3P" };
  const CELL_CLASS = { N: "", L: "dl", T: "tl", D: "dw", W: "tw", C: "center" };

  // --- Valori e distribuzione (Scarabeo) --------------------------------
  const VALUES = {
    A: 1, B: 4, C: 1, D: 4, E: 1, F: 4, G: 4, H: 8, I: 1, L: 2,
    M: 2, N: 2, O: 1, P: 3, Q: 10, R: 1, S: 1, T: 1, U: 4, V: 4, Z: 8,
    "?": 0,
  };
  const COUNTS = {
    A: 12, B: 4, C: 7, D: 4, E: 12, F: 4, G: 4, H: 2, I: 12, L: 6,
    M: 6, N: 6, O: 12, P: 4, Q: 2, R: 7, S: 7, T: 7, U: 4, V: 4, Z: 2,
    "?": 2,
  };
  const LETTERS = Object.keys(VALUES);
  const SCARABEO_BONUS = 100;

  const letterMult = (t) => (t === "L" ? 2 : t === "T" ? 3 : 1);
  const wordMult = (t) => (t === "D" ? 2 : t === "W" ? 3 : 1);
  const lengthBonus = (n) => (n === 6 ? 10 : n === 7 ? 30 : n === 8 ? 50 : 0);

  // --- Tileset SVG (definizioni <symbol>) -------------------------------
  function buildDefs() {
    let s = '<defs><linearGradient id="ivory" x1="0" y1="0" x2="0" y2="1">' +
      '<stop offset="0" stop-color="#fbf3dd"/><stop offset="1" stop-color="#eadfbf"/>' +
      "</linearGradient></defs>";
    for (const L of LETTERS) {
      const glyph = L === "?"
        ? '<g fill="#c0392b"><ellipse cx="50" cy="56" rx="19" ry="25"/>' +
          '<circle cx="50" cy="30" r="9"/>' +
          '<path d="M31 42 L17 34 M31 56 L15 56 M31 70 L17 78 M69 42 L83 34 M69 56 L85 56 M69 70 L83 78" ' +
          'stroke="#c0392b" stroke-width="5" fill="none" stroke-linecap="round"/>' +
          '<line x1="50" y1="38" x2="50" y2="79" stroke="#8e2b1f" stroke-width="3"/></g>'
        : `<text class="tile-letter" x="50" y="66" text-anchor="middle">${L}</text>` +
          `<text class="tile-value" x="82" y="90" text-anchor="middle">${VALUES[L]}</text>`;
      s += `<symbol id="tile-${L}" viewBox="0 0 100 100">` +
        '<rect x="4" y="4" width="92" height="92" rx="11" fill="url(#ivory)" stroke="#b39a5f" stroke-width="3"/>' +
        glyph + "</symbol>";
    }
    return s;
  }

  function tileSVG(letter) {
    const svg = document.createElementNS(NS, "svg");
    svg.setAttribute("viewBox", "0 0 100 100");
    svg.setAttribute("class", "tile");
    const use = document.createElementNS(NS, "use");
    use.setAttribute("href", "#tile-" + letter);
    use.setAttributeNS("http://www.w3.org/1999/xlink", "xlink:href", "#tile-" + letter);
    svg.appendChild(use);
    return svg;
  }

  // --- Stato ------------------------------------------------------------
  let bag = [];
  let board = [];        // board[r][c] = {type} | {letter,value,type,jolly,assigned}
  let players = [];      // {name, rack:[], score}
  let current = 0;
  let move = [];         // celle posate in questo turno
  let moveSet = new Set();
  let selected = -1;
  let gameOver = false;

  const $ = (id) => document.getElementById(id);

  function newBag() {
    const b = [];
    for (const L of LETTERS) for (let i = 0; i < COUNTS[L]; i++) b.push(L);
    for (let i = b.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [b[i], b[j]] = [b[j], b[i]];
    }
    return b;
  }

  function emptyBoard() {
    return Array.from({ length: SIZE }, (_, r) =>
      Array.from({ length: SIZE }, (_, c) => ({ type: LAYOUT[r][c] }))
    );
  }

  function boardIsEmpty() {
    for (let r = 0; r < SIZE; r++)
      for (let c = 0; c < SIZE; c++)
        if (board[r][c].letter !== undefined && !moveSet.has(r + "," + c)) return false;
    return true;
  }

  // --- Partita ----------------------------------------------------------
  function startGame(n) {
    bag = newBag();
    board = emptyBoard();
    players = Array.from({ length: n }, (_, i) => ({ name: `Giocatore ${i + 1}`, rack: [], score: 0 }));
    current = 0;
    move = [];
    moveSet = new Set();
    selected = -1;
    gameOver = false;
    for (const p of players) drawFor(p);
    renderBoard();
    renderPlayers();
    renderRack();
    updateScore();
    setStatus(`Inizia ${players[current].name}. La prima parola deve coprire il centro (🪲).`);
  }

  function drawFor(p) {
    while (p.rack.length < RACK_SIZE && bag.length > 0) p.rack.push(bag.pop());
  }

  // --- Rendering --------------------------------------------------------
  function renderBoard() {
    const el = $("board");
    el.innerHTML = "";
    for (let r = 0; r < SIZE; r++) {
      for (let c = 0; c < SIZE; c++) {
        const cell = document.createElement("div");
        const type = LAYOUT[r][c];
        cell.className = "cell " + CELL_CLASS[type];
        cell.dataset.r = r;
        cell.dataset.c = c;
        cell.setAttribute("role", "gridcell");
        if (type === "C") cell.textContent = "🪲";
        else if (CELL_LABEL[type]) cell.textContent = CELL_LABEL[type];
        cell.addEventListener("click", () => onCellClick(r, c));
        el.appendChild(cell);
      }
    }
    paintTiles();
  }

  const cellEl = (r, c) => $("board").children[r * SIZE + c];

  function paintTiles() {
    for (let r = 0; r < SIZE; r++) {
      for (let c = 0; c < SIZE; c++) {
        const el = cellEl(r, c);
        const t = board[r][c];
        el.classList.remove("filled", "locked", "pending");
        el.querySelectorAll("svg.tile").forEach((n) => n.remove());
        if (t && t.letter !== undefined) {
          const letter = t.jolly ? (t.assigned || "?") : t.letter;
          el.appendChild(tileSVG(letter));
          if (moveSet.has(r + "," + c)) el.classList.add("pending");
          else el.classList.add("filled", "locked");
        }
      }
    }
  }

  function renderPlayers() {
    const ul = $("players");
    ul.innerHTML = "";
    players.forEach((p, i) => {
      const li = document.createElement("li");
      li.className = i === current && !gameOver ? "current" : "";
      li.innerHTML =
        `<span class="pname">${p.name}</span>` +
        `<span class="pmeta">${p.score} pt · ${p.rack.length} tessere</span>`;
      ul.appendChild(li);
    });
    $("current-player").textContent = players[current] ? players[current].name : "—";
  }

  function renderRack() {
    const el = $("rack");
    el.innerHTML = "";
    const rack = players[current] ? players[current].rack : [];
    for (let i = 0; i < RACK_SIZE; i++) {
      const L = rack[i];
      const btn = document.createElement("button");
      btn.className = "tile-btn" + (i === selected ? " selected" : "");
      btn.type = "button";
      if (L) {
        btn.appendChild(tileSVG(L));
        btn.title = L === "?" ? "Scarabeo (jolly)" : `Tessera ${L} (${VALUES[L]})`;
      } else {
        btn.disabled = true;
      }
      btn.addEventListener("click", () => onRackClick(i));
      el.appendChild(btn);
    }
    $("bag-count").textContent = String(bag.length);
  }

  // --- Interazione ------------------------------------------------------
  function onRackClick(i) {
    if (gameOver) return;
    const rack = players[current].rack;
    if (!rack[i]) return;
    selected = selected === i ? -1 : i;
    renderRack();
  }

  function onCellClick(r, c) {
    if (gameOver) return;
    const t = board[r][c];
    const key = r + "," + c;

    // riprende una tessera del turno corrente
    if (t && t.letter !== undefined && moveSet.has(key)) {
      players[current].rack.push(t.jolly ? "?" : t.letter);
      board[r][c] = { type: LAYOUT[r][c] };
      move = move.filter((m) => !(m.r === r && m.c === c));
      moveSet.delete(key);
      paintTiles();
      renderRack();
      updateScore();
      return;
    }
    // parola già giocata: bloccata
    if (t && t.letter !== undefined) {
      setStatus("Questa parola è già stata giocata e non può essere modificata.");
      return;
    }

    const rack = players[current].rack;
    if (selected < 0 || !rack[selected]) {
      setStatus("Seleziona prima una tessera dal rack.");
      return;
    }

    let letter = rack[selected];
    let jolly = false;
    let assigned = null;
    if (letter === "?") {
      const ans = (prompt("Scarabeo (jolly): quale lettera rappresenta? (A-Z)") || "").trim().toUpperCase();
      if (!/^[A-Z]$/.test(ans) || !(ans in VALUES)) {
        setStatus("Lettera non valida per lo scarabeo.");
        return;
      }
      jolly = true;
      assigned = ans;
    }

    board[r][c] = {
      type: LAYOUT[r][c],
      letter,
      value: jolly ? VALUES[assigned] : VALUES[letter],
      jolly,
      assigned,
    };
    rack.splice(selected, 1);
    selected = -1;
    move.push({ r, c });
    moveSet.add(key);

    paintTiles();
    renderRack();
    updateScore();
    setStatus("");
  }

  // --- Punteggio e parole ----------------------------------------------
  function cellAt(r, c) {
    if (r < 0 || c < 0 || r >= SIZE || c >= SIZE) return null;
    const t = board[r][c];
    return t && t.letter !== undefined ? t : null;
  }

  function run(r, c, dir) {
    const dr = dir === "V" ? 1 : 0;
    const dc = dir === "H" ? 1 : 0;
    let sr = r, sc = c;
    while (cellAt(sr - dr, sc - dc)) { sr -= dr; sc -= dc; }
    const cells = [];
    let cr = sr, cc = sc;
    while (cellAt(cr, cc)) { cells.push({ r: cr, c: cc }); cr += dr; cc += dc; }
    return cells;
  }

  const letterOf = (cell) => (cell.jolly ? (cell.assigned || "") : cell.letter);

  function wordsForMove() {
    const seen = new Set();
    const words = [];
    for (const { r, c } of move) {
      for (const dir of ["H", "V"]) {
        const cells = run(r, c, dir);
        if (cells.length < 2) continue;
        const key = dir + ":" + cells.map((x) => x.r + "," + x.c).join(";");
        if (seen.has(key)) continue;
        seen.add(key);
        words.push({ dir, cells, text: cells.map(({ r, c }) => letterOf(board[r][c])).join("") });
      }
    }
    return words;
  }

  function scoreWord(cells) {
    let sum = 0, mult = 1;
    for (const { r, c } of cells) {
      const cell = board[r][c];
      if (moveSet.has(r + "," + c)) {
        sum += cell.value * letterMult(cell.type);
        mult *= wordMult(cell.type);
      } else {
        sum += cell.value;
      }
    }
    return sum * mult;
  }

  function updateScore() {
    const words = wordsForMove();
    let score = words.reduce((s, w) => s + scoreWord(w.cells), 0);
    const bonus = lengthBonus(move.length);
    const scarabeo = words.some((w) => w.text === "SCARABEO") ? SCARABEO_BONUS : 0;
    score += bonus + scarabeo;
    const parts = [];
    if (bonus) parts.push(`bonus lunghezza +${bonus}`);
    if (scarabeo) parts.push(`SCARABEO +${scarabeo}`);
    $("move-score").textContent = "Punteggio mossa: " + score + (parts.length ? ` (${parts.join(", ")})` : "");
    return score;
  }

  // --- Validazione delle regole ----------------------------------------
  function validateMove() {
    if (move.length === 0) return "Nessuna tessera posata.";

    const rows = new Set(move.map((m) => m.r));
    const cols = new Set(move.map((m) => m.c));
    if (rows.size > 1 && cols.size > 1) {
      return "Le tessere devono essere tutte sulla stessa riga o colonna.";
    }
    const horizontal = rows.size === 1;
    const sorted = [...move].sort((a, b) => (horizontal ? a.c - b.c : a.r - b.r));
    const fixed = horizontal ? sorted[0].r : sorted[0].c;
    const start = horizontal ? sorted[0].c : sorted[0].r;
    const end = horizontal ? sorted[sorted.length - 1].c : sorted[sorted.length - 1].r;
    for (let i = start; i <= end; i++) {
      const r = horizontal ? fixed : i;
      const c = horizontal ? i : fixed;
      if (!cellAt(r, c)) return "Le tessere devono essere adiacenti, senza spazi vuoti.";
    }

    if (wordsForMove().length === 0) {
      return "La mossa non forma nessuna parola (servono almeno 2 lettere).";
    }

    if (boardIsEmpty()) {
      if (!move.some((m) => m.r === CENTER && m.c === CENTER)) {
        return "La prima parola deve coprire il centro (🪲).";
      }
    } else {
      const connected = move.some((m) =>
        [[1, 0], [-1, 0], [0, 1], [0, -1]].some(([dr, dc]) => {
          const rr = m.r + dr, cc = m.c + dc;
          const t = cellAt(rr, cc);
          return t && !moveSet.has(rr + "," + cc);
        })
      );
      if (!connected) return "La mossa deve collegarsi ad almeno una lettera già presente.";
    }
    return null;
  }

  // --- Azioni -----------------------------------------------------------
  function commit() {
    if (gameOver) return;
    const err = validateMove();
    if (err) { setStatus("⚠ " + err); return; }

    const score = updateScore();
    const player = players[current];
    player.score += score;

    move = [];
    moveSet = new Set();
    paintTiles();
    updateScore();

    drawFor(player);
    const who = player.name;
    current = (current + 1) % players.length;
    renderPlayers();
    renderRack();
    updateScore();

    if (bag.length === 0 && players.every((p) => p.rack.length === 0)) {
      gameOver = true;
      const best = players.reduce((a, b) => (b.score > a.score ? b : a));
      setStatus(`Partita finita! Vince ${best.name} con ${best.score} punti.`);
    } else {
      setStatus(`${who}: +${score} punti. Tocca a ${players[current].name}.`);
    }
  }

  function undo() {
    if (gameOver) return;
    for (const { r, c } of move) {
      const t = board[r][c];
      if (t && t.letter !== undefined) players[current].rack.push(t.jolly ? "?" : t.letter);
      board[r][c] = { type: LAYOUT[r][c] };
    }
    move = [];
    moveSet = new Set();
    selected = -1;
    paintTiles();
    renderRack();
    updateScore();
    setStatus("Mossa annullata.");
  }

  function setStatus(msg) {
    $("status").textContent = msg;
  }

  // --- Tileset (pannello) ----------------------------------------------
  function renderTileset() {
    const el = $("tileset");
    el.innerHTML = "";
    for (const L of LETTERS) {
      const wrap = document.createElement("div");
      wrap.title = L === "?" ? "Scarabeo (jolly)" : `${L} · ${VALUES[L]} punti · ${COUNTS[L]} tessere`;
      wrap.appendChild(tileSVG(L));
      const cap = document.createElement("div");
      cap.style.cssText = "font-size:10px;text-align:center;color:#6b5a33";
      cap.textContent = `×${COUNTS[L]}`;
      wrap.appendChild(cap);
      el.appendChild(wrap);
    }
  }

  // --- Avvio ------------------------------------------------------------
  function init() {
    const defsHost = document.createElementNS(NS, "svg");
    defsHost.setAttribute("width", "0");
    defsHost.setAttribute("height", "0");
    defsHost.style.position = "absolute";
    defsHost.innerHTML = buildDefs();
    document.body.insertBefore(defsHost, document.body.firstChild);

    renderTileset();
    startGame(parseInt($("player-count").value, 10) || 2);

    $("commit").addEventListener("click", commit);
    $("undo").addEventListener("click", undo);
    $("new-game").addEventListener("click", () =>
      startGame(parseInt($("player-count").value, 10) || 2)
    );
    $("player-count").addEventListener("change", () =>
      startGame(parseInt($("player-count").value, 10) || 2)
    );

    // Hook di test opzionale: attivo solo se window.__SCARABEO_TEST__ è true.
    if (window.__SCARABEO_TEST__) {
      window.__scarabeo = {
        startGame,
        commit,
        undo,
        setRack(letters) { players[current].rack = letters.slice(); selected = -1; renderRack(); },
        place(r, c, letter) {
          board[r][c] = { type: LAYOUT[r][c], letter, value: VALUES[letter] };
          move.push({ r, c });
          moveSet.add(r + "," + c);
          paintTiles();
          updateScore();
        },
        currentScore: () => players[current].score,
        validate: () => validateMove(),
        moveScore: () => updateScore(),
      };
    }
  }

  document.addEventListener("DOMContentLoaded", init);
})();
