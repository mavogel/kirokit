import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, readdirSync } from 'node:fs';
import { teamsView } from './teams.js';
import { agentsView } from './agents.js';
import { installView, syncArgs, uninstallArgs } from './install.js';
import { updateArgs, saveAndMaybeClone, initSettings } from './settings.js';
import { skillsView } from './skills.js';
import { welcomeView, showWelcome } from './welcome.js';

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
  assert.deepEqual(syncArgs('alpha'), ['sync', 'alpha', '--latest-tag']);
  assert.deepEqual(uninstallArgs('alpha'), ['uninstall', 'alpha']);
  assert.deepEqual(updateArgs([]), ['sync', '--latest-tag']);
  assert.deepEqual(updateArgs([{ Team: 'alpha' }]), ['sync', 'alpha', '--latest-tag']);
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

test('A clone started from Settings keeps the pages showing', () => {
  const view = { cloned: true, welcome: { busy: true, fromSettings: true } };
  assert.equal(showWelcome(view), false);
  assert.equal(showWelcome({ ...view, cloned: false }), false);
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
function settingsPage({ stored, cloneRejects }) {
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
  const page = initSettings(() => {}, async (address) => {
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

test('No page source names a sample team', () => {
  const dir = new URL('.', import.meta.url);
  for (const f of readdirSync(dir).filter((n) => n.endsWith('.js') && !n.endsWith('.test.js'))) {
    assert.doesNotMatch(readFileSync(new URL(f, dir), 'utf8'), /platform-team|data-team(?!=)/, f);
  }
});
