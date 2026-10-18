import { esc, chips, problem } from './html.js';

// The rows of the agents table, or the message that replaces it.
export function agentsView(view) {
  const blocked = problem(view);
  if (blocked) return { blocked: true, html: blocked };
  const rows = [];
  for (const t of view.catalog.teams || []) {
    for (const a of t.agents) {
      rows.push(`<tr>
  <td class="mono">${esc(a.name)}</td>
  <td>${esc(t.name)}</td>
  <td class="desc">${esc(a.description)}</td>
  <td>${chips(a.tools)}${chips(a.servers.map((s) => '@' + s).filter((s) => !a.tools.includes(s)))}</td>
  <td>${a.needsKey ? '<span class="badge warn">needs key</span>' : '<span class="badge off">none</span>'}</td>
</tr>`);
    }
  }
  if (!rows.length) return { blocked: true, html: '<p class="note">No agents yet.</p>' };
  return { blocked: false, html: rows.join('\n') };
}

export function initAgents(section) {
  const body = section.querySelector('tbody');
  const table = section.querySelector('table');
  const message = document.createElement('div');
  table.after(message);
  return (view) => {
    const out = agentsView(view);
    table.hidden = out.blocked;
    message.innerHTML = out.blocked ? out.html : '';
    if (!out.blocked) body.innerHTML = out.html;
  };
}
