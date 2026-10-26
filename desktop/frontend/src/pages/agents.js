import { esc, chips, problem } from './html.js';
import { ownerCell, feedbackCell, noFeedbackNote, onFeedbackClick } from './feedback.js';

// The rows of the agents table, or the message that replaces it, and the note
// under the table for teams whose Feedback buttons are off.
export function agentsView(view) {
  const blocked = problem(view);
  if (blocked) return { blocked: true, html: blocked, note: '' };
  const rows = [];
  const off = [];
  for (const t of view.catalog.teams || []) {
    if (t.agents.length && !t.feedback) off.push(t);
    for (const a of t.agents) {
      rows.push(`<tr>
  <td class="mono">${esc(a.name)}</td>
  <td>${esc(t.name)}</td>
  <td class="desc">${esc(a.description)}</td>
  <td>${chips(a.tools)}${chips(a.servers.map((s) => '@' + s).filter((s) => !a.tools.includes(s)))}</td>
  <td>${a.needsKey ? '<span class="badge warn">needs key</span>' : '<span class="badge off">none</span>'}</td>
  ${ownerCell(a)}
  ${feedbackCell(view, t, 'agent', a)}
</tr>`);
    }
  }
  if (!rows.length) return { blocked: true, html: '<p class="note">No agents yet.</p>', note: '' };
  return { blocked: false, html: rows.join('\n'), note: noFeedbackNote(off) };
}

// onFeedback({ team, kind, name }) runs when a row's Feedback button is pressed.
export function initAgents(section, onFeedback) {
  const body = section.querySelector('tbody');
  const table = section.querySelector('table');
  const message = document.createElement('div');
  table.after(message);
  onFeedbackClick(section, onFeedback);
  return (view) => {
    const out = agentsView(view);
    table.hidden = out.blocked;
    message.innerHTML = out.blocked ? out.html : out.note;
    if (!out.blocked) body.innerHTML = out.html;
  };
}
