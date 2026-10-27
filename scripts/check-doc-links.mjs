#!/usr/bin/env node
// Checks that the user docs still point at things that exist:
//   - every relative link and image in README.md, CONTRIBUTING.md, AGENTS.md
//     and docs/*.md names a file of the repository, and its #anchor a heading
//     (or an id="" attribute) of that file;
//   - every section name quoted in a tracked file next to the README or a
//     docs/ page ("X" in the README, docs/secrets.md: X, ...) is a heading of
//     that page.
// Prints "all links resolve" and exits 0, or lists each broken one and exits 1.
//
// Usage: node scripts/check-doc-links.mjs (from anywhere; paths are relative to the repository)
import { execFileSync } from 'node:child_process';
import { existsSync, readFileSync, readdirSync, statSync } from 'node:fs';
import { dirname, join, relative, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const docs = ['README.md', 'CONTRIBUTING.md', 'AGENTS.md', ...readdirSync(join(root, 'docs')).filter((f) => f.endsWith('.md')).map((f) => `docs/${f}`)];

// withoutCode drops fenced code blocks and inline code, which hold examples, not links.
function withoutCode(text) {
  return text.replace(/^```[\s\S]*?^```/gm, '').replace(/`[^`\n]*`/g, '');
}

// slug is GitHub's heading anchor: lower case, punctuation dropped, spaces to hyphens.
export function slug(heading) {
  return heading
    .replace(/`/g, '')
    .replace(/\[([^\]]*)\]\([^)]*\)/g, '$1')
    .trim()
    .toLowerCase()
    .replace(/[^\p{L}\p{N} _-]/gu, '')
    .replace(/ /g, '-');
}

const anchorCache = new Map();
function anchorsOf(file) {
  if (anchorCache.has(file)) return anchorCache.get(file);
  const text = readFileSync(join(root, file), 'utf8');
  const anchors = new Set();
  const headings = new Set();
  const counts = new Map();
  for (const m of text.replace(/^```[\s\S]*?^```/gm, '').matchAll(/^#{1,6}\s+(.+?)\s*#*\s*$/gm)) {
    headings.add(m[1].replace(/`/g, ''));
    const s = slug(m[1]);
    const n = counts.get(s) || 0;
    anchors.add(n ? `${s}-${n}` : s);
    counts.set(s, n + 1);
  }
  for (const m of text.matchAll(/<h[1-6][^>]*\bid="([^"]+)"[^>]*>([^<]*)</g)) {
    anchors.add(m[1]);
    headings.add(m[2]);
  }
  for (const m of text.matchAll(/\bid="([^"]+)"/g)) anchors.add(m[1]);
  const result = { anchors, headings };
  anchorCache.set(file, result);
  return result;
}

const problems = [];
let checked = 0;

// Links and images: [text](target), <a href>, <img src>.
for (const doc of docs) {
  const text = withoutCode(readFileSync(join(root, doc), 'utf8'));
  const targets = [
    ...[...text.matchAll(/\]\(([^)\s]+)(?:\s+"[^"]*")?\)/g)].map((m) => m[1]),
    ...[...text.matchAll(/\b(?:href|src|srcset)="([^"]+)"/g)].map((m) => m[1]),
  ];
  for (const target of targets) {
    if (/^[a-z][a-z0-9+.-]*:/i.test(target)) continue; // https:, mailto: and the like
    checked++;
    const [path, anchor] = target.split('#');
    const file = path ? relative(root, resolve(join(root, dirname(doc)), decodeURIComponent(path))) : doc;
    if (file.startsWith('..') || !existsSync(join(root, file))) {
      problems.push(`${doc}: ${target}: no such file ${file}`);
      continue;
    }
    if (anchor === undefined) continue;
    if (statSync(join(root, file)).isDirectory() || !file.endsWith('.md')) {
      problems.push(`${doc}: ${target}: an anchor needs a Markdown file`);
      continue;
    }
    if (!anchorsOf(file).anchors.has(anchor)) problems.push(`${doc}: ${target}: ${file} has no heading #${anchor}`);
  }
}

// Quoted section names elsewhere in the repository.
const tracked = execFileSync('git', ['ls-files', '-z'], { cwd: root, encoding: 'utf8' })
  .split('\0')
  .filter((f) => f && f !== 'scripts/check-doc-links.mjs' && !f.startsWith('docs/plans/') && !f.startsWith('docs/research/') && !f.startsWith('testdata/') && !/\.(png|gif|mp4|webm|svg|ico|icns|lock|sum)$/.test(f));
const patterns = [
  { re: /README's "([^"]+)"/g, page: () => 'README.md' },
  { re: /"([^"]+)" (?:section )?in the README/g, page: () => 'README.md' },
  { re: /README: ([A-Z][^)\n]*?)\)/g, page: () => 'README.md' },
  { re: /"([^"]+)" in\s+(?:#\s*)?(?:the repository's )?(docs\/[\w-]+\.md)/g, page: (m) => m[2] },
  { re: /(docs\/[\w-]+\.md): ([A-Z][^)\n]*?)\)/g, page: (m) => m[1], name: (m) => m[2] },
];
for (const file of tracked) {
  if (!existsSync(join(root, file)) || statSync(join(root, file)).isDirectory()) continue;
  const text = readFileSync(join(root, file), 'utf8');
  if (!text.includes('README') && !text.includes('docs/')) continue;
  for (const { re, page, name } of patterns) {
    for (const m of text.matchAll(re)) {
      checked++;
      const target = page(m);
      const section = (name ? name(m) : m[1]).replace(/\s+/g, ' ').trim();
      if (!existsSync(join(root, target))) {
        problems.push(`${file}: "${section}": no such page ${target}`);
      } else if (!anchorsOf(target).headings.has(section)) {
        problems.push(`${file}: "${section}" is not a heading of ${target}`);
      }
    }
  }
}

if (problems.length) {
  for (const p of problems) console.error(p);
  console.error(`${problems.length} of ${checked} links and section names do not resolve`);
  process.exit(1);
}
console.log(`all links resolve (${checked} checked)`);
