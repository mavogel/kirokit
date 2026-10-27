import { esc, plural, timeTag, problem, installedBy, busyAttr, actionError } from './html.js';

// The kirokit arguments Export for Quick runs; dir comes from the app
// (api.quickDir), never from the page.
export const quickArgs = (team, dir) => ['render', '--quick', team, '--out', dir];

// quickRuns matches the one Export for Quick run with its exit. started
// records the run id; exited takes each term:kirokit:exit. Either returns
// { team, ok } once, when the started run has ended: a fast command can end
// before its id arrives, so the last exit is kept for started to find.
export function quickRuns() {
  let pending = null;
  let last = null;
  return {
    started(team, id) {
      if (last && last.id === id) return { team, ok: last.code === 0 };
      pending = { team, id };
      return null;
    },
    exited(id, code) {
      last = { id, code };
      if (!pending || pending.id !== id) return null;
      const done = { team: pending.team, ok: code === 0 };
      pending = null;
      return done;
    },
  };
}

// The Export for Quick part of a card: the result of the last export, the
// button, or why the team has none.
function quickParts(view, t) {
  if (!t.quick) return { result: '<p class="note">No Quick export: this team lists no skills or agents for Quick.</p>', button: '' };
  const b = (view.quick || {})[t.name];
  const result = b
    ? `<p class="quick-result" role="status"><span>Quick bundle ready: <b>${b.skills}</b> ${b.skills === 1 ? 'skill' : 'skills'}, <b>${b.agents}</b> ${b.agents === 1 ? 'agent' : 'agents'}.</span><button class="link-btn" type="button" data-action="quick-folder" data-team="${esc(t.name)}">Show folder</button><span class="path">${esc(b.dir)}</span></p>`
    : '';
  const label = view.running && view.exporting === t.name ? 'Exporting…' : 'Export for Quick';
  const button = `<button class="btn" data-action="quick" data-team="${esc(t.name)}" title="Render the team's skills and agents for Amazon Quick and open the folder"${busyAttr(view)}>${label}</button>${actionError(view, 'quick:' + t.name)}`;
  return { result, button };
}

// view = { catalog, installs, error, cloned, running, exporting, quick }
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
      const quick = quickParts(view, t);
      return `<article class="card" data-team="${esc(t.name)}">
  <div class="card-head"><h3>${esc(t.name)}</h3>${badge}</div>
  <p class="facts"><span><b>${t.agents.length}</b> ${t.agents.length === 1 ? 'agent' : 'agents'}</span><span><b>${t.skills.length}</b> ${t.skills.length === 1 ? 'skill' : 'skills'}</span><span>${plural(t.prompts, 'prompt', 'prompts')}</span><span>${plural(t.steerings.length, 'steering', 'steerings')}</span></p>
  ${changed}${quick.result}
  <div class="actions"><button class="btn primary" data-action="install" data-team="${esc(t.name)}"${busyAttr(view)}>Install</button>${actionError(view, 'install:' + t.name)}${quick.button}</div>
</article>`;
    })
    .join('\n');
}

// onInstall(team) starts the install flow, onQuick(team) the export and
// onShowFolder(team) opens the exported bundle.
export function initTeams(el, onInstall, onQuick, onShowFolder) {
  const actions = { install: onInstall, quick: onQuick, 'quick-folder': onShowFolder };
  el.addEventListener('click', (e) => {
    const b = e.target.closest('button[data-action]');
    const act = b && actions[b.dataset.action];
    if (act && !b.disabled) act(b.dataset.team);
  });
  return (view) => {
    el.innerHTML = teamsView(view);
  };
}
