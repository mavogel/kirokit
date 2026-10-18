import { esc, plural, timeTag, problem, installedBy, busyAttr, actionError } from './html.js';

// view = { catalog, installs, error, cloned }
export function teamsView(view) {
  const blocked = problem(view);
  if (blocked) return blocked;
  const installed = installedBy(view.installs);
  const teams = view.catalog.teams || [];
  if (!teams.length) return '<p class="note">The repository has no teams yet.</p>';

  return teams
    .map((t) => {
      const rec = installed.get(t.name);
      const badge = rec
        ? `<span class="badge ok">installed</span>`
        : `<span class="badge off">not installed</span>`;
      const changed = rec ? `<p class="note">Last change ${timeTag(rec.Modified)}</p>` : '';
      return `<article class="card" data-team="${esc(t.name)}">
  <div class="card-head"><h3>${esc(t.name)}</h3>${badge}</div>
  <p class="facts"><span><b>${t.agents.length}</b> ${t.agents.length === 1 ? 'agent' : 'agents'}</span><span><b>${t.skills.length}</b> ${t.skills.length === 1 ? 'skill' : 'skills'}</span><span>${plural(t.prompts, 'prompt', 'prompts')}</span><span>${plural(t.steerings.length, 'steering', 'steerings')}</span></p>
  ${changed}
  <div class="actions"><button class="btn primary" data-action="install" data-team="${esc(t.name)}"${busyAttr(view)}>Install</button>${actionError(view, 'install:' + t.name)}</div>
</article>`;
    })
    .join('\n');
}

// onInstall(team) starts the install flow for the card's team.
export function initTeams(el, onInstall) {
  el.addEventListener('click', (e) => {
    const b = e.target.closest('button[data-action="install"]');
    if (b && !b.disabled) onInstall(b.dataset.team);
  });
  return (view) => {
    el.innerHTML = teamsView(view);
  };
}
