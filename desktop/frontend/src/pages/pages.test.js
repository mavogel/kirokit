import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, readdirSync } from 'node:fs';
import { teamsView, quickArgs, quickRuns } from './teams.js';
import { agentsView } from './agents.js';
import { installView, syncArgs, uninstallArgs } from './install.js';
import { updateArgs, updateLabel, saveAndMaybeClone, initSettings } from './settings.js';
import { refArgs, versionOptions, sourceLine } from './version.js';
import { skillsView } from './skills.js';
import { welcomeView, showWelcome } from './welcome.js';
import { feedbackFor } from './feedback.js';

const catalog = {
  tag: 'v1.0.0',
  pins: { 'jira-mcp': '1.4.2', 'unused-mcp': '0.1.0' },
  teams: [
    {
      name: 'alpha',
      prompts: 2,
      steerings: ['style'],
      agents: [{ name: 'ops', description: 'Does <ops> & more', tools: ['read', '@jira'], servers: ['jira'], pins: ['jira-mcp'], needsKey: true }],
      skills: [
        { name: 'own', source: 'Written by the team' },
        { name: 'ext', source: 'https://example.com/s.git @ v1' },
      ],
    },
    { name: 'beta', prompts: 0, steerings: [], agents: [], skills: [] },
  ],
};
const installs = [{ Team: 'alpha', Files: 7, Modified: '2026-10-17T09:30:00Z' }];
const ready = { catalog, installs, error: '', cloned: true };

test('Teams marks only the team with a record as installed and gives the change time', () => {
  const html = teamsView(ready);
  const [alpha, beta] = html.split('</article>');
  assert.match(alpha, /data-team="alpha"/);
  assert.match(alpha, /badge ok">installed</);
  assert.match(alpha, /<time datetime="2026-10-17T09:30:00.000Z">/);
  assert.match(beta, /not installed/);
  assert.doesNotMatch(beta, /<time/);
});

test('Install lists installed teams with their file counts', () => {
  const out = installView(ready);
  assert.equal(out.blocked, false);
  assert.match(out.html, /data-team="alpha"/);
  assert.match(out.html, /<td class="num">7<\/td>/);
  assert.doesNotMatch(out.html, /beta/);
});

test('Install says so when nothing is installed', () => {
  const out = installView({ ...ready, installs: [] });
  assert.equal(out.blocked, true);
  assert.match(out.html, /Nothing is installed/);
});

test('Agents escapes text and flags the agent that needs a key', () => {
  const out = agentsView(ready);
  assert.match(out.html, /Does &lt;ops&gt; &amp; more/);
  assert.match(out.html, /needs key/);
  assert.match(out.html, /<span class="chip">read<\/span>/);
});

test('Skills shows each source and the teams that use a pinned server', () => {
  const out = skillsView(ready);
  assert.match(out.skills, /Written by the team/);
  assert.match(out.skills, /https:\/\/example.com\/s.git @ v1/);
  assert.match(out.mcp, /jira-mcp<\/td><td class="mono">1.4.2<\/td><td>alpha<\/td>/);
  assert.match(out.mcp, /unused-mcp[\s\S]*no agent yet/);
});

test('A render error replaces every table with its message', () => {
  const view = { catalog: { teams: [], pins: {} }, installs, error: 'team alpha: render: boom <x>', cloned: true };
  assert.match(teamsView(view), /role="alert">team alpha: render: boom &lt;x&gt;</);
  for (const out of [agentsView(view), skillsView(view)]) {
    assert.equal(out.blocked, true);
    assert.match(out.html ?? out.skills, /boom/);
  }
});

test('Without a clone the pages point at Settings', () => {
  const view = { catalog: { teams: [], pins: {} }, installs: [], error: '', cloned: false };
  assert.match(teamsView(view), /Enter its address in Settings/);
  assert.equal(agentsView(view).blocked, true);
});

test('Each team card has an Install button that is disabled while a command runs', () => {
  const idle = teamsView(ready);
  assert.match(idle, /<button class="btn primary" data-action="install" data-team="alpha">Install<\/button>/);
  assert.match(idle, /data-action="install" data-team="beta"/);

  const busy = teamsView({ ...ready, running: true });
  assert.match(busy, /data-action="install" data-team="alpha" disabled/);
});

test('A still-running error shows beside the button it came from', () => {
  const view = { ...ready, actionError: { where: 'install:beta', message: 'kirokit is still running a command' } };
  const [alpha, beta] = teamsView(view).split('</article>');
  assert.doesNotMatch(alpha, /still running/);
  assert.match(beta, /role="alert">kirokit is still running a command</);
});

test('Install rows have Sync and Uninstall buttons for their team', () => {
  const html = installView(ready).html;
  assert.match(html, /data-action="sync" data-team="alpha">Sync</);
  assert.match(html, /data-action="uninstall" data-team="alpha">Uninstall</);
  assert.match(installView({ ...ready, running: true }).html, /data-action="sync" data-team="alpha" disabled/);
});

test('Buttons run the commands the design names', () => {
  assert.deepEqual(syncArgs('alpha', ''), ['sync', 'alpha', '--latest-tag']);
  assert.deepEqual(uninstallArgs('alpha'), ['uninstall', 'alpha']);
  assert.deepEqual(updateArgs([], ''), ['sync', '--latest-tag']);
  assert.deepEqual(updateArgs([{ Team: 'alpha' }], ''), ['sync', 'alpha', '--latest-tag']);
});

test('Update and Sync keep the saved version instead of the newest release', () => {
  assert.deepEqual(syncArgs('alpha', 'feature'), ['sync', 'alpha', '--branch', 'feature']);
  assert.deepEqual(updateArgs([{ Team: 'alpha' }], 'v1.0.0'), ['sync', 'alpha', '--branch', 'v1.0.0']);
  assert.deepEqual(updateArgs([], 'main'), ['sync', '--branch', 'main']);
  assert.deepEqual(refArgs(''), ['--latest-tag']);
});

const refs = {
  current: { kind: 'branch', name: 'main', commit: 'b81d0a7' },
  branches: ['feature/<x>', 'main', 'both'],
  tags: ['v1.3.0', 'v1.2.0', 'both'],
  newest: 'v1.3.0',
};

test('The version list offers Newest release, then the branches, then the tags', () => {
  const html = versionOptions(refs, '', '');
  assert.match(html, /^<option value="" selected>Newest release · follows new releases \(now v1.3.0\)<\/option>/);
  assert.match(html, /<optgroup label="Branches"><option value="feature\/&lt;x&gt;">feature\/&lt;x&gt;<\/option><option value="main">main<\/option><option value="both">both<\/option><\/optgroup>/);
  // A branch wins over a tag of the same name, so that tag cannot be chosen.
  assert.match(html, /<optgroup label="Tags"><option value="v1.3.0">v1.3.0<\/option><option value="v1.2.0">v1.2.0<\/option><\/optgroup>$/);
  assert.match(versionOptions(refs, '', 'v1.2.0'), /<option value="v1.2.0" selected>/);
  assert.match(versionOptions({ ...refs, newest: '' }, '', 'main'), /^<option value="">Newest release \(none yet\)<\/option>/);
});

test('A saved version that is gone from origin still shows in the list', () => {
  assert.match(versionOptions(refs, 'old', 'old'), /<optgroup label="Saved"><option value="old" selected>old \(not on origin\)<\/option><\/optgroup>/);
  // Before the clone's refs are read nothing is known to be gone.
  assert.doesNotMatch(versionOptions({ branches: [], tags: [] }, 'old', 'old'), /not on origin/);
});

test('The source line names the branch, tag or newest release and marks anything else', () => {
  const at = (current, ref) => sourceLine({ checkout: { ...refs, current }, ref }, 'Templates from');
  assert.equal(at({ kind: 'branch', name: 'main', commit: 'b81d0a7' }, 'main'),
    'Templates from branch <b>main</b> at <b>b81d0a7</b> <span class="badge warn">not a release</span>');
  assert.equal(at({ kind: 'tag', name: 'v1.3.0', commit: '4f2c9e1' }, ''),
    'Templates from <span class="badge off">newest release</span> <b>v1.3.0</b> at <b>4f2c9e1</b>');
  // A pinned tag is a pin, even when it is today's newest release.
  assert.equal(at({ kind: 'tag', name: 'v1.3.0', commit: '4f2c9e1' }, 'v1.3.0'),
    'Templates from tag <b>v1.3.0</b> at <b>4f2c9e1</b> <span class="badge warn">pinned · does not follow new releases</span>');
  // What git says is checked out is shown, even when it differs from the choice.
  assert.match(at({ kind: 'tag', name: 'v1.2.0', commit: '1111111' }, ''), /tag <b>v1.2.0<\/b>.*not the newest release/);
  assert.match(at({ kind: 'branch', name: 'x<y>', commit: 'c' }, ''), /branch <b>x&lt;y&gt;<\/b>.*not a release/);
  assert.equal(at({ kind: '', name: '', commit: 'abc1234' }, ''),
    'Templates from commit <b>abc1234</b> <span class="badge warn">not a release</span>');
  assert.equal(sourceLine({ checkout: { current: { kind: '', name: '', commit: '' } }, ref: '' }, 'In use:'), '');
});

test('Settings names the switch on its Update button', () => {
  assert.equal(updateLabel('', ''), 'Update');
  assert.equal(updateLabel('feature', 'feature'), 'Update');
  assert.equal(updateLabel('feature', ''), 'Switch to feature and update');
  assert.equal(updateLabel('', 'feature'), 'Switch to Newest release and update');
});

test('First launch asks for the address, prefilled from Settings', () => {
  const view = { catalog: { teams: [], pins: {} }, installs: [], error: '', cloned: false, welcome: { address: 'https://example.com/x.git?a=1&b="2"' } };
  const html = welcomeView(view);
  assert.match(html, /<h1>Set up kirokit<\/h1>/);
  assert.match(html, /id="clone-address"[^>]*value="https:\/\/example.com\/x.git\?a=1&amp;b=&quot;2&quot;"/);
  assert.match(html, /<button class="btn primary" type="submit">Clone<\/button>/);
  assert.doesNotMatch(html, /role="alert">[^<]/);
});

test('A refused or failed address stays in the field with its error beside it', () => {
  const view = { cloned: false, welcome: { address: '-oops', error: 'Use an https, ssh or file address, <or> a path' } };
  const html = welcomeView(view);
  assert.match(html, /value="-oops"/);
  assert.match(html, /aria-invalid="true"/);
  assert.match(html, /id="clone-error" role="alert">Use an https, ssh or file address, &lt;or&gt; a path</);
});

test('While the clone runs the button says so and cannot be pressed twice', () => {
  const html = welcomeView({ cloned: false, welcome: { address: '/x', busy: true } });
  assert.match(html, /<button class="btn primary" type="submit" disabled>Cloning…<\/button>/);
});

test('The sidebar is hidden until a clone exists', () => {
  assert.equal(showWelcome({ cloned: false }), true);
  assert.equal(showWelcome({ cloned: true }), false);
  // git has made the folder before the first sync ends; the page stays until clone:done.
  assert.equal(showWelcome({ cloned: true, welcome: { busy: true } }), true);
});

const choosing = (welcome, extra = {}) => ({
  cloned: true,
  checkout: refs,
  ref: '',
  welcome: { address: 'https://example.com/x.git', choosing: true, ...welcome },
  ...extra,
});

test('First launch stays up after the clone to ask for the version', () => {
  assert.equal(showWelcome({ cloned: true, welcome: { choosing: true } }), true);
  const html = welcomeView(choosing({}));
  assert.match(html, /<h1>Set up kirokit<\/h1>/);
  assert.match(html, /Cloned <b>https:\/\/example.com\/x.git<\/b>/);
  assert.match(html, /<select id="welcome-version"[^>]*><option value="" selected>Newest release · follows new releases/);
  assert.match(html, /<button class="btn primary" type="submit">Continue<\/button>/);
  assert.doesNotMatch(html, /id="clone-address"/);
});

test('First launch names the switch, then says it is switching', () => {
  assert.match(welcomeView(choosing({ pick: 'main' })), /<option value="main" selected>main<\/option>[\s\S]*type="submit">Switch to main and continue</);
  const busy = welcomeView(choosing({ pick: 'main' }, { switching: 'welcome' }));
  assert.match(busy, /<select id="welcome-version"[^>]* disabled>/);
  assert.match(busy, /type="submit" disabled>Switching…</);
});

test('First launch shows a failed switch under the list and lets the person pick again', () => {
  const failed = welcomeView(choosing({ pick: 'gone' }, { refError: { where: 'welcome', message: 'kirokit sync: --branch gone: <no> branch' } }));
  assert.match(failed, /aria-invalid="true"/);
  assert.match(failed, /id="welcome-version-error" role="alert">kirokit sync: --branch gone: &lt;no&gt; branch</);
  assert.match(failed, /type="submit">Switch to gone and continue</);
  // An error from Settings does not show here.
  assert.doesNotMatch(welcomeView(choosing({}, { refError: { where: 'settings', message: 'x' } })), /role="alert">x</);
});

test('First launch without a release says so and preselects the checked-out branch', () => {
  const html = welcomeView(choosing({}, { checkout: { ...refs, newest: '' } }));
  assert.match(html, /no release yet/);
  assert.match(html, /<option value="main" selected>/);
  assert.match(html, /Switch to main and continue/);
});

test('A clone started from Settings keeps the pages showing', () => {
  const view = { cloned: true, welcome: { busy: true, fromSettings: true } };
  assert.equal(showWelcome(view), false);
  assert.equal(showWelcome({ ...view, cloned: false }), false);
  assert.equal(showWelcome({ cloned: true, welcome: { fromSettings: true, choosing: true } }), false);
});

test('Saving a different address in Settings clones it; the same one does not', async () => {
  const calls = [];
  const save = async (s) => calls.push(['save', s]);
  const clone = async (a) => calls.push(['clone', a]);
  const saved = { repoUrl: 'https://example.com/old.git', kiroHome: '/h', repoDir: '/r' };

  const changed = await saveAndMaybeClone(save, clone, saved, { repoUrl: 'https://example.com/new.git', kiroHome: '/h2', repoDir: '/r' });
  assert.equal(changed, true);
  // The old address is saved with the home folder; Clone records the new one
  // only once the clone succeeds, so a refused or failed address is never remembered.
  assert.deepEqual(calls, [['save', { repoUrl: 'https://example.com/old.git', kiroHome: '/h2', repoDir: '/r' }], ['clone', 'https://example.com/new.git']]);

  calls.length = 0;
  assert.equal(await saveAndMaybeClone(save, clone, saved, { ...saved }), false);
  assert.deepEqual(calls, [['save', saved]]);
});

test('A refused save does not start a clone, and a refused clone rejects', async () => {
  const saved = { repoUrl: 'https://example.com/old.git', kiroHome: '/h' };
  let cloned = false;
  await assert.rejects(
    saveAndMaybeClone(async () => { throw new Error('KiroHome: no'); }, async () => { cloned = true; }, saved, { repoUrl: 'https://example.com/n.git', kiroHome: '/' }),
    /KiroHome/,
  );
  assert.equal(cloned, false);
  await assert.rejects(
    saveAndMaybeClone(async () => {}, async () => { throw new Error('kirokit is still running a command'); }, saved, { repoUrl: 'https://example.com/n.git', kiroHome: '/h' }),
    /still running/,
  );
});

// A Settings page on a fake document: elements keep their handlers and text.
function settingsPage({ stored, cloneRejects, onUpdate = () => {} }) {
  const els = {};
  const el = (id) => (els[id] ||= {
    id, value: '', textContent: '', disabled: false, attrs: {}, handlers: {},
    addEventListener(type, fn) { this.handlers[type] = fn; },
    setAttribute(k, v) { this.attrs[k] = v; },
    removeAttribute(k) { delete this.attrs[k]; },
    focus() {},
  });
  globalThis.document = { getElementById: el };
  const calls = [];
  globalThis.window = {
    go: { main: { App: {
      State: async () => ({ settings: { kiroHome: '/h', repoUrl: stored.repoUrl }, env: { ready: true, defaultKiroHome: '/h', note: '' } }),
      SaveSettings: async (s) => { calls.push(['save', s.repoUrl]); stored.repoUrl = s.repoUrl; },
    } } },
  };
  const page = initSettings(onUpdate, async (address) => {
    calls.push(['clone', address]);
    if (cloneRejects) throw new Error(cloneRejects);
  });
  return { els, el, calls, page };
}

test('A failed clone from Settings shows beside the address and leaves Save able to retry', async () => {
  const stored = { repoUrl: '/old/origin' };
  const { el, calls, page } = settingsPage({ stored });
  await page.load();
  el('repo').value = '/missing/origin';

  await el('save').handlers.click();
  assert.match(el('save-status').textContent, /Cloning/);

  page.cloneDone({ ok: false, error: 'git clone failed (exit 128).' });
  assert.equal(el('repo-error').textContent, 'git clone failed (exit 128).');
  assert.equal(el('repo').attrs['aria-invalid'], 'true');
  assert.doesNotMatch(el('save-status').textContent, /Cloning/);
  assert.equal(stored.repoUrl, '/old/origin');

  calls.length = 0;
  await el('save').handlers.click();
  assert.deepEqual(calls, [['save', '/old/origin'], ['clone', '/missing/origin']]);
  assert.equal(el('repo-error').textContent, '');
});

test('A finished clone clears the Cloning status in Settings', async () => {
  const { el, page } = settingsPage({ stored: { repoUrl: '/old/origin' } });
  await page.load();
  el('repo').value = '/new/origin';
  await el('save').handlers.click();

  page.cloneDone({ ok: true, error: '' });
  assert.doesNotMatch(el('save-status').textContent, /Cloning/);
  assert.match(el('save-status').textContent, /Cloned/);
  assert.equal(el('repo-error').textContent, '');

  page.cloneDone({ ok: true, error: 'This repository has no release yet.' });
  assert.match(el('save-status').textContent, /no release yet/);
});

test('Settings keeps a picked version across redraws, switches with it, and shows a failed switch', () => {
  const updates = [];
  const { el, page } = settingsPage({ stored: { repoUrl: '' }, onUpdate: (picked) => updates.push(picked) });
  const view = { checkout: { ...refs, current: { kind: 'tag', name: 'v1.3.0', commit: '4f2c9e1' } }, ref: '', installs: [] };

  page.paint(view);
  assert.match(el('version').innerHTML, /<option value="" selected>Newest release/);
  assert.equal(el('update').textContent, 'Update');
  assert.match(el('version-source').innerHTML, /^In use: <span class="badge off">newest release<\/span> <b>v1.3.0<\/b>/);

  el('version').value = 'main';
  el('version').handlers.change();
  page.paint({ ...view }); // a poll redraws: the pick stays
  assert.equal(el('version').value, 'main');
  assert.equal(el('update').textContent, 'Switch to main and update');
  el('update').handlers.click();
  assert.deepEqual(updates, ['main']);

  page.paint({ ...view, running: true, switching: 'settings' });
  assert.equal(el('version').disabled, true);
  assert.equal(el('update').disabled, true);
  assert.equal(el('update').textContent, 'Switching…');

  page.paint({ ...view, refError: { where: 'settings', message: 'kirokit sync: --branch main: no branch' } });
  assert.equal(el('version-error').textContent, 'kirokit sync: --branch main: no branch');
  assert.equal(el('version').attrs['aria-invalid'], 'true');
  assert.equal(el('version').value, 'main', 'the failed pick stays so it can be retried');

  // Once the switch is saved the button is plain Update again.
  page.paint({ ...view, ref: 'main', checkout: { ...refs } });
  assert.equal(el('update').textContent, 'Update');
  assert.equal(el('version-error').textContent, '');
  assert.equal(el('version').attrs['aria-invalid'], undefined);
});

test('Settings Update keeps the saved version when nothing else is picked, and Newest release is the way back', () => {
  const updates = [];
  const { el, page } = settingsPage({ stored: { repoUrl: '' }, onUpdate: (picked) => updates.push(picked) });
  const view = { checkout: { ...refs, current: { kind: 'branch', name: 'main', commit: 'b81d0a7' } }, ref: 'main', installs: [{ Team: 'alpha' }] };

  // A fresh page with a saved branch: Update follows it, never the newest release.
  page.paint(view);
  assert.equal(el('version').value, 'main');
  assert.equal(el('update').textContent, 'Update');
  el('update').handlers.click();
  assert.deepEqual(updates, ['main']);

  // Picking Newest release is how the person goes back to releases.
  el('version').value = '';
  el('version').handlers.change();
  assert.equal(el('update').textContent, 'Switch to Newest release and update');
  el('update').handlers.click();
  assert.deepEqual(updates, ['main', '']);
});

// Owners and feedback: alpha routes feedback to a repository, gamma does not.
const owned = {
  tag: '',
  pins: {},
  teams: [
    {
      name: 'alpha',
      feedback: 'acme/cfg',
      prompts: 0,
      steerings: [],
      agents: [
        { name: 'po', description: 'd', tools: [], servers: [], owner: '@nick', ownerFromTeam: false },
        { name: 'eng', description: 'd', tools: [], servers: [], owner: '@jane', ownerFromTeam: true },
      ],
      skills: [{ name: 'vendored', source: 'https://x @ v1', owner: '@jane', ownerFromTeam: true }],
    },
    {
      name: 'gamma',
      feedback: '',
      prompts: 0,
      steerings: [],
      agents: [{ name: 'g1', description: 'd', tools: [], servers: [], owner: '', ownerFromTeam: false }],
      skills: [{ name: 'gs', source: 'Written by the team', owner: '<b>@x</b>', ownerFromTeam: false }],
    },
  ],
};
const ownedView = { catalog: owned, installs: [], error: '', cloned: true };
const row = (html, name) => html.split('</tr>').find((r) => r.includes(`>${name}<`));

test('Agents shows each owner, marking the inherited one and the missing one', () => {
  const html = agentsView(ownedView).html;
  assert.match(row(html, 'po'), /<span class="mono">@nick<\/span><\/span>/);
  assert.doesNotMatch(row(html, 'po'), /team owner/);
  assert.match(row(html, 'eng'), /@jane<\/span><span class="from">team owner<\/span>/);
  assert.match(row(html, 'g1'), /No owner/);
});

test('Each agent and skill has a Feedback button naming it, off for a team without a feedback repository', () => {
  const agents = agentsView(ownedView);
  assert.match(row(agents.html, 'po'), /<button class="btn" data-feedback="agent" data-team="alpha" data-name="po" aria-label="Give feedback on po">Feedback<\/button>/);
  assert.match(row(agents.html, 'g1'), /<button class="btn" disabled aria-label="Give feedback on g1 \(no feedback repository\)">Feedback<\/button>/);
  assert.match(agents.note, /gamma names no feedback repository in its manifest.yaml, so its Feedback buttons are off/);
  assert.doesNotMatch(agents.note, /alpha/);

  const skills = skillsView(ownedView);
  assert.match(row(skills.skills, 'vendored'), /@jane<\/span><span class="from">team owner<\/span>/);
  assert.match(row(skills.skills, 'vendored'), /data-feedback="skill" data-team="alpha" data-name="vendored"/);
  assert.match(row(skills.skills, 'gs'), /&lt;b&gt;@x&lt;\/b&gt;/);
  assert.match(row(skills.skills, 'gs'), /disabled/);
  assert.match(skills.note, /gamma names no feedback repository/);
  const allRouted = { ...ownedView, catalog: { ...owned, teams: [owned.teams[0]] } };
  assert.equal(agentsView(allRouted).note, '', 'no note when no row has a disabled button');
  assert.equal(skillsView(allRouted).note, '');
});

test('A refused open shows beside the Feedback button it came from', () => {
  const view = { ...ownedView, actionError: { where: 'feedback:alpha:agent:po', message: 'refusing to open x' } };
  const html = agentsView(view).html;
  assert.match(row(html, 'po'), /role="alert">refusing to open x</);
  assert.doesNotMatch(row(html, 'eng'), /role="alert"/);
});

test('Feedback builds the issue for the pressed item from the catalog', () => {
  const url = new URL(feedbackFor(owned, { team: 'alpha', kind: 'agent', name: 'po' }));
  assert.equal(url.pathname, '/acme/cfg/issues/new');
  assert.match(url.searchParams.get('body'), /\*\*Owner:\*\* @nick/);
  assert.throws(() => feedbackFor(owned, { team: 'alpha', kind: 'skill', name: 'gone' }), /gone is no longer in alpha/);
  assert.throws(() => feedbackFor(owned, { team: 'gamma', kind: 'agent', name: 'g1' }), /gamma names no feedback repository/);
});

const quickView = {
  ...ready,
  catalog: { ...catalog, teams: [{ ...catalog.teams[0], quick: true }, catalog.teams[1]] },
};

test('Export for Quick runs render --quick into the given folder', () => {
  assert.deepEqual(quickArgs('alpha', '/data/quick/alpha'), ['render', '--quick', 'alpha', '--out', '/data/quick/alpha']);
});

test('Only a team that lists Quick content has Export for Quick; the others say why', () => {
  const [alpha, beta] = teamsView(quickView).split('</article>');
  assert.match(alpha, /<button class="btn" data-action="quick" data-team="alpha"[^>]*>Export for Quick<\/button>/);
  assert.doesNotMatch(alpha, /No Quick export/);
  assert.doesNotMatch(beta, /data-action="quick"/);
  assert.match(beta, /No Quick export: this team lists no skills or agents for Quick\./);
});

test('While a command runs Export for Quick is disabled, and says it is exporting for its own team', () => {
  const html = teamsView({ ...quickView, running: true, exporting: 'alpha' });
  assert.match(html, /data-action="quick" data-team="alpha"[^>]* disabled>Exporting…<\/button>/);
  const other = teamsView({ ...quickView, running: true, exporting: '' });
  assert.match(other, /data-action="quick" data-team="alpha"[^>]* disabled>Export for Quick<\/button>/);
});

test('A finished export shows the counts, the folder and Show folder', () => {
  const one = teamsView({ ...quickView, quick: { alpha: { dir: '/data/quick/<alpha>', skills: 1, agents: 1 } } });
  assert.match(one, /role="status"><span>Quick bundle ready: <b>1<\/b> skill, <b>1<\/b> agent\.<\/span>/);
  assert.match(one, /<button class="link-btn" type="button" data-action="quick-folder" data-team="alpha">Show folder<\/button>/);
  assert.match(one, /<span class="path">\/data\/quick\/&lt;alpha&gt;<\/span>/);
  const many = teamsView({ ...quickView, quick: { alpha: { dir: '/d', skills: 2, agents: 0 } } });
  assert.match(many, /<b>2<\/b> skills, <b>0<\/b> agents\./);
  assert.doesNotMatch(teamsView(quickView), /Quick bundle ready/);
});

test('A failed export shows its error beside the button', () => {
  const html = teamsView({ ...quickView, actionError: { where: 'quick:alpha', message: 'kirokit render failed: see the kirokit tab below.' } });
  const [alpha] = html.split('</article>');
  assert.match(alpha, /Export for Quick<\/button><span class="error" role="alert">kirokit render failed: see the kirokit tab below\.<\/span>/);
});

test('The Quick result comes only from the exit of the run that was started', () => {
  const runs = quickRuns();
  assert.equal(runs.exited(4, 0), null); // some other command
  assert.equal(runs.started('alpha', 5), null);
  assert.equal(runs.exited(6, 0), null);
  assert.deepEqual(runs.exited(5, 1), { team: 'alpha', ok: false });
  assert.equal(runs.exited(5, 0), null); // taken once

  // A fast command can end before its id arrives.
  const fast = quickRuns();
  assert.equal(fast.exited(9, 0), null);
  assert.deepEqual(fast.started('beta', 9), { team: 'beta', ok: true });
});

test('No page source names a sample team', () => {
  const dir = new URL('.', import.meta.url);
  for (const f of readdirSync(dir).filter((n) => n.endsWith('.js') && !n.endsWith('.test.js'))) {
    assert.doesNotMatch(readFileSync(new URL(f, dir), 'utf8'), /platform-team|data-team(?!=)/, f);
  }
});
