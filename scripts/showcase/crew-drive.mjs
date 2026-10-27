// Shows kirocrew listing the team's agents from the showcase's Kiro home and
// records its dashboard with headless Chrome:
//   start `kirocrew gateway` on a throwaway HOME, open the dashboard, click
//   through its first-run dialogs, open the chat's agent picker, wait for every
//   agent name. The gateway is stopped on every exit.
//
// The throwaway HOME gets a link (never a copy) to kiro-cli's sign-in folder,
// so kirocrew finds a signed-in Kiro CLI; nothing is written to the real HOME.
// The dashboard token travels only in the URL and must never show as text.
//
// Writes to <outdir>:
//   crew.webm        the recording (1180x760)
//   crew-text.txt    every line of page text shown during the run, for the privacy check
//   crew-marks.json  seconds from the start of the video to each step, for captions and trimming
//
// Usage: node crew-drive.mjs <outdir> <kiro-home> <crew-home> <agent>...
// KIROKIT_CREW_BIN overrides the kirocrew command, KIROKIT_CREW_PORT the port (34117).
import { spawn } from 'node:child_process';
import { existsSync } from 'node:fs';
import { mkdir, rename, symlink, writeFile } from 'node:fs/promises';
import { homedir } from 'node:os';
import { join } from 'node:path';
import { chromium } from 'playwright-core';

const [outdir, kiroHome, crewHome, ...agents] = process.argv.slice(2);
if (!outdir || !kiroHome || !crewHome || agents.length === 0) {
  console.error('usage: node crew-drive.mjs <outdir> <kiro-home> <crew-home> <agent>...');
  process.exit(2);
}
const BIN = process.env.KIROKIT_CREW_BIN || 'kirocrew';
const PORT = Number(process.env.KIROKIT_CREW_PORT || 34117);
const size = { width: 1180, height: 760 };
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

let step = 'gateway';
const marks = {};
let start = 0;
const mark = (name) => { marks[name] = Number(((Date.now() - start) / 1000).toFixed(2)); };

// startGateway starts kirocrew in its own process group and resolves to the
// token from its ready line, or rejects when it exits or stays silent.
function startGateway() {
  const gw = spawn(BIN, ['gateway', '--no-open', '--no-tunnel', '--no-crons', '--port', String(PORT), '--json-ready'], {
    env: { ...process.env, HOME: crewHome, KIRO_HOME: kiroHome, KIROCREW_HOME: join(crewHome, 'crew') },
    cwd: crewHome,
    detached: true,
    stdio: ['ignore', 'pipe', 'pipe'],
  });
  const ready = new Promise((resolve, reject) => {
    let buf = '';
    const timer = setTimeout(() => reject(new Error('kirocrew printed no KIROCREW_READY line within 90 s')), 90_000);
    gw.on('error', (err) => {
      clearTimeout(timer);
      reject(new Error(err.code === 'ENOENT' ? `${BIN} not found` : err.message));
    });
    gw.on('exit', (code) => {
      clearTimeout(timer);
      reject(new Error(`kirocrew gateway exited (${code}) before it was ready`));
    });
    gw.stdout.on('data', (chunk) => {
      buf += chunk;
      const m = buf.match(/KIROCREW_READY:(\{.*\})/);
      if (!m) return;
      clearTimeout(timer);
      try {
        resolve(JSON.parse(m[1]).token);
      } catch {
        reject(new Error('kirocrew printed an unreadable ready line'));
      }
    });
    gw.stderr.resume();
  });
  return { gw, ready };
}

function stop(gw) {
  if (!gw || gw.exitCode !== null) return;
  try {
    process.kill(-gw.pid, 'SIGTERM');
  } catch {
    gw.kill('SIGTERM');
  }
}

await mkdir(join(crewHome, 'Library', 'Application Support'), { recursive: true });
const signIn = join(homedir(), 'Library', 'Application Support', 'kiro-cli');
const linked = join(crewHome, 'Library', 'Application Support', 'kiro-cli');
if (existsSync(signIn) && !existsSync(linked)) await symlink(signIn, linked);

let gw;
let token = '';
let browser;
let failed = false;
let sampling = false;
let sampler = Promise.resolve();
for (const sig of ['SIGINT', 'SIGTERM']) process.on(sig, () => { stop(gw); process.exit(1); });
try {
  const started = startGateway();
  gw = started.gw;
  token = await started.ready;

  step = 'dashboard';
  browser = await chromium.launch({ channel: 'chrome', headless: true });
  const context = await browser.newContext({ viewport: size, deviceScaleFactor: 2, recordVideo: { dir: outdir, size } });
  const page = await context.newPage();
  start = Date.now();

  const seen = new Set();
  sampling = true;
  sampler = (async () => {
    while (sampling) {
      const text = await page.evaluate(() => (document.body ? document.body.innerText : '')).catch(() => '');
      for (const line of text.split('\n')) if (line.trim()) seen.add(line);
      await sleep(200);
    }
  })();

  await page.goto(`http://127.0.0.1:${PORT}/?token=${encodeURIComponent(token)}`, { timeout: 60_000 });
  const picker = page.getByRole('button', { name: /^Agent: / });
  await picker.waitFor({ timeout: 90_000 });

  step = 'first-run dialogs';
  for (let i = 0; i < 6; i++) {
    const dialog = page.locator('[role=dialog][aria-modal=true]').first();
    if (!(await dialog.count())) break;
    const next = dialog.getByRole('button', { name: /^(Continue|Skip|Got it|Done|Not now)/ }).first();
    if (!(await next.count())) break;
    await next.click();
    await sleep(700);
  }
  mark('opened');
  await sleep(1200);

  step = 'agent picker';
  mark('picker');
  await picker.click();
  const until = Date.now() + 30_000;
  let missing = agents;
  while (Date.now() < until) {
    const lines = (await page.evaluate(() => document.body.innerText)).split('\n').map((l) => l.trim());
    missing = agents.filter((a) => !lines.includes(a));
    if (missing.length === 0) break;
    await sleep(250);
  }
  if (missing.length) throw new Error(`agents did not appear: ${missing.join(', ')}`);
  // kirocrew lists its own agents among the team's; show the team's last one.
  await page.getByText(agents[agents.length - 1], { exact: true }).first().scrollIntoViewIfNeeded().catch(() => {});
  mark('agents');
  await sleep(3000);
  mark('end');

  step = 'text';
  sampling = false;
  await sampler;
  if ([...seen].some((line) => line.includes(token))) throw new Error('the dashboard token showed as page text');
  await writeFile(join(outdir, 'crew-text.txt'), `${[...seen].join('\n')}\n`);
  await writeFile(join(outdir, 'crew-marks.json'), `${JSON.stringify(marks, null, 2)}\n`);

  step = 'save video';
  const video = page.video();
  await context.close();
  await rename(await video.path(), join(outdir, 'crew.webm'));
} catch (err) {
  failed = true;
  // Playwright errors quote the URL, which holds the token.
  const reason = err.message.split('\n')[0];
  console.error(`crew-drive: ${step}: ${token ? reason.replaceAll(token, '<token>').replaceAll(encodeURIComponent(token), '<token>') : reason}`);
} finally {
  sampling = false;
  await sampler.catch(() => {});
  if (browser) await browser.close().catch(() => {});
  stop(gw);
}
process.exit(failed ? 1 : 0);
