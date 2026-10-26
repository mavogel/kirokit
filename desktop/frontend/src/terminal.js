// The terminal dock: two xterm.js terminals, "kirokit" for button commands and
// one for the person's own shell (zsh, bash, PowerShell), in one resizable panel. State lives in
// tabs.js; this file draws it and talks to Go through api.js.
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';
import * as api from './api.js';
import { MIN_HEIGHT, MAX_HEIGHT, TABS } from './tabs.js';

const css = (name) => getComputedStyle(document.documentElement).getPropertyValue(name).trim();

function theme() {
  return {
    background: css('--term-bg'),
    foreground: css('--term-text'),
    cursor: css('--term-prompt'),
    selectionBackground: css('--accent-soft'),
  };
}

// Go sends terminal bytes as base64 so no escape sequence is mangled in JSON.
function decode(b64) {
  const raw = atob(b64);
  const bytes = new Uint8Array(raw.length);
  for (let i = 0; i < raw.length; i++) bytes[i] = raw.charCodeAt(i);
  return bytes;
}

function makeTerm(host) {
  const term = new Terminal({
    fontFamily: css('--mono'),
    fontSize: 13,
    lineHeight: 1.2,
    cursorBlink: true,
    scrollback: 5000,
    theme: theme(),
  });
  const fit = new FitAddon();
  term.loadAddon(fit);
  term.open(host);
  return { term, fit, host };
}

// initDock wires the dock markup in index.html to the dock state.
export function initDock(dock) {
  const panel = document.getElementById('terminal');
  const toggle = document.getElementById('term-open');
  const handle = document.getElementById('resize');
  const tabButtons = Object.fromEntries(TABS.map((t) => [t, document.getElementById('tab-' + t)]));
  const terms = {
    kirokit: makeTerm(document.getElementById('term-kirokit')),
    shell: makeTerm(document.getElementById('term-shell')),
  };
  let runId = 0;
  let running = false;

  // The kirokit tab takes input only while a command runs (git sign-in).
  const setRunning = (on) => { running = on; terms.kirokit.term.options.disableStdin = !on; };
  setRunning(false);

  // The shell tab is named after its shell once the app's state says which;
  // the kirokit tab's first line names it then too.
  let greeted = false;
  function setShell(name) {
    tabButtons.shell.textContent = name;
    if (greeted) return;
    greeted = true;
    terms.kirokit.term.writeln(`\x1b[2mButtons run their kirokit command here. Type your own commands in the ${name} tab.\x1b[0m`);
  }

  terms.kirokit.term.onData((data) => { if (running) api.runWrite(runId, data); });
  terms.shell.term.onData((data) => api.shellWrite(data));
  terms.shell.term.onResize(({ cols, rows }) => api.shellResize(cols, rows));

  // Fitting a hidden terminal measures nothing, so only the front tab of an
  // open dock is fitted. A fit that changes the size reaches Go through onResize.
  function refit() {
    if (!dock.open) return;
    const front = terms[dock.tab];
    front.fit.fit();
    // The shell has not heard its size when the fit changed nothing yet.
    if (dock.tab === 'shell') api.shellResize(front.term.cols, front.term.rows);
  }

  function render({ open, tab, height }) {
    panel.hidden = !open;
    toggle.hidden = open;
    panel.style.height = height + 'px';
    handle.setAttribute('aria-valuenow', height);
    for (const t of TABS) {
      tabButtons[t].setAttribute('aria-selected', String(t === tab));
      tabButtons[t].tabIndex = t === tab ? 0 : -1;
      terms[t].host.hidden = t !== tab;
    }
    requestAnimationFrame(() => {
      refit();
      if (open) terms[tab].term.focus();
    });
  }
  dock.subscribe(render);
  render(dock);

  toggle.addEventListener('click', () => dock.show());
  document.getElementById('term-close').addEventListener('click', () => dock.hide());
  document.getElementById('term-clear').addEventListener('click', () => terms[dock.tab].term.clear());
  for (const t of TABS) tabButtons[t].addEventListener('click', () => dock.select(t));
  document.getElementById('term-tabs').addEventListener('keydown', (e) => {
    if (e.key !== 'ArrowLeft' && e.key !== 'ArrowRight') return;
    const next = TABS[(TABS.indexOf(dock.tab) + (e.key === 'ArrowRight' ? 1 : TABS.length - 1)) % TABS.length];
    dock.select(next);
    tabButtons[next].focus();
    e.preventDefault();
  });
  document.addEventListener('keydown', (e) => {
    if (e.ctrlKey && !e.metaKey && e.key === '`') {
      dock.toggle();
      e.preventDefault();
    }
  });

  // Dragging the top edge, or Arrow Up and Down on it, resizes the dock.
  handle.addEventListener('pointerdown', (e) => {
    const startY = e.clientY;
    const startH = dock.height;
    handle.setPointerCapture(e.pointerId);
    const move = (m) => dock.resize(startH + (startY - m.clientY));
    const up = () => {
      handle.removeEventListener('pointermove', move);
      handle.removeEventListener('pointerup', up);
    };
    handle.addEventListener('pointermove', move);
    handle.addEventListener('pointerup', up);
  });
  handle.addEventListener('keydown', (e) => {
    if (e.key === 'ArrowUp') dock.resize(dock.height + 20);
    else if (e.key === 'ArrowDown') dock.resize(dock.height - 20);
    else return;
    e.preventDefault();
  });
  handle.setAttribute('aria-valuemin', MIN_HEIGHT);
  handle.setAttribute('aria-valuemax', MAX_HEIGHT);

  // A window resize changes the panel's width; a dock resize is handled by
  // render. Either way the front terminal is refitted.
  new ResizeObserver(() => refit()).observe(document.getElementById('term-bodies'));

  api.on('term:shell', ({ b64 }) => terms.shell.term.write(decode(b64)));
  api.on('term:kirokit', ({ id, b64 }) => {
    if (id !== runId) {
      runId = id;
      setRunning(true);
      dock.select('kirokit');
    }
    terms.kirokit.term.write(decode(b64));
  });
  api.on('term:kirokit:exit', ({ id, code }) => {
    if (id !== runId) return;
    setRunning(false);
    if (code !== 0) terms.kirokit.term.write(`\r\n\x1b[2mexited with code ${code}\x1b[0m\r\n`);
  });

  return { setShell };
}
