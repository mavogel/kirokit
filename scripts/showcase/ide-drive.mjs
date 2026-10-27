// Shows Kiro IDE listing the team's agents from the showcase's Kiro home and
// records it with playwright-core's Electron driver:
//   start Kiro on a fresh profile with KIRO_HOME set, Skip All on the import
//   screen, open the chat panel's agent selector, wait for every agent name.
//
// Kiro IDE uses this Mac's existing Kiro sign-in through the real HOME; only
// KIRO_HOME and the profile folder are the showcase's own. The profile folder
// must have a short path: Electron's socket inside it fails past ~100 bytes.
//
// Writes to <outdir>:
//   ide.webm        the recording
//   ide-text.txt    every line of window and chat text shown during the run, for the privacy check
//   ide-marks.json  seconds from the start of the video to each step, for captions and trimming
//
// A window that stays blank (no chat panel within 60 s) is closed and the
// scene tried once more on a fresh profile (<profile-dir>-2). Every wait is
// bounded, and the whole run gives up after KIROKIT_IDE_DEADLINE seconds
// (default 300), so a stuck Kiro never hangs the recording.
//
// Usage: node ide-drive.mjs <outdir> <kiro-home> <profile-dir> <workspace> <agent>...
// KIROKIT_IDE_APP overrides the Kiro executable.
import { existsSync } from 'node:fs';
import { access, mkdir, rename, rm, writeFile } from 'node:fs/promises';
import { join } from 'node:path';
import { _electron as electron } from 'playwright-core';

const [outdir, kiroHome, profile, workspace, ...agents] = process.argv.slice(2);
if (!outdir || !kiroHome || !profile || !workspace || agents.length === 0) {
  console.error('usage: node ide-drive.mjs <outdir> <kiro-home> <profile-dir> <workspace> <agent>...');
  process.exit(2);
}
const APP = process.env.KIROKIT_IDE_APP || '/Applications/Kiro.app/Contents/MacOS/Kiro';
const size = { width: 1280, height: 800 };
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

try {
  await access(APP);
} catch {
  console.error(`ide-drive: Kiro IDE not found at ${APP}`);
  process.exit(1);
}

let step = 'launch';
let marks = {};
let start = 0;
const mark = (name) => { marks[name] = Number(((Date.now() - start) / 1000).toFixed(2)); };

// within rejects when promise has not settled after ms: a window that never
// finishes loading can leave an evaluate or a close waiting forever.
function within(promise, ms, what) {
  let timer;
  const late = new Promise((_, reject) => { timer = setTimeout(() => reject(new Error(`${what} did not finish within ${ms / 1000} s`)), ms); });
  return Promise.race([promise, late]).finally(() => clearTimeout(timer));
}

// textOf reads the visible text of the window and every frame in it (the
// chat panel is a webview frame).
async function textOf(page) {
  const lines = [];
  for (const f of page.frames()) {
    const t = await within(f.evaluate(() => (document.body ? document.body.innerText : '')), 2000, 'reading the text').catch(() => '');
    lines.push(...t.split('\n'));
  }
  return lines;
}

// The whole run has one deadline (KIROKIT_IDE_DEADLINE seconds, default 300),
// so it always ends with a message instead of hanging the recording.
let app;
const deadline = Number(process.env.KIROKIT_IDE_DEADLINE || 300);
const watchdog = setTimeout(() => {
  console.error(`ide-drive: ${step}: gave up after ${deadline} s`);
  try { app?.process()?.kill('SIGKILL'); } catch { /* already gone */ }
  process.exit(1);
}, deadline * 1000);

// closeApp closes Kiro, and kills it when it does not close in 10 s.
async function closeApp() {
  if (!app) return;
  const a = app;
  app = null;
  await within(a.close(), 10_000, 'closing Kiro IDE').catch(() => {
    try { a.process()?.kill('SIGKILL'); } catch { /* already gone */ }
  });
}

// chatFrame finds the frame that holds the agent selector.
async function chatFrame(page, timeout) {
  const until = Date.now() + timeout;
  while (Date.now() < until) {
    for (const f of page.frames()) {
      if (await f.getByText('Default', { exact: true }).first().isVisible().catch(() => false)) return f;
    }
    await sleep(250);
  }
  throw new Error('the chat panel with its agent selector did not appear');
}

await mkdir(workspace, { recursive: true });
const seen = new Set();

// attempt runs the scene once on a profile folder of its own. A window that
// stays blank fails at the chat panel step; the caller then tries again.
async function attempt(profileDir) {
  await mkdir(profileDir, { recursive: true });
  marks = {};
  let sampling = false;
  let sampler = Promise.resolve();
  let video = null;
  try {
    step = 'launch';
    app = await electron.launch({
      executablePath: APP,
      args: [`--user-data-dir=${profileDir}`, '--new-window', workspace],
      env: { ...process.env, KIRO_HOME: kiroHome },
      recordVideo: { dir: outdir, size },
      timeout: 90_000,
    });
    const page = await app.firstWindow({ timeout: 90_000 });
    video = page.video();
    start = Date.now();
    await within(page.setViewportSize(size), 5000, 'sizing the window').catch(() => {});

    sampling = true;
    sampler = (async () => {
      while (sampling) {
        for (const line of await textOf(page).catch(() => [])) if (line.trim()) seen.add(line);
        await sleep(200);
      }
    })();

    step = 'import screen';
    const skip = page.getByRole('button', { name: 'Skip All' });
    await skip.waitFor({ timeout: 60_000 }).then(() => skip.click({ timeout: 10_000 })).catch(() => {});
    mark('opened');

    step = 'chat panel';
    const chat = await chatFrame(page, 60_000);
    await sleep(1200);

    step = 'agent selector';
    mark('picker');
    await chat.getByText('Default', { exact: true }).first().click({ timeout: 10_000 });
    const until = Date.now() + 30_000;
    let missing = agents;
    while (Date.now() < until) {
      const text = (await within(chat.evaluate(() => document.body.innerText), 2000, 'reading the chat panel').catch(() => ''))
        .split('\n').map((l) => l.trim());
      missing = agents.filter((a) => !text.includes(a));
      if (missing.length === 0) break;
      await sleep(250);
    }
    if (missing.length) throw new Error(`agents did not appear: ${missing.join(', ')}`);
    // The team's agents are listed under GLOBAL, below the built-in ones.
    await chat.getByText(agents[agents.length - 1], { exact: true }).first().scrollIntoViewIfNeeded({ timeout: 5000 }).catch(() => {});
    mark('agents');
    await sleep(3000);
    mark('end');

    step = 'save video';
    sampling = false;
    await sampler;
    await closeApp();
    await rename(await within(video.path(), 10_000, 'saving the video'), join(outdir, 'ide.webm'));
  } finally {
    sampling = false;
    await within(sampler, 5000, 'stopping the text sampler').catch(() => {});
    await closeApp();
    // A failed attempt's recording is never used.
    if (!existsSync(join(outdir, 'ide.webm'))) await within(video?.path() ?? Promise.resolve(''), 5000, 'the video').then((p) => p && rm(p, { force: true })).catch(() => {});
  }
}

let failed = false;
try {
  try {
    await attempt(profile);
  } catch (err) {
    if (step !== 'chat panel') throw err;
    console.error(`ide-drive: ${step}: ${err.message.split('\n')[0]}; trying once more on a fresh profile`);
    await attempt(`${profile}-2`);
  }
  step = 'text';
  await writeFile(join(outdir, 'ide-text.txt'), `${[...seen].join('\n')}\n`);
  await writeFile(join(outdir, 'ide-marks.json'), `${JSON.stringify(marks, null, 2)}\n`);
} catch (err) {
  failed = true;
  console.error(`ide-drive: ${step}: ${err.message.split('\n')[0]}`);
}
clearTimeout(watchdog);
process.exit(failed ? 1 : 0);
