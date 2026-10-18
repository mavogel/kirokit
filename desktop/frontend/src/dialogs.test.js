import test from 'node:test';
import assert from 'node:assert/strict';
import { setupState, installPlan, replaceMessage, setupRows } from './dialogs.js';

const found = (name) => ({ name, found: true, detail: '/usr/bin/' + name });
const missing = (name) => ({ name, found: false, detail: 'Install it with: brew install ' + name });

test('Install stays disabled while a check fails, and the status names how many', () => {
  const state = setupState([found('SOPS'), missing('Decryption key'), missing('Kiro CLI')]);
  assert.equal(state.canInstall, false);
  assert.equal(state.failing, 2);
  assert.match(state.status, /2 of 3 checks need attention/);
});

test('One failing check reads in the singular', () => {
  assert.match(setupState([found('SOPS'), missing('Kiro CLI')]).status, /1 of 2 checks needs attention|1 check needs attention/);
});

test('All three found enables Install', () => {
  const state = setupState([found('SOPS'), found('Decryption key'), found('Kiro CLI')]);
  assert.equal(state.canInstall, true);
  assert.equal(state.failing, 0);
  assert.match(state.status, /ready/i);
});

test('No answer from the checks yet keeps Install disabled', () => {
  assert.equal(setupState([]).canInstall, false);
  assert.equal(setupState(null).canInstall, false);
});

test('Each check row shows found or missing with its detail, escaped', () => {
  const rows = setupRows([found('SOPS'), { name: 'Kiro <CLI>', found: false, detail: 'sign in & retry' }]);
  assert.match(rows, /data-found="true"[^>]*>[\s\S]*SOPS[\s\S]*\/usr\/bin\/SOPS/);
  assert.match(rows, /data-found="false"[\s\S]*Kiro &lt;CLI&gt;[\s\S]*sign in &amp; retry/);
});

test('Installing a team plans a setup check, a replace when another team is installed, or nothing', () => {
  assert.deepEqual(installPlan('b', []), { kind: 'setup' });
  assert.deepEqual(installPlan('b', [{ Team: 'a' }]), { kind: 'replace', installed: ['a'] });
  assert.deepEqual(installPlan('a', [{ Team: 'a' }]), { kind: 'skip' });
});

test('The replace message names both teams', () => {
  const msg = replaceMessage('platform', ['data']);
  assert.match(msg, /data/);
  assert.match(msg, /platform/);
});
