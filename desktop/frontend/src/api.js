// The only module that reaches Go. Wails injects window.go and window.runtime
// at run time, so nothing generated is committed and `npm run build` works
// before any Wails run. Pages import this file and nothing else from Go.

const app = () => {
  const bound = window.go && window.go.main && window.go.main.App;
  if (!bound) throw new Error('kirokit is not running inside its window');
  return bound;
};

export const state = () => app().State();

// Stores only the clone folder (empty restores the default); rejects with a
// "RepoDir: ..." line when the folder is refused.
export const setRepoDir = (dir) => app().SetRepoDir(dir);

export const saveSettings = (settings) => app().SaveSettings(settings);

export const catalog = () => app().Catalog();

export const installs = () => app().Installs();

// Captures the login environment afresh each time it is called.
export const prerequisites = (team) => app().Prerequisites(team);

// Runs kirokit with args in the kirokit tab and resolves to the run id.
export const run = (args) => app().Run(args);

// Where Export for Quick renders a team's bundle (<app data>/quick/<team>),
// what that bundle holds ({ dir, skills, agents }), and showing it in the
// file manager. Each takes only the team name; Go builds the path.
export const quickDir = (team) => app().QuickDir(team);
export const quickBundle = (team) => app().QuickBundle(team);
export const openFolder = (team) => app().OpenFolder(team);

// Clones the repository at address and syncs to its newest release; resolves
// when the clone starts, rejects with the reason an address was refused.
// clone:done reports the end.
export const clone = (address) => app().Clone(address);

// The clone's current version and the branches and tags it can switch to,
// read from its own refs (no network). Empty lists before a clone.
export const checkout = () => app().Checkout();

// Switches the clone to a branch or tag ('' for Newest release) by running
// kirokit sync in the kirokit tab; resolves when it starts, rejects with the
// reason it was refused. ref:done reports the end, and the choice is saved only
// when the switch succeeded.
export const switchRef = (ref) => app().SwitchRef(ref);

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

// The only addresses the app opens in the system browser: a GitHub
// repository's new-issue form. The address comes from a cloned repository's
// manifest, so anything else (another host, another page, javascript:) is
// refused. The parsed check also catches what the pattern alone would let
// through, such as a "/../" that leaves the repository.
const ISSUE_FORM = /^https:\/\/github\.com\/[A-Za-z0-9-]+\/[A-Za-z0-9._-]+\/issues\/new\?/;
const ISSUE_PATH = /^\/[A-Za-z0-9-]+\/[A-Za-z0-9._-]+\/issues\/new$/;

function isIssueForm(url) {
  if (typeof url !== 'string' || !ISSUE_FORM.test(url)) return false;
  try {
    const u = new URL(url);
    return u.protocol === 'https:' && u.host === 'github.com' && !u.username && ISSUE_PATH.test(u.pathname);
  } catch {
    return false;
  }
}

// Opens url in the system browser; rejects with "refusing to open <url>" for
// any address that is not a GitHub new-issue form.
export async function openURL(url) {
  if (!isIssueForm(url)) throw new Error(`refusing to open ${url}`);
  const rt = window.runtime;
  if (!rt || !rt.BrowserOpenURL) throw new Error('kirokit is not running inside its window');
  rt.BrowserOpenURL(url);
}

// Subscribes to a Go event; returns the function that unsubscribes.
export function on(event, handler) {
  const rt = window.runtime;
  return rt && rt.EventsOn ? rt.EventsOn(event, handler) : () => {};
}
