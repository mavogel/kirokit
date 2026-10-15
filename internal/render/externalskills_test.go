package render

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/mavogel/kiro-cli-kit/internal/externalskills"
)

const (
	externalSkillURI = "skill:///kh/skills/pdf/SKILL.md"
	manifestWithPDF  = "team: fixture-team\nglobalSteerings:\n  - security-baseline\nexternalSkills:\n  - source: https://h/r.git\n    ref: v1\n    path: skills/pdf\n"
)

// vendorPDF declares skills/pdf in the fixture team's manifest and vendors it
// the way `kirokit skills fetch` would (files, plus a lock entry whose hash
// matches), without touching the network.
func vendorPDF(t *testing.T, teamsRoot string) (teamDir string) {
	t.Helper()
	teamDir = filepath.Join(teamsRoot, "fixture-team")
	write := func(rel, content string, mode os.FileMode) {
		t.Helper()
		p := filepath.Join(externalskills.SkillsDir(teamDir), "pdf", filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), mode); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(p, mode); err != nil {
			t.Fatal(err)
		}
	}
	write("SKILL.md", "---\nname: pdf\ndescription: PDFs.\n---\n", 0o644)
	write("scripts/run.sh", "#!/bin/sh\necho {{ not a template }}\n", 0o755)
	write("notes.md.tmpl", "literal {{ .Nope }} stays\n", 0o644)

	if err := os.WriteFile(filepath.Join(teamDir, "manifest.yaml"), []byte(manifestWithPDF), 0o600); err != nil {
		t.Fatal(err)
	}
	sum, err := externalskills.HashDir(filepath.Join(externalskills.SkillsDir(teamDir), "pdf"))
	if err != nil {
		t.Fatal(err)
	}
	lock := externalskills.Lock{Skills: []externalskills.LockEntry{{
		Name: "pdf", Source: "https://h/r.git", Ref: "v1", Path: "skills/pdf", Commit: "abc", SHA256: sum,
	}}}
	if err := externalskills.WriteLock(externalskills.LockPath(teamDir), lock); err != nil {
		t.Fatal(err)
	}
	return teamDir
}

func TestRenderTeam_VendoredExternalSkillIsCopiedVerbatim(t *testing.T) {
	teamsRoot, globalRoot, envVersionsPath := setupFixture(t)
	teamDir := vendorPDF(t, teamsRoot)

	outDir, _, _ := renderedResources(t, teamsRoot, globalRoot, envVersionsPath, "agent1")

	for _, rel := range []string{"SKILL.md", "scripts/run.sh", "notes.md.tmpl"} {
		want, err := os.ReadFile(filepath.Join(externalskills.SkillsDir(teamDir), "pdf", filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(filepath.Join(outDir, "skills", "pdf", filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("expected skills/pdf/%s in the output: %v", rel, err)
		}
		if string(got) != string(want) {
			t.Errorf("skills/pdf/%s = %q, want the vendored bytes %q (never template-rendered)", rel, got, want)
		}
	}
	if _, err := os.Stat(filepath.Join(outDir, "skills", "pdf", "notes.md")); err == nil {
		t.Error("a vendored *.tmpl file was rendered and renamed; it must be copied as is")
	}

	script, err := os.Stat(filepath.Join(outDir, "skills", "pdf", "scripts", "run.sh"))
	if err != nil || script.Mode().Perm()&0o100 == 0 {
		t.Errorf("rendered script mode = %v (%v), want owner-executable", script.Mode().Perm(), err)
	}
	plain, err := os.Stat(filepath.Join(outDir, "skills", "pdf", "SKILL.md"))
	if err != nil || plain.Mode().Perm() != 0o600 {
		t.Errorf("rendered SKILL.md mode = %v (%v), want 600", plain.Mode().Perm(), err)
	}
	for _, leaked := range []string{"external", "skills.lock.yaml"} {
		if _, err := os.Stat(filepath.Join(outDir, leaked)); err == nil {
			t.Errorf("%s leaked into the output tree", leaked)
		}
	}
}

func TestRenderTeam_ExternalSkillIsWiredIntoEveryAgentExactlyOnce(t *testing.T) {
	teamsRoot, globalRoot, envVersionsPath := setupFixture(t)
	vendorPDF(t, teamsRoot)
	agent2 := filepath.Join(teamsRoot, "fixture-team", "agents", "agent2", "agent.json.tmpl")
	if err := os.MkdirAll(filepath.Dir(agent2), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(agent2, []byte(`{"name": "agent2", "resources": ["`+externalSkillURI+`"]}`), 0o600); err != nil {
		t.Fatal(err)
	}

	_, res1, _ := renderedResources(t, teamsRoot, globalRoot, envVersionsPath, "agent1")
	_, res2, _ := renderedResources(t, teamsRoot, globalRoot, envVersionsPath, "agent2")

	if want := []string{fixtureSteeringURI, externalSkillURI, workspaceSteeringURI, workspaceSkillsURI}; !slices.Equal(res1, want) {
		t.Errorf("agent1 resources = %v, want %v (global steering, external skill, workspace entries)", res1, want)
	}
	if got := strings.Count(strings.Join(res2, "\n"), externalSkillURI); got != 1 {
		t.Errorf("agent2 lists the external skill %d times (%v), want exactly once", got, res2)
	}
}

func TestRenderTeam_ExternalSkillDriftFailsTheRender(t *testing.T) {
	tests := []struct {
		name      string
		break_    func(t *testing.T, teamDir string)
		wantInErr []string
	}{
		{"vendored file modified", func(t *testing.T, d string) {
			p := filepath.Join(externalskills.SkillsDir(d), "pdf", "SKILL.md")
			if err := os.WriteFile(p, []byte("tampered"), 0o644); err != nil {
				t.Fatal(err)
			}
		}, []string{`"pdf"`, "does not match the lock file"}},
		{"manifest entry without lock entry", func(t *testing.T, d string) {
			if err := externalskills.WriteLock(externalskills.LockPath(d), externalskills.Lock{}); err != nil {
				t.Fatal(err)
			}
		}, []string{`"pdf"`, "missing from the lock file"}},
		{"vendored directory without manifest entry", func(t *testing.T, d string) {
			stray := filepath.Join(externalskills.SkillsDir(d), "stray", "SKILL.md")
			if err := os.MkdirAll(filepath.Dir(stray), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(stray, []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
		}, []string{`"stray"`, "no entry in manifest.yaml"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teamsRoot, globalRoot, envVersionsPath := setupFixture(t)
			tc.break_(t, vendorPDF(t, teamsRoot))

			err := RenderTeam(teamsRoot, "fixture-team", globalRoot, envVersionsPath, filepath.Join(t.TempDir(), "out"), "/kh")
			if err == nil {
				t.Fatal("RenderTeam succeeded, want a drift error")
			}
			for _, want := range tc.wantInErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("err = %q, want it to contain %q", err, want)
				}
			}
		})
	}
}

func TestRenderTeam_ExternalSkillNameCollidingWithATeamSkillFails(t *testing.T) {
	teamsRoot, globalRoot, envVersionsPath := setupFixture(t)
	vendorPDF(t, teamsRoot)
	own := filepath.Join(teamsRoot, "fixture-team", "skills", "pdf", "SKILL.md.tmpl")
	if err := os.MkdirAll(filepath.Dir(own), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(own, []byte("---\nname: pdf\ndescription: ours.\n---\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := RenderTeam(teamsRoot, "fixture-team", globalRoot, envVersionsPath, filepath.Join(t.TempDir(), "out"), "/kh")
	if err == nil || !strings.Contains(err.Error(), `"pdf"`) || !strings.Contains(err.Error(), "collides") {
		t.Fatalf("err = %v, want a collision error naming the skill", err)
	}
}

// vendorCollection declares an all: true entry in the fixture team's manifest
// and vendors the given skills the way a fetch would (the manifest excludes
// "b", which therefore has no lock entry or files).
func vendorCollection(t *testing.T, teamsRoot string, names ...string) {
	t.Helper()
	teamDir := filepath.Join(teamsRoot, "fixture-team")
	manifestYAML := "team: fixture-team\nglobalSteerings:\n  - security-baseline\nexternalSkills:\n" +
		"  - source: https://h/r.git\n    ref: v1\n    path: skills\n    all: true\n    exclude: [b]\n"
	if err := os.WriteFile(filepath.Join(teamDir, "manifest.yaml"), []byte(manifestYAML), 0o600); err != nil {
		t.Fatal(err)
	}
	var lock externalskills.Lock
	for _, name := range names {
		dir := filepath.Join(externalskills.SkillsDir(teamDir), name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("skill "+name+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		sum, err := externalskills.HashDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		lock.Skills = append(lock.Skills, externalskills.LockEntry{
			Name: name, Source: "https://h/r.git", Ref: "v1", Path: "skills/" + name,
			From: "skills", Commit: "abc", SHA256: sum,
		})
	}
	if err := externalskills.WriteLock(externalskills.LockPath(teamDir), lock); err != nil {
		t.Fatal(err)
	}
}

func TestRenderTeam_CollectionSkillsAreCopiedAndWiredIntoEveryAgent(t *testing.T) {
	teamsRoot, globalRoot, envVersionsPath := setupFixture(t)
	vendorCollection(t, teamsRoot, "a", "c")

	outDir, resources, _ := renderedResources(t, teamsRoot, globalRoot, envVersionsPath, "agent1")

	want := []string{fixtureSteeringURI, "skill:///kh/skills/a/SKILL.md", "skill:///kh/skills/c/SKILL.md", workspaceSteeringURI, workspaceSkillsURI}
	if !slices.Equal(resources, want) {
		t.Errorf("agent1 resources = %v, want %v", resources, want)
	}
	for _, name := range []string{"a", "c"} {
		got, err := os.ReadFile(filepath.Join(outDir, "skills", name, "SKILL.md"))
		if err != nil || string(got) != "skill "+name+"\n" {
			t.Errorf("skills/%s/SKILL.md = %q (%v), want the vendored content", name, got, err)
		}
	}
	if _, err := os.Stat(filepath.Join(outDir, "skills", "b")); err == nil {
		t.Error("the excluded skill b was rendered")
	}
	if _, err := os.Stat(filepath.Join(outDir, "skills", "skills")); err == nil {
		t.Error("the collection directory itself was rendered as a skill")
	}
}
