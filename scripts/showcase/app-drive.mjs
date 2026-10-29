// Clicks the showcase app story in the real desktop UI (the Wails server build
// that app-server.sh starts serves it at http://localhost:34115) and records it
// with headless Chrome:
//   Clone the config repository, Continue on its version step (Newest
//   release), Install platform-team, Install in the setup dialog, wait for the
//   "installed" badge.
//
// Writes to <outdir>:
//   app.webm        the recording (1180x760)
//   app-dock.txt    every line of page text, input value and dock row shown during the run, for the privacy check
//   app-marks.json  seconds from the start of the video to each step, for captions and trimming
//
// Usage: node app-drive.mjs <outdir>
import { rename, writeFile } from 'node:fs/promises';
import { join } from 'node:path';
import { chromium } from 'playwright-core';

const outdir = process.argv[2];
if (!outdir) {
  console.error('usage: node app-drive.mjs <outdir>');
  process.exit(2);
}

const URL = 'http://localhost:34115';
const REPO = 'https://github.com/mavogel/kirokit.git';
const TEAM = 'platform-team';
const size = { width: 1180, height: 760 };
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

let step = 'launch';
const marks = {};
let start = 0;
const mark = (name) => { marks[name] = Number(((Date.now() - start) / 1000).toFixed(2)); };

const browser = await chromium.launch({ channel: 'chrome', headless: true });
let failed = false;
try {
  const context = await browser.newContext({
    viewport: size,
    deviceScaleFactor: 2,
    recordVideo: { dir: outdir, size },
  });
  const page = await context.newPage();
  start = Date.now();

  // Dock lines scroll away, so the text for the privacy check is sampled
  // every 200 ms over the whole run, not read once at the end.
  const seen = new Set();
  let sampling = true;
  const sampler = (async () => {
    while (sampling) {
      const { lines, inputs } = await page.evaluate(() => ({
        lines: [
          ...document.body.innerText.split('\n'),
          ...[...document.querySelectorAll('.xterm-rows > div')].map((r) => r.textContent),
        ],
        inputs: [...document.querySelectorAll('input')].map((i) => i.value || i.placeholder),
      })).catch(() => ({ lines: [], inputs: [] }));
      // A half-typed repository address is our own public text, not output.
      const typed = inputs.filter((v) => !(REPO.startsWith(v) && v !== REPO));
      for (const line of [...lines, ...typed]) if (line.trim()) seen.add(line);
      await sleep(200);
    }
  })();

  step = 'welcome';
  await page.goto(URL, { timeout: 30_000 });
  await page.locator('#clone-address').waitFor({ timeout: 60_000 });
  await sleep(1000);
  mark('welcome');

  step = 'type address';
  await page.locator('#clone-address').pressSequentially(REPO, { delay: 45 });
  await sleep(600);

  step = 'clone';
  mark('clone');
  await page.locator('#clone-form button[type=submit]').click();

  // After the clone the setup page asks which version to use, with Newest
  // release preselected; Continue keeps it.
  step = 'version';
  const proceed = page.locator('#version-form button[type=submit]:not([disabled])');
  await proceed.waitFor({ timeout: 120_000 });
  await sleep(800);
  await proceed.click();

  step = 'teams';
  const install = page.locator(`button[data-action="install"][data-team="${TEAM}"]`);
  await install.waitFor({ timeout: 120_000 });
  mark('teams');
  await sleep(1500);

  step = 'install';
  mark('install');
  await install.click();

  // A team installed before shows "Replace the installed team?" first.
  step = 'setup dialog';
  const ready = page.locator('#setup-install:not([disabled])');
  const replace = page.locator('#replace-dialog[open] button[value="ok"]');
  // The branch that loses the race must not reject later as an unhandled error;
  // the final waitFor below is the real check.
  await Promise.race([
    ready.waitFor({ timeout: 60_000 }),
    replace.waitFor({ timeout: 60_000 }).then(() => replace.click()),
  ]).catch(() => {});
  await ready.waitFor({ timeout: 60_000 });
  mark('setup');
  await sleep(1800);

  step = 'confirm install';
  await ready.click();
  mark('installing');

  step = 'installed';
  await page.locator(`.card[data-team="${TEAM}"] .badge.ok`).waitFor({ timeout: 120_000 });
  await page.waitForFunction(
    (team) => [...document.querySelectorAll('#term-kirokit .xterm-rows > div')]
      .some((row) => row.textContent.includes(`installed ${team} into`)),
    TEAM,
    { timeout: 30_000 },
  );
  mark('installed');
  await sleep(2500);
  mark('end');

  step = 'dock text';
  sampling = false;
  await sampler;
  await writeFile(join(outdir, 'app-dock.txt'), `${[...seen].join('\n')}\n`);
  await writeFile(join(outdir, 'app-marks.json'), `${JSON.stringify(marks, null, 2)}\n`);

  step = 'save video';
  const video = page.video();
  await context.close();
  await rename(await video.path(), join(outdir, 'app.webm'));
} catch (err) {
  failed = true;
  console.error(`app-drive: ${step}: ${err.message.split('\n')[0]}`);
} finally {
  await browser.close();
}
process.exit(failed ? 1 : 0);
