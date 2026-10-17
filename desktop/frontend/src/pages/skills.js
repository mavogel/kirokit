import { esc, problem } from './html.js';

// Two tables: every skill with its team and source, and the central MCP
// version pins with the teams whose agents run that server.
export function skillsView(view) {
  const blocked = problem(view);
  if (blocked) return { blocked: true, skills: blocked, mcp: '' };

  const skillRows = [];
  const users = new Map();
  for (const t of view.catalog.teams || []) {
    for (const s of t.skills) {
      skillRows.push(`<tr><td class="mono">${esc(s.name)}</td><td>${esc(t.name)}</td><td class="desc">${esc(s.source)}</td></tr>`);
    }
    for (const a of t.agents) {
      for (const pin of a.pins || []) {
        if (!users.has(pin)) users.set(pin, new Set());
        users.get(pin).add(t.name);
      }
    }
  }
  const pins = Object.entries(view.catalog.pins || {}).sort(([a], [b]) => a.localeCompare(b));
  const mcpRows = pins.map(([server, version]) => {
    const by = [...(users.get(server) || [])].sort().join(', ');
    return `<tr><td class="mono">${esc(server)}</td><td class="mono">${esc(version)}</td><td>${by ? esc(by) : '<span class="note">no agent yet</span>'}</td></tr>`;
  });
  return {
    blocked: false,
    skills: skillRows.join('\n') || '<tr><td colspan="3" class="note">No skills yet.</td></tr>',
    mcp: mcpRows.join('\n') || '<tr><td colspan="3" class="note">No pinned versions.</td></tr>',
  };
}

export function initSkills(section) {
  const tables = [...section.querySelectorAll('table')];
  const bodies = section.querySelectorAll('tbody');
  const headings = [...section.querySelectorAll('h2')];
  const message = document.createElement('div');
  tables[tables.length - 1].after(message);
  return (view) => {
    const out = skillsView(view);
    for (const el of [...tables, ...headings]) el.hidden = out.blocked;
    message.innerHTML = out.blocked ? out.skills : '';
    if (!out.blocked) {
      bodies[0].innerHTML = out.skills;
      bodies[1].innerHTML = out.mcp;
    }
  };
}
