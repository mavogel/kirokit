import { esc } from './html.js';

// First launch: until a clone exists the sidebar pages stay hidden and this
// page asks where the team keeps its repository.
// view.welcome = { address, dir, error, busy }
// It also stays while the clone runs: git creates the folder before the first
// sync has finished.
// A clone started from Settings (welcome.fromSettings) never hides the pages:
// the old clone keeps serving them until the new one is ready.
export const showWelcome = (view) =>
  !view.welcome?.fromSettings && (!view.cloned || Boolean(view.welcome?.busy));

export function welcomeView(view) {
  const w = view.welcome || {};
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
// typing away from the field. onClone(address, dir) starts the clone.
export function initWelcome(section, onClone) {
  let last = '';
  section.addEventListener('submit', (e) => {
    e.preventDefault();
    onClone(section.querySelector('#clone-address').value, section.querySelector('#clone-dir').value);
  });
  return (view) => {
    const html = welcomeView(view);
    if (html === last) return;
    last = html;
    section.innerHTML = html;
    if (view.welcome?.dirError) section.querySelector('#clone-dir').focus();
    else if (view.welcome?.error) section.querySelector('#clone-address').focus();
  };
}
