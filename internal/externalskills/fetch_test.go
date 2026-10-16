package externalskills

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mavogel/kirokit/internal/manifest"
)

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t",
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

func commitAll(t *testing.T, repo, msg string) string {
	t.Helper()
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-q", "-m", msg)
	return git(t, repo, "rev-parse", "HEAD")
}

// sourceRepo builds a git repository holding skills/pdf (SKILL.md, a nested
// executable script and a plain reference file) and returns it with the
// commits of its two versions; v1 is tagged on the first.
func sourceRepo(t *testing.T) (repo, v1, v2 string) {
	t.Helper()
	repo = t.TempDir()
	git(t, repo, "init", "-q", "-b", "main")
	writeFile(t, filepath.Join(repo, "skills/pdf/SKILL.md"), "pdf v1\n", 0o644)
	writeFile(t, filepath.Join(repo, "skills/pdf/scripts/run.sh"), "echo run\n", 0o755)
	writeFile(t, filepath.Join(repo, "skills/pdf/reference.md"), "ref\n", 0o644)
	writeFile(t, filepath.Join(repo, "skills/other/SKILL.md"), "other\n", 0o644)
	v1 = commitAll(t, repo, "v1")
	git(t, repo, "tag", "-a", "v1", "-m", "v1")
	writeFile(t, filepath.Join(repo, "skills/pdf/SKILL.md"), "pdf v2\n", 0o644)
	v2 = commitAll(t, repo, "v2")
	return repo, v1, v2
}

func pdfEntry(repo, ref string) manifest.ExternalSkill {
	return manifest.ExternalSkill{Source: repo, Ref: ref, Path: "skills/pdf"}
}

func teamDirFor(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "demo-team")
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestFetch_TagPinVendorsThatTagsTreeAndRecordsItsCommit(t *testing.T) {
	repo, v1, _ := sourceRepo(t)
	teamDir := teamDirFor(t)

	lock, err := Fetch(teamDir, []manifest.ExternalSkill{pdfEntry(repo, "v1")})
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	skill := filepath.Join(SkillsDir(teamDir), "pdf")
	if got := readFile(t, filepath.Join(skill, "SKILL.md")); got != "pdf v1\n" {
		t.Errorf("SKILL.md = %q, want the v1 content", got)
	}
	if _, err := os.Stat(filepath.Join(skill, "skills")); err == nil {
		t.Error("vendored dir contains the repository path prefix; want the skill directory itself")
	}
	if len(lock.Skills) != 1 || lock.Skills[0].Commit != v1 {
		t.Fatalf("lock = %+v, want one entry at commit %s (the tag's commit, not the tag object)", lock.Skills, v1)
	}
	if lock.Skills[0].SHA256 != mustHash(t, skill) {
		t.Error("lock hash does not match the vendored files")
	}
	onDisk, err := ReadLock(LockPath(teamDir))
	if err != nil || len(onDisk.Skills) != 1 || onDisk.Skills[0] != lock.Skills[0] {
		t.Errorf("lock on disk = %+v (%v), want the returned lock", onDisk, err)
	}
	if err := Verify(teamDir, []manifest.ExternalSkill{pdfEntry(repo, "v1")}); err != nil {
		t.Errorf("Verify after Fetch: %v", err)
	}
}

func TestFetch_PreservesExecutableBitOfNestedScript(t *testing.T) {
	repo, _, _ := sourceRepo(t)
	teamDir := teamDirFor(t)
	if _, err := Fetch(teamDir, []manifest.ExternalSkill{pdfEntry(repo, "v1")}); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(SkillsDir(teamDir), "pdf", "scripts", "run.sh")
	info, err := os.Stat(script)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("run.sh mode = %v, want owner-executable", info.Mode().Perm())
	}
	if got := readFile(t, filepath.Join(SkillsDir(teamDir), "pdf", "reference.md")); got != "ref\n" {
		t.Errorf("reference.md = %q", got)
	}
}

func TestFetch_CommitSHAPin(t *testing.T) {
	repo, v1, _ := sourceRepo(t)
	teamDir := teamDirFor(t)

	lock, err := Fetch(teamDir, []manifest.ExternalSkill{pdfEntry(repo, v1)})
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if lock.Skills[0].Commit != v1 {
		t.Errorf("commit = %s, want %s", lock.Skills[0].Commit, v1)
	}
}

func TestFetch_CommitSHAPinFallsBackWhenShallowFetchIsRefused(t *testing.T) {
	repo, v1, _ := sourceRepo(t)

	// Protocol v0 only serves advertised refs, like a host that does not
	// allow wants of arbitrary commits.
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "protocol.version")
	t.Setenv("GIT_CONFIG_VALUE_0", "0")
	probe := t.TempDir()
	git(t, probe, "init", "-q")
	probeCmd := exec.Command("git", "fetch", "--depth", "1", repo, v1)
	probeCmd.Dir = probe
	if err := probeCmd.Run(); err == nil {
		t.Skip("this git serves unadvertised commits even over protocol v0; the fallback cannot be forced here")
	}

	teamDir := teamDirFor(t)
	lock, err := Fetch(teamDir, []manifest.ExternalSkill{pdfEntry(repo, v1)})
	if err != nil {
		t.Fatalf("Fetch should fall back to a full fetch: %v", err)
	}
	if lock.Skills[0].Commit != v1 {
		t.Errorf("commit = %s, want %s", lock.Skills[0].Commit, v1)
	}
}

func TestFetch_BranchRefFollowsTheBranchAndRerunIsIdempotent(t *testing.T) {
	repo, _, v2 := sourceRepo(t)
	teamDir := teamDirFor(t)
	entries := []manifest.ExternalSkill{pdfEntry(repo, "main")}

	first, err := Fetch(teamDir, entries)
	if err != nil || first.Skills[0].Commit != v2 {
		t.Fatalf("first Fetch = %+v, %v; want commit %s", first, err, v2)
	}
	lockBytes := readFile(t, LockPath(teamDir))

	second, err := Fetch(teamDir, entries)
	if err != nil {
		t.Fatal(err)
	}
	if readFile(t, LockPath(teamDir)) != lockBytes || second.Skills[0] != first.Skills[0] {
		t.Error("re-running Fetch with no upstream change altered the lock")
	}

	writeFile(t, filepath.Join(repo, "skills/pdf/SKILL.md"), "pdf v3\n", 0o644)
	v3 := commitAll(t, repo, "v3")
	third, err := Fetch(teamDir, entries)
	if err != nil || third.Skills[0].Commit != v3 {
		t.Fatalf("after the branch moved, Fetch = %+v, %v; want commit %s", third, err, v3)
	}
	if got := readFile(t, filepath.Join(SkillsDir(teamDir), "pdf", "SKILL.md")); got != "pdf v3\n" {
		t.Errorf("SKILL.md = %q, want the moved branch's content", got)
	}
}

func TestFetch_ExportIgnoredFilesAreStillVendored(t *testing.T) {
	repo, _, _ := sourceRepo(t)
	writeFile(t, filepath.Join(repo, ".gitattributes"), "skills/pdf/reference.md export-ignore\n", 0o644)
	commitAll(t, repo, "export-ignore reference")
	teamDir := teamDirFor(t)

	if _, err := Fetch(teamDir, []manifest.ExternalSkill{pdfEntry(repo, "main")}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(SkillsDir(teamDir), "pdf", "reference.md")); err != nil {
		t.Errorf("export-ignore file was dropped: %v", err)
	}
}

func TestFetch_FailuresNameTheSkillAndLeaveEverythingUntouched(t *testing.T) {
	repo, _, _ := sourceRepo(t)
	writeFile(t, filepath.Join(repo, "skills/nomanifest/readme.md"), "no SKILL.md\n", 0o644)
	writeFile(t, filepath.Join(repo, "skills/linked/SKILL.md"), "x\n", 0o644)
	if err := os.Symlink("SKILL.md", filepath.Join(repo, "skills/linked/alias.md")); err != nil {
		t.Fatal(err)
	}
	commitAll(t, repo, "broken skills")

	good := pdfEntry(repo, "v1")
	tests := []struct {
		name      string
		bad       manifest.ExternalSkill
		wantInErr []string
	}{
		{"missing path", manifest.ExternalSkill{Source: repo, Ref: "main", Path: "skills/nope"}, []string{`"nope"`, repo}},
		{"missing SKILL.md", manifest.ExternalSkill{Source: repo, Ref: "main", Path: "skills/nomanifest"}, []string{`"nomanifest"`, "SKILL.md"}},
		{"symlink in tree", manifest.ExternalSkill{Source: repo, Ref: "main", Path: "skills/linked"}, []string{`"linked"`, "symbolic link"}},
		{"unreachable source", manifest.ExternalSkill{Source: filepath.Join(repo, "does-not-exist"), Ref: "main", Path: "skills/x"}, []string{`"x"`, "authentication is handled by git"}},
		{"unknown ref", manifest.ExternalSkill{Source: repo, Ref: "no-such-ref", Path: "skills/other"}, []string{`"other"`, "no-such-ref"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teamDir := teamDirFor(t)
			if _, err := Fetch(teamDir, []manifest.ExternalSkill{good}); err != nil {
				t.Fatal(err)
			}
			lockBefore := readFile(t, LockPath(teamDir))
			hashBefore := mustHash(t, filepath.Join(SkillsDir(teamDir), "pdf"))

			_, err := Fetch(teamDir, []manifest.ExternalSkill{good, tc.bad})
			if err == nil {
				t.Fatal("Fetch succeeded, want an error")
			}
			for _, want := range tc.wantInErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("err = %q, want it to contain %q", err, want)
				}
			}
			if readFile(t, LockPath(teamDir)) != lockBefore {
				t.Error("lock file changed after a failed Fetch")
			}
			if mustHash(t, filepath.Join(SkillsDir(teamDir), "pdf")) != hashBefore {
				t.Error("vendored files changed after a failed Fetch")
			}
			if entries, _ := filepath.Glob(filepath.Join(teamDir, "external", ".fetch-*")); len(entries) != 0 {
				t.Errorf("staging left behind: %v", entries)
			}
			if _, err := os.Stat(filepath.Join(SkillsDir(teamDir), filepath.Base(tc.bad.Path))); err == nil {
				t.Error("a partial copy of the failed skill was left behind")
			}
		})
	}
}

func TestFetch_PrunesSkillsRemovedFromTheManifest(t *testing.T) {
	repo, _, _ := sourceRepo(t)
	teamDir := teamDirFor(t)
	other := manifest.ExternalSkill{Source: repo, Ref: "main", Path: "skills/other"}

	if _, err := Fetch(teamDir, []manifest.ExternalSkill{pdfEntry(repo, "v1"), other}); err != nil {
		t.Fatal(err)
	}
	lock, err := Fetch(teamDir, []manifest.ExternalSkill{pdfEntry(repo, "v1")})
	if err != nil {
		t.Fatal(err)
	}

	if len(lock.Skills) != 1 || lock.Skills[0].Name != "pdf" {
		t.Errorf("lock = %+v, want only pdf", lock.Skills)
	}
	if _, err := os.Stat(filepath.Join(SkillsDir(teamDir), "other")); !os.IsNotExist(err) {
		t.Errorf("removed skill still vendored (err=%v)", err)
	}
}

func TestFetch_NoEntriesAndNothingVendoredDoesNothing(t *testing.T) {
	teamDir := teamDirFor(t)
	lock, err := Fetch(teamDir, nil)
	if err != nil || len(lock.Skills) != 0 {
		t.Fatalf("Fetch(nil) = %+v, %v", lock, err)
	}
	if _, err := os.Stat(filepath.Join(teamDir, "external")); !os.IsNotExist(err) {
		t.Errorf("external/ created for a team with no external skills (err=%v)", err)
	}
}

func TestFetch_WorksWithARelativeTeamDir(t *testing.T) {
	repo, _, _ := sourceRepo(t)
	base := t.TempDir()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(base); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })

	// The CLI passes teams/<team> relative to the repository root.
	if _, err := Fetch(filepath.Join("teams", "demo"), []manifest.ExternalSkill{pdfEntry(repo, "v1")}); err != nil {
		t.Fatalf("Fetch with a relative team dir: %v", err)
	}
	if got := readFile(t, filepath.Join(base, "teams", "demo", "external", "skills", "pdf", "SKILL.md")); got != "pdf v1\n" {
		t.Errorf("SKILL.md = %q, want the v1 content", got)
	}
}

func TestFetch_FailureOnAFirstFetchDoesNotCreateTheExternalDirectory(t *testing.T) {
	repo, _, _ := sourceRepo(t)
	teamDir := teamDirFor(t)

	_, err := Fetch(teamDir, []manifest.ExternalSkill{{Source: repo, Ref: "main", Path: "skills/nope"}})
	if err == nil {
		t.Fatal("Fetch succeeded, want an error")
	}
	if _, err := os.Stat(filepath.Join(teamDir, "external")); !os.IsNotExist(err) {
		t.Errorf("external/ left behind after a failed first fetch (err=%v)", err)
	}
}

func TestFetch_RefusesEntriesThatWouldInjectGitOptions(t *testing.T) {
	// Fetch is a public API, so it must not trust that the manifest loader
	// already screened its input.
	teamDir := teamDirFor(t)
	marker := filepath.Join(t.TempDir(), "PWNED")
	hostile := manifest.ExternalSkill{
		Source: "--upload-pack=touch " + marker,
		Ref:    ".",
		Path:   "skills/pdf",
	}

	_, err := Fetch(teamDir, []manifest.ExternalSkill{hostile})
	if err == nil {
		t.Fatal("Fetch accepted an entry whose source starts with a dash")
	}
	if _, statErr := os.Stat(marker); statErr == nil {
		t.Fatal("the --upload-pack command ran: argument injection reached git")
	}
}

// collectionRepo holds skills a, b, c under skills/, plus skills/notes which
// is a directory without a SKILL.md and therefore not a skill.
func collectionRepo(t *testing.T) (repo, commit string) {
	t.Helper()
	repo = t.TempDir()
	git(t, repo, "init", "-q", "-b", "main")
	for _, name := range []string{"a", "b", "c"} {
		writeFile(t, filepath.Join(repo, "skills", name, "SKILL.md"), "skill "+name+"\n", 0o644)
	}
	writeFile(t, filepath.Join(repo, "skills/a/run.sh"), "echo a\n", 0o755)
	writeFile(t, filepath.Join(repo, "skills/notes/readme.md"), "not a skill\n", 0o644)
	return repo, commitAll(t, repo, "collection")
}

func collectionEntry(repo string, exclude ...string) manifest.ExternalSkill {
	return manifest.ExternalSkill{Source: repo, Ref: "main", Path: "skills", All: true, Exclude: exclude}
}

func lockNames(l Lock) string {
	var names []string
	for _, e := range l.Skills {
		names = append(names, e.Name)
	}
	return strings.Join(names, ",")
}

func TestFetch_CollectionVendorsEverySkillAndIgnoresDirectoriesWithoutSkillMD(t *testing.T) {
	repo, commit := collectionRepo(t)
	teamDir := teamDirFor(t)
	entries := []manifest.ExternalSkill{collectionEntry(repo)}

	lock, err := Fetch(teamDir, entries)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	if got := lockNames(lock); got != "a,b,c" {
		t.Fatalf("locked skills = %q, want a,b,c (notes has no SKILL.md)", got)
	}
	for _, le := range lock.Skills {
		if le.From != "skills" || le.Path != "skills/"+le.Name || le.Commit != commit {
			t.Errorf("lock entry %+v, want From=skills Path=skills/%s Commit=%s", le, le.Name, commit)
		}
		if le.SHA256 != mustHash(t, filepath.Join(SkillsDir(teamDir), le.Name)) {
			t.Errorf("%s: lock hash does not match the vendored files", le.Name)
		}
	}
	if _, err := os.Stat(filepath.Join(SkillsDir(teamDir), "notes")); err == nil {
		t.Error("notes (no SKILL.md) was vendored")
	}
	info, err := os.Stat(filepath.Join(SkillsDir(teamDir), "a", "run.sh"))
	if err != nil || info.Mode().Perm()&0o100 == 0 {
		t.Errorf("a/run.sh = %v (%v), want owner-executable", info, err)
	}
	if err := Verify(teamDir, entries); err != nil {
		t.Errorf("Verify after a collection Fetch: %v", err)
	}
	if entries, _ := filepath.Glob(filepath.Join(teamDir, "external", ".fetch-*")); len(entries) != 0 {
		t.Errorf("staging left behind: %v", entries)
	}
}

func TestFetch_CollectionExcludeSkipsSkillsAndRefetchRemovesNewlyExcludedOnes(t *testing.T) {
	repo, _ := collectionRepo(t)
	teamDir := teamDirFor(t)

	lock, err := Fetch(teamDir, []manifest.ExternalSkill{collectionEntry(repo, "b")})
	if err != nil || lockNames(lock) != "a,c" {
		t.Fatalf("Fetch = %q, %v; want a,c", lockNames(lock), err)
	}
	if _, err := os.Stat(filepath.Join(SkillsDir(teamDir), "b")); err == nil {
		t.Error("excluded skill b was vendored")
	}

	lock, err = Fetch(teamDir, []manifest.ExternalSkill{collectionEntry(repo, "b", "c")})
	if err != nil || lockNames(lock) != "a" {
		t.Fatalf("re-fetch = %q, %v; want a", lockNames(lock), err)
	}
	if _, err := os.Stat(filepath.Join(SkillsDir(teamDir), "c")); !os.IsNotExist(err) {
		t.Errorf("newly excluded skill c is still vendored (err=%v)", err)
	}
}

func TestFetch_CollectionFailuresNameTheProblemAndLeaveEverythingUntouched(t *testing.T) {
	repo, _ := collectionRepo(t)
	writeFile(t, filepath.Join(repo, "skills/linked/SKILL.md"), "x\n", 0o644)
	if err := os.Symlink("SKILL.md", filepath.Join(repo, "skills/linked/alias.md")); err != nil {
		t.Fatal(err)
	}
	commitAll(t, repo, "add a skill with a symlink")

	tests := []struct {
		name      string
		entries   func(repo string) []manifest.ExternalSkill
		wantInErr []string
	}{
		{"exclude naming a skill that does not exist", func(r string) []manifest.ExternalSkill {
			return []manifest.ExternalSkill{collectionEntry(r, "linked", "nope")}
		}, []string{`"nope"`, "exclude"}},
		{"collection with no skills in it", func(r string) []manifest.ExternalSkill {
			return []manifest.ExternalSkill{{Source: r, Ref: "main", Path: "skills/notes", All: true}}
		}, []string{`"notes"`, "no skills"}},
		{"explicit entry clashing with a collection skill", func(r string) []manifest.ExternalSkill {
			return []manifest.ExternalSkill{
				collectionEntry(r, "linked"),
				{Source: r, Ref: "main", Path: "skills/a"},
			}
		}, []string{`"a"`, "more than one"}},
		{"symbolic link inside a collected skill", func(r string) []manifest.ExternalSkill {
			return []manifest.ExternalSkill{collectionEntry(r)}
		}, []string{`"linked"`, "symbolic link"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teamDir := teamDirFor(t)
			if _, err := Fetch(teamDir, []manifest.ExternalSkill{collectionEntry(repo, "linked", "c")}); err != nil {
				t.Fatalf("setup fetch: %v", err)
			}
			lockBefore := readFile(t, LockPath(teamDir))
			hashBefore := mustHash(t, filepath.Join(SkillsDir(teamDir), "a"))

			_, err := Fetch(teamDir, tc.entries(repo))
			if err == nil {
				t.Fatal("Fetch succeeded, want an error")
			}
			for _, want := range tc.wantInErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("err = %q, want it to contain %q", err, want)
				}
			}
			if readFile(t, LockPath(teamDir)) != lockBefore {
				t.Error("lock changed after a failed Fetch")
			}
			if mustHash(t, filepath.Join(SkillsDir(teamDir), "a")) != hashBefore {
				t.Error("vendored files changed after a failed Fetch")
			}
			if _, err := os.Stat(filepath.Join(SkillsDir(teamDir), "b")); err != nil {
				t.Errorf("previously vendored skill b disappeared: %v", err)
			}
		})
	}
}

func TestFetch_NameClashIsReportedWhicheverEntryComesFirst(t *testing.T) {
	repo, _ := collectionRepo(t)
	explicit := manifest.ExternalSkill{Source: repo, Ref: "main", Path: "skills/a"}
	teamDir := teamDirFor(t)

	for name, entries := range map[string][]manifest.ExternalSkill{
		"explicit entry first":   {explicit, collectionEntry(repo)},
		"collection entry first": {collectionEntry(repo), explicit},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Fetch(teamDir, entries)
			if err == nil {
				t.Fatal("Fetch succeeded, want a name clash error")
			}
			for _, want := range []string{`"a"`, "more than one"} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("err = %q, want it to contain %q", err, want)
				}
			}
			if strings.Contains(err.Error(), "rename") {
				t.Errorf("err = %q leaks an internal staging rename failure", err)
			}
			if _, statErr := os.Stat(filepath.Join(teamDir, "external")); !os.IsNotExist(statErr) {
				t.Errorf("a failed first fetch left external/ behind (err=%v)", statErr)
			}
		})
	}
}
