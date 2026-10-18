import { esc, timeTag, problem, busyAttr, actionError } from './html.js';

// The rows of the install table: one per team with an install record. The
// records are on this computer, so a catalog error or a missing clone does
// not hide them; the catalog only flags a team the repository no longer has.
export function installView(view) {
  const known = new Set((view.catalog?.teams || []).map((t) => t.name));
  const records = view.installs || [];
  if (!records.length) {
    const blocked = view.error ? problem(view) : '';
    return { blocked: true, html: blocked + '<p class="note">Nothing is installed yet. Pick a team on the Teams page.</p>' };
  }
  const rows = records.map((r) => {
    const gone = view.cloned && !view.error && !known.has(r.Team);
    return `<tr data-team="${esc(r.Team)}">
  <td class="mono">${esc(r.Team)}</td>
  <td><span class="badge ok">installed</span>${gone ? ' <span class="badge warn">not in the repository</span>' : ''}</td>
  <td class="num">${r.Files}</td>
  <td>${timeTag(r.Modified)}</td>
  <td><div class="row-actions"><button class="btn" data-action="sync" data-team="${esc(r.Team)}"${busyAttr(view)}>Sync</button><button class="btn" data-action="uninstall" data-team="${esc(r.Team)}"${busyAttr(view)}>Uninstall</button>${actionError(view, 'sync:' + r.Team)}${actionError(view, 'uninstall:' + r.Team)}</div></td>
</tr>`;
  });
  return { blocked: false, html: rows.join('\n') };
}

export const syncArgs = (team) => ['sync', team, '--latest-tag'];
export const uninstallArgs = (team) => ['uninstall', team];

// onAction(action, team) runs the button's command; action is "sync" or
// "uninstall".
export function initInstall(section, onAction) {
  const body = section.querySelector('tbody');
  const table = section.querySelector('table');
  const message = document.createElement('div');
  table.after(message);
  body.addEventListener('click', (e) => {
    const b = e.target.closest('button[data-action]');
    if (b && !b.disabled) onAction(b.dataset.action, b.dataset.team);
  });
  return (view) => {
    const out = installView(view);
    table.hidden = out.blocked;
    message.innerHTML = out.blocked ? out.html : '';
    if (!out.blocked) body.innerHTML = out.html;
  };
}
