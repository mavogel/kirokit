package quick

import (
	"maps"
	"path/filepath"
	"strings"
	"testing"
)

// Pinned skills are looked up by the skill's frontmatter name (Quick's match
// key, and the ids file's key), not by its folder; they keep the agent's
// manifest order, and a pinned skill without an id is left out rather than
// written as an empty entry.
func TestRenderTeam_PinsByFrontmatterNameInManifestOrder(t *testing.T) {
	quick := `quick:
  skills: [other, story-dir]
  agents:
    - name: po
      skills: [story-dir, other]
`
	extra := map[string]string{
		"teams/t/skills/story-dir/SKILL.md.tmpl": "---\nname: renamed\ndescription: Folder and name differ.\n---\nBody.\n",
	}

	f := newTeam(t, quick, extra)
	out, res, err := f.render(t, map[string]string{"renamed": idStory, "other": idOther})
	if err != nil {
		t.Fatalf("RenderTeam: %v", err)
	}
	if len(res.MissingIDs) != 0 {
		t.Errorf("MissingIDs = %v, want none", res.MissingIDs)
	}
	agent := read(t, filepath.Join(out, "agents", "po.json"))
	if want := "\"skills\": [\n    \"" + idStory + "\",\n    \"" + idOther + "\"\n  ],"; !strings.Contains(agent, want) {
		t.Errorf("po.json skills not in manifest order by frontmatter name:\n%s", agent)
	}
	if skill := read(t, filepath.Join(out, "skills", "story-dir", "SKILL.md")); !strings.Contains(skill, "\nid: "+idStory+"\n---\n") {
		t.Errorf("story-dir SKILL.md lacks the id of its frontmatter name:\n%s", skill)
	}

	out, res, err = f.render(t, map[string]string{"other": idOther})
	if err != nil {
		t.Fatalf("RenderTeam: %v", err)
	}
	if strings.Join(res.MissingIDs, ",") != "renamed" {
		t.Errorf("MissingIDs = %v, want the frontmatter name renamed", res.MissingIDs)
	}
	agent = read(t, filepath.Join(out, "agents", "po.json"))
	if want := "\"skills\": [\n    \"" + idOther + "\"\n  ],"; !strings.Contains(agent, want) {
		t.Errorf("po.json should pin only the skill that has an id:\n%s", agent)
	}
}

// A frontmatter whose closing --- ends the file (no newline, no body) still
// gets the id as its last frontmatter line.
func TestRenderTeam_FrontmatterClosedAtEndOfFile(t *testing.T) {
	f := newTeam(t, "quick:\n  skills: [other]\n", map[string]string{
		"teams/t/skills/other/SKILL.md.tmpl": "---\nname: other\ndescription: Another skill.\n---",
	})
	out, res, err := f.render(t, map[string]string{"other": idOther})
	if err != nil {
		t.Fatalf("RenderTeam: %v", err)
	}
	if len(res.MissingIDs) != 0 {
		t.Errorf("MissingIDs = %v", res.MissingIDs)
	}
	got := read(t, filepath.Join(out, "skills", "other", "SKILL.md"))
	want := "---\nname: other\ndescription: Another skill.\nid: " + idOther + "\n---"
	if strings.TrimSuffix(got, "\n") != want {
		t.Errorf("SKILL.md = %q, want %q (optionally newline-terminated)", got, want)
	}
	// The rendered file reads back as an export with that id.
	if name, id, err := exportedID([]byte(got)); err != nil || name != "other" || id != idOther {
		t.Errorf("exportedID = %q, %q, %v", name, id, err)
	}
}

// An empty or comment-only ids file is a valid first state, like a missing one.
func TestLoadIDs_EmptyOrCommentOnlyFileIsEmpty(t *testing.T) {
	for name, content := range map[string]string{
		"empty":        "",
		"comment only": "# Amazon Quick skill ids\n",
		"null skills":  "skills:\n",
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			writeFiles(t, dir, map[string]string{"quick-ids.yaml": content})
			ids, err := LoadIDs(filepath.Join(dir, "quick-ids.yaml"))
			if err != nil || len(ids) != 0 {
				t.Errorf("LoadIDs = %v, %v; want an empty map", ids, err)
			}
		})
	}
}

// A Quick export zip holds the skill in a folder, beside reference files; only
// SKILL.md entries count, wherever they sit in the zip.
func TestReadExports_ZipWithNestedSkillFolder(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "user-story.zip")
	writeZip(t, zipPath, map[string]string{
		"user-story/":                    "",
		"user-story/SKILL.md":            exported("user-story", idStory),
		"user-story/references/notes.md": exported("notes", idOther),
	})
	got, skipped, err := ReadExports([]string{zipPath})
	if err != nil {
		t.Fatalf("ReadExports: %v", err)
	}
	if want := map[string]string{"user-story": idStory}; !maps.Equal(got, want) {
		t.Errorf("ids = %v, want %v", got, want)
	}
	if len(skipped) != 0 {
		t.Errorf("skipped = %v", skipped)
	}
}
