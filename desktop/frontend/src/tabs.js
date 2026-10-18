// Dock state: whether the terminal dock is showing, which tab is in front and
// how tall it is. No DOM here, so node's test runner covers it; terminal.js
// renders it. Each tab keeps its own xterm.js, so switching loses no output.

export const TABS = ['kirokit', 'shell'];
export const MIN_HEIGHT = 120;
export const MAX_HEIGHT = 480;
const DEFAULT_HEIGHT = 220;

const clamp = (h) => Math.max(MIN_HEIGHT, Math.min(MAX_HEIGHT, Math.round(h)));

export function createDock({ height = DEFAULT_HEIGHT, open = false, tab = 'kirokit' } = {}) {
  const state = { open, tab, height: clamp(height) };
  const listeners = new Set();
  const emit = () => listeners.forEach((fn) => fn(state));

  return {
    get open() { return state.open; },
    get tab() { return state.tab; },
    get height() { return state.height; },
    show() {
      state.open = true;
      emit();
    },
    hide() {
      state.open = false;
      emit();
    },
    toggle() {
      state.open = !state.open;
      emit();
    },
    // Selecting a tab brings the dock up, so a command's output is never hidden.
    select(next) {
      if (!TABS.includes(next)) return;
      state.tab = next;
      state.open = true;
      emit();
    },
    resize(h) {
      if (!Number.isFinite(h)) return;
      state.height = clamp(h);
      emit();
    },
    subscribe(fn) {
      listeners.add(fn);
      return () => listeners.delete(fn);
    },
  };
}
