import './style.css';
import logoUrl from '../../../assets/logo.svg';
import { createSidebar, PAGES } from './sidebar.js';
import * as api from './api.js';
import { initSettings } from './pages/settings.js';
import { initTeams } from './pages/teams.js';
import { initAgents } from './pages/agents.js';
import { initInstall } from './pages/install.js';
import { initSkills } from './pages/skills.js';
import { createDock } from './tabs.js';
import { initDock } from './terminal.js';

document.getElementById('logo').src = logoUrl;

const app = document.getElementById('app');
const collapse = document.getElementById('collapse');
const navButtons = [...document.querySelectorAll('.nav button')];
const sidebar = createSidebar();

function render({ collapsed, page }) {
  app.classList.toggle('rail', collapsed);
  collapse.setAttribute('aria-expanded', String(!collapsed));
  collapse.title = collapsed ? 'Expand sidebar' : 'Collapse sidebar';
  collapse.querySelector('.label').textContent = collapsed ? 'Expand' : 'Collapse';
  for (const p of PAGES) document.getElementById('page-' + p).hidden = p !== page;
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

// The four data pages draw from one view: the catalog, the install records and
// whether the clone exists. A catalog error is kept as text for the pages.
const view = { catalog: { teams: [], pins: {}, tag: '' }, installs: [], error: '', cloned: false };
const paint = [
  initTeams(document.getElementById('teams-grid')),
  initAgents(document.getElementById('page-agents')),
  initInstall(document.getElementById('page-install')),
  initSkills(document.getElementById('page-skills')),
];
const message = (e) => String((e && e.message) || e);

function draw() {
  for (const p of paint) p(view);
  for (const [name, n] of [['teams', view.catalog.teams.length], ['agents', view.catalog.teams.reduce((sum, t) => sum + t.agents.length, 0)]]) {
    document.querySelector(`[data-count="${name}"]`).textContent = n || '';
  }
}

async function loadCatalog() {
  try {
    view.cloned = (await api.state()).cloned;
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

initDock(createDock());
initSettings().load();
loadCatalog().then(loadInstalls);
api.on('installs:changed', (records) => {
  view.installs = records || [];
  draw();
});
api.on('clone:done', loadCatalog);
