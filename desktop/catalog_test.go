package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/mavogel/kirokit/internal/externalskills"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func findTeam(t *testing.T, c Catalog, name string) Team {
	t.Helper()
	for _, tm := range c.Teams {
		if tm.Name == name {
			return tm
		}
	}
	t.Fatalf("catalog has no team %q: %+v", name, c.Teams)
	return Team{}
}

func findAgent(t *testing.T, tm Team, name string) Agent {
	t.Helper()
	for _, a := range tm.Agents {
		if a.Name == name {
			return a
		}
	}
	t.Fatalf("team %s has no agent %q: %+v", tm.Name, name, tm.Agents)
	return Agent{}
}

// The repository's own two teams are the fixture: names, descriptions, tools
// and servers come from the rendered agent files, not from the page.
func TestCatalog_ReadsRealTeams(t *testing.T) {
	cat, err := BuildCatalog("..")
	if err != nil {
		t.Fatalf("BuildCatalog: %v", err)
	}

	names := []string{}
	for _, tm := range cat.Teams {
		names = append(names, tm.Name)
	}
	if !slices.Equal(names, []string{"data-team", "platform-team", "product-team"}) {
		t.Fatalf("teams = %v, want data-team, platform-team and product-team", names)
	}

	product := findTeam(t, cat, "product-team")
	roles := []string{}
	for _, a := range product.Agents {
		roles = append(roles, a.Name)
		if !slices.Contains(a.Tools, "subagent") || a.NeedsKey || len(a.Servers) != 0 {
			t.Errorf("product-team/%s tools = %v, needsKey = %v, servers = %v; want subagent, no key, no servers", a.Name, a.Tools, a.NeedsKey, a.Servers)
		}
	}
	if !slices.Equal(roles, []string{"customer", "engineer", "product-owner", "qa", "ux-designer"}) {
		t.Errorf("product-team agents = %v, want the five roles", roles)
	}

	devops := findAgent(t, findTeam(t, cat, "platform-team"), "devops")
	if !strings.HasPrefix(devops.Description, "Triages and resolves Jira tickets") {
		t.Errorf("devops description = %q", devops.Description)
	}
	if !slices.Contains(devops.Tools, "shell") || !slices.Contains(devops.Tools, "@atlassian") {
		t.Errorf("devops tools = %v, want shell and @atlassian", devops.Tools)
	}
	if !slices.Equal(devops.Servers, []string{"atlassian", "context7"}) {
		t.Errorf("devops servers = %v, want atlassian and context7", devops.Servers)
	}
	if cat.Pins["jira-mcp"] != "1.4.2" {
		t.Errorf("Pins[jira-mcp] = %q, want 1.4.2 from env/mcp-versions.yaml", cat.Pins["jira-mcp"])
	}
}

// Pins lists the central version pins an agent's template resolves, which is
// not the same as its server names: devops runs "atlassian" through the
// pinned mcp-remote proxy and "context7" through the pinned context7-mcp.
func TestCatalog_AgentPinsFollowMcpVersionCalls(t *testing.T) {
	cat, err := BuildCatalog("..")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]string{
		"platform-team/devops":      {"context7-mcp", "mcp-remote"},
		"platform-team/platformdev": {"confluence-mcp"},
		"data-team/javadev":         {"snowflake-mcp"},
		"data-team/godev":           {"pipeline-mcp"},
	}
	for key, pins := range want {
		team, agent, _ := strings.Cut(key, "/")
		if got := findAgent(t, findTeam(t, cat, team), agent).Pins; !slices.Equal(got, pins) {
			t.Errorf("%s Pins = %v, want %v", key, got, pins)
		}
	}
}

// NeedsKey follows the encrypted secrets file in the source tree.
func TestCatalog_NeedsKeyExactlyForShippedSecrets(t *testing.T) {
	cat, err := BuildCatalog("..")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"platform-team/devops": true, "platform-team/platformdev": true, "data-team/javadev": true, "data-team/godev": false}
	for key, needs := range want {
		team, agent, _ := strings.Cut(key, "/")
		if got := findAgent(t, findTeam(t, cat, team), agent).NeedsKey; got != needs {
			t.Errorf("%s NeedsKey = %v, want %v", key, got, needs)
		}
	}
}

func fixtureRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write(t, filepath.Join(root, "env", "mcp-versions.yaml"), "tool-mcp: \"2.0.0\"\n")
	write(t, filepath.Join(root, "teams", "t1", "manifest.yaml"), "team: t1\nglobalSteerings: []\n")
	write(t, filepath.Join(root, "teams", "t1", "agents", "a1", "agent.json.tmpl"),
		`{"name": "a1", "description": "d", "tools": ["read"], "mcpServers": {"tool-mcp": {"command": "x", "args": ["{{ mcpVersion "tool-mcp" }}"]}}}`)
	write(t, filepath.Join(root, "teams", "t1", "skills", "own-skill", "SKILL.md"), "---\nname: own-skill\n---\n")
	return root
}

// lockSkill declares a vendored skill in the team's manifest and lock, with
// the real hash of its files: render refuses a lock that does not match.
func lockSkill(t *testing.T, root, team, name, source, ref string) {
	t.Helper()
	teamDir := filepath.Join(root, "teams", team)
	write(t, filepath.Join(teamDir, "manifest.yaml"),
		"team: "+team+"\nglobalSteerings: []\nexternalSkills:\n  - source: "+source+"\n    ref: "+ref+"\n    paths:\n      - skills/"+name+"\n")
	sum, err := externalskills.HashDir(filepath.Join(externalskills.SkillsDir(teamDir), name))
	if err != nil {
		t.Fatal(err)
	}
	lock := externalskills.Lock{Skills: []externalskills.LockEntry{{Name: name, Source: source, Ref: ref, Path: "skills/" + name, Commit: "abc123", SHA256: sum}}}
	if err := externalskills.WriteLock(externalskills.LockPath(teamDir), lock); err != nil {
		t.Fatal(err)
	}
}

func TestCatalog_SkillSources(t *testing.T) {
	root := fixtureRepo(t)
	write(t, filepath.Join(root, "teams", "t1", "external", "skills", "ext-skill", "SKILL.md"), "---\nname: ext-skill\n---\n")
	lockSkill(t, root, "t1", "ext-skill", "https://example.com/skills.git", "v1.2.0")

	cat, err := BuildCatalog(root)
	if err != nil {
		t.Fatalf("BuildCatalog: %v", err)
	}
	got := map[string]string{}
	for _, s := range findTeam(t, cat, "t1").Skills {
		got[s.Name] = s.Source
	}
	want := map[string]string{"own-skill": "Written by the team", "ext-skill": "https://example.com/skills.git @ v1.2.0"}
	for name, src := range want {
		if got[name] != src {
			t.Errorf("skill %s source = %q, want %q (all: %v)", name, got[name], src, got)
		}
	}
}

// A template that cannot render reaches the page as an error with its
// message; nothing panics and no half-built table is returned.
func TestCatalog_RenderErrorCarriesMessage(t *testing.T) {
	root := fixtureRepo(t)
	write(t, filepath.Join(root, "teams", "t1", "agents", "a1", "agent.json.tmpl"), `{"v": "{{ mcpVersion "undefined-key" }}"}`)

	_, err := BuildCatalog(root)

	if err == nil || !strings.Contains(err.Error(), "undefined-key") || !strings.Contains(err.Error(), "t1") {
		t.Errorf("err = %v, want one naming team t1 and the undefined key", err)
	}
}

// Catalog() serves the cached result until a clone or sync ends.
func TestCatalog_CachedUntilRefresh(t *testing.T) {
	a := NewAppAt(t.TempDir())
	root := fixtureRepo(t)
	if err := os.Rename(root, a.repoDir()); err != nil {
		t.Fatal(err)
	}

	first, err := a.Catalog()
	if err != nil || len(first.Teams) != 1 {
		t.Fatalf("Catalog = %+v, %v; want one team", first, err)
	}
	write(t, filepath.Join(a.repoDir(), "teams", "t2", "manifest.yaml"), "team: t2\nglobalSteerings: []\n")

	again, _ := a.Catalog()
	if len(again.Teams) != 1 {
		t.Errorf("a second call re-read the clone: %d teams, want the cached 1", len(again.Teams))
	}
	a.refreshCatalog()
	fresh, _ := a.Catalog()
	if len(fresh.Teams) != 2 {
		t.Errorf("after refreshCatalog %d teams, want 2", len(fresh.Teams))
	}
}

func TestCatalog_NoCloneIsEmptyNotAnError(t *testing.T) {
	cat, err := NewAppAt(t.TempDir()).Catalog()
	if err != nil || len(cat.Teams) != 0 {
		t.Errorf("Catalog = %+v, %v; want empty and no error before a clone", cat, err)
	}
}
