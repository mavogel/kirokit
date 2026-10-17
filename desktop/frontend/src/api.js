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

// Runs kirokit with args in the kirokit tab and resolves to the run id.
export const run = (args) => app().Run(args);

// Clones the repository at address and syncs to its newest release; resolves
// when the clone starts, rejects with the reason an address was refused.
// clone:done reports the end.
export const clone = (address) => app().Clone(address);

// Wails runs every bound call on its own goroutine, so keystrokes sent without
// waiting can reach a terminal out of order. Each terminal sends its input
// through one line: a call starts after the one before it has ended, whether
// that succeeded or not.
function line() {
  let tail = Promise.resolve();
  return (call) => {
    const result = tail.then(() => call());
    tail = result.catch(() => {});
    return result;
  };
}
const runLine = line();
const shellLine = line();

// Keystrokes for the command running in the kirokit tab.
export const runWrite = (id, data) => runLine(() => app().RunWrite(id, data));

// Keystrokes and the dock size for the zsh tab.
export const shellWrite = (data) => shellLine(() => app().ShellWrite(data));

export const shellResize = (cols, rows) => app().ShellResize(cols, rows);

// Subscribes to a Go event; returns the function that unsubscribes.
export function on(event, handler) {
  const rt = window.runtime;
  return rt && rt.EventsOn ? rt.EventsOn(event, handler) : () => {};
}
