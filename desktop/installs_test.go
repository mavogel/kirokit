package main

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/mavogel/kirokit/internal/install"
)

type fakeEmitter struct {
	mu     sync.Mutex
	events []string
	last   []install.Record
}

func (f *fakeEmitter) emit(event string, data any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, event)
	f.last, _ = data.([]install.Record)
}

func (f *fakeEmitter) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.events)
}

// installFixture installs a one-agent team into home.
func installFixture(t *testing.T, home string) {
	t.Helper()
	root := t.TempDir()
	roots := install.Roots{
		TeamsRoot:       filepath.Join(root, "teams"),
		GlobalRoot:      filepath.Join(root, "global", "steerings"),
		EnvVersionsPath: filepath.Join(root, "env", "mcp-versions.yaml"),
	}
	write(t, roots.EnvVersionsPath, "x-mcp: \"1\"\n")
	write(t, filepath.Join(roots.TeamsRoot, "t1", "manifest.yaml"), "team: t1\nglobalSteerings: []\n")
	write(t, filepath.Join(roots.TeamsRoot, "t1", "agents", "a1", "agent.json.tmpl"), `{"name": "a1"}`)
	if err := install.Install(roots, "t1", home, nil); err != nil {
		t.Fatal(err)
	}
}

func TestInstalls_EmitsOnlyOnChange(t *testing.T) {
	home := t.TempDir()
	em := &fakeEmitter{}
	p := newInstallPoller(func() string { return home }, em.emit, time.Hour)
	defer p.Stop()

	p.Check()
	if em.count() != 0 {
		t.Fatalf("an unchanged empty state emitted %d events, want none", em.count())
	}

	installFixture(t, home)
	p.Check()
	if em.count() != 1 || len(em.last) != 1 || em.last[0].Team != "t1" {
		t.Fatalf("after install: events %v last %+v, want one with t1", em.events, em.last)
	}
	p.Check()
	if em.count() != 1 {
		t.Errorf("an unchanged state emitted again: %d events", em.count())
	}

	// A newer file time on the record is a change (a reinstall rewrites it).
	record := filepath.Join(home, ".kirokit-installed", "t1.json")
	later := time.Now().Add(time.Hour)
	if err := os.Chtimes(record, later, later); err != nil {
		t.Fatal(err)
	}
	p.Check()
	if em.count() != 2 {
		t.Errorf("a changed file time emitted %d events in total, want 2", em.count())
	}

	if err := install.Uninstall(home, "t1"); err != nil {
		t.Fatal(err)
	}
	p.Check()
	if em.count() != 3 || len(em.last) != 0 {
		t.Errorf("after uninstall: events %d last %+v, want a third event with no records", em.count(), em.last)
	}
}

// Settings can change the Kiro home between checks; the next check reads the
// new folder.
func TestInstalls_FollowsSettingsChange(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	installFixture(t, second)
	home := first
	em := &fakeEmitter{}
	p := newInstallPoller(func() string { return home }, em.emit, time.Hour)
	defer p.Stop()

	home = second
	p.Check()

	if em.count() != 1 || len(em.last) != 1 {
		t.Errorf("after the home changed: events %d last %+v, want one event with t1", em.count(), em.last)
	}
}

// The goroutine ticks on its own, and Stop ends it: nothing is emitted after
// Stop returns.
func TestInstalls_TicksThenStops(t *testing.T) {
	home := t.TempDir()
	em := &fakeEmitter{}
	p := newInstallPoller(func() string { return home }, em.emit, 5*time.Millisecond)

	installFixture(t, home)
	deadline := time.Now().Add(5 * time.Second)
	for em.count() == 0 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if em.count() == 0 {
		t.Fatal("the ticker never noticed the new record")
	}

	p.Stop()
	before := em.count()
	if err := install.Uninstall(home, "t1"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	if em.count() != before {
		t.Error("an event arrived after Stop returned")
	}
	p.Stop() // a second Stop must not block or panic
}
