// Helpers the page views share. Views return HTML strings built only from
// escaped data, so node's test runner covers them without a DOM.

export function esc(value) {
  return String(value ?? '')
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;');
}

export function plural(n, one, many) {
  return `${n} ${n === 1 ? one : many}`;
}

// A time as a <time> element: the machine value in datetime, the person's
// local rendering as text.
export function timeTag(iso) {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '';
  const text = d.toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' });
  return `<time datetime="${esc(d.toISOString())}">${esc(text)}</time>`;
}

export function chips(items) {
  return (items || []).map((i) => `<span class="chip">${esc(i)}</span>`).join('');
}

// What replaces a table when the catalog could not be read, or there is no
// repository yet. Returns '' when the data is ready to render.
export function problem({ error, cloned }) {
  if (error) return `<p class="error" role="alert">${esc(error)}</p>`;
  if (!cloned) return '<p class="note">No repository yet. Enter its address in Settings to see your teams.</p>';
  return '';
}

export function installedBy(installs) {
  return new Map((installs || []).map((r) => [r.Team, r]));
}
