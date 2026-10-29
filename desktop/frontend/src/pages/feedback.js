import { esc, actionError } from './html.js';

// Where a refused open is shown: beside the Feedback button of that item.
export const feedbackWhere = (team, kind, name) => `feedback:${team}:${kind}:${name}`;

// The Owner cell: the handle, "team owner" under an inherited one, or
// "No owner" when neither the item nor its team names one.
export function ownerCell(item) {
  if (!item.owner) return '<td><span class="note">No owner</span></td>';
  const from = item.ownerFromTeam ? '<span class="from">team owner</span>' : '';
  return `<td><span class="owner"><span class="mono">${esc(item.owner)}</span>${from}</span></td>`;
}

// The Feedback cell of one agent or skill (kind "agent" or "skill"). The
// button is off when the team names no feedback repository.
export function feedbackCell(view, team, kind, item) {
  if (!team.feedback) {
    return `<td><button class="btn" disabled aria-label="Give feedback on ${esc(item.name)} (no feedback repository)">Feedback</button></td>`;
  }
  const button = `<button class="btn" data-feedback="${esc(kind)}" data-team="${esc(team.name)}" data-name="${esc(item.name)}" aria-label="Give feedback on ${esc(item.name)}">Feedback</button>`;
  return `<td><div class="row-actions">${button}${actionError(view, feedbackWhere(team.name, kind, item.name))}</div></td>`;
}

// The note under a table naming each team whose rows have their Feedback
// button off; '' when there is none.
export function noFeedbackNote(teams) {
  return teams
    .map((t) => `<p class="note">${esc(t.name)} names no feedback repository in its manifest.yaml, so its Feedback buttons are off.</p>`)
    .join('');
}

// The issue address for the pressed button, looked up in the catalog the
// page was drawn from. Throws when the item or its repository is gone.
export function feedbackFor(catalog, { team, kind, name }) {
  const t = (catalog.teams || []).find((x) => x.name === team);
  const item = t && (kind === 'skill' ? t.skills : t.agents).find((i) => i.name === name);
  if (!item) throw new Error(`${name} is no longer in ${team}`);
  if (!t.feedback) throw new Error(`${team} names no feedback repository in its manifest.yaml`);
  return feedbackURL(t.feedback, team, kind, name, item.owner);
}

// Calls onFeedback({ team, kind, name }) when a Feedback button in section
// is pressed.
export function onFeedbackClick(section, onFeedback) {
  section.addEventListener('click', (e) => {
    const b = e.target.closest && e.target.closest('button[data-feedback]');
    if (b && !b.disabled) onFeedback({ team: b.dataset.team, kind: b.dataset.feedback, name: b.dataset.name });
  });
}

// The GitHub new-issue address a Feedback button opens: the team's feedback
// repository, prefilled with a title, a body to complete and the labels.
// kind is "agent" or "skill"; owner is a GitHub handle or '' for none, in
// which case the body mentions nobody.
export function feedbackURL(repo, team, kind, name, owner) {
  const label = kind === 'skill' ? 'Skill' : 'Agent';
  const lines = [`**Team:** ${team}`, `**${label}:** ${name}`];
  if (owner) lines.push(`**Owner:** ${owner}`);
  lines.push(
    '',
    '**What happened**',
    `<describe what the ${kind} did>`,
    '',
    '**What you expected**',
    '<describe what should have happened>',
    '',
    '_Sent from the kirokit desktop app._',
  );
  const labels = ['feedback', `team:${team}`, `${kind}:${name}`].map(encode).join(',');
  const title = encode(`Feedback: ${team} ${kind} ${name}`);
  const body = encode(lines.join('\n'));
  return `https://github.com/${repo}/issues/new?title=${title}&body=${body}&labels=${labels}`;
}

// encodeURIComponent, plus the characters it leaves alone (! ' ( ) * ~), which
// Wails v2 refused as shell metacharacters; v3 accepts them, and encoding them
// still keeps every issue address the same as before.
function encode(s) {
  return encodeURIComponent(s).replace(/[!'()*~]/g, (c) => '%' + c.charCodeAt(0).toString(16).toUpperCase());
}
