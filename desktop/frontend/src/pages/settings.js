import * as api from '../api.js';

// SaveSettings refuses with one "KiroHome: ..." or "RepoURL: ..." line per bad
// field. Anything else is a failure that belongs to no field.
export function parseFieldErrors(message) {
  const found = { kiroHome: '', repoUrl: '', other: '' };
  for (const line of String(message).split('\n')) {
    if (line.startsWith('KiroHome: ')) found.kiroHome = line.slice('KiroHome: '.length);
    else if (line.startsWith('RepoURL: ')) found.repoUrl = line.slice('RepoURL: '.length);
    else if (line.trim()) found.other = line.trim();
  }
  return found;
}

// Saves the form and, when it holds a different repository address, starts the
// fresh-folder clone. The old address is what gets saved with the form: Clone
// records the new one itself once the clone succeeds, so a refused or failed
// address is never remembered. Resolves to whether a clone started; rejects with the save's or
// the clone's refusal (a save refusal means no clone started).
export async function saveAndMaybeClone(save, clone, saved, form) {
  const changed = form.repoUrl !== '' && form.repoUrl !== saved.repoUrl;
  await save({ repoUrl: changed ? saved.repoUrl : form.repoUrl, kiroHome: form.kiroHome });
  if (changed) await clone(form.repoUrl);
  return changed;
}

// Update fetches the newest release: for the installed team when there is one,
// else only the checkout.
export const updateArgs = (installs) => {
  const team = (installs || [])[0]?.Team;
  return team ? ['sync', team, '--latest-tag'] : ['sync', '--latest-tag'];
};

// onUpdate() runs Update; onClone(address) starts a clone of a new address and
// rejects with the reason it was refused. The returned paint(view) keeps the button disabled
// while a command runs and shows a refusal beside it.
export function initSettings(onUpdate, onClone) {
  const updateBtn = document.getElementById('update');
  const updateErr = document.getElementById('update-error');
  updateBtn.addEventListener('click', () => onUpdate());

  const home = document.getElementById('kirohome');
  const repo = document.getElementById('repo');
  const homeErr = document.getElementById('kirohome-error');
  const repoErr = document.getElementById('repo-error');
  const status = document.getElementById('save-status');

  const show = (input, errEl, message) => {
    errEl.textContent = message;
    if (message) input.setAttribute('aria-invalid', 'true');
    else input.removeAttribute('aria-invalid');
  };

  const envNote = document.getElementById('env-note');
  const saved = { repoUrl: '' }; // the address Settings last held
  let edited = false;
  home.addEventListener('input', () => { edited = true; });

  // Shows the saved values, or the default Kiro home before anything is saved.
  // The window opens before the login shell has been read, so until it has
  // (env.ready) the default is provisional: ask again, and replace it by the
  // captured KIRO_HOME unless the person has typed in the field.
  async function load() {
    try {
      const st = await api.state();
      if (!edited) home.value = st.settings.kiroHome || st.env.defaultKiroHome;
      saved.repoUrl = st.settings.repoUrl || '';
      if (!repo.value) repo.value = st.settings.repoUrl;
      envNote.textContent = st.env.note || '';
      if (!st.env.ready) setTimeout(load, 300);
    } catch (e) {
      status.textContent = String(e.message || e);
    }
  }

  document.getElementById('save').addEventListener('click', async () => {
    status.textContent = '';
    show(home, homeErr, '');
    show(repo, repoErr, '');
    let cloning = false;
    try {
      const form = { repoUrl: repo.value.trim(), kiroHome: home.value.trim() };
      const changed = await saveAndMaybeClone(
        api.saveSettings,
        async (address) => {
          cloning = true;
          await onClone(address);
        },
        saved,
        form,
      );
      // saved.repoUrl stays: it is the address the app uses until a clone of
      // the new one succeeds (load() then reads it), so Save can retry.
      status.textContent = changed ? 'Saved. Cloning the new repository; follow it in the kirokit tab.' : 'Saved on this computer';
    } catch (e) {
      const err = parseFieldErrors(e && e.message ? e.message : e);
      // A refused clone start carries no field prefix: it belongs beside the address.
      if (cloning && !err.repoUrl) {
        err.repoUrl = err.other;
        err.other = '';
      }
      show(home, homeErr, err.kiroHome);
      show(repo, repoErr, err.repoUrl);
      status.textContent = err.other;
      if (err.kiroHome) home.focus();
      else if (err.repoUrl) repo.focus();
    }
  });

  const paint = (view) => {
    updateBtn.disabled = Boolean(view.running);
    const e = view.actionError;
    updateErr.textContent = e && e.where === 'update' ? e.message : '';
  };

  // The end of a clone this page started: a failure shows beside the address,
  // and the status stops saying the clone is running.
  const cloneDone = ({ ok, error }) => {
    if (ok) {
      show(repo, repoErr, '');
      status.textContent = error || 'Cloned. The new repository is in use.';
    } else {
      show(repo, repoErr, error || 'The clone failed. See the kirokit tab.');
      status.textContent = '';
    }
  };

  return { load, paint, cloneDone };
}
