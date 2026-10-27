// Renders the showcase's still images with headless Chrome (this ffmpeg has no
// text filter) and computes when each caption shows.
//
// Reads <outdir>/cli.cast, ide-marks.json, crew-marks.json and app-marks.json;
// writes to <outdir>:
//   cap-1.png to cap-7.png 1280x96 captions on a transparent background
//   title.png, end.png     1280x896 opening and closing cards
//   frame.png              1180x760 rounded frame laid over each clip
//   captions.txt           every word the images show, for the privacy check
//   cues.env               clip lengths, cut points and caption times (seconds), sourced by record.sh
//
// Usage: node caption.mjs <outdir>
import { readFile, writeFile } from 'node:fs/promises';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import { chromium } from 'playwright-core';

const outdir = process.argv[2];
if (!outdir) {
  console.error('usage: node caption.mjs <outdir>');
  process.exit(2);
}

// agg renders the cast with this idle limit (record.sh); the cue times replay
// the cast the same way so they match the CLI clip.
const IDLE_LIMIT = 3;
const LAST_FRAME = 2;

// In order: three over the CLI clip, one each over the Kiro IDE and kirocrew
// clips, two over the app clip.
const CAPTIONS = [
  "Install your team's Kiro config",
  "Kiro starts with the team's agent and skills",
  'Or export it for Amazon Quick',
  'Kiro IDE lists the same agents',
  'So does kirocrew',
  'Or click: clone the config repo',
  'Install the team, see every command',
];
const TITLE = ['kirokit', 'One set of team templates for every Kiro tool'];
const END = [
  'Get started',
  'brew install --cask mavogel/tap/kirokit',
  'brew install --cask mavogel/tap/kirokit-desktop',
  'github.com/mavogel/kirokit',
];

async function input(name) {
  try {
    return await readFile(join(outdir, name), 'utf8');
  } catch {
    console.error(`caption: missing input ${join(outdir, name)}`);
    process.exit(1);
  }
}

// Times in the rendered CLI clip of the first output containing each text.
function castCues(cast, texts) {
  const [header, ...events] = cast.trim().split('\n').map((l) => JSON.parse(l));
  const relative = header.version >= 3; // v3 stores intervals, v2 absolute times
  const found = {};
  let t = 0;
  let prev = 0;
  for (const [time, kind, data] of events) {
    t += Math.min(relative ? time : time - prev, IDLE_LIMIT);
    prev = time;
    if (kind !== 'o') continue;
    for (const text of texts.filter((x) => !(x in found) && data.includes(x))) found[text] = t;
  }
  const missing = texts.find((x) => !(x in found));
  if (missing) {
    console.error(`caption: cli.cast never shows "${missing}"`);
    process.exit(1);
  }
  return { found, duration: t + LAST_FRAME };
}

async function marksOf(name, keys) {
  const marks = JSON.parse(await input(name));
  for (const key of keys) {
    if (typeof marks[key] !== 'number') {
      console.error(`caption: ${name} has no "${key}"`);
      process.exit(1);
    }
  }
  return marks;
}

const cast = await input('cli.cast');
const marks = await marksOf('app-marks.json', ['welcome', 'clone', 'teams', 'install', 'end']);
const ide = await marksOf('ide-marks.json', ['opened', 'picker', 'agents', 'end']);
const crew = await marksOf('crew-marks.json', ['opened', 'picker', 'agents', 'end']);

const cli = castCues(cast, ['installed product-team', 'Session ended', 'rendered Quick bundle for product-team']);
// Caption 2 takes over once the install line has been on screen a moment, and
// caption 3 once the Kiro session has ended and the Quick render is typed.
const cap2 = cli.found['installed product-team'] + 1.2;
const cap3 = cli.found['Session ended'] + 0.5;

// The IDE and kirocrew clips start once the window is up (at most 2.5 s
// before the agent picker opens) and keep the agent list to the end.
const clipCut = (m) => ({ t0: Math.max(m.opened, m.picker - 2.5), t1: m.end });
const ideCut = clipCut(ide);
const crewCut = clipCut(crew);

// The app clip keeps the welcome page from just before it settles, cuts the
// clone wait down to 1.5 s, and keeps everything from just before the teams show.
const appA0 = Math.max(0, marks.welcome - 0.5);
const appA1 = marks.clone + 1.5;
const appB0 = Math.max(appA1, marks.teams - 1.0);
const appB1 = marks.end;
const appDuration = appA1 - appA0 + (appB1 - appB0);
const cap7 = appA1 - appA0 + (marks.install - 0.3 - appB0);

const fmt = (n) => n.toFixed(2);
const cues = {
  CLI_DURATION: cli.duration,
  CAP2_AT: cap2,
  CAP3_AT: cap3,
  IDE_T0: ideCut.t0,
  IDE_T1: ideCut.t1,
  IDE_DURATION: ideCut.t1 - ideCut.t0,
  CREW_T0: crewCut.t0,
  CREW_T1: crewCut.t1,
  CREW_DURATION: crewCut.t1 - crewCut.t0,
  APP_A0: appA0,
  APP_A1: appA1,
  APP_B0: appB0,
  APP_B1: appB1,
  APP_DURATION: appDuration,
  CAP7_AT: cap7,
};
const shortest = Math.min(
  cap2, cap3 - cap2, cli.duration - cap3,
  cues.IDE_DURATION, cues.CREW_DURATION,
  cap7, appDuration - cap7,
);
if (shortest < 2.5) {
  console.error(`caption: a caption would show only ${fmt(shortest)} s (at least 2.5 s)`);
  process.exit(1);
}

const logo = await readFile(join(dirname(fileURLToPath(import.meta.url)), '../../assets/logo.svg'), 'utf8');
const esc = (s) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;');
const base = `<style>
  * { margin: 0; box-sizing: border-box; }
  html, body { background: transparent; font-family: -apple-system, "SF Pro Text", "Helvetica Neue", sans-serif; color: #fff; }
  code { font-family: "SF Mono", Menlo, monospace; }
</style>`;

const pages = [
  ...CAPTIONS.map((text, i) => ({
    file: `cap-${i + 1}.png`, width: 1280, height: 96, transparent: true,
    html: `${base}<style>
      body { height: 96px; display: flex; align-items: center; justify-content: center; }
      .pill { display: flex; align-items: center; gap: 14px; padding: 12px 28px 12px 14px; border-radius: 999px; background: #5b21b6; font-size: 30px; font-weight: 600; letter-spacing: -0.01em; }
      .n { display: grid; place-items: center; width: 44px; height: 44px; border-radius: 50%; background: #ede9fe; color: #4c1d95; font-size: 24px; font-weight: 700; }
    </style><div class="pill"><span class="n">${i + 1}</span>${esc(text)}</div>`,
  })),
  {
    file: 'title.png', width: 1280, height: 896,
    html: `${base}<style>
      body { height: 896px; background: #13111f; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 28px; }
      svg { width: 148px; height: 148px; }
      h1 { font-size: 88px; font-weight: 700; letter-spacing: -0.03em; }
      p { font-size: 32px; color: #c4b5fd; }
    </style>${logo}<h1>${esc(TITLE[0])}</h1><p>${esc(TITLE[1])}</p>`,
  },
  {
    file: 'end.png', width: 1280, height: 896,
    html: `${base}<style>
      body { height: 896px; background: #13111f; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 26px; }
      svg { width: 96px; height: 96px; }
      h2 { font-size: 56px; font-weight: 700; letter-spacing: -0.02em; margin-bottom: 8px; }
      code { display: block; padding: 18px 30px; border-radius: 14px; background: #221d38; border: 1px solid #3a3450; font-size: 30px; color: #ede9fe; }
      code b { color: #a78bfa; font-weight: 400; }
      p { font-size: 28px; color: #c4b5fd; margin-top: 10px; }
    </style>${logo}<h2>${esc(END[0])}</h2><code><b>$</b> ${esc(END[1])}</code><code><b>$</b> ${esc(END[2])}</code><p>${esc(END[3])}</p>`,
  },
  {
    // Laid over each clip: clear inside the rounded rectangle, the page
    // background in the corners outside it, and a 1 px border.
    file: 'frame.png', width: 1180, height: 760, transparent: true,
    html: `<style>* { margin: 0; box-sizing: border-box; } html, body { background: transparent; } div { width: 1180px; height: 760px; border-radius: 12px; border: 1px solid #3a3450; box-shadow: 0 0 0 40px #13111f; }</style><div></div>`,
  },
];

// One fresh page per image, sized at creation: resizing a reused page let
// screenshots catch stale paint (repeated rows, a blank caption).
const browser = await chromium.launch({ channel: 'chrome', headless: true });
try {
  for (const p of pages) {
    const page = await browser.newPage({ viewport: { width: p.width, height: p.height }, deviceScaleFactor: 1 });
    await page.setContent(p.html, { waitUntil: 'load' });
    await page.evaluate(async () => {
      await document.fonts.ready;
      await new Promise((r) => requestAnimationFrame(() => requestAnimationFrame(r)));
    });
    await page.screenshot({ path: join(outdir, p.file), omitBackground: Boolean(p.transparent) });
    await page.close();
  }
} finally {
  await browser.close();
}

await writeFile(join(outdir, 'captions.txt'), `${[...CAPTIONS, ...TITLE, ...END].join('\n')}\n`);
await writeFile(join(outdir, 'cues.env'), Object.entries(cues).map(([k, v]) => `${k}=${fmt(v)}\n`).join(''));
console.log(`caption: ${pages.length} images, CLI ${fmt(cli.duration)} s, IDE ${fmt(cues.IDE_DURATION)} s, kirocrew ${fmt(cues.CREW_DURATION)} s, app ${fmt(appDuration)} s`);
