import { esc } from './html.js';

// The version a sync uses: a branch or tag with --branch, or the newest
// release ('' ) with --latest-tag. No button falls back to the newest release
// on its own; the saved choice decides.
export const refArgs = (ref) => (ref ? ['--branch', ref] : ['--latest-tag']);

// The entries of a Version select: Newest release, then origin's branches,
// then the tags. A tag named like a branch is left out, since --branch takes
// the branch. A saved choice no longer on origin stays listed so the select
// can show it, once the clone's refs have been read. picked is selected.
export function versionOptions(refs, saved, picked) {
  const branches = refs?.branches || [];
  const tags = (refs?.tags || []).filter((t) => !branches.includes(t));
  const option = (value, label = value) =>
    `<option value="${esc(value)}"${value === picked ? ' selected' : ''}>${esc(label)}</option>`;
  const group = (label, values) =>
    values.length ? `<optgroup label="${label}">${values.map((v) => option(v)).join('')}</optgroup>` : '';

  let html = option('', refs?.newest ? `Newest release · follows new releases (now ${refs.newest})` : 'Newest release (none yet)');
  const read = Boolean(refs?.current?.commit);
  if (read && saved && !branches.includes(saved) && !tags.includes(saved)) {
    html += `<optgroup label="Saved">${option(saved, `${saved} (not on origin)`)}</optgroup>`;
  }
  return html + group('Branches', branches) + group('Tags', tags);
}

// Where the templates come from, after lead ("Templates from", "In use:"):
// what git says is checked out. Anything but the newest release under the
// Newest release choice carries a badge, so a forgotten branch or pin stands
// out. '' until the clone has been read.
export function sourceLine(view, lead) {
  const refs = view.checkout || {};
  const c = refs.current;
  if (!c || !c.commit) return '';
  const at = `at <b>${esc(c.commit)}</b>`;
  const badge = (text) => `<span class="badge warn">${text}</span>`;
  if (c.kind === 'tag' && !view.ref && c.name === refs.newest) {
    return `${lead} <span class="badge off">newest release</span> <b>${esc(c.name)}</b> ${at}`;
  }
  if (c.kind === 'tag') {
    return `${lead} tag <b>${esc(c.name)}</b> ${at} ${badge(view.ref ? 'pinned · does not follow new releases' : 'not the newest release')}`;
  }
  if (c.kind === 'branch') return `${lead} branch <b>${esc(c.name)}</b> ${at} ${badge('not a release')}`;
  return `${lead} commit <b>${esc(c.commit)}</b> ${badge('not a release')}`;
}
