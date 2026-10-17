import test from 'node:test';
import assert from 'node:assert/strict';

// Wails runs each bound call on its own goroutine, so keystrokes sent without
// waiting can reach the shell out of order (`commadn` for `command`).
function fakeApp(log, delays) {
  const call = (name) => (...args) =>
    new Promise((resolve, reject) => {
      const key = args[args.length - 1];
      log.push(`start ${name} ${key}`);
      setTimeout(() => {
        log.push(`end ${name} ${key}`);
        if (key === 'boom') reject(new Error('boom'));
        else resolve();
      }, delays[key] ?? 1);
    });
  return { ShellWrite: call('shell'), RunWrite: call('run') };
}

test('Keystrokes reach Go one call at a time, in the order they were typed', async () => {
  const log = [];
  globalThis.window = { go: { main: { App: fakeApp(log, { c: 30, o: 1, m: 10 }) } } };
  const api = await import('./api.js');

  await Promise.all(['c', 'o', 'm'].map((k) => api.shellWrite(k)));

  assert.deepEqual(log, ['start shell c', 'end shell c', 'start shell o', 'end shell o', 'start shell m', 'end shell m']);
});

test('A failed keystroke call does not hold back the next one, and each target has its own line', async () => {
  const log = [];
  globalThis.window = { go: { main: { App: fakeApp(log, { boom: 20, x: 1, y: 1 }) } } };
  const api = await import('./api.js');

  const failed = api.runWrite(1, 'boom');
  const next = api.runWrite(1, 'x');
  const shell = api.shellWrite('y');
  await assert.rejects(failed, /boom/);
  await Promise.all([next, shell]);

  assert.ok(log.indexOf('start shell y') < log.indexOf('end run boom'), 'the shell does not wait for the run line');
  assert.ok(log.indexOf('start run x') > log.indexOf('end run boom'), 'x waits for boom to end');
  assert.ok(log.includes('end run x'));
});
