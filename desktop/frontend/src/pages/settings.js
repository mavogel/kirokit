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

export function initSettings() {
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

  // Shows the saved values, or the default Kiro home before anything is saved.
  async function load() {
    try {
      const st = await api.state();
      home.value = st.settings.kiroHome || st.env.defaultKiroHome;
      repo.value = st.settings.repoUrl;
    } catch (e) {
      status.textContent = String(e.message || e);
    }
  }

  document.getElementById('save').addEventListener('click', async () => {
    status.textContent = '';
    show(home, homeErr, '');
    show(repo, repoErr, '');
    try {
      await api.saveSettings({ repoUrl: repo.value.trim(), kiroHome: home.value.trim() });
      status.textContent = 'Saved on this computer';
    } catch (e) {
      const err = parseFieldErrors(e && e.message ? e.message : e);
      show(home, homeErr, err.kiroHome);
      show(repo, repoErr, err.repoUrl);
      status.textContent = err.other;
      if (err.kiroHome) home.focus();
      else if (err.repoUrl) repo.focus();
    }
  });

  return { load };
}
