// The two dialogs of Install: replacing the installed team, and the setup
// check. The state builders are plain functions so node's test runner covers
// them; createDialogs draws them in the <dialog> elements of index.html.
import { esc } from './pages/html.js';

// What the setup dialog shows for the checks Go reported.
export function setupState(checks) {
  const list = checks || [];
  const failing = list.filter((c) => !c.found).length;
  const canInstall = list.length > 0 && failing === 0;
  let status = 'Checking...';
  if (list.length) {
    status = failing === 0
      ? 'Everything is ready. You can install.'
      : `${failing} of ${list.length} checks ${failing === 1 ? 'needs' : 'need'} attention. Fix them, then press Check again.`;
  }
  return { checks: list, failing, canInstall, status };
}

export function setupRows(checks) {
  return (checks || [])
    .map((c) => `<li data-found="${c.found}"><span class="badge ${c.found ? 'ok' : 'warn'}">${c.found ? 'found' : 'missing'}</span> <b>${esc(c.name)}</b><span class="detail">${esc(c.detail)}</span></li>`)
    .join('\n');
}

// What pressing Install on a team leads to: nothing when it is already
// installed, the replace dialog when another team is, else the setup check.
export function installPlan(team, installs) {
  const teams = (installs || []).map((r) => r.Team);
  if (teams.includes(team)) return { kind: 'skip' };
  return teams.length ? { kind: 'replace', installed: teams } : { kind: 'setup' };
}

export function replaceMessage(team, installed) {
  return `${installed.join(' and ')} is installed. A Kiro home holds one team, so installing ${team} uninstalls ${installed.join(' and ')} first.`;
}

// createDialogs wires the dialog markup in index.html. Each question resolves
// true when the person confirms, false on Cancel or Escape.
export function createDialogs(doc = document) {
  const ask = (dialog, ready) =>
    new Promise((resolve) => {
      dialog.returnValue = '';
      dialog.addEventListener('close', () => resolve(dialog.returnValue === 'ok'), { once: true });
      dialog.showModal();
      ready?.();
    });

  const replace = doc.getElementById('replace-dialog');
  const setup = doc.getElementById('setup-dialog');

  return {
    confirmReplace(team, installed) {
      doc.getElementById('replace-text').textContent = replaceMessage(team, installed);
      return ask(replace);
    },

    // loadChecks() resolves to the checks; Check again calls it afresh.
    confirmSetup(team, loadChecks) {
      doc.getElementById('setup-title').textContent = `Install ${team}`;
      const rows = doc.getElementById('setup-checks');
      const status = doc.getElementById('setup-status');
      const install = doc.getElementById('setup-install');
      const again = doc.getElementById('setup-again');

      const show = (checks) => {
        const s = setupState(checks);
        rows.innerHTML = setupRows(s.checks);
        status.textContent = s.status;
        install.disabled = !s.canInstall;
      };
      const check = async () => {
        again.disabled = true;
        install.disabled = true;
        status.textContent = 'Checking...';
        try {
          show(await loadChecks());
        } catch (e) {
          rows.innerHTML = '';
          status.textContent = String((e && e.message) || e);
        } finally {
          again.disabled = false;
        }
      };
      again.onclick = check;
      return ask(setup, check);
    },
  };
}
