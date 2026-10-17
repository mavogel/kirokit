import './style.css';
import logoUrl from '../../../assets/logo.svg';
import { createSidebar, PAGES } from './sidebar.js';
import { initSettings } from './pages/settings.js';

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

initSettings().load();
