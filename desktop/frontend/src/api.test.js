import test from 'node:test';
import assert from 'node:assert/strict';

// api.js reaches Go through @wailsio/runtime; globalThis.__kirokitRuntime
// replaces that runtime in these tests. Calls are { name, args } entries.
function fakeRuntime({ byName = () => Promise.resolve(), calls = [], emitted = [], opened = [], listeners = {} } = {}) {
  return {
    Call: {
      ByName: (name, ...args) => {
        calls.push({ name, args });
        return byName(name, ...args);
      },
    },
    Events: {
      On: (event, cb) => {
        (listeners[event] ||= []).push(cb);
        return () => {
          listeners[event] = listeners[event].filter((c) => c !== cb);
        };
      },
      Emit: (name, data) => {
        emitted.push(data === undefined ? { name } : { name, data });
        return Promise.resolve();
      },
    },
    Browser: { OpenURL: (url) => opened.push(url) },
  };
}

const api = await import('./api.js');

// Wails runs each bound call on its own goroutine, so keystrokes sent without
// waiting can reach the shell out of order (`commadn` for `command`).
function delayed(log, delays) {
  return (name, ...args) =>
    new Promise((resolve, reject) => {
      const key = args[args.length - 1];
      const who = name === 'main.App.ShellWrite' ? 'shell' : 'run';
      log.push(`start ${who} ${key}`);
      setTimeout(() => {
        log.push(`end ${who} ${key}`);
        if (key === 'boom') reject(new Error('boom'));
        else resolve();
      }, delays[key] ?? 1);
    });
}

test('Keystrokes reach Go one call at a time, in the order they were typed', async () => {
  const log = [];
  globalThis.__kirokitRuntime = fakeRuntime({ byName: delayed(log, { c: 30, o: 1, m: 10 }) });

  await Promise.all(['c', 'o', 'm'].map((k) => api.shellWrite(k)));

  assert.deepEqual(log, ['start shell c', 'end shell c', 'start shell o', 'end shell o', 'start shell m', 'end shell m']);
});

test('A failed keystroke call does not hold back the next one, and each target has its own line', async () => {
  const log = [];
  globalThis.__kirokitRuntime = fakeRuntime({ byName: delayed(log, { boom: 20, x: 1, y: 1 }) });

  const failed = api.runWrite(1, 'boom');
  const next = api.runWrite(1, 'x');
  const shell = api.shellWrite('y');
  await assert.rejects(failed, /boom/);
  await Promise.all([next, shell]);

  assert.ok(log.indexOf('start shell y') < log.indexOf('end run boom'), 'the shell does not wait for the run line');
  assert.ok(log.indexOf('start run x') > log.indexOf('end run boom'), 'x waits for boom to end');
  assert.ok(log.includes('end run x'));
});

// Wails v3 addresses a bound method as <Go package>.<Type>.<Method>; App is in
// package main, so every call is main.App.<Method>.
test('Every call reaches main.App.<Method> with its arguments', async () => {
  const calls = [];
  globalThis.__kirokitRuntime = fakeRuntime({ calls });

  await api.state();
  await api.setRepoDir('/r');
  await api.saveSettings({ ref: 'v1' });
  await api.catalog();
  await api.installs();
  await api.prerequisites('team-a');
  await api.run(['sync']);
  await api.quickDir('team-a');
  await api.quickBundle('team-a');
  await api.openFolder('team-a');
  await api.clone('https://example.test/r.git');
  await api.checkout();
  await api.switchRef('v2');
  await api.runWrite(7, 'k');
  await api.shellWrite('s');
  await api.shellResize(80, 24);

  assert.deepEqual(calls, [
    { name: 'main.App.State', args: [] },
    { name: 'main.App.SetRepoDir', args: ['/r'] },
    { name: 'main.App.SaveSettings', args: [{ ref: 'v1' }] },
    { name: 'main.App.Catalog', args: [] },
    { name: 'main.App.Installs', args: [] },
    { name: 'main.App.Prerequisites', args: ['team-a'] },
    { name: 'main.App.Run', args: [['sync']] },
    { name: 'main.App.QuickDir', args: ['team-a'] },
    { name: 'main.App.QuickBundle', args: ['team-a'] },
    { name: 'main.App.OpenFolder', args: ['team-a'] },
    { name: 'main.App.Clone', args: ['https://example.test/r.git'] },
    { name: 'main.App.Checkout', args: [] },
    { name: 'main.App.SwitchRef', args: ['v2'] },
    { name: 'main.App.RunWrite', args: [7, 'k'] },
    { name: 'main.App.ShellWrite', args: ['s'] },
    { name: 'main.App.ShellResize', args: [80, 24] },
  ]);
});

// A Go error rejects as Wails' RuntimeError carrying err.Error(); the pages
// show its message (and strip "RepoDir: ") themselves.
test('A Go error reaches the page with its own message', async () => {
  const goErr = Object.assign(new Error('RepoDir: the folder is not empty'), { name: 'RuntimeError' });
  globalThis.__kirokitRuntime = fakeRuntime({ byName: () => Promise.reject(goErr) });

  await assert.rejects(api.setRepoDir('/x'), { name: 'RuntimeError', message: 'RepoDir: the folder is not empty' });
});

// v3 hands listeners an event object; the pages keep receiving the payload.
test('on hands the handler the event payload and returns the unsubscribe function', () => {
  const listeners = {};
  globalThis.__kirokitRuntime = fakeRuntime({ listeners });
  const got = [];

  const off = api.on('term:kirokit:exit', (payload) => got.push(payload));
  listeners['term:kirokit:exit'].forEach((cb) => cb({ name: 'term:kirokit:exit', data: { id: 3, code: 0 } }));
  off();
  (listeners['term:kirokit:exit'] || []).forEach((cb) => cb({ name: 'term:kirokit:exit', data: { id: 4, code: 1 } }));

  assert.deepEqual(got, [{ id: 3, code: 0 }]);
});

// The smoke start (KIROKIT_DESKTOP_SMOKE=1) quits only on smoke:ready, which
// proves the page loaded and a bound call answered.
test('start sends smoke:ready only after State has answered, and resolves to the state', async () => {
  const emitted = [];
  let answer;
  const pending = new Promise((resolve) => (answer = resolve));
  globalThis.__kirokitRuntime = fakeRuntime({ emitted, byName: () => pending });

  const started = api.start();
  await new Promise((r) => setTimeout(r, 10));
  assert.deepEqual(emitted, [], 'nothing is sent before State answers');

  answer({ cloned: true });
  assert.deepEqual(await started, { cloned: true });
  assert.deepEqual(emitted, [{ name: 'smoke:ready' }]);
});

test('start sends nothing when State fails, and rejects with its error', async () => {
  const emitted = [];
  globalThis.__kirokitRuntime = fakeRuntime({ emitted, byName: () => Promise.reject(new Error('no state')) });

  await assert.rejects(api.start(), /no state/);
  assert.deepEqual(emitted, []);
});

// The app opens the system browser only for a GitHub new-issue form, since the
// address is built from a cloned repository's manifest.
test('openURL opens a GitHub new-issue form and refuses every other address', async () => {
  const opened = [];
  globalThis.__kirokitRuntime = fakeRuntime({ opened });

  const ok = 'https://github.com/acme/kiro.config_1/issues/new?title=x&body=y';
  await api.openURL(ok);
  assert.deepEqual(opened, [ok]);

  for (const bad of [
    'javascript:alert(1)',
    'https://evil.example/issues/new?x',
    'https://github.com/a/b/pulls?x',
    'https://github.com/a/b/issues/new',
    'https://github.com.evil.example/a/b/issues/new?x',
    'https://github.com@evil.example/a/b/issues/new?x',
    'http://github.com/a/b/issues/new?x',
    'https://github.com/a/../issues/new?x',
    ' https://github.com/a/b/issues/new?x',
    undefined,
  ]) {
    await assert.rejects(api.openURL(bad), new RegExp(`^Error: refusing to open ${String(bad).replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}$`), String(bad));
  }
  assert.equal(opened.length, 1, 'no refused address reached the browser');
});

// Without a runtime a call or a valid address fails with a message the page
// can show, instead of a TypeError.
test('Without the runtime, calls and openURL reject with a readable message', async () => {
  globalThis.__kirokitRuntime = null;

  await assert.rejects(api.openURL('https://github.com/acme/cfg/issues/new?title=x'), /^Error: kirokit is not running inside its window$/);
  await assert.rejects(api.state(), /^Error: kirokit is not running inside its window$/);
  assert.equal(typeof api.on('x', () => {}), 'function');
});
