import './style.css';
import logoUrl from '../../../assets/logo.svg';
import { createSidebar, PAGES } from './sidebar.js';
import * as api from './api.js';
import { initSettings, updateArgs } from './pages/settings.js';
import { initTeams, quickArgs, quickRuns } from './pages/teams.js';
import { initAgents } from './pages/agents.js';
import { initInstall, syncArgs, uninstallArgs } from './pages/install.js';
import { initSkills } from './pages/skills.js';
import { feedbackFor, feedbackWhere } from './pages/feedback.js';
import { initWelcome, showWelcome } from './pages/welcome.js';
import { createDialogs, installPlan } from './dialogs.js';
import { createDock, shellLabel } from './tabs.js';
import { initDock } from './terminal.js';

document.getElementById('logo').src = logoUrl;

const app = document.getElementById('app');
const collapse = document.getElementById('collapse');
const navButtons = [...document.querySelectorAll('.nav button')];
const sidebar = createSidebar({ isHidden: () => setup() });

// What the pages know of the clone's versions before (or without) a read.
const noCheckout = { current: { kind: '', name: '', commit: '' }, branches: [], tags: [], newest: '' };

// The pages draw from one view: the catalog, the install records, whether the
// clone exists, and the state of the buttons. A catalog error is kept as text
// for the pages. Until the first State arrives (loaded) nothing is decided.
const view = {
  catalog: { teams: [], pins: {}, tag: '' },
  installs: [],
  error: '',
  cloned: false,
  loaded: false,
  running: false, // a button command is running in the kirokit tab
  actionError: null, // { where, message } of a refused command
  notice: '', // the no-release message, or why the last clone failed
  welcome: { address: '', error: '', busy: false },
  ref: '', // the saved version every sync uses; '' is Newest release
  checkout: noCheckout,
  switching: '', // where a version switch was started: 'settings' or 'welcome'
  refError: null, // { where, message } of the last failed switch
  exporting: '', // the team whose Export for Quick is running
  quick: {}, // team -> { dir, skills, agents } of its last export
};
const setup = () => view.loaded && showWelcome(view);

function render({ collapsed, page }) {
  const welcoming = setup();
  app.classList.toggle('rail', collapsed);
  app.classList.toggle('setup', welcoming);
  document.getElementById('page-welcome').hidden = !welcoming;
  collapse.setAttribute('aria-expanded', String(!collapsed));
  collapse.title = collapsed ? 'Expand sidebar' : 'Collapse sidebar';
  collapse.querySelector('.label').textContent = collapsed ? 'Expand' : 'Collapse';
  for (const p of PAGES) document.getElementById('page-' + p).hidden = welcoming || p !== page;
  for (const b of navButtons) {
    if (b.dataset.page === page) b.setAttribute('aria-current', 'page');
    else b.removeAttribute('aria-current');
  }
}

sidebar.subscribe((s) => {
  render(s);
  document.getElementById('main').scrollTop = 0;
});
collapse.addEventListener('click', () => sidebar.toggle());
for (const b of navButtons) b.addEventListener('click', () => sidebar.select(b.dataset.page));
document.addEventListener('keydown', (e) => sidebar.handleKey(e));
render(sidebar);

const dialogs = createDialogs();
const message = (e) => String((e && e.message) || e);

// Runs kirokit with args in the kirokit tab. A refusal, such as "kirokit is
// still running a command", shows beside the button at where.
let lastExit = 0;
async function runCommand(args, where) {
  view.actionError = null;
  try {
    const id = await api.run(args);
    // A fast command may already have ended before the id arrives.
    if (id > lastExit) view.running = true;
  } catch (e) {
    view.actionError = { where, message: message(e) };
  }
  draw();
}

async function installTeam(team) {
  const plan = installPlan(team, view.installs);
  if (plan.kind === 'skip') {
    view.actionError = { where: 'install:' + team, message: `${team} is already installed.` };
    draw();
    return;
  }
  if (plan.kind === 'replace' && !(await dialogs.confirmReplace(team, plan.installed))) return;
  if (!(await dialogs.confirmSetup(team, () => api.prerequisites(team)))) return;
  await runCommand(['install', team], 'install:' + team);
}

// Export for Quick renders the team's Quick bundle in the kirokit tab; when
// that run ends well the card shows what it holds and the folder opens once.
const quickRun = quickRuns();

async function exportQuick(team) {
  const where = 'quick:' + team;
  view.actionError = null;
  try {
    const id = await api.run(quickArgs(team, await api.quickDir(team)));
    if (id > lastExit) view.running = true;
    view.exporting = team;
    const done = quickRun.started(team, id);
    if (done) await quickDone(done);
  } catch (e) {
    view.actionError = { where, message: message(e) };
  }
  draw();
}

async function quickDone({ team, ok }) {
  view.exporting = '';
  if (!ok) {
    view.actionError = { where: 'quick:' + team, message: 'kirokit render failed: see the kirokit tab below.' };
    return;
  }
  try {
    view.quick = { ...view.quick, [team]: await api.quickBundle(team) };
    await api.openFolder(team);
  } catch (e) {
    view.actionError = { where: 'quick:' + team, message: message(e) };
  }
}

async function showQuickFolder(team) {
  view.actionError = null;
  try {
    await api.openFolder(team);
  } catch (e) {
    view.actionError = { where: 'quick:' + team, message: message(e) };
  }
  draw();
}

// Opens the prefilled GitHub issue for an agent or skill in the browser. A
// refusal shows beside that row's Feedback button.
async function giveFeedback(item) {
  view.actionError = null;
  try {
    await api.openURL(feedbackFor(view.catalog, item));
  } catch (e) {
    view.actionError = { where: feedbackWhere(item.team, item.kind, item.name), message: message(e) };
  }
  draw();
}

async function startClone(address, dir = '') {
  const typed = { address: address.trim(), dir: dir.trim() };
  view.welcome = { ...typed, error: '', dirError: '', busy: true };
  draw();
  try {
    try {
      await api.setRepoDir(typed.dir);
    } catch (e) {
      view.welcome = { ...typed, error: '', dirError: message(e).replace(/^RepoDir: /, ''), busy: false };
      draw();
      return;
    }
    await api.clone(address);
  } catch (e) {
    view.welcome = { ...typed, error: message(e), dirError: '', busy: false };
    draw();
  }
}

// A new address saved in Settings clones into a fresh folder while the pages
// keep using the old clone. A refusal rejects back to the Settings form.
async function cloneFromSettings(address) {
  const before = view.welcome;
  view.notice = '';
  view.welcome = { address, error: '', busy: true, fromSettings: true };
  try {
    await api.clone(address);
  } catch (e) {
    view.welcome = before;
    throw e;
  } finally {
    draw();
  }
}

// Switches to another version in the kirokit tab; ref:done reports the end.
// where ('settings' or 'welcome') is the page that shows a failure.
async function switchVersion(ref, where) {
  view.refError = null;
  view.actionError = null;
  // Marked before the call: ref:done may arrive before switchRef resolves.
  view.switching = where;
  try {
    await api.switchRef(ref);
  } catch (e) {
    view.switching = '';
    view.refError = { where, message: message(e) };
  }
  draw();
}

// Settings' Update runs the saved version, or switches when another is picked.
const settings = initSettings(
  (picked) => (picked === view.ref ? runCommand(updateArgs(view.installs, view.ref), 'update') : switchVersion(picked, 'settings')),
  cloneFromSettings,
);
const paint = [
  initWelcome(
    document.getElementById('page-welcome'),
    startClone,
    (pick) => {
      view.welcome = { ...view.welcome, pick };
      draw();
    },
    (pick) => {
      // The clone is on the newest release, which is already the saved choice.
      if (pick === '') {
        view.welcome = { ...view.welcome, choosing: false };
        view.refError = null;
        draw();
      } else switchVersion(pick, 'welcome');
    },
  ),
  initTeams(document.getElementById('teams-grid'), installTeam, exportQuick, showQuickFolder),
  initAgents(document.getElementById('page-agents'), giveFeedback),
  initInstall(document.getElementById('page-install'), (action, team) =>
    runCommand(action === 'sync' ? syncArgs(team, view.ref) : uninstallArgs(team), action + ':' + team)),
  initSkills(document.getElementById('page-skills'), giveFeedback),
  settings.paint,
];

function draw() {
  for (const p of paint) p(view);
  const notice = document.getElementById('notice');
  notice.textContent = view.notice;
  notice.hidden = !view.notice;
  render(sidebar);
  for (const [name, n] of [['teams', view.catalog.teams.length], ['agents', view.catalog.teams.reduce((sum, t) => sum + t.agents.length, 0)]]) {
    document.querySelector(`[data-count="${name}"]`).textContent = n || '';
  }
}

// The first load reads the state through api.start, which also tells Go the
// page answered (a smoke start waits for that); later loads use api.state.
async function loadCatalog(readState = api.state) {
  try {
    const st = await readState();
    terminalDock.setShell(shellLabel(st));
    view.cloned = st.cloned;
    view.loaded = true;
    // The address field starts from the saved one; typing replaces it.
    if (!view.welcome.address) view.welcome = { ...view.welcome, address: st.settings.repoUrl || '' };
    if (!view.welcome.busy && !view.welcome.dirError) view.welcome = { ...view.welcome, dir: st.settings.repoDir || '' };
    view.defaultRepoDir = st.env.defaultRepoDir || '';
    view.ref = st.settings.ref || '';
    view.catalog = await api.catalog();
    view.error = '';
    try {
      view.checkout = await api.checkout();
    } catch {
      // A clone git cannot read shows no version line; the catalog reports the clone's problems.
      view.checkout = noCheckout;
    }
  } catch (e) {
    view.error = message(e);
    terminalDock.setShell(shellLabel());
  }
  draw();
}

async function loadInstalls() {
  try {
    view.installs = (await api.installs()) || [];
  } catch (e) {
    view.error = view.error || message(e);
  }
  draw();
}

const refresh = () => loadCatalog().then(loadInstalls);

const terminalDock = initDock(createDock());
settings.load();
loadCatalog(api.start).then(loadInstalls);
api.on('installs:changed', (records) => {
  view.installs = records || [];
  draw();
});
api.on('term:kirokit', ({ id }) => {
  if (id > lastExit && !view.running) {
    view.running = true;
    draw();
  }
});
// The native menu's Settings… (⌘, / Ctrl+,) opens Settings as its sidebar
// button does. The first-launch page has no Settings to open, so it is ignored there.
api.on('menu:settings', () => {
  if (!setup()) sidebar.select('settings');
});
// A command that ends may have synced a new team or pin or installed one.
api.on('term:kirokit:exit', async ({ id, code }) => {
  lastExit = Math.max(lastExit, id);
  view.running = false;
  const done = quickRun.exited(id, code);
  if (done) await quickDone(done);
  refresh();
});
api.on('clone:done', ({ ok, error }) => {
  const fromSettings = view.welcome.fromSettings;
  if (fromSettings) settings.cloneDone({ ok, error });
  // A first clone keeps the setup page up to ask which version to use.
  view.welcome = { ...view.welcome, busy: false, fromSettings: false, error: ok ? '' : error, choosing: ok && !fromSettings, pick: null };
  view.notice = error || '';
  settings.load(); // the Settings address field now shows the cloned address
  refresh();
});
// A switch saved its choice only when ok; the pages read it back in refresh.
api.on('ref:done', ({ ok, error }) => {
  const where = view.switching;
  view.switching = '';
  if (ok && where === 'welcome') view.welcome = { ...view.welcome, choosing: false };
  if (!ok) view.refError = { where, message: error || 'The switch failed. See the kirokit tab.' };
  refresh();
});
