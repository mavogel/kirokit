import { esc } from './html.js';

// First launch: until a clone exists the sidebar pages stay hidden and this
// page asks where the team keeps its repository.
// view.welcome = { address, error, busy }
// It also stays while the clone runs: git creates the folder before the first
// sync has finished.
export const showWelcome = (view) => !view.cloned || Boolean(view.welcome?.busy);

export function welcomeView(view) {
  const w = view.welcome || {};
  const invalid = w.error ? ' aria-invalid="true"' : '';
  return `<div class="page-head"><div><h1>Set up kirokit</h1><p class="lede">Enter the address of your team's kirokit repository. kirokit copies it to this computer and uses its newest release.</p></div></div>
<form class="welcome" id="clone-form" novalidate>
  <div class="field"><label for="clone-address">Repository address</label><input id="clone-address" value="${esc(w.address)}"${invalid} placeholder="https://github.com/your-org/kiro-configs.git" spellcheck="false" autocomplete="off" aria-describedby="clone-hint clone-error"><span class="hint" id="clone-hint">An https, ssh or file address, git@host:path, or an absolute path. If git asks you to sign in, answer in the kirokit tab below.</span><span class="error" id="clone-error" role="alert">${esc(w.error)}</span></div>
  <div class="actions"><button class="btn primary" type="submit"${w.busy ? ' disabled' : ''}>${w.busy ? 'Cloning…' : 'Clone'}</button></div>
</form>`;
}

// initWelcome returns the paint function. The form is redrawn only when its
// markup changes, so a poll that redraws every page never takes focus or
// typing away from the field. onClone(address) starts the clone.
export function initWelcome(section, onClone) {
  let last = '';
  section.addEventListener('submit', (e) => {
    e.preventDefault();
    onClone(section.querySelector('#clone-address').value);
  });
  return (view) => {
    const html = welcomeView(view);
    if (html === last) return;
    last = html;
    section.innerHTML = html;
    if (view.welcome?.error) section.querySelector('#clone-address').focus();
  };
}
