// Behaviour tests for the showcase recorder's own logic: the publish checks,
// caption cue times, and the error paths of the drivers and record.sh.
// Run: node --test scripts/showcase/
// Tests that need tmux, ffmpeg or Google Chrome skip when it is missing.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import {
  chmodSync, existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, symlinkSync, truncateSync, writeFileSync,
} from 'node:fs';
import { createServer } from 'node:net';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { inflateSync } from 'node:zlib';

const here = dirname(fileURLToPath(import.meta.url));
const has = (tool) => spawnSync('sh', ['-c', `command -v ${tool}`]).status === 0;
const hasChrome = existsSync('/Applications/Google Chrome.app');

function scratch(t) {
  const dir = mkdtempSync(join(tmpdir(), 'showcase-test-'));
  t.after(() => rmSync(dir, { recursive: true, force: true }));
  return dir;
}

function run(cmd, args, opts = {}) {
  const r = spawnSync(cmd, args, { encoding: 'utf8', timeout: 90_000, ...opts });
  return { status: r.status, stdout: r.stdout ?? '', stderr: r.stderr ?? '' };
}

// --- checks.sh --------------------------------------------------------------

// A fixed identity, independent of the machine running the test: the user name
// equals the GitHub owner, as on the maintainer's Mac.
function checksEnv(dir) {
  const gitconfig = join(dir, 'gitconfig');
  writeFileSync(gitconfig, '[user]\n\temail = dev@example.com\n');
  return {
    PATH: process.env.PATH,
    HOME: '/Users/demo-person',
    USER: 'mavogel',
    GIT_CONFIG_GLOBAL: gitconfig,
    GIT_CONFIG_NOSYSTEM: '1',
  };
}

function checks(dir, args) {
  return run(join(here, 'checks.sh'), args, { cwd: dir, env: checksEnv(dir) });
}

function file(dir, name, content) {
  const path = join(dir, name);
  writeFileSync(path, content);
  return path;
}

test('privacy: a home path fails naming the file', (t) => {
  const dir = scratch(t);
  const f = file(dir, 'cast.txt', 'installed platform-team into /Users/demo-person/.kiro\n');
  const r = checks(dir, ['privacy', f]);
  assert.equal(r.status, 1);
  assert.match(r.stderr, new RegExp(`showcase: privacy: home path in ${f}`));
});

test('privacy: the public clone address and brew tap pass although they hold the user name', (t) => {
  const dir = scratch(t);
  const f = file(dir, 'end.txt',
    'git clone https://github.com/mavogel/kirokit.git\nbrew install --cask mavogel/tap/kirokit\n');
  const r = checks(dir, ['privacy', f]);
  assert.equal(r.status, 0, r.stderr);
});

test('privacy: the user name anywhere else fails', (t) => {
  const dir = scratch(t);
  const f = file(dir, 'dock.txt',
    'git clone https://github.com/mavogel/kirokit.git\nsigned in as mavogel\n');
  const r = checks(dir, ['privacy', f]);
  assert.equal(r.status, 1);
  assert.match(r.stderr, new RegExp(`showcase: privacy: user name in ${f}`));
});

test('privacy: the git email fails, whatever its case', (t) => {
  const dir = scratch(t);
  const f = file(dir, 'dock.txt', 'Logged in: Dev@Example.com\n');
  const r = checks(dir, ['privacy', f]);
  assert.equal(r.status, 1);
  assert.match(r.stderr, new RegExp(`showcase: privacy: git email in ${f}`));
});

test('privacy: every file is checked, and the failing one is named', (t) => {
  const dir = scratch(t);
  const clean = file(dir, 'a.txt', 'kirokit install platform-team\n');
  const dirty = file(dir, 'b.txt', 'HOME=/Users/demo-person\n');
  const r = checks(dir, ['privacy', clean, dirty]);
  assert.equal(r.status, 1);
  assert.match(r.stderr, new RegExp(`home path in ${dirty}`));
});

test('privacy: a missing file or no file at all fails instead of passing', (t) => {
  const dir = scratch(t);
  const missing = join(dir, 'nope.txt');
  let r = checks(dir, ['privacy', missing]);
  assert.equal(r.status, 1);
  assert.match(r.stderr, new RegExp(`showcase: privacy: missing ${missing}`));
  r = checks(dir, ['privacy']);
  assert.equal(r.status, 1);
  assert.match(r.stderr, /showcase: privacy: no files to check/);
});

test('size: the GIF may be exactly 8 MiB and the MP4 exactly 4 MiB, one byte more fails', (t) => {
  const dir = scratch(t);
  const gif = join(dir, 's.gif');
  const mp4 = join(dir, 's.mp4');
  writeFileSync(gif, '');
  writeFileSync(mp4, '');
  truncateSync(gif, 8 * 1024 * 1024);
  truncateSync(mp4, 4 * 1024 * 1024);
  assert.equal(checks(dir, ['size', gif, mp4]).status, 0);

  truncateSync(gif, 8 * 1024 * 1024 + 1);
  let r = checks(dir, ['size', gif, mp4]);
  assert.equal(r.status, 1);
  assert.match(r.stderr, new RegExp(`showcase: size: ${gif} .*limit 8 MB`));

  truncateSync(gif, 1024);
  truncateSync(mp4, 4 * 1024 * 1024 + 1);
  r = checks(dir, ['size', gif, mp4]);
  assert.equal(r.status, 1);
  assert.match(r.stderr, new RegExp(`showcase: size: ${mp4} .*limit 4 MB`));
});

test('length: 30 and 60 s pass, just outside fails naming the file', { skip: !has('ffmpeg') }, (t) => {
  const dir = scratch(t);
  const clip = (secs) => {
    const path = join(dir, `${secs}.wav`);
    const r = run('ffmpeg', ['-v', 'error', '-y', '-f', 'lavfi', '-i', 'anullsrc=r=8000:cl=mono', '-t', String(secs), path]);
    assert.equal(r.status, 0, r.stderr);
    return path;
  };
  assert.equal(checks(dir, ['length', clip(30), clip(60)]).status, 0);
  for (const secs of [29.9, 60.1]) {
    const path = clip(secs);
    const r = checks(dir, ['length', path]);
    assert.equal(r.status, 1, `${secs} s passed`);
    assert.match(r.stderr, new RegExp(`showcase: length: ${path} is .* s \\(needs 30 to 60 s\\)`));
  }
});

test('length: a file that is not media fails with the check\'s message', { skip: !has('ffprobe') }, (t) => {
  const dir = scratch(t);
  const f = file(dir, 'notes.gif', 'not a video\n');
  const r = checks(dir, ['length', f]);
  assert.equal(r.status, 1);
  assert.match(r.stderr, new RegExp(`showcase: length: cannot read the length of ${f}`));
});

test('an unknown check prints the usage', (t) => {
  const dir = scratch(t);
  const r = checks(dir, ['speed', 'x']);
  assert.equal(r.status, 2);
  assert.match(r.stderr, /usage: checks.sh privacy\|size\|length/);
});

// --- caption.mjs ------------------------------------------------------------

const v3 = (events) => [
  JSON.stringify({ version: 3, term: { cols: 110, rows: 32 } }),
  ...events.map((e) => JSON.stringify(e)),
].join('\n');

// Intervals of 0.5, 10, 2 and 20 s: the two long waits are capped at agg's 3 s
// idle limit, so "installed" shows at 3.5 s and "Session ended" at 8.5 s of the
// rendered clip, which agg holds 2 s more (record.sh: --last-frame-duration 2).
const STORY = [
  [0.5, 'o', '$ '],
  [10, 'o', 'installed platform-team into /tmp/kirokit-demo/kiro\r\n'],
  [2, 'o', '$ kiro-cli chat --agent devops --v3\r\n'],
  [20, 'o', 'Session ended\r\n'],
];

// A slow clone: the teams show 34 s after Clone is pressed.
const MARKS = { welcome: 2, clone: 6, teams: 40, install: 42.5, end: 55 };

function captionDir(t, cast, marks = MARKS) {
  const dir = scratch(t);
  if (cast !== undefined) writeFileSync(join(dir, 'cli.cast'), cast);
  if (marks !== undefined) writeFileSync(join(dir, 'app-marks.json'), JSON.stringify(marks));
  return dir;
}

const caption = (dir) => run('node', [join(here, 'caption.mjs'), dir]);

function cues(dir) {
  return Object.fromEntries(readFileSync(join(dir, 'cues.env'), 'utf8').trim().split('\n')
    .map((l) => l.split('=')).map(([k, v]) => [k, Number(v)]));
}

// Width, height, colour type and the alpha of the top-left pixel of an 8-bit
// RGBA PNG (the first pixel of the first row is stored raw for every filter).
function png(path) {
  const buf = readFileSync(path);
  const width = buf.readUInt32BE(16);
  const height = buf.readUInt32BE(20);
  const colorType = buf[25];
  const idat = [];
  for (let off = 8; off < buf.length;) {
    const len = buf.readUInt32BE(off);
    if (buf.toString('ascii', off + 4, off + 8) === 'IDAT') idat.push(buf.subarray(off + 8, off + 8 + len));
    off += 12 + len;
  }
  const corner = colorType === 6 ? inflateSync(Buffer.concat(idat))[4] : 255;
  return { width, height, colorType, corner };
}

test('caption: a missing recording exits naming the file', (t) => {
  const dir = captionDir(t, undefined, undefined);
  const r = caption(dir);
  assert.equal(r.status, 1);
  assert.match(r.stderr, new RegExp(`caption: missing input ${join(dir, 'cli.cast')}`));
});

test('caption: a cast that never ends the session fails instead of guessing a length', (t) => {
  const dir = captionDir(t, v3(STORY.slice(0, 3)));
  const r = caption(dir);
  assert.equal(r.status, 1);
  assert.match(r.stderr, /caption: cli.cast never shows "Session ended"/);
});

test('caption: app marks without a step fail naming it', (t) => {
  const { teams, ...rest } = MARKS;
  const dir = captionDir(t, v3(STORY), rest);
  const r = caption(dir);
  assert.equal(r.status, 1);
  assert.match(r.stderr, /caption: app-marks.json has no "teams"/);
});

test('caption: a caption shorter than 2.5 s fails the run', (t) => {
  // "installed" at 0.5 s leaves caption 1 on screen for under 2.5 s.
  const dir = captionDir(t, v3([[0.5, 'o', 'installed platform-team into x\r\n'], [5, 'o', 'Session ended\r\n']]));
  const r = caption(dir);
  assert.equal(r.status, 1);
  assert.match(r.stderr, /caption: a caption would show only .* s \(at least 2.5 s\)/);
});

test('caption: cue times follow the rendered CLI clip and the clone wait is cut', { skip: !hasChrome }, (t) => {
  const dir = captionDir(t, v3(STORY));
  const r = caption(dir);
  assert.equal(r.status, 0, r.stderr);
  const c = cues(dir);

  // The idle limit and the 2 s last frame make the clip 10.5 s, not 34.5 s.
  assert.equal(c.CLI_DURATION, 10.5);
  // Caption 2 takes over after the install line, and both captions last 2.5 s.
  assert.ok(c.CAP2_AT > 3.5, `CAP2_AT ${c.CAP2_AT}`);
  assert.ok(c.CAP2_AT >= 2.5 && c.CLI_DURATION - c.CAP2_AT >= 2.5);

  // At most 1.5 s of the 34 s clone wait is kept; the teams-to-end part (15 s) is all kept.
  assert.equal(c.APP_A1, MARKS.clone + 1.5);
  assert.ok(c.APP_DURATION >= MARKS.end - MARKS.teams && c.APP_DURATION < 25, `APP_DURATION ${c.APP_DURATION}`);
  assert.ok(c.CAP4_AT >= 2.5 && c.APP_DURATION - c.CAP4_AT >= 2.5);

  for (let i = 1; i <= 4; i++) {
    const p = png(join(dir, `cap-${i}.png`));
    assert.deepEqual([p.width, p.height, p.colorType, p.corner], [1280, 96, 6, 0], `cap-${i}.png`);
  }
  const frame = png(join(dir, 'frame.png'));
  assert.deepEqual([frame.width, frame.height, frame.corner], [1180, 760, 255]);
  for (const card of ['title.png', 'end.png']) {
    const p = png(join(dir, card));
    assert.deepEqual([p.width, p.height], [1280, 896], card);
  }
  const words = readFileSync(join(dir, 'captions.txt'), 'utf8');
  assert.match(words, /Install your team's Kiro config/);
  assert.match(words, /brew install --cask mavogel\/tap\/kirokit-desktop/);
});

test('caption: an asciicast v2 (absolute times) gives the same cues as v3', { skip: !hasChrome }, (t) => {
  let at = 0;
  const v2 = [
    JSON.stringify({ version: 2, width: 110, height: 32 }),
    ...STORY.map(([dt, kind, data]) => JSON.stringify([(at += dt), kind, data])),
  ].join('\n');
  const a = captionDir(t, v3(STORY));
  const b = captionDir(t, v2);
  assert.equal(caption(a).status, 0);
  assert.equal(caption(b).status, 0);
  assert.equal(cues(b).CLI_DURATION, cues(a).CLI_DURATION);
  assert.equal(cues(b).CAP2_AT, cues(a).CAP2_AT);
});

// One pixel as rgba hex, read with ffmpeg so any PNG filter decodes.
function pixel(path, x, y) {
  const r = spawnSync('ffmpeg', ['-v', 'error', '-i', path, '-vf', `crop=1:1:${x}:${y},format=rgba`, '-f', 'rawvideo', '-'],
    { timeout: 30_000 });
  assert.equal(r.status, 0, String(r.stderr));
  return r.stdout.toString('hex');
}

// A reused, resized Chrome page once caught stale paint: the frame's top rows
// repeated over its bottom border, or a caption came out blank. Every render
// must be complete, so render several times.
test('caption: every render paints the whole frame and every caption', { skip: !hasChrome || !has('ffmpeg'), timeout: 120_000 }, (t) => {
  const dir = captionDir(t, v3(STORY));
  for (let run = 1; run <= 5; run++) {
    const r = caption(dir);
    assert.equal(r.status, 0, r.stderr);
    const frame = join(dir, 'frame.png');
    const edges = { top: [590, 0], bottom: [590, 759], left: [0, 380], right: [1179, 380] };
    for (const [edge, [x, y]] of Object.entries(edges)) {
      assert.equal(pixel(frame, x, y), '3a3450ff', `run ${run}: frame ${edge} border`);
    }
    assert.equal(pixel(frame, 590, 380), '00000000', `run ${run}: frame centre is clear`);
    assert.equal(pixel(frame, 0, 0), '13111fff', `run ${run}: frame corner is the background`);
    for (let i = 1; i <= 4; i++) {
      assert.equal(pixel(join(dir, `cap-${i}.png`), 640, 48).slice(6), 'ff', `run ${run}: cap-${i}.png centre is painted`);
    }
  }
});

// --- app-drive.mjs ----------------------------------------------------------

async function portFree(port) {
  return new Promise((resolve) => {
    const s = createServer().once('error', () => resolve(false))
      .listen(port, '127.0.0.1', () => s.close(() => resolve(true)));
  });
}

test('app-drive: no output folder prints the usage', () => {
  const r = run('node', [join(here, 'app-drive.mjs')]);
  assert.equal(r.status, 2);
  assert.match(r.stderr, /usage: node app-drive.mjs <outdir>/);
});

test('app-drive: without a dev server it stops at the welcome step', { skip: !hasChrome }, async (t) => {
  if (!(await portFree(34115))) return t.skip('port 34115 is in use');
  const dir = scratch(t);
  const r = run('node', [join(here, 'app-drive.mjs'), dir]);
  assert.equal(r.status, 1);
  assert.match(r.stderr, /app-drive: welcome: /);
  assert.equal(existsSync(join(dir, 'app.webm')), false);
});

// --- cli-drive.sh -----------------------------------------------------------

test('cli-drive: no session prints the usage', () => {
  const r = run(join(here, 'cli-drive.sh'), []);
  assert.notEqual(r.status, 0);
  assert.match(r.stderr, /usage: cli-drive.sh <tmux-socket> <session>/);
});

test('cli-drive: a shell without kirokit times out naming the awaited text', { skip: !has('tmux'), timeout: 60_000 }, (t) => {
  const sock = `kirokit-showcase-test-${process.pid}`;
  const dir = scratch(t);
  t.after(() => run('tmux', ['-L', sock, 'kill-server']));
  const started = run('tmux', ['-u', '-L', sock, '-f', '/dev/null', 'new-session', '-d', '-s', 'kk', '-x', '110', '-y', '32', '-c', dir,
    "env -i PATH=/usr/bin:/bin HOME=/nonexistent BASH_SILENCE_DEPRECATION_WARNING=1 PS1='$ ' /bin/bash --norc --noprofile"]);
  assert.equal(started.status, 0, started.stderr);
  const t0 = Date.now();
  const r = run(join(here, 'cli-drive.sh'), [sock, 'kk']);
  assert.equal(r.status, 1);
  assert.match(r.stderr, /timeout waiting for installed platform-team/);
  assert.ok(Date.now() - t0 < 45_000, 'the 30 s limit was not applied');
});

// --- record.sh preflight ----------------------------------------------------

const TOOLS = ['tmux', 'asciinema', 'agg', 'ffmpeg', 'ffprobe', 'node', 'npm', 'go', 'git', 'kiro-cli'];

// A PATH of no-op stand-ins, so preflight runs without recording anything.
function stubPath(t, without = []) {
  const bin = scratch(t);
  for (const tool of TOOLS.filter((x) => !without.includes(x))) {
    writeFileSync(join(bin, tool), '#!/bin/sh\nexit 0\n');
    chmodSync(join(bin, tool), 0o755);
  }
  symlinkSync('/usr/sbin/lsof', join(bin, 'lsof'));
  return `${bin}:/usr/bin:/bin`;
}

// record.sh always works in /tmp/kirokit-demo; never disturb a kept or running one.
const demoBusy = () => existsSync('/tmp/kirokit-demo');

function record(t, env, args = []) {
  return run(join(here, 'record.sh'), args, { env: { HOME: process.env.HOME, ...env } });
}

test('record.sh: an unknown flag prints the usage', () => {
  const r = run(join(here, 'record.sh'), ['--bogus']);
  assert.equal(r.status, 2);
  assert.match(r.stderr, /usage: scripts\/showcase\/record.sh \[--keep\]/);
});

test('record.sh: a missing tool fails preflight and leaves no work folder', { skip: demoBusy() }, (t) => {
  const r = record(t, { PATH: stubPath(t, ['agg']) });
  assert.equal(r.status, 1);
  assert.match(r.stderr, /showcase: preflight: agg not found/);
  assert.equal(existsSync('/tmp/kirokit-demo'), false);
});

test('record.sh: a missing age key fails preflight naming its path', { skip: demoBusy() || !hasChrome }, (t) => {
  const r = record(t, { PATH: stubPath(t), SOPS_AGE_KEY_FILE: '/nonexistent/keys.txt' });
  assert.equal(r.status, 1);
  assert.match(r.stderr, /showcase: preflight: SOPS age key not found at \/nonexistent\/keys.txt/);
});

test('record.sh: a second run leaves a live run\'s folder alone', { skip: demoBusy() || !hasChrome }, async (t) => {
  if (!(await portFree(34115)) || !(await portFree(5173))) return t.skip('a dev server port is in use');
  const dir = scratch(t);
  const key = file(dir, 'keys.txt', 'not a key\n');
  // This test process stands in for the live run.
  mkdirSync('/tmp/kirokit-demo/out', { recursive: true });
  t.after(() => rmSync('/tmp/kirokit-demo', { recursive: true, force: true }));
  writeFileSync('/tmp/kirokit-demo/pid', `${process.pid}\n`);
  writeFileSync('/tmp/kirokit-demo/out/clip.txt', 'in progress\n');
  const r = record(t, { PATH: stubPath(t), SOPS_AGE_KEY_FILE: key });
  assert.equal(r.status, 1);
  assert.match(r.stderr, new RegExp(`showcase: preflight: another showcase run \\(pid ${process.pid}\\) is using /tmp/kirokit-demo`));
  assert.equal(readFileSync('/tmp/kirokit-demo/out/clip.txt', 'utf8'), 'in progress\n');
});

test('record.sh: a busy dev-server port fails preflight', { skip: demoBusy() || !hasChrome }, async (t) => {
  if (!(await portFree(34115))) return t.skip('port 34115 is in use');
  const dir = scratch(t);
  const key = file(dir, 'keys.txt', 'not a key\n');
  const server = createServer();
  await new Promise((resolve) => server.listen(34115, '127.0.0.1', resolve));
  t.after(() => server.close());
  const r = await new Promise((resolve) => {
    // spawnSync would block the event loop that serves the port; run it async.
    import('node:child_process').then(({ execFile }) => execFile(join(here, 'record.sh'), [], {
      env: { HOME: process.env.HOME, PATH: stubPath(t), SOPS_AGE_KEY_FILE: key }, encoding: 'utf8',
    }, (err, stdout, stderr) => resolve({ status: err ? err.code : 0, stderr })));
  });
  assert.equal(r.status, 1);
  assert.match(r.stderr, /showcase: preflight: port 34115 is busy: close the running wails dev first/);
  assert.equal(existsSync('/tmp/kirokit-demo'), false);
});
