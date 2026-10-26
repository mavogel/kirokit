import { esc } from './html.js';
import { versionOptions } from './version.js';

// First launch: until a clone exists the sidebar pages stay hidden and this
// page asks where the team keeps its repository.
// view.welcome = { address, dir, error, busy, choosing, pick }
// It also stays while the clone runs: git creates the folder before the first
// sync has finished. Once the clone is done (choosing) the same page asks which
// version to use before setup ends.
// A clone started from Settings (welcome.fromSettings) never hides the pages:
// the old clone keeps serving them until the new one is ready.
export const showWelcome = (view) =>
  !view.welcome?.fromSettings && (!view.cloned || Boolean(view.welcome?.busy) || Boolean(view.welcome?.choosing));

// The entry the version list starts on: Newest release, or the checked-out
// branch when the repository has no release yet.
export function welcomePick(view) {
  const w = view.welcome || {};
  if (w.pick != null) return w.pick;
  const refs = view.checkout || {};
  if (refs.newest || !refs.current?.commit) return '';
  return refs.current.kind === 'branch' ? refs.current.name : (refs.branches || [])[0] || '';
}

function chooseView(view) {
  const w = view.welcome || {};
  const refs = view.checkout || {};
  const pick = welcomePick(view);
  const busy = view.switching === 'welcome' || Boolean(view.running);
  const error = view.refError?.where === 'welcome' ? view.refError.message : '';
  const lede = !refs.newest && refs.current?.commit
    ? 'This repository has no release yet (no v* tag). Pick a branch to use until your team lead tags one.'
    : 'Your team\'s repository is on this computer. Choose which version kirokit uses; you can change it later in Settings.';
  const label = busy ? 'Switching…' : pick === '' ? 'Continue' : `Switch to ${pick} and continue`;
  return `<div class="page-head"><div><h1>Set up kirokit</h1><p class="lede">${esc(lede)}</p></div></div>
<form class="welcome" id="version-form" novalidate>
  <p class="source">Cloned <b>${esc(w.address)}</b></p>
  <div class="field"><label for="welcome-version">Version</label><select id="welcome-version"${error ? ' aria-invalid="true"' : ''} aria-describedby="welcome-version-hint welcome-version-error"${busy ? ' disabled' : ''}>${versionOptions(refs, '', pick)}</select><span class="hint" id="welcome-version-hint">Newest release follows each new release. Pick a branch to try changes that are not released yet, or a tag to stay on one release.</span><span class="error" id="welcome-version-error" role="alert">${esc(error)}</span></div>
  <div class="actions"><button class="btn primary" type="submit"${busy ? ' disabled' : ''}>${esc(label)}</button></div>
</form>`;
}

export function welcomeView(view) {
  const w = view.welcome || {};
  if (w.choosing) return chooseView(view);
  const invalid = w.error ? ' aria-invalid="true"' : '';
  return `<div class="page-head"><div><h1>Set up kirokit</h1><p class="lede">Enter the address of your team's kirokit repository. kirokit copies it to this computer and uses its newest release.</p></div></div>
<form class="welcome" id="clone-form" novalidate>
  <div class="field"><label for="clone-address">Repository address</label><input id="clone-address" value="${esc(w.address)}"${invalid} placeholder="https://github.com/your-org/kiro-configs.git" spellcheck="false" autocomplete="off" aria-describedby="clone-hint clone-error"><span class="hint" id="clone-hint">An https, ssh or file address, git@host:path, or an absolute path. If git asks you to sign in, answer in the kirokit tab below.</span><span class="error" id="clone-error" role="alert">${esc(w.error)}</span></div>
  <div class="field"><label for="clone-dir">Repository folder</label><input id="clone-dir" value="${esc(w.dir)}" placeholder="${esc(view.defaultRepoDir)}" spellcheck="false" autocomplete="off" aria-describedby="clone-dir-hint clone-dir-error"><span class="hint" id="clone-dir-hint">Where the repository is cloned. Leave empty to use the default.</span><span class="error" id="clone-dir-error" role="alert">${esc(w.dirError)}</span></div>
  <div class="actions"><button class="btn primary" type="submit"${w.busy ? ' disabled' : ''}>${w.busy ? 'Cloning…' : 'Clone'}</button></div>
</form>`;
}

// initWelcome returns the paint function. The form is redrawn only when its
// markup changes, so a poll that redraws every page never takes focus or
// typing away from the field. onClone(address, dir) starts the clone;
// onPick(ref) records a pick in the version list and onChoose(ref) ends setup
// with it.
export function initWelcome(section, onClone, onPick, onChoose) {
  let last = '';
  section.addEventListener('submit', (e) => {
    e.preventDefault();
    if (e.target.id === 'version-form') onChoose(section.querySelector('#welcome-version').value);
    else onClone(section.querySelector('#clone-address').value, section.querySelector('#clone-dir').value);
  });
  section.addEventListener('change', (e) => {
    if (e.target.id === 'welcome-version') onPick(e.target.value);
  });
  return (view) => {
    const html = welcomeView(view);
    if (html === last) return;
    const refocus = document.activeElement?.id === 'welcome-version';
    last = html;
    section.innerHTML = html;
    if (view.welcome?.dirError) section.querySelector('#clone-dir').focus();
    else if (view.welcome?.error) section.querySelector('#clone-address').focus();
    else if (view.refError?.where === 'welcome' || refocus) section.querySelector('#welcome-version')?.focus();
  };
}
