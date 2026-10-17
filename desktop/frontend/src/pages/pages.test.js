import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, readdirSync } from 'node:fs';
import { teamsView } from './teams.js';
import { agentsView } from './agents.js';
import { installView } from './install.js';
import { skillsView } from './skills.js';

const catalog = {
  tag: 'v1.0.0',
  pins: { 'jira-mcp': '1.4.2', 'unused-mcp': '0.1.0' },
  teams: [
    {
      name: 'alpha',
      prompts: 2,
      steerings: ['style'],
      agents: [{ name: 'ops', description: 'Does <ops> & more', tools: ['read', '@jira'], servers: ['jira-mcp'], needsKey: true }],
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

test('No page source names a sample team', () => {
  const dir = new URL('.', import.meta.url);
  for (const f of readdirSync(dir).filter((n) => n.endsWith('.js') && !n.endsWith('.test.js'))) {
    assert.doesNotMatch(readFileSync(new URL(f, dir), 'utf8'), /platform-team|data-team(?!=)/, f);
  }
});
