// The only module that reaches Go. Wails injects window.go and window.runtime
// at run time, so nothing generated is committed and `npm run build` works
// before any Wails run. Pages import this file and nothing else from Go.

const app = () => {
  const bound = window.go && window.go.main && window.go.main.App;
  if (!bound) throw new Error('kirokit is not running inside its window');
  return bound;
};

export const state = () => app().State();

export const saveSettings = (settings) => app().SaveSettings(settings);

export const catalog = () => app().Catalog();

export const installs = () => app().Installs();

// Captures the login environment afresh each time it is called.
export const prerequisites = (team) => app().Prerequisites(team);

// Subscribes to a Go event; returns the function that unsubscribes.
export function on(event, handler) {
  const rt = window.runtime;
  return rt && rt.EventsOn ? rt.EventsOn(event, handler) : () => {};
}
