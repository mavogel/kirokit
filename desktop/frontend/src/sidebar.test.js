import test from 'node:test';
import assert from 'node:assert/strict';
import { createSidebar, PAGES } from './sidebar.js';

const key = (init) => ({ key: 'b', metaKey: false, ctrlKey: false, preventDefault() { this.prevented = true; }, ...init });

test('toggle collapses to the rail and expands again', () => {
  const sb = createSidebar();
  assert.equal(sb.collapsed, false);
  sb.toggle();
  assert.equal(sb.collapsed, true);
  sb.toggle();
  assert.equal(sb.collapsed, false);
});

test('Command-B toggles and is consumed', () => {
  const sb = createSidebar();
  const e = key({ metaKey: true });
  assert.equal(sb.handleKey(e), true);
  assert.equal(e.prevented, true);
  assert.equal(sb.collapsed, true);
  sb.handleKey(key({ metaKey: true, key: 'B' }));
  assert.equal(sb.collapsed, false);
});

test('plain B and Control-B do nothing', () => {
  const sb = createSidebar();
  for (const init of [{}, { ctrlKey: true }]) {
    const e = key(init);
    assert.equal(sb.handleKey(e), false);
    assert.equal(e.prevented, undefined);
  }
  assert.equal(sb.collapsed, false);
});

test('Command-B is ignored while the sidebar is not showing', () => {
  let hidden = true;
  const sb = createSidebar({ isHidden: () => hidden });
  const e = key({ metaKey: true });
  assert.equal(sb.handleKey(e), false);
  assert.equal(sb.collapsed, false);
  hidden = false;
  assert.equal(sb.handleKey(key({ metaKey: true })), true);
  assert.equal(sb.collapsed, true);
});

test('select switches among the five pages and ignores unknown ones', () => {
  const sb = createSidebar();
  assert.deepEqual(PAGES, ['teams', 'agents', 'install', 'skills', 'settings']);
  assert.equal(sb.page, 'teams');
  for (const p of PAGES) {
    sb.select(p);
    assert.equal(sb.page, p);
  }
  sb.select('nope');
  assert.equal(sb.page, 'settings');
});

test('subscribers hear every change', () => {
  const sb = createSidebar();
  const seen = [];
  sb.subscribe((s) => seen.push([s.collapsed, s.page]));
  sb.toggle();
  sb.select('agents');
  assert.deepEqual(seen, [[true, 'teams'], [true, 'agents']]);
});
