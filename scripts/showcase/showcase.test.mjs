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
const hasKiroIDE = existsSync('/Applications/Kiro.app');

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

// The IDE and kirocrew scenes write their sampled text to files of their own.
test('privacy: the Kiro IDE and kirocrew text files are checked like the others', (t) => {
  const dir = scratch(t);
  const ide = file(dir, 'ide-text.txt', 'GLOBAL\nproduct-owner\n');
  const crew = file(dir, 'crew-text.txt', 'Project: /Users/demo-person/work\n');
  assert.equal(checks(dir, ['privacy', ide]).status, 0);
  const r = checks(dir, ['privacy', ide, crew]);
  assert.equal(r.status, 1);
  assert.match(r.stderr, new RegExp(`showcase: privacy: home path in ${crew}`));
  const signedIn = file(dir, 'ide-text2.txt', 'Signed in as dev@example.com\n');
  assert.match(checks(dir, ['privacy', signedIn]).stderr, /git email in/);
});

test('size: the GIF may be exactly 10 MiB and the MP4 exactly 6 MiB, one byte more fails', (t) => {
  const dir = scratch(t);
  const gif = join(dir, 's.gif');
  const mp4 = join(dir, 's.mp4');
  writeFileSync(gif, '');
  writeFileSync(mp4, '');
  truncateSync(gif, 10 * 1024 * 1024);
  truncateSync(mp4, 6 * 1024 * 1024);
  assert.equal(checks(dir, ['size', gif, mp4]).status, 0);

  truncateSync(gif, 10 * 1024 * 1024 + 1);
  let r = checks(dir, ['size', gif, mp4]);
  assert.equal(r.status, 1);
  assert.match(r.stderr, new RegExp(`showcase: size: ${gif} .*limit 10 MB`));

  truncateSync(gif, 9 * 1024 * 1024);
  truncateSync(mp4, 6 * 1024 * 1024 + 1);
  r = checks(dir, ['size', gif, mp4]);
  assert.equal(r.status, 1);
  assert.match(r.stderr, new RegExp(`showcase: size: ${mp4} .*limit 6 MB`));
});

test('length: 30, 75 and 90 s pass, just outside fails naming the file', { skip: !has('ffmpeg') }, (t) => {
  const dir = scratch(t);
  const clip = (secs) => {
    const path = join(dir, `${secs}.wav`);
    const r = run('ffmpeg', ['-v', 'error', '-y', '-f', 'lavfi', '-i', 'anullsrc=r=8000:cl=mono', '-t', String(secs), path]);
    assert.equal(r.status, 0, r.stderr);
    return path;
  };
  assert.equal(checks(dir, ['length', clip(30), clip(75), clip(90)]).status, 0);
  for (const secs of [29.9, 90.1, 95]) {
    const path = clip(secs);
    const r = checks(dir, ['length', path]);
    assert.equal(r.status, 1, `${secs} s passed`);
    assert.match(r.stderr, new RegExp(`showcase: length: ${path} is .* s \\(needs 30 to 90 s\\)`));
  }
});

test('length: a file that is not media fails with the check\'s message', { skip: !has('ffprobe') }, (t) => {
  const dir = scratch(t);
  const f = file(dir, 'notes.gif', 'not a video\n');
  const r = checks(dir, ['length', f]);
  assert.equal(r.status, 1);
  assert.match(r.stderr, new RegExp(`showcase: length: cannot read the length of ${f}`));
});

// ffprobe reads a still image with exit 0 but prints "N/A" for its length: an
// exit-status-only guard would hand "N/A" to awk and report "is N/A s".
test('length: a still image without a length fails with the check\'s message', { skip: !has('ffmpeg') }, (t) => {
  const dir = scratch(t);
  const still = join(dir, 'still.png');
  const r0 = run('ffmpeg', ['-v', 'error', '-y', '-f', 'lavfi', '-i', 'color=c=red:s=16x16', '-frames:v', '1', still]);
  assert.equal(r0.status, 0, r0.stderr);
  const r = checks(dir, ['length', still]);
  assert.equal(r.status, 1);
  assert.match(r.stderr, new RegExp(`showcase: length: cannot read the length of ${still}`));
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

// Intervals of 0.5, 10, 2, 20, 4 and 2 s: the long waits are capped at agg's
// 3 s idle limit, so "installed" shows at 3.5 s, "Session ended" at 8.5 s, the
// Quick render at 11.5 s and the bundle listing at 13.5 s of the rendered clip,
// which agg holds 2 s more (record.sh: --last-frame-duration 2).
const STORY = [
  [0.5, 'o', '$ '],
  [10, 'o', 'installed product-team into /tmp/kirokit-demo/kiro\r\n'],
  [2, 'o', '$ kiro-cli chat --agent product-owner --v3\r\n'],
  [20, 'o', 'Session ended\r\n'],
  [4, 'o', 'rendered Quick bundle for product-team: 1 skills, 1 agents -> quick\r\n'],
  [2, 'o', 'quick/skills/user-story:\r\nSKILL.md\r\n'],
];

// A slow clone: the teams show 34 s after Clone is pressed.
const MARKS = { welcome: 2, clone: 6, teams: 40, install: 42.5, end: 55 };
// The IDE window is up at 1.2 s; its picker opens at 4.3 s. kirocrew's
// dashboard is past its first-run dialogs at 2.7 s.
const IDE_MARKS = { opened: 1.2, picker: 4.3, agents: 4.4, end: 7.4 };
const CREW_MARKS = { opened: 2.7, picker: 3.9, agents: 4.3, end: 7.3 };

function captionDir(t, cast, marks = MARKS, ide = IDE_MARKS, crew = CREW_MARKS) {
  const dir = scratch(t);
  if (cast !== undefined) writeFileSync(join(dir, 'cli.cast'), cast);
  if (marks !== undefined) writeFileSync(join(dir, 'app-marks.json'), JSON.stringify(marks));
  if (ide !== undefined) writeFileSync(join(dir, 'ide-marks.json'), JSON.stringify(ide));
  if (crew !== undefined) writeFileSync(join(dir, 'crew-marks.json'), JSON.stringify(crew));
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

test('caption: a cast without the Quick render fails naming it', (t) => {
  const dir = captionDir(t, v3(STORY.slice(0, 4)));
  const r = caption(dir);
  assert.equal(r.status, 1);
  assert.match(r.stderr, /caption: cli.cast never shows "rendered Quick bundle for product-team"/);
});

test('caption: IDE or kirocrew marks without a step fail naming the file and step', (t) => {
  const { picker, ...noPicker } = IDE_MARKS;
  let r = caption(captionDir(t, v3(STORY), MARKS, noPicker));
  assert.equal(r.status, 1);
  assert.match(r.stderr, /caption: ide-marks.json has no "picker"/);
  const noCrew = captionDir(t, v3(STORY));
  rmSync(join(noCrew, 'crew-marks.json'));
  r = caption(noCrew);
  assert.equal(r.status, 1);
  assert.match(r.stderr, /caption: missing input .*crew-marks.json/);
});

test('caption: a caption shorter than 2.5 s fails the run', (t) => {
  // "installed" at 0.5 s leaves caption 1 on screen for under 2.5 s.
  const short = [[0.5, 'o', 'installed product-team into x\r\n'], [5, 'o', 'Session ended\r\n'], [5, 'o', 'rendered Quick bundle for product-team\r\n']];
  let r = caption(captionDir(t, v3(short)));
  assert.equal(r.status, 1);
  assert.match(r.stderr, /caption: a caption would show only .* s \(at least 2.5 s\)/);
  // A kirocrew clip that ends 2 s after its dashboard opened is too short for its caption.
  r = caption(captionDir(t, v3(STORY), MARKS, IDE_MARKS, { opened: 3, picker: 3.5, agents: 4, end: 5 }));
  assert.equal(r.status, 1);
  assert.match(r.stderr, /caption: a caption would show only 2.00 s/);
});

test('caption: cue times follow the rendered CLI clip, the IDE and kirocrew marks, and the clone wait is cut', { skip: !hasChrome }, (t) => {
  const dir = captionDir(t, v3(STORY));
  const r = caption(dir);
  assert.equal(r.status, 0, r.stderr);
  const c = cues(dir);

  // The idle limit and the 2 s last frame make the clip 15.5 s, not 40.5 s.
  assert.equal(c.CLI_DURATION, 15.5);
  // Caption 2 takes over after the install line, caption 3 once the session
  // ended, before the Quick render shows; each lasts at least 2.5 s.
  assert.ok(c.CAP2_AT > 3.5 && c.CAP3_AT > 8.5 && c.CAP3_AT < 11.5, `CAP2_AT ${c.CAP2_AT} CAP3_AT ${c.CAP3_AT}`);
  assert.ok(c.CAP2_AT >= 2.5 && c.CAP3_AT - c.CAP2_AT >= 2.5 && c.CLI_DURATION - c.CAP3_AT >= 2.5);

  // The IDE clip starts 2.5 s before its picker opens (after the window is up);
  // kirocrew's starts once its dialogs are gone. Both keep the list to the end.
  assert.deepEqual([c.IDE_T0, c.IDE_T1, c.IDE_DURATION], [1.8, 7.4, 5.6]);
  assert.deepEqual([c.CREW_T0, c.CREW_T1, c.CREW_DURATION], [2.7, 7.3, 4.6]);

  // At most 1.5 s of the 34 s clone wait is kept; the teams-to-end part (15 s) is all kept.
  assert.equal(c.APP_A1, MARKS.clone + 1.5);
  assert.ok(c.APP_DURATION >= MARKS.end - MARKS.teams && c.APP_DURATION < 25, `APP_DURATION ${c.APP_DURATION}`);
  assert.ok(c.CAP7_AT >= 2.5 && c.APP_DURATION - c.CAP7_AT >= 2.5);

  for (let i = 1; i <= 7; i++) {
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
  assert.match(words, /Or export it for Amazon Quick\nKiro IDE lists the same agents\nSo does kirocrew\n/);
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
    for (let i = 1; i <= 7; i++) {
      assert.equal(pixel(join(dir, `cap-${i}.png`), 640, 48).slice(6), 'ff', `run ${run}: cap-${i}.png centre is painted`);
    }
  }
});

// --- app-drive.mjs ----------------------------------------------------------

// The same test record.sh's preflight uses, so a listener on any address
// counts as busy.
async function portFree(port) {
  return spawnSync('lsof', ['-nP', `-iTCP:${port}`, '-sTCP:LISTEN']).status !== 0;
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

// --- app-server.sh ----------------------------------------------------------

// The server build answers every Go method the page can call, the shell tab's
// included, so it must only ever listen on loopback.
const appServer = join(here, 'app-server.sh');
const canBuildApp = has('go') && has('npm') && has('curl');

// Runs app-server.sh without blocking the event loop (a test server may need it).
function startAppServer(t, env = {}) {
  const out = scratch(t);
  const home = scratch(t);
  return new Promise((resolve) => {
    import('node:child_process').then(({ execFile }) => execFile(appServer, [out, home], {
      env: { ...process.env, ...env }, encoding: 'utf8', timeout: 300_000,
    }, (err, stdout, stderr) => resolve({ status: err ? err.code : 0, stdout: stdout.trim(), stderr, out })));
  });
}

// Waits up to 10 s for nothing to listen on 34115.
function waitPortClosed() {
  for (let i = 0; i < 40 && !spawnSync('lsof', ['-nP', '-iTCP:34115', '-sTCP:LISTEN']).status; i++) {
    spawnSync('sleep', ['0.25']);
  }
}

// Stops the server's process group; pid must be the number app-server.sh printed.
function stopAppServer(pid) {
  if (!/^[1-9]\d*$/.test(pid)) return;
  spawnSync('sh', ['-c', `kill -TERM -- -${pid} 2>/dev/null || kill -TERM ${pid} 2>/dev/null; true`]);
  waitPortClosed();
}

const listeners = () => spawnSync('lsof', ['-nP', '-iTCP:34115', '-sTCP:LISTEN', '-Fn'], { encoding: 'utf8' })
  .stdout.split('\n').filter((l) => l.startsWith('n')).map((l) => l.slice(1));

test('app-server: missing arguments print the usage', () => {
  const r = run(appServer, []);
  assert.equal(r.status, 2);
  assert.match(r.stderr, /usage: scripts\/showcase\/app-server.sh <out-dir> <home>/);
});

test('app-server: an inherited WAILS_SERVER_HOST=0.0.0.0 still serves on loopback only', { skip: !canBuildApp, timeout: 320_000 }, async (t) => {
  if (!(await portFree(34115))) return t.skip('port 34115 is in use');
  const r = await startAppServer(t, { WAILS_SERVER_HOST: '0.0.0.0', WAILS_SERVER_PORT: '34999' });
  assert.equal(r.status, 0, r.stderr);
  t.after(() => stopAppServer(r.stdout));
  assert.match(r.stdout, /^\d+$/, 'prints the server pid');
  assert.deepEqual(listeners(), ['127.0.0.1:34115']);
  const page = spawnSync('curl', ['-fsS', 'http://localhost:34115/'], { encoding: 'utf8' });
  assert.equal(page.status, 0);
  assert.match(page.stdout, /<title>kirokit<\/title>/);
});

test('app-server: a server that cannot start fails naming its log', { skip: !canBuildApp, timeout: 320_000 }, async (t) => {
  if (!(await portFree(34115))) return t.skip('port 34115 is in use');
  // Something else holds the port and does not answer like the app.
  const blocker = (await import('node:http')).createServer((req, res) => { res.statusCode = 500; res.end(); });
  await new Promise((resolve) => blocker.listen(34115, '127.0.0.1', resolve));
  t.after(() => blocker.close());
  const r = await startAppServer(t);
  assert.equal(r.status, 1);
  assert.match(r.stderr, /showcase: app-server: the app server exited; see .*app-server\.log/);
});

test('app-server: a listener beyond loopback stops the run and the server', { skip: !canBuildApp, timeout: 320_000 }, async (t) => {
  if (!(await portFree(34115))) return t.skip('port 34115 is in use');
  // lsof reports the server's own listener on every interface.
  const bin = scratch(t);
  writeFileSync(join(bin, 'lsof'), '#!/bin/sh\n/usr/sbin/lsof "$@" | sed "s/^n127\\.0\\.0\\.1:/n*:/"\n');
  chmodSync(join(bin, 'lsof'), 0o755);
  const r = await startAppServer(t, { PATH: `${bin}:${process.env.PATH}` });
  assert.equal(r.status, 1);
  assert.match(r.stderr, /showcase: app-server: the app server listens on \*:34115, not only on 127\.0\.0\.1/);
  waitPortClosed();
  assert.deepEqual(listeners(), [], 'the server was stopped');
});

// --- ide-drive.mjs ----------------------------------------------------------

test('ide-drive: missing arguments print the usage', () => {
  const r = run('node', [join(here, 'ide-drive.mjs'), '/tmp/out', '/tmp/kiro']);
  assert.equal(r.status, 2);
  assert.match(r.stderr, /usage: node ide-drive.mjs <outdir> <kiro-home> <profile-dir> <workspace> <agent>\.\.\./);
});

test('ide-drive: without Kiro IDE it stops before starting anything', (t) => {
  const dir = scratch(t);
  const r = run('node', [join(here, 'ide-drive.mjs'), dir, join(dir, 'kiro'), join(dir, 'ide'), join(dir, 'ws'), 'qa'],
    { env: { ...process.env, KIROKIT_IDE_APP: join(dir, 'no-such-Kiro') } });
  assert.equal(r.status, 1);
  assert.match(r.stderr, new RegExp(`ide-drive: Kiro IDE not found at ${join(dir, 'no-such-Kiro')}`));
  assert.equal(existsSync(join(dir, 'ide.webm')), false);
});

// A Kiro that starts but never opens a window must not hang the recording.
test('ide-drive: a Kiro that never starts gives up at the deadline and is stopped', (t) => {
  const dir = scratch(t);
  const pidFile = join(dir, 'kiro.pid');
  const fake = file(dir, 'Kiro', `#!/bin/sh\necho $$ > '${pidFile}'\nexec sleep 300\n`);
  chmodSync(fake, 0o755);
  const t0 = Date.now();
  const r = run('node', [join(here, 'ide-drive.mjs'), dir, join(dir, 'kiro'), join(dir, 'ide'), join(dir, 'ws'), 'qa'],
    { env: { ...process.env, KIROKIT_IDE_APP: fake, KIROKIT_IDE_DEADLINE: '4' } });
  assert.equal(r.status, 1);
  assert.match(r.stderr, /ide-drive: launch: gave up after 4 s/);
  assert.ok(Date.now() - t0 < 20_000, 'the deadline was not applied');
  let alive = true;
  for (let i = 0; i < 20 && alive; i++) {
    try {
      process.kill(Number(readFileSync(pidFile, 'utf8')), 0);
      spawnSync('sleep', ['0.1']);
    } catch {
      alive = false;
    }
  }
  assert.equal(alive, false, 'the fake Kiro still runs');
});

// --- crew-drive.mjs ---------------------------------------------------------

const crewArgs = (dir) => [join(here, 'crew-drive.mjs'), dir, join(dir, 'kiro'), join(dir, 'crew-home'), 'qa'];

test('crew-drive: missing arguments print the usage', () => {
  const r = run('node', [join(here, 'crew-drive.mjs'), '/tmp/out']);
  assert.equal(r.status, 2);
  assert.match(r.stderr, /usage: node crew-drive.mjs <outdir> <kiro-home> <crew-home> <agent>\.\.\./);
});

test('crew-drive: without kirocrew it fails at the gateway step', (t) => {
  const dir = scratch(t);
  const r = run('node', crewArgs(dir), { env: { ...process.env, KIROKIT_CREW_BIN: join(dir, 'no-kirocrew') } });
  assert.equal(r.status, 1);
  assert.match(r.stderr, new RegExp(`crew-drive: gateway: ${join(dir, 'no-kirocrew')} not found`));
});

test('crew-drive: a gateway that exits before it is ready fails naming that', (t) => {
  const dir = scratch(t);
  const bin = file(dir, 'kirocrew', '#!/bin/sh\necho "starting"\nexit 3\n');
  chmodSync(bin, 0o755);
  const r = run('node', crewArgs(dir), { env: { ...process.env, KIROKIT_CREW_BIN: bin } });
  assert.equal(r.status, 1);
  assert.match(r.stderr, /crew-drive: gateway: kirocrew gateway exited \(3\) before it was ready/);
});

// The gateway is stopped on every exit, also when the dashboard never loads,
// and its throwaway HOME gets only a link to kiro-cli's sign-in.
test('crew-drive: the gateway is stopped when the dashboard step fails', { skip: !hasChrome }, async (t) => {
  if (!(await portFree(34999))) return t.skip('port 34999 is in use');
  const dir = scratch(t);
  const pidFile = join(dir, 'gateway.pid');
  const bin = file(dir, 'kirocrew', `#!/bin/sh\necho $$ > '${pidFile}'\necho 'KIROCREW_READY:{"port": 34999, "token": "secret-token"}'\nexec sleep 300\n`);
  chmodSync(bin, 0o755);
  const r = run('node', crewArgs(dir), { env: { ...process.env, KIROKIT_CREW_BIN: bin, KIROKIT_CREW_PORT: '34999' } });
  assert.equal(r.status, 1);
  assert.match(r.stderr, /crew-drive: dashboard: /);
  assert.doesNotMatch(r.stderr, /secret-token/);
  const pid = Number(readFileSync(pidFile, 'utf8'));
  let alive = true;
  for (let i = 0; i < 20 && alive; i++) {
    try {
      process.kill(pid, 0);
      spawnSync('sleep', ['0.1']);
    } catch {
      alive = false;
    }
  }
  assert.equal(alive, false, `gateway ${pid} still runs`);
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
  assert.match(r.stderr, /timeout waiting for installed product-team/);
  assert.ok(Date.now() - t0 < 45_000, 'the 30 s limit was not applied');
});

// --- record.sh preflight ----------------------------------------------------

const TOOLS = ['tmux', 'asciinema', 'agg', 'ffmpeg', 'ffprobe', 'node', 'npm', 'go', 'git', 'kiro-cli', 'kirocrew'];

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

test('record.sh: a missing kirocrew fails preflight', { skip: demoBusy() }, (t) => {
  const r = record(t, { PATH: stubPath(t, ['kirocrew']) });
  assert.equal(r.status, 1);
  assert.match(r.stderr, /showcase: preflight: kirocrew not found/);
  assert.equal(existsSync('/tmp/kirokit-demo'), false);
});

test('record.sh: a missing age key fails preflight naming its path', { skip: demoBusy() || !hasChrome || !hasKiroIDE }, (t) => {
  const r = record(t, { PATH: stubPath(t), SOPS_AGE_KEY_FILE: '/nonexistent/keys.txt' });
  assert.equal(r.status, 1);
  assert.match(r.stderr, /showcase: preflight: SOPS age key not found at \/nonexistent\/keys.txt/);
});

test('record.sh: a second run leaves a live run\'s folder alone', { skip: demoBusy() || !hasChrome || !hasKiroIDE }, async (t) => {
  if (!(await portFree(34115))) return t.skip('port 34115 is in use');
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

// The run-ownership guard must only refuse a live run: a folder left by a run
// that has ended (crashed, or kept with --keep) is replaced, and once this run
// owns it, a later failure still removes it.
test('record.sh: a dead run\'s folder is taken over and removed when the run fails', { skip: demoBusy() || !hasChrome || !hasKiroIDE }, async (t) => {
  if (!(await portFree(34115))) return t.skip('port 34115 is in use');
  const dir = scratch(t);
  const key = file(dir, 'keys.txt', 'not a key\n');
  const dead = spawnSync('sh', ['-c', 'echo $$']).stdout.toString().trim();
  mkdirSync('/tmp/kirokit-demo/out', { recursive: true });
  t.after(() => rmSync('/tmp/kirokit-demo', { recursive: true, force: true }));
  writeFileSync('/tmp/kirokit-demo/pid', `${dead}\n`);
  writeFileSync('/tmp/kirokit-demo/out/clip.txt', 'left over\n');
  // A failing go stops the run in its build step, right after it took the folder.
  const path = stubPath(t);
  writeFileSync(join(path.split(':')[0], 'go'), '#!/bin/sh\nexit 1\n');
  const r = record(t, { PATH: path, SOPS_AGE_KEY_FILE: key });
  assert.equal(r.status, 1);
  assert.doesNotMatch(r.stderr, /another showcase run/);
  assert.match(r.stderr, /showcase: build: /);
  assert.equal(existsSync('/tmp/kirokit-demo'), false);
});

test('record.sh: a busy dev-server port fails preflight', { skip: demoBusy() || !hasChrome || !hasKiroIDE }, async (t) => {
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
  assert.match(r.stderr, /showcase: preflight: port 34115 is busy: close the running kirokit app server or kirocrew gateway first/);
  assert.equal(existsSync('/tmp/kirokit-demo'), false);
});
