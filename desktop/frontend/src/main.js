import './style.css';
import logoUrl from '../../../assets/logo.svg';
import { createSidebar, PAGES } from './sidebar.js';
import * as api from './api.js';
import { initSettings, updateArgs } from './pages/settings.js';
import { initTeams } from './pages/teams.js';
import { initAgents } from './pages/agents.js';
import { initInstall, syncArgs, uninstallArgs } from './pages/install.js';
import { initSkills } from './pages/skills.js';
import { initWelcome, showWelcome } from './pages/welcome.js';
import { createDialogs, installPlan } from './dialogs.js';
import { createDock } from './tabs.js';
import { initDock } from './terminal.js';

document.getElementById('logo').src = logoUrl;

const app = document.getElementById('app');
const collapse = document.getElementById('collapse');
const navButtons = [...document.querySelectorAll('.nav button')];
const sidebar = createSidebar();

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

async function startClone(address) {
  view.welcome = { address: address.trim(), error: '', busy: true };
  draw();
  try {
    await api.clone(address);
  } catch (e) {
    view.welcome = { address: address.trim(), error: message(e), busy: false };
    draw();
  }
}

const settings = initSettings(() => runCommand(updateArgs(view.installs), 'update'));
const paint = [
  initWelcome(document.getElementById('page-welcome'), startClone),
  initTeams(document.getElementById('teams-grid'), installTeam),
  initAgents(document.getElementById('page-agents')),
  initInstall(document.getElementById('page-install'), (action, team) =>
    runCommand(action === 'sync' ? syncArgs(team) : uninstallArgs(team), action + ':' + team)),
  initSkills(document.getElementById('page-skills')),
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

async function loadCatalog() {
  try {
    const st = await api.state();
    view.cloned = st.cloned;
    view.loaded = true;
    // The address field starts from the saved one; typing replaces it.
    if (!view.welcome.address) view.welcome = { ...view.welcome, address: st.settings.repoUrl || '' };
    view.catalog = await api.catalog();
    view.error = '';
  } catch (e) {
    view.error = message(e);
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

initDock(createDock());
settings.load();
refresh();
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
// A command that ends may have synced a new team or pin or installed one.
api.on('term:kirokit:exit', ({ id }) => {
  lastExit = Math.max(lastExit, id);
  view.running = false;
  refresh();
});
api.on('clone:done', ({ ok, error }) => {
  view.welcome = { ...view.welcome, busy: false, error: ok ? '' : error };
  view.notice = error || '';
  settings.load(); // the Settings address field now shows the cloned address
  refresh();
});
