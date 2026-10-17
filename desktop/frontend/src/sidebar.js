// Sidebar state: which page is showing and whether the sidebar is an icon
// rail. No DOM here, so node's test runner covers it; main.js renders it.

export const PAGES = ['teams', 'agents', 'install', 'skills', 'settings'];

// isHidden() is true while the sidebar is not showing (the first-launch page);
// Command-B then does nothing instead of changing a sidebar nobody can see.
export function createSidebar({ collapsed = false, page = 'teams', isHidden = () => false } = {}) {
  const state = { collapsed, page };
  const listeners = new Set();
  const emit = () => listeners.forEach((fn) => fn(state));

  return {
    get collapsed() { return state.collapsed; },
    get page() { return state.page; },
    toggle() {
      state.collapsed = !state.collapsed;
      emit();
    },
    select(next) {
      if (!PAGES.includes(next)) return;
      state.page = next;
      emit();
    },
    // Command-B only: plain B and Control-B belong to the page's inputs.
    handleKey(e) {
      if (!e.metaKey || e.ctrlKey || String(e.key).toLowerCase() !== 'b') return false;
      if (isHidden()) return false;
      e.preventDefault();
      this.toggle();
      return true;
    },
    subscribe(fn) {
      listeners.add(fn);
      return () => listeners.delete(fn);
    },
  };
}
