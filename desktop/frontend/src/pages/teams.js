import { esc, plural, timeTag, problem, installedBy } from './html.js';

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
</article>`;
    })
    .join('\n');
}

export function initTeams(el) {
  return (view) => {
    el.innerHTML = teamsView(view);
  };
}
