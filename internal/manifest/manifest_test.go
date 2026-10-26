package manifest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFixture(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}

func TestLoad_ValidManifest(t *testing.T) {
	path := writeFixture(t, "team: platform-team\nglobalSteerings:\n  - security-baseline\n")

	m, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if m.Team != "platform-team" {
		t.Errorf("Team = %q, want %q", m.Team, "platform-team")
	}
	if len(m.GlobalSteerings) != 1 || m.GlobalSteerings[0] != "security-baseline" {
		t.Errorf("GlobalSteerings = %v, want [security-baseline]", m.GlobalSteerings)
	}
}

func TestLoad_EmptyGlobalSteeringsIsValidOptOut(t *testing.T) {
	path := writeFixture(t, "team: data-team\nglobalSteerings: []\n")

	m, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(m.GlobalSteerings) != 0 {
		t.Errorf("GlobalSteerings = %v, want empty", m.GlobalSteerings)
	}
}

func TestLoad_MissingTeamErrors(t *testing.T) {
	path := writeFixture(t, "globalSteerings: []\n")

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load returned no error for missing team field")
	}
}

func TestLoad_ExternalSkills(t *testing.T) {
	path := writeFixture(t, `team: t
externalSkills:
  - source: https://github.com/anthropics/skills.git
    ref: abc123
    paths: [skills/pdf]
  - source: git@github.com:acme/private.git
    ref: v1.2.0
    paths: [team/sql]
    name: warehouse-sql
`)

	m, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(m.ExternalSkills) != 2 {
		t.Fatalf("ExternalSkills = %v, want 2 entries", m.ExternalSkills)
	}
	if got := m.ExternalSkills[0].EffectiveName(); got != "pdf" {
		t.Errorf("EffectiveName() = %q, want %q", got, "pdf")
	}
	if got := m.ExternalSkills[1].EffectiveName(); got != "warehouse-sql" {
		t.Errorf("EffectiveName() with name override = %q, want %q", got, "warehouse-sql")
	}
}

func TestLoad_ExternalSkillsAcceptsKeyOnlySSHSources(t *testing.T) {
	for _, source := range []string{"ssh://git@host/r.git", "git@host:org/r.git", "https://host/r.git"} {
		path := writeFixture(t, "team: t\nexternalSkills:\n  - source: "+source+"\n    ref: main\n    paths: [skills/x]\n")
		if _, err := Load(path); err != nil {
			t.Errorf("source %q rejected: %v", source, err)
		}
	}
}

func TestLoad_ExternalSkillsRejectsInvalidEntries(t *testing.T) {
	entry := func(source, ref, p, name string) string {
		s := "team: t\nexternalSkills:\n  - source: " + source + "\n    ref: " + ref + "\n    paths: [" + p + "]\n"
		if name != "" {
			s += "    name: " + name + "\n"
		}
		return s
	}
	tests := []struct {
		name, yaml, wantInErr string
	}{
		{"missing source", entry("''", "main", "skills/x", ""), "source"},
		{"missing ref", entry("https://h/r.git", "''", "skills/x", ""), "ref"},
		{"missing path", entry("https://h/r.git", "main", "''", ""), "path"},
		{"parent path", entry("https://h/r.git", "main", "../x", ""), "path"},
		{"absolute path", entry("https://h/r.git", "main", "/etc/x", ""), "path"},
		{"dot path", entry("https://h/r.git", "main", ".", ""), "path"},
		{"name with slash", entry("https://h/r.git", "main", "skills/x", "a/b"), "name"},
		{"user and password", entry("https://u:p@h/r.git", "main", "skills/x", ""), "credentials"},
		{"bare token userinfo", entry("https://TOKEN@h/r.git", "main", "skills/x", ""), "credentials"},
		{"ssh password", entry("ssh://u:p@h/r.git", "main", "skills/x", ""), "credentials"},
		{"duplicate names", entry("https://h/r.git", "main", "skills/x", "") +
			"  - source: https://h/o.git\n    ref: main\n    paths: [other/x]\n", "duplicate"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Load(writeFixture(t, tc.yaml))
			if err == nil {
				t.Fatalf("Load accepted %q", tc.yaml)
			}
			if !strings.Contains(err.Error(), tc.wantInErr) {
				t.Errorf("err = %v, want it to mention %q", err, tc.wantInErr)
			}
			if strings.Contains(err.Error(), "TOKEN") || strings.Contains(err.Error(), ":p@") {
				t.Errorf("err = %v leaks the credential", err)
			}
		})
	}
}

func TestLoad_ExternalSkillsRejectsGitOptionAndTransportInjection(t *testing.T) {
	entry := func(source, ref string) string {
		return "team: t\nexternalSkills:\n  - source: '" + source + "'\n    ref: '" + ref + "'\n    paths: [skills/x]\n"
	}
	tests := []struct {
		name, source, ref, wantInErr string
	}{
		{"source that git would parse as an option", "--upload-pack=touch /tmp/pwned", "main", "source"},
		{"source starting with a dash", "-oProxyCommand=x", "main", "source"},
		{"ext transport helper", "ext::sh -c touch", "main", "source"},
		{"fd transport helper", "fd::3", "main", "source"},
		{"non-git URL scheme", "ftp://host/r.git", "main", "source"},
		{"ref that git would parse as an option", "https://host/r.git", "--upload-pack=x", "ref"},
		{"ref with whitespace", "https://host/r.git", "main branch", "ref"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Load(writeFixture(t, entry(tc.source, tc.ref)))
			if err == nil {
				t.Fatalf("Load accepted source %q ref %q", tc.source, tc.ref)
			}
			if !strings.Contains(err.Error(), tc.wantInErr) {
				t.Errorf("err = %v, want it to mention %q", err, tc.wantInErr)
			}
		})
	}
}

func TestLoad_ExternalSkillsAcceptsLocalPathsAndCommitRefs(t *testing.T) {
	path := writeFixture(t, "team: t\nexternalSkills:\n  - source: /srv/git/skills.git\n    ref: 0123456789abcdef0123456789abcdef01234567\n    paths: [skills/x]\n  - source: ../sibling\n    ref: release/1.2\n    paths: [skills/y]\n")
	if _, err := Load(path); err != nil {
		t.Fatalf("Load rejected local-path sources or ordinary refs: %v", err)
	}
}

func TestLoad_ExternalSkillsCollectionEntry(t *testing.T) {
	m, err := Load(writeFixture(t, `team: t
externalSkills:
  - source: https://github.com/anthropics/skills.git
    ref: abc123
    paths: [skills]
    all: true
    exclude: [docx, pdf]
  - source: https://github.com/anthropics/skills.git
    ref: abc123
    paths: [skills/skill-creator]
`))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	collection, single := m.ExternalSkills[0], m.ExternalSkills[1]
	if !collection.IsCollection() || single.IsCollection() {
		t.Errorf("IsCollection() = %v, %v; want true, false", collection.IsCollection(), single.IsCollection())
	}
	if len(collection.Exclude) != 2 || collection.Exclude[0] != "docx" {
		t.Errorf("Exclude = %v, want [docx pdf]", collection.Exclude)
	}
}

func TestLoad_ExternalSkillsCollectionDoesNotClashWithASingleEntryNamedLikeItsPath(t *testing.T) {
	// A collection at path "skills" has no skill name of its own, so a single
	// entry whose directory is also called "skills" is not a duplicate.
	path := writeFixture(t, `team: t
externalSkills:
  - {source: "https://h/r.git", ref: main, paths: [skills], all: true}
  - {source: "https://h/o.git", ref: main, paths: [other/skills]}
`)
	if _, err := Load(path); err != nil {
		t.Fatalf("Load rejected a collection beside a single skill: %v", err)
	}
}

func TestLoad_ExternalSkillsRejectsInvalidCollectionEntries(t *testing.T) {
	tests := []struct {
		name, entry, wantInErr string
	}{
		{"all with a name", `{source: "https://h/r.git", ref: main, paths: [skills], all: true, name: mine}`, "name"},
		{"exclude without all", `{source: "https://h/r.git", ref: main, paths: [skills/x], exclude: [y]}`, "exclude"},
		{"exclude with a slash", `{source: "https://h/r.git", ref: main, paths: [skills], all: true, exclude: ["a/b"]}`, "exclude"},
		{"empty exclude element", `{source: "https://h/r.git", ref: main, paths: [skills], all: true, exclude: [""]}`, "exclude"},
		{"collection at the repository root", `{source: "https://h/r.git", ref: main, paths: ["."], all: true}`, "path"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Load(writeFixture(t, "team: t\nexternalSkills:\n  - "+tc.entry+"\n"))
			if err == nil {
				t.Fatalf("Load accepted %s", tc.entry)
			}
			if !strings.Contains(err.Error(), tc.wantInErr) || !strings.Contains(err.Error(), "externalSkills[0]") {
				t.Errorf("err = %v, want it to name externalSkills[0] and mention %q", err, tc.wantInErr)
			}
		})
	}
}

func TestLoad_ExternalSkillsPathsEntryExpandsToOneSkillPerPath(t *testing.T) {
	m, err := Load(writeFixture(t, `team: t
externalSkills:
  - source: https://github.com/anthropics/skills.git
    ref: abc123
    paths: [skills/pdf, skills/xlsx]
  - source: https://h/o.git
    ref: main
    paths: [skills/other]
`))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	var got []string
	for _, e := range m.ExternalSkills {
		if e.Source == "" || e.Ref == "" || len(e.Paths) != 0 {
			t.Errorf("entry %+v, want source and ref set and paths expanded away", e)
		}
		got = append(got, e.Source+" "+e.Ref+" "+e.Path+" "+e.EffectiveName())
	}
	want := []string{
		"https://github.com/anthropics/skills.git abc123 skills/pdf pdf",
		"https://github.com/anthropics/skills.git abc123 skills/xlsx xlsx",
		"https://h/o.git main skills/other other",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("entries = %q, want %q", got, want)
	}
}

func TestLoad_ExternalSkillsRejectsInvalidPathsEntries(t *testing.T) {
	tests := []struct {
		name, entry, wantInErr string
	}{
		{"legacy path key", `{source: "https://h/r.git", ref: main, path: skills/a}`, "paths"},
		{"name with several paths", `{source: "https://h/r.git", ref: main, paths: [skills/a, skills/b], name: mine}`, "name"},
		{"exclude with several paths", `{source: "https://h/r.git", ref: main, paths: [skills, more], all: true, exclude: [x]}`, "exclude"},
		{"escaping path", `{source: "https://h/r.git", ref: main, paths: [skills/a, ../b]}`, "path"},
		{"empty path", `{source: "https://h/r.git", ref: main, paths: [""]}`, "path"},
		{"duplicate skill name", `{source: "https://h/r.git", ref: main, paths: [a/pdf, b/pdf]}`, "pdf"},
		{"bad source", `{source: "-oops", ref: main, paths: [skills/a]}`, "source"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Load(writeFixture(t, "team: t\nexternalSkills:\n  - "+tc.entry+"\n"))
			if err == nil {
				t.Fatalf("Load accepted %s", tc.entry)
			}
			if !strings.Contains(err.Error(), tc.wantInErr) || !strings.Contains(err.Error(), "externalSkills[0]") {
				t.Errorf("err = %v, want it to name externalSkills[0] and mention %q", err, tc.wantInErr)
			}
		})
	}
}

func TestLoad_ExternalSkillsAllExpandsEachPathToACollection(t *testing.T) {
	m, err := Load(writeFixture(t, `team: t
externalSkills:
  - source: https://h/r.git
    ref: main
    all: true
    paths:
      - skills
      - extra/skills
`))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(m.ExternalSkills) != 2 || !m.ExternalSkills[0].IsCollection() || !m.ExternalSkills[1].IsCollection() ||
		m.ExternalSkills[0].Path != "skills" || m.ExternalSkills[1].Path != "extra/skills" {
		t.Errorf("entries = %+v, want two collections at skills and extra/skills", m.ExternalSkills)
	}
}

func TestLoad_OwnersAndFeedback(t *testing.T) {
	path := writeFixture(t, `team: t
owner: "@jane"
owners:
  agents:
    qa: "@acme/qa-team"
  skills:
    jira: "@bob"
feedback:
  github: acme/kiro-config
`)
	m, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if m.Owner != "@jane" || m.Feedback.GitHub != "acme/kiro-config" {
		t.Errorf("Owner = %q, Feedback = %q", m.Owner, m.Feedback.GitHub)
	}
	cases := []struct {
		kind, name, owner string
		fromTeam          bool
	}{
		{"agent", "qa", "@acme/qa-team", false},
		{"skill", "jira", "@bob", false},
		{"agent", "engineer", "@jane", true},
		{"skill", "vendored", "@jane", true},
		// An agent override does not leak to a skill of the same name.
		{"skill", "qa", "@jane", true},
	}
	for _, c := range cases {
		owner, fromTeam := m.OwnerOf(c.kind, c.name)
		if owner != c.owner || fromTeam != c.fromTeam {
			t.Errorf("OwnerOf(%s, %s) = %q, %v; want %q, %v", c.kind, c.name, owner, fromTeam, c.owner, c.fromTeam)
		}
	}
}

func TestLoad_NoOwnerIsValidAndOwnsNothing(t *testing.T) {
	m, err := Load(writeFixture(t, "team: t\nowners:\n  agents:\n    qa: \"@x\"\n"))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if owner, fromTeam := m.OwnerOf("agent", "other"); owner != "" || fromTeam {
		t.Errorf("OwnerOf without a team owner = %q, %v; want empty", owner, fromTeam)
	}
	if owner, _ := m.OwnerOf("agent", "qa"); owner != "@x" {
		t.Errorf("override = %q, want @x", owner)
	}
}

func TestLoad_RejectsMalformedOwnersAndFeedback(t *testing.T) {
	cases := []struct {
		yaml, want string
	}{
		{"owner: jane", `owner: "jane" is not a GitHub handle (write @jane)`},
		{`owner: "@"`, `owner: "@" is not a GitHub handle`},
		{`owner: "@-x"`, `owner: "@-x" is not a GitHub handle`},
		{`owner: "@a b"`, `owner: "@a b" is not a GitHub handle`},
		{"owners:\n  agents:\n    qa: jane", `owners.agents.qa: "jane" is not a GitHub handle (write @jane)`},
		{"owners:\n  skills:\n    s: \"@a/\"", `owners.skills.s: "@a/" is not a GitHub handle`},
		{"feedback:\n  github: owner", `feedback.github: "owner" is not a GitHub repository (write owner/repo)`},
		{"feedback:\n  github: a/b/c", `feedback.github: "a/b/c" is not a GitHub repository`},
		{"feedback:\n  github: https://github.com/a/b", `feedback.github: "https://github.com/a/b" is not a GitHub repository`},
		{"feedback:\n  github: \"javascript:x\"", `feedback.github: "javascript:x" is not a GitHub repository`},
		{"feedback:\n  github: a/..", `feedback.github: "a/.." is not a GitHub repository`},
	}
	for _, c := range cases {
		path := writeFixture(t, "team: t\n"+c.yaml+"\n")
		_, err := Load(path)
		if err == nil || !strings.Contains(err.Error(), c.want) || !strings.Contains(err.Error(), path) {
			t.Errorf("Load(%q) err = %v, want it to name %s and contain %q", c.yaml, err, path, c.want)
		}
	}
	for _, ok := range []string{`owner: "@jane"`, `owner: "@acme/platform-leads"`, `owner: "@a1-b"`, "feedback:\n  github: acme/kiro.config_1"} {
		if _, err := Load(writeFixture(t, "team: t\n"+ok+"\n")); err != nil {
			t.Errorf("Load(%q) = %v, want ok", ok, err)
		}
	}
}

// GitHub logins are at most 39 characters; a handle one longer is refused.
// An override written as an empty string names nobody by mistake, so it fails
// like any other malformed handle. A repository named "." would let the issue
// address step out of owner/repo.
func TestLoad_OwnerHandleBoundariesAndEmptyOverride(t *testing.T) {
	login39 := "a" + strings.Repeat("b", 38)
	if _, err := Load(writeFixture(t, "team: t\nowner: \"@"+login39+"\"\n")); err != nil {
		t.Errorf("39-character login: %v, want ok", err)
	}
	for _, c := range []struct{ yaml, want string }{
		{"owner: \"@" + login39 + "c\"", `owner: "@` + login39 + `c" is not a GitHub handle`},
		{"owners:\n  agents:\n    qa: \"\"", `owners.agents.qa: "" is not a GitHub handle`},
		{"feedback:\n  github: a/.", `feedback.github: "a/." is not a GitHub repository`},
	} {
		_, err := Load(writeFixture(t, "team: t\n"+c.yaml+"\n"))
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("Load(%q) err = %v, want it to contain %q", c.yaml, err, c.want)
		}
	}
}
