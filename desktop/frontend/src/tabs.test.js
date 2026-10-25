import test from 'node:test';
import assert from 'node:assert/strict';
import { createDock, MIN_HEIGHT, MAX_HEIGHT, TABS, shellLabel } from './tabs.js';

test('the shell tab is named after the shell the app reports, zsh until it does', () => {
  assert.equal(shellLabel({ env: { shell: 'PowerShell' } }), 'PowerShell');
  assert.equal(shellLabel({ env: { shell: 'bash' } }), 'bash');
  assert.equal(shellLabel({ env: { shell: '' } }), 'zsh');
  assert.equal(shellLabel({ env: {} }), 'zsh');
  assert.equal(shellLabel(undefined), 'zsh');
});

test('the dock starts closed on the kirokit tab', () => {
  const dock = createDock();
  assert.equal(dock.open, false);
  assert.equal(dock.tab, 'kirokit');
  assert.deepEqual(TABS, ['kirokit', 'shell']);
});

test('open, close and toggle change only the visibility', () => {
  const dock = createDock();
  dock.show();
  assert.equal(dock.open, true);
  dock.hide();
  assert.equal(dock.open, false);
  dock.toggle();
  assert.equal(dock.open, true);
  dock.toggle();
  assert.equal(dock.open, false);
});

test('height stays within the minimum and maximum', () => {
  const dock = createDock();
  dock.resize(10);
  assert.equal(dock.height, MIN_HEIGHT);
  dock.resize(5000);
  assert.equal(dock.height, MAX_HEIGHT);
  dock.resize(300.4);
  assert.equal(dock.height, 300);
  dock.resize(Number.NaN);
  assert.equal(dock.height, 300);
});

test('closing and reopening keeps the height and the tab', () => {
  const dock = createDock();
  dock.show();
  dock.select('shell');
  dock.resize(333);
  dock.hide();
  dock.show();
  assert.equal(dock.tab, 'shell');
  assert.equal(dock.height, 333);
});

test('an unknown tab is ignored', () => {
  const dock = createDock();
  dock.select('nope');
  assert.equal(dock.tab, 'kirokit');
});

test('selecting a tab shows the dock and listeners see every change', () => {
  const dock = createDock();
  const seen = [];
  const off = dock.subscribe((s) => seen.push({ ...s }));
  dock.select('shell');
  assert.deepEqual(seen.at(-1), { open: true, tab: 'shell', height: dock.height });
  off();
  dock.resize(200);
  assert.equal(seen.length, 1);
});
