/* Scarabeo — client WebSocket.
 * Tutta la validazione delle regole e il punteggio vivono sul server.
 * Il client mostra lo stato ricevuto e invia le mosse.
 */
(() => {
  "use strict";

  const NS = "http://www.w3.org/2000/svg";
  const SIZE = 17;

  // --- Layout della plancia (statico, uguale al server) -----------------
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

  const VALUES = {
    A: 1, B: 4, C: 1, D: 4, E: 1, F: 4, G: 4, H: 8, I: 1, L: 2,
    M: 2, N: 2, O: 1, P: 3, Q: 10, R: 1, S: 1, T: 1, U: 4, V: 4, Z: 8, "?": 0,
  };
  const COUNTS = {
    A: 12, B: 4, C: 7, D: 4, E: 12, F: 4, G: 4, H: 2, I: 12, L: 6,
    M: 6, N: 6, O: 12, P: 4, Q: 2, R: 7, S: 7, T: 7, U: 4, V: 4, Z: 2, "?": 2,
  };
  const LETTERS = Object.keys(VALUES);

  const $ = (id) => document.getElementById(id);

  // --- Tileset SVG ------------------------------------------------------
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

  // --- Stato del client -------------------------------------------------
  let ws = null;
  let myId = null;
  let host = false;
  let gameCode = "";
  let state = null;        // ultimo stato ricevuto dal server
  let pending = [];        // posizionamenti locali non ancora confermati
  let preview = null;      // ultima preview del server
  let selected = -1;
  let previewTimer = null;
  let awaitingMove = false;

  // --- Connessione ------------------------------------------------------
  function connect() {
    const proto = location.protocol === "https:" ? "wss" : "ws";
    const url = `${proto}://${location.host}/ws`;
    setConn(`Connessione a ${url}…`);
    ws = new WebSocket(url);
    ws.onopen = () => setConn("Connesso. Inserisci un nome e crea/entra in una partita.");
    ws.onclose = () => { setConn("Connessione chiusa. Ricarica la pagina."); };
    ws.onerror = () => setConn("Errore di connessione.");
    ws.onmessage = (ev) => {
      let m;
      try { m = JSON.parse(ev.data); } catch { return; }
      handle(m);
    };
  }

  function send(obj) {
    if (ws && ws.readyState === WebSocket.OPEN) ws.send(JSON.stringify(obj));
  }

  function handle(m) {
    switch (m.type) {
      case "joined":
        myId = m.playerId;
        host = !!m.host;
        gameCode = m.game;
        $("lobby-code").textContent = m.game;
        $("game-card").classList.remove("hidden");
        $("rack-card").classList.remove("hidden");
        setStatus(host ? "Sei l'host: avvia la partita quando ci sono almeno 2 giocatori." : "In attesa dell'host…");
        break;
      case "state":
        onState(m.state);
        break;
      case "preview":
        preview = m.preview;
        updateMoveScore();
        break;
      case "error":
        setStatus("⚠ " + m.message);
        awaitingMove = false;
        renderControls();
        break;
      default:
        break;
    }
  }

  function onState(s) {
    state = s;
    // se la mossa è stata accettata, il turno è passato o `last` è nostro
    if (pending.length && (s.current !== myId || (s.last && s.last.playerId === myId))) {
      pending = [];
      preview = null;
      selected = -1;
    }
    if (s.last && s.last.playerId === myId && awaitingMove) {
      awaitingMove = false;
    }
    render();
  }

  // --- Rendering --------------------------------------------------------
  function render() {
    renderPlayers();
    renderBoard();
    renderRack();
    renderControls();
    updateMoveScore();
    $("bag-count").textContent = state ? String(state.bag) : "0";
    $("current-player").textContent = currentName();
    if (state && state.phase === "over") {
      const w = state.players.find((p) => p.id === state.winner);
      setStatus(`Partita finita! Vince ${w ? w.name : "?"}.`);
    }
  }

  function currentName() {
    if (!state) return "—";
    const p = state.players.find((x) => x.id === state.current);
    return p ? p.name : "—";
  }

  function renderPlayers() {
    const ul = $("players");
    ul.innerHTML = "";
    if (!state) return;
    for (const p of state.players) {
      const li = document.createElement("li");
      const isMe = p.id === myId;
      li.className = (p.isCurrent ? "current " : "") + (isMe ? "me" : "");
      li.innerHTML =
        `<span class="pname">${escapeHtml(p.name)}${isMe ? " (tu)" : ""}` +
        `${p.connected ? "" : " ⚫"}</span>` +
        `<span class="pmeta">${p.score} pt · ${p.tiles} tessere</span>`;
      ul.appendChild(li);
    }
    const startBtn = $("start");
    const restartBtn = $("restart");
    startBtn.classList.toggle("hidden", !(host && state.phase === "lobby"));
    startBtn.disabled = !state.canStart;
    restartBtn.classList.toggle("hidden", !(host && state.phase !== "lobby"));
  }

  function renderBoard() {
    // pulisce le tessere
    const cells = $("board").children;
    for (let i = 0; i < cells.length; i++) {
      cells[i].classList.remove("filled", "locked", "pending");
      cells[i].querySelectorAll("svg.tile").forEach((n) => n.remove());
    }
    const put = (r, c, letter, cls) => {
      const el = cells[r * SIZE + c];
      el.appendChild(tileSVG(letter));
      el.classList.add(cls);
    };
    if (state) {
      for (const t of state.board) put(t.row, t.col, t.letter, "filled locked");
    }
    for (const p of pending) put(p.r, p.c, p.jolly ? p.assigned : p.letter, "pending");
  }

  function renderRack() {
    const el = $("rack");
    el.innerHTML = "";
    const rack = state ? state.rack : [];
    const myTurn = isMyTurn();
    for (let i = 0; i < 8; i++) {
      const L = rack[i];
      const btn = document.createElement("button");
      btn.className = "tile-btn" + (i === selected ? " selected" : "");
      btn.type = "button";
      if (L) {
        btn.appendChild(tileSVG(L));
        btn.title = L === "?" ? "Scarabeo (jolly)" : `Tessera ${L} (${VALUES[L]})`;
        btn.disabled = !myTurn;
      } else {
        btn.disabled = true;
      }
      btn.addEventListener("click", () => onRackClick(i));
      el.appendChild(btn);
    }
  }

  function renderControls() {
    const myTurn = isMyTurn();
    const playing = state && state.phase === "playing";
    $("commit").disabled = !(myTurn && pending.length > 0 && !awaitingMove);
    $("undo").disabled = !(myTurn && pending.length > 0);
    $("pass").disabled = !(myTurn && playing);
  }

  function updateMoveScore() {
    if (!pending.length) {
      $("move-score").textContent = "Punteggio mossa: 0";
      return;
    }
    if (preview) {
      if (preview.valid) {
        const extra = preview.words && preview.words.length ? ` (${preview.words.join(", ")})` : "";
        $("move-score").textContent = "Punteggio mossa: " + preview.score + extra;
      } else {
        $("move-score").textContent = "⚠ " + preview.error;
      }
    } else {
      $("move-score").textContent = "Punteggio mossa: …";
    }
  }

  // --- Interazione ------------------------------------------------------
  function isMyTurn() {
    return !!(state && state.phase === "playing" && state.current === myId);
  }

  function onRackClick(i) {
    if (!isMyTurn()) return;
    const rack = state.rack;
    if (!rack[i]) return;
    selected = selected === i ? -1 : i;
    renderRack();
  }

  function onCellClick(r, c) {
    if (!isMyTurn()) return;
    const key = r + "," + c;
    const pi = pending.findIndex((p) => p.r === r && p.c === c);
    if (pi >= 0) {
      pending.splice(pi, 1);
      schedulePreview();
      renderBoard();
      renderControls();
      updateMoveScore();
      return;
    }
    if (state.board.some((t) => t.row === r && t.col === c)) return; // occupata

    if (selected < 0 || !state.rack[selected]) {
      setStatus("Seleziona prima una tessera dal rack.");
      return;
    }
    const letter = state.rack[selected];
    let jolly = false;
    let assigned = "";
    if (letter === "?") {
      const ans = (prompt("Scarabeo (jolly): quale lettera rappresenta? (A-Z)") || "").trim().toUpperCase();
      if (!/^[A-Z]$/.test(ans) || !(ans in VALUES)) {
        setStatus("Lettera non valida per lo scarabeo.");
        return;
      }
      jolly = true;
      assigned = ans;
    }
    pending.push({ r, c, letter, jolly, assigned });
    selected = -1;
    schedulePreview();
    renderBoard();
    renderRack();
    renderControls();
    updateMoveScore();
    setStatus("");
  }

  function placements() {
    return pending.map((p) => ({
      row: p.r, col: p.c, letter: p.letter, assigned: p.assigned || undefined,
    }));
  }

  function schedulePreview() {
    if (previewTimer) clearTimeout(previewTimer);
    previewTimer = setTimeout(() => {
      if (pending.length) send({ type: "preview", placements: placements() });
      else preview = null;
    }, 120);
  }

  // --- Azioni -----------------------------------------------------------
  function commit() {
    if (!isMyTurn() || !pending.length) return;
    awaitingMove = true;
    renderControls();
    send({ type: "move", placements: placements() });
  }

  function undo() {
    if (!isMyTurn()) return;
    pending = [];
    preview = null;
    selected = -1;
    renderBoard();
    renderRack();
    renderControls();
    updateMoveScore();
    setStatus("Mossa annullata.");
  }

  function pass() {
    if (!isMyTurn()) return;
    send({ type: "pass" });
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

  // --- Utilità ----------------------------------------------------------
  function setStatus(msg) { $("status").textContent = msg; }
  function setConn(msg) { $("conn-status").textContent = msg; }
  function escapeHtml(s) {
    return String(s).replace(/[&<>"']/g, (c) => (
      { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]
    ));
  }

  // --- Avvio ------------------------------------------------------------
  function init() {
    const defsHost = document.createElementNS(NS, "svg");
    defsHost.setAttribute("width", "0");
    defsHost.setAttribute("height", "0");
    defsHost.style.position = "absolute";
    defsHost.innerHTML = buildDefs();
    document.body.insertBefore(defsHost, document.body.firstChild);

    // plancia statica
    const board = $("board");
    for (let r = 0; r < SIZE; r++) {
      for (let c = 0; c < SIZE; c++) {
        const cell = document.createElement("div");
        const type = LAYOUT[r][c];
        cell.className = "cell " + CELL_CLASS[type];
        cell.dataset.r = r;
        cell.dataset.c = c;
        if (type === "C") cell.textContent = "🪲";
        else if (CELL_LABEL[type]) cell.textContent = CELL_LABEL[type];
        cell.addEventListener("click", () => onCellClick(r, c));
        board.appendChild(cell);
      }
    }

    renderTileset();
    $("join").addEventListener("click", () => {
      send({ type: "join", name: $("name").value, game: $("game-code").value.trim().toUpperCase() });
    });
    $("start").addEventListener("click", () => send({ type: "start" }));
    $("restart").addEventListener("click", () => send({ type: "restart" }));
    $("commit").addEventListener("click", commit);
    $("undo").addEventListener("click", undo);
    $("pass").addEventListener("click", pass);

    connect();
  }

  document.addEventListener("DOMContentLoaded", init);
})();
