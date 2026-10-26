import * as api from '../api.js';
import { refArgs, versionOptions, sourceLine } from './version.js';

// SaveSettings refuses with one "KiroHome: ...", "RepoURL: ..." or "RepoDir: ..." line per bad
// field. Anything else is a failure that belongs to no field.
export function parseFieldErrors(message) {
  const found = { kiroHome: '', repoUrl: '', repoDir: '', other: '' };
  for (const line of String(message).split('\n')) {
    if (line.startsWith('KiroHome: ')) found.kiroHome = line.slice('KiroHome: '.length);
    else if (line.startsWith('RepoDir: ')) found.repoDir = line.slice('RepoDir: '.length);
    else if (line.startsWith('RepoURL: ')) found.repoUrl = line.slice('RepoURL: '.length);
    else if (line.trim()) found.other = line.trim();
  }
  return found;
}

// Saves the form and, when it holds a different repository address, starts the
// fresh-folder clone. The old address is what gets saved with the form: Clone
// records the new one itself once the clone succeeds, so a refused or failed
// address is never remembered. Resolves to whether a clone started; rejects
// with the save's or the clone's refusal (a save refusal means no clone
// started).
export async function saveAndMaybeClone(save, clone, saved, form) {
  const changed = form.repoUrl !== '' && form.repoUrl !== saved.repoUrl;
  await save({ repoUrl: changed ? saved.repoUrl : form.repoUrl, kiroHome: form.kiroHome, repoDir: form.repoDir });
  if (changed) await clone(form.repoUrl);
  return changed;
}

// Update fetches the saved version (ref, '' for Newest release): for the
// installed team when there is one, else only the checkout.
export const updateArgs = (installs, ref) => {
  const team = (installs || [])[0]?.Team;
  return team ? ['sync', team, ...refArgs(ref)] : ['sync', ...refArgs(ref)];
};

// Update's label says when pressing it switches to another version.
export const updateLabel = (picked, saved) =>
  picked === saved ? 'Update' : `Switch to ${picked || 'Newest release'} and update`;

// onUpdate(picked) runs Update with the version picked in the select;
// onClone(address) starts a clone of a new address and rejects with the reason
// it was refused. The returned paint(view) keeps the controls disabled while a
// command runs and shows a refusal beside them.
export function initSettings(onUpdate, onClone) {
  const updateBtn = document.getElementById('update');
  const updateErr = document.getElementById('update-error');
  const version = document.getElementById('version');
  const versionErr = document.getElementById('version-error');
  const source = document.getElementById('version-source');
  // The entry the person picked; null follows the saved choice, so a redraw
  // never takes a pick away and a finished switch shows the new choice.
  let picked = null;
  let lastView = null;
  let lastOptions = '';
  version.addEventListener('change', () => {
    picked = version.value;
    if (lastView) paint(lastView);
  });
  updateBtn.addEventListener('click', () => onUpdate(picked ?? lastView?.ref ?? ''));

  const home = document.getElementById('kirohome');
  const repo = document.getElementById('repo');
  const homeErr = document.getElementById('kirohome-error');
  const repoErr = document.getElementById('repo-error');
  const dir = document.getElementById('repodir');
  const dirErr = document.getElementById('repodir-error');
  const status = document.getElementById('save-status');

  const show = (input, errEl, message) => {
    errEl.textContent = message;
    if (message) input.setAttribute('aria-invalid', 'true');
    else input.removeAttribute('aria-invalid');
  };

  const envNote = document.getElementById('env-note');
  const saved = { repoUrl: '' }; // the address Settings last held
  let edited = false;
  let dirEdited = false;
  home.addEventListener('input', () => { edited = true; });
  dir.addEventListener('input', () => { dirEdited = true; });

  // Shows the saved values, or the default Kiro home before anything is saved.
  // The window opens before the login shell has been read, so until it has
  // (env.ready) the default is provisional: ask again, and replace it by the
  // captured KIRO_HOME unless the person has typed in the field.
  async function load() {
    try {
      const st = await api.state();
      if (!edited) home.value = st.settings.kiroHome || st.env.defaultKiroHome;
      if (!dirEdited) dir.value = st.settings.repoDir || st.env.defaultRepoDir || '';
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
    show(dir, dirErr, '');
    let cloning = false;
    try {
      const form = { repoUrl: repo.value.trim(), kiroHome: home.value.trim(), repoDir: dir.value.trim() };
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
      show(dir, dirErr, err.repoDir);
      status.textContent = err.other;
      if (err.kiroHome) home.focus();
      else if (err.repoDir) dir.focus();
      else if (err.repoUrl) repo.focus();
    }
  });

  function paint(view) {
    lastView = view;
    const saved = view.ref || '';
    if (picked === saved) picked = null;
    const current = picked ?? saved;
    const options = versionOptions(view.checkout, saved, current);
    if (options !== lastOptions) {
      lastOptions = options;
      version.innerHTML = options;
    }
    version.value = current;
    source.innerHTML = sourceLine(view, 'In use:');

    const busy = Boolean(view.running) || Boolean(view.switching);
    version.disabled = busy;
    updateBtn.disabled = busy;
    updateBtn.textContent = view.switching === 'settings' ? 'Switching…' : updateLabel(current, saved);
    const e = view.actionError;
    updateErr.textContent = e && e.where === 'update' ? e.message : '';
    const r = view.refError;
    const failed = r && r.where === 'settings' ? r.message : '';
    versionErr.textContent = failed;
    if (failed) version.setAttribute('aria-invalid', 'true');
    else version.removeAttribute('aria-invalid');
  }

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
