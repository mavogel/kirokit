package quick

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mavogel/kirokit/internal/externalskills"
	"github.com/mavogel/kirokit/internal/testutil"
)

const (
	idStory = "0123456789abcdef0123456789abcdef"
	idOther = "fedcba9876543210fedcba9876543210"
)

type fixture struct {
	teamsRoot, globalRoot, versions string
}

func writeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

const storySkill = "---\nname: user-story\ndescription: Write a user story with acceptance criteria.\n---\n\n# User story for {{ .Team.Team }}\n\nAsk for {{\"{{\"}}feature}}.\n"

// newTeam writes a team "t" with the given manifest quick section and a
// skill, an inline-prompt agent and a file-prompt agent.
func newTeam(t *testing.T, quick string, extra map[string]string) fixture {
	t.Helper()
	root := t.TempDir()
	f := fixture{
		teamsRoot:  filepath.Join(root, "teams"),
		globalRoot: filepath.Join(root, "global", "steerings"),
		versions:   filepath.Join(root, "env", "mcp-versions.yaml"),
	}
	files := map[string]string{
		"env/mcp-versions.yaml":                   "x: \"1\"\n",
		"global/steerings/.keep":                  "",
		"teams/t/manifest.yaml":                   "team: t\nglobalSteerings: []\n" + quick,
		"teams/t/skills/user-story/SKILL.md.tmpl": storySkill,
		"teams/t/skills/other/SKILL.md.tmpl":      "---\nname: other\ndescription: Another skill.\n---\nBody.\n",
		"teams/t/agents/po/agent.json.tmpl":       `{"name": "po", "description": "Product owner <lead> & friends.", "prompt": "You own the vision.\nBe kind.", "tools": ["read"]}`,
		"teams/t/agents/filer/agent.json.tmpl":    `{"name": "filer", "description": "Reads its prompt from a file.", "prompt": "file://../prompts/filer.md", "tools": []}`,
		"teams/t/prompts/filer.md.tmpl":           "Prompt of {{ .Team.Team }}.\n",
	}
	for k, v := range extra {
		files[k] = v
	}
	writeFiles(t, root, files)
	return f
}

// writeLock records the vendored skill name the way `kirokit skills fetch`
// would, with a hash that matches the files already written.
func writeLock(t *testing.T, f fixture, name string) {
	t.Helper()
	teamDir := filepath.Join(f.teamsRoot, "t")
	sum, err := externalskills.HashDir(filepath.Join(externalskills.SkillsDir(teamDir), name))
	if err != nil {
		t.Fatal(err)
	}
	lock := externalskills.Lock{Skills: []externalskills.LockEntry{{
		Name: name, Source: "https://example.com/r.git", Ref: "v1", Path: "skills/" + name, Commit: "abc", SHA256: sum,
	}}}
	if err := externalskills.WriteLock(externalskills.LockPath(teamDir), lock); err != nil {
		t.Fatal(err)
	}
}

func (f fixture) render(t *testing.T, ids map[string]string) (string, Result, error) {
	t.Helper()
	out := filepath.Join(t.TempDir(), "bundle")
	res, err := RenderTeam(f.teamsRoot, "t", f.globalRoot, f.versions, out, Options{IDs: ids})
	return out, res, err
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

const quickPO = `quick:
  skills: [user-story]
  agents:
    - name: po
      welcomeMessage: "What should we build next?"
      starterPrompts: ["Write a user story for <login>"]
      skills: [user-story]
`

func TestRenderTeam_FirstPassHasNoIDs(t *testing.T) {
	f := newTeam(t, quickPO, nil)
	out, res, err := f.render(t, nil)
	if err != nil {
		t.Fatalf("RenderTeam: %v", err)
	}
	if res.Skills != 1 || res.Agents != 1 || strings.Join(res.MissingIDs, ",") != "user-story" {
		t.Errorf("Result = %+v, want 1 skill, 1 agent, user-story missing", res)
	}

	wantSkill := "---\nname: user-story\ndescription: Write a user story with acceptance criteria.\n---\n\n# User story for t\n\nAsk for {{feature}}.\n"
	if got := read(t, filepath.Join(out, "skills", "user-story", "SKILL.md")); got != wantSkill {
		t.Errorf("SKILL.md =\n%s\nwant\n%s", got, wantSkill)
	}

	wantAgent := `{
  "name": "po",
  "description": "Product owner <lead> & friends.",
  "system_prompt": "You own the vision.\nBe kind.",
  "welcome_message": "What should we build next?",
  "starter_prompts": [
    "Write a user story for <login>"
  ],
  "capabilities": [],
  "connectors": [],
  "skills": [],
  "spaces": [],
  "files": []
}
`
	if got := read(t, filepath.Join(out, "agents", "po.json")); got != wantAgent {
		t.Errorf("po.json =\n%s\nwant\n%s", got, wantAgent)
	}

	// Only what the manifest lists is in the bundle.
	for _, p := range []string{"skills/other", "agents/filer.json", "prompts", "steering"} {
		if _, err := os.Stat(filepath.Join(out, p)); err == nil {
			t.Errorf("%s is in the bundle, want only listed content", p)
		}
	}
}

func TestRenderTeam_SecondPassWritesIDs(t *testing.T) {
	f := newTeam(t, quickPO, nil)
	out, res, err := f.render(t, map[string]string{"user-story": idStory, "unrelated": idOther})
	if err != nil {
		t.Fatalf("RenderTeam: %v", err)
	}
	if len(res.MissingIDs) != 0 {
		t.Errorf("MissingIDs = %v, want none", res.MissingIDs)
	}
	skill := read(t, filepath.Join(out, "skills", "user-story", "SKILL.md"))
	if !strings.HasPrefix(skill, "---\nname: user-story\ndescription: Write a user story with acceptance criteria.\nid: "+idStory+"\n---\n\n# User story") {
		t.Errorf("SKILL.md does not end its frontmatter with the id line:\n%s", skill)
	}
	agent := read(t, filepath.Join(out, "agents", "po.json"))
	if !strings.Contains(agent, "\"skills\": [\n    \""+idStory+"\"\n  ],") {
		t.Errorf("po.json does not pin the id:\n%s", agent)
	}
}

func TestRenderTeam_FilePromptAndDefaults(t *testing.T) {
	f := newTeam(t, "quick:\n  agents:\n    - name: filer\n", nil)
	out, res, err := f.render(t, nil)
	if err != nil {
		t.Fatalf("RenderTeam: %v", err)
	}
	if res.Skills != 0 || res.Agents != 1 {
		t.Errorf("Result = %+v", res)
	}
	got := read(t, filepath.Join(out, "agents", "filer.json"))
	for _, want := range []string{`"system_prompt": "Prompt of t.\n"`, `"welcome_message": ""`, `"starter_prompts": []`, `"skills": []`} {
		if !strings.Contains(got, want) {
			t.Errorf("filer.json lacks %s:\n%s", want, got)
		}
	}
}

func TestRenderTeam_VendoredSkillIsCopiedWhole(t *testing.T) {
	f := newTeam(t, "externalSkills:\n  - source: https://example.com/r.git\n    ref: v1\n    paths: [skills/vend]\nquick:\n  skills: [vend]\n", nil)
	vendDir := filepath.Join(f.teamsRoot, "t", "external", "skills", "vend")
	writeFiles(t, vendDir, map[string]string{
		"SKILL.md":        "---\nname: vend\ndescription: Vendored.\n---\nUse {{input}} as is.\n",
		"references/a.md": "reference\n",
	})
	writeLock(t, f, "vend")

	out, _, err := f.render(t, map[string]string{"vend": idStory})
	if err != nil {
		t.Fatalf("RenderTeam: %v", err)
	}
	if got := read(t, filepath.Join(out, "skills", "vend", "SKILL.md")); got != "---\nname: vend\ndescription: Vendored.\nid: "+idStory+"\n---\nUse {{input}} as is.\n" {
		t.Errorf("vend SKILL.md = %q", got)
	}
	if got := read(t, filepath.Join(out, "skills", "vend", "references", "a.md")); got != "reference\n" {
		t.Errorf("reference = %q", got)
	}
}

// The bundle keeps a vendored script's owner-execute bit, as the Kiro render does.
func TestRenderTeam_VendoredScriptStaysExecutable(t *testing.T) {
	testutil.SkipWithoutExecBits(t)
	f := newTeam(t, "externalSkills:\n  - source: https://example.com/r.git\n    ref: v1\n    paths: [skills/vend]\nquick:\n  skills: [vend]\n", nil)
	vendDir := filepath.Join(f.teamsRoot, "t", "external", "skills", "vend")
	writeFiles(t, vendDir, map[string]string{
		"SKILL.md":       "---\nname: vend\ndescription: Vendored.\n---\nRun scripts/run.sh.\n",
		"scripts/run.sh": "#!/bin/sh\necho ok\n",
	})
	if err := os.Chmod(filepath.Join(vendDir, "scripts", "run.sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeLock(t, f, "vend")

	out, _, err := f.render(t, nil)
	if err != nil {
		t.Fatalf("RenderTeam: %v", err)
	}
	for file, want := range map[string]os.FileMode{"scripts/run.sh": 0o700, "SKILL.md": 0o600} {
		info, err := os.Stat(filepath.Join(out, "skills", "vend", filepath.FromSlash(file)))
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != want {
			t.Errorf("%s mode = %v, want %v", file, got, want)
		}
	}
}

func TestRenderTeam_ReplacesAnIDTheSkillAlreadyCarries(t *testing.T) {
	f := newTeam(t, "quick:\n  skills: [other]\n", map[string]string{
		"teams/t/skills/other/SKILL.md.tmpl": "---\nname: other\nid: " + idOther + "\ndescription: Another skill.\n---\nBody.\n",
	})
	out, res, err := f.render(t, nil)
	if err != nil {
		t.Fatalf("RenderTeam: %v", err)
	}
	if got := read(t, filepath.Join(out, "skills", "other", "SKILL.md")); got != "---\nname: other\ndescription: Another skill.\n---\nBody.\n" {
		t.Errorf("SKILL.md = %q, want the foreign id dropped", got)
	}
	if strings.Join(res.MissingIDs, ",") != "other" {
		t.Errorf("MissingIDs = %v", res.MissingIDs)
	}
}

func TestRenderTeam_Errors(t *testing.T) {
	long := strings.Repeat("x", MaxPromptChars+1)
	for name, tc := range map[string]struct {
		quick string
		extra map[string]string
		ids   map[string]string
		want  string
	}{
		"no quick section":    {"", nil, nil, "lists nothing for Quick"},
		"empty quick section": {"quick: {}\n", nil, nil, "lists nothing for Quick"},
		"unknown skill":       {"quick:\n  skills: [nope]\n", nil, nil, `skill "nope"`},
		"unknown agent":       {"quick:\n  agents:\n    - name: nope\n", nil, nil, `agent "nope"`},
		"pin not listed":      {"quick:\n  skills: [other]\n  agents:\n    - name: po\n      skills: [user-story]\n", nil, nil, `"user-story" is not in quick.skills`},
		"skill without desc":  {"quick:\n  skills: [other]\n", map[string]string{"teams/t/skills/other/SKILL.md.tmpl": "---\nname: other\n---\nBody.\n"}, nil, "description"},
		"skill without fm":    {"quick:\n  skills: [other]\n", map[string]string{"teams/t/skills/other/SKILL.md.tmpl": "Body.\n"}, nil, "frontmatter"},
		"prompt leaves tree":  {"quick:\n  agents:\n    - name: filer\n", map[string]string{"teams/t/agents/filer/agent.json.tmpl": `{"name": "filer", "description": "d", "prompt": "file://../../secret.md", "tools": []}`}, nil, "leaves the rendered tree"},
		"absolute prompt":     {"quick:\n  agents:\n    - name: filer\n", map[string]string{"teams/t/agents/filer/agent.json.tmpl": `{"name": "filer", "description": "d", "prompt": "file:///etc/passwd", "tools": []}`}, nil, "leaves the rendered tree"},
		"prompt too long":     {"quick:\n  agents:\n    - name: po\n", map[string]string{"teams/t/agents/po/agent.json.tmpl": `{"name": "po", "description": "d", "prompt": "` + long + `", "tools": []}`}, nil, "50000"},
		"malformed id":        {quickPO, nil, map[string]string{"user-story": "ABC"}, "32 lowercase hex"},
		// Quick matches skills by name, so two folders with one name would overwrite each other.
		"two folders, one name": {"quick:\n  skills: [user-story, other]\n", map[string]string{"teams/t/skills/other/SKILL.md.tmpl": "---\nname: user-story\ndescription: Copy.\n---\nBody.\n"}, nil, `skills "user-story" and "other" share the name "user-story"`},
	} {
		t.Run(name, func(t *testing.T) {
			f := newTeam(t, tc.quick, tc.extra)
			out, _, err := f.render(t, tc.ids)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want it to contain %q", err, tc.want)
			}
			if _, statErr := os.Stat(out); statErr == nil {
				t.Errorf("a failed render wrote %s", out)
			}
		})
	}
}

func TestRenderTeam_NoQuickSectionIsErrNoQuick(t *testing.T) {
	f := newTeam(t, "", nil)
	if _, _, err := f.render(t, nil); !errors.Is(err, ErrNoQuick) {
		t.Errorf("err = %v, want ErrNoQuick", err)
	}
}

func TestRenderTeam_RerenderReplacesTheBundle(t *testing.T) {
	f := newTeam(t, quickPO, nil)
	out := filepath.Join(t.TempDir(), "bundle")
	for _, ids := range []map[string]string{nil, {"user-story": idStory}} {
		if _, err := RenderTeam(f.teamsRoot, "t", f.globalRoot, f.versions, out, Options{IDs: ids}); err != nil {
			t.Fatalf("RenderTeam: %v", err)
		}
	}
	if got := read(t, filepath.Join(out, "skills", "user-story", "SKILL.md")); !strings.Contains(got, "id: "+idStory) {
		t.Errorf("second render did not replace the bundle:\n%s", got)
	}
}
