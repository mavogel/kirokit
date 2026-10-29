// The only module that reaches Go, through Wails' @wailsio/runtime: methods are
// called by name (main.App.<Method>), so nothing generated is committed and
// `npm run build` works before any Wails run. Pages import this file and
// nothing else from Go.
import * as wails from '@wailsio/runtime';

const NOT_IN_WINDOW = 'kirokit is not running inside its window';

// Tests set globalThis.__kirokitRuntime to a fake runtime (null for none).
function runtime() {
  const rt = '__kirokitRuntime' in globalThis ? globalThis.__kirokitRuntime : wails;
  if (!rt) throw new Error(NOT_IN_WINDOW);
  return rt;
}

// A Go error rejects as Wails' RuntimeError whose message is err.Error().
const call = async (method, ...args) => runtime().Call.ByName(`main.App.${method}`, ...args);

export const state = () => call('State');

// Calls State and then tells Go the page has loaded and the bridge answers
// (smoke:ready, which a smoke start waits for); sends nothing when State fails.
// The page never waits on that message: if it is lost, only a smoke start
// notices (it fails after its time limit).
export async function start() {
  const st = await state();
  Promise.resolve()
    .then(() => runtime().Events.Emit('smoke:ready'))
    .catch(() => {});
  return st;
}

// Stores only the clone folder (empty restores the default); rejects with a
// "RepoDir: ..." line when the folder is refused.
export const setRepoDir = (dir) => call('SetRepoDir', dir);

export const saveSettings = (settings) => call('SaveSettings', settings);

export const catalog = () => call('Catalog');

export const installs = () => call('Installs');

// Captures the login environment afresh each time it is called.
export const prerequisites = (team) => call('Prerequisites', team);

// Runs kirokit with args in the kirokit tab and resolves to the run id.
export const run = (args) => call('Run', args);

// Where Export for Quick renders a team's bundle (<app data>/quick/<team>),
// what that bundle holds ({ dir, skills, agents }), and showing it in the
// file manager. Each takes only the team name; Go builds the path.
export const quickDir = (team) => call('QuickDir', team);
export const quickBundle = (team) => call('QuickBundle', team);
export const openFolder = (team) => call('OpenFolder', team);

// Clones the repository at address and syncs to its newest release; resolves
// when the clone starts, rejects with the reason an address was refused.
// clone:done reports the end.
export const clone = (address) => call('Clone', address);

// The clone's current version and the branches and tags it can switch to,
// read from its own refs (no network). Empty lists before a clone.
export const checkout = () => call('Checkout');

// Switches the clone to a branch or tag ('' for Newest release) by running
// kirokit sync in the kirokit tab; resolves when it starts, rejects with the
// reason it was refused. ref:done reports the end, and the choice is saved only
// when the switch succeeded.
export const switchRef = (ref) => call('SwitchRef', ref);

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
export const runWrite = (id, data) => runLine(() => call('RunWrite', id, data));

// Keystrokes and the dock size for the zsh tab.
export const shellWrite = (data) => shellLine(() => call('ShellWrite', data));

export const shellResize = (cols, rows) => call('ShellResize', cols, rows);

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
  await runtime().Browser.OpenURL(url);
}

// Subscribes to a Go event and hands handler the event's payload (Wails passes
// an event object { name, data }); returns the function that unsubscribes.
export function on(event, handler) {
  let rt;
  try {
    rt = runtime();
  } catch {
    return () => {};
  }
  return rt.Events.On(event, (e) => handler(e.data));
}
