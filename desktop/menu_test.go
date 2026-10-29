package main

import (
	"bytes"
	"errors"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
)

func titles(groups []menuGroup) []string {
	var out []string
	for _, g := range groups {
		out = append(out, g.Title)
	}
	return out
}

// describe renders a group's items as "role:x", "action:y" or "-" (separator).
func describe(g menuGroup) []string {
	var out []string
	for _, it := range g.Items {
		switch {
		case it.Role != "":
			out = append(out, "role:"+it.Role)
		case it.Action != "":
			out = append(out, "action:"+it.Action)
		default:
			out = append(out, "-")
		}
	}
	return out
}

func group(t *testing.T, groups []menuGroup, title string) menuGroup {
	t.Helper()
	for _, g := range groups {
		if g.Title == title {
			return g
		}
	}
	t.Fatalf("no %q menu in %v", title, titles(groups))
	return menuGroup{}
}

// macOS: the app menu, then Edit, Window and Help. Windows and Linux show the
// menu in the window: File instead of the app menu, About under Help.
func TestMenu_GroupsPerSystem(t *testing.T) {
	for _, tc := range []struct {
		goos string
		want []string
	}{
		{"darwin", []string{"kirokit", "Edit", "Window", "Help"}},
		{"windows", []string{"File", "Edit", "Window", "Help"}},
		{"linux", []string{"File", "Edit", "Window", "Help"}},
	} {
		t.Run(tc.goos, func(t *testing.T) {
			groups := menuFor(tc.goos)
			if got := titles(groups); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("menus = %v, want %v", got, tc.want)
			}
			// Edit and Window are the system's own menus (undo, cut, copy, paste,
			// select all; minimise, zoom).
			if g := group(t, groups, "Edit"); g.Role != "edit" || len(g.Items) != 0 {
				t.Errorf("Edit = %+v, want the edit role", g)
			}
			if g := group(t, groups, "Window"); g.Role != "window" || len(g.Items) != 0 {
				t.Errorf("Window = %+v, want the window role", g)
			}
		})
	}
}

func TestMenu_MacAppMenuAndHelp(t *testing.T) {
	groups := menuFor("darwin")

	app := group(t, groups, "kirokit")
	want := []string{"role:about", "-", "action:settings", "-", "role:services", "-", "role:hide", "role:hideOthers", "role:showAll", "-", "role:quit"}
	if got := describe(app); !reflect.DeepEqual(got, want) {
		t.Errorf("kirokit menu = %v, want %v", got, want)
	}
	// kirokit's own Help: Wails' Help role would add a wails.io link.
	if got := describe(group(t, groups, "Help")); !reflect.DeepEqual(got, []string{"action:docs", "action:report"}) {
		t.Errorf("Help = %v, want the docs and report entries only", got)
	}
}

func TestMenu_WindowsAndLinuxFileAndHelp(t *testing.T) {
	for _, goos := range []string{"windows", "linux"} {
		t.Run(goos, func(t *testing.T) {
			groups := menuFor(goos)
			if got := describe(group(t, groups, "File")); !reflect.DeepEqual(got, []string{"action:settings", "-", "role:quit"}) {
				t.Errorf("File = %v, want Settings…, separator, Quit", got)
			}
			if got := describe(group(t, groups, "Help")); !reflect.DeepEqual(got, []string{"action:docs", "action:report", "-", "role:about"}) {
				t.Errorf("Help = %v, want docs, report, separator, About", got)
			}
		})
	}
}

// Every entry kirokit adds has a label and does something; only Settings has
// its own shortcut (⌘, / Ctrl+,), every other shortcut is the system's.
func TestMenu_EntriesHaveLabelsActionsAndOneShortcut(t *testing.T) {
	a := NewAppAt(t.TempDir())
	for _, goos := range []string{"darwin", "windows", "linux"} {
		for _, g := range menuFor(goos) {
			for _, it := range g.Items {
				if it.Action == "" {
					if it.Shortcut != "" || it.Label != "" && it.Role == "" {
						t.Errorf("%s %s: item %+v has a shortcut or label without an action", goos, g.Title, it)
					}
					continue
				}
				if it.Label == "" || a.menuAction(it.Action, func(string) error { return nil }) == nil {
					t.Errorf("%s %s: action %q has no label or no handler", goos, g.Title, it.Action)
				}
				wantShortcut := ""
				if it.Action == "settings" {
					wantShortcut = "CmdOrCtrl+,"
				}
				if it.Shortcut != wantShortcut {
					t.Errorf("%s %s: %q shortcut = %q, want %q", goos, g.Title, it.Label, it.Shortcut, wantShortcut)
				}
			}
		}
	}
}

func TestMenu_SettingsAsksThePageToOpenSettings(t *testing.T) {
	a := NewAppAt(t.TempDir())
	rec := &eventLog{}
	a.setEmitter(rec.emit)

	a.menuAction("settings", func(string) error { t.Fatal("settings opened a URL"); return nil })()

	if !reflect.DeepEqual(rec.events, []string{"menu:settings"}) {
		t.Errorf("events = %v, want [menu:settings]", rec.events)
	}
}

func TestMenu_HelpOpensTheFixedKirokitAddresses(t *testing.T) {
	a := NewAppAt(t.TempDir())
	var opened []string
	open := func(u string) error { opened = append(opened, u); return nil }

	a.menuAction("docs", open)()
	a.menuAction("report", open)()

	want := []string{
		"https://github.com/mavogel/kirokit/blob/main/docs/desktop-app.md",
		"https://github.com/mavogel/kirokit/issues/new",
	}
	if !reflect.DeepEqual(opened, want) {
		t.Errorf("opened %v, want %v", opened, want)
	}
}

// A browser that cannot be started must not take the app down; the reason is logged.
func TestMenu_OpenFailureIsLogged(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	a := NewAppAt(t.TempDir())

	a.menuAction("docs", func(string) error { return errors.New("no browser") })()

	if !strings.Contains(buf.String(), "no browser") {
		t.Errorf("log = %q, want the open error", buf.String())
	}
}

func TestMenu_UnknownActionHasNoHandler(t *testing.T) {
	if NewAppAt(t.TempDir()).menuAction("bogus", nil) != nil {
		t.Error("an unknown action got a handler")
	}
}

// Every role menuFor names must reach the native menu: addRole only logs an
// unknown name and leaves the entry out, so a typo would silently drop Quit or
// Edit from one system's menu.
func TestMenu_EveryRoleMapsToAWailsRole(t *testing.T) {
	for _, goos := range []string{"darwin", "windows", "linux"} {
		for _, g := range menuFor(goos) {
			if g.Role != "" {
				if _, ok := menuRoles[g.Role]; !ok {
					t.Errorf("%s %s: group role %q has no Wails role", goos, g.Title, g.Role)
				}
			}
			for _, it := range g.Items {
				if it.Role == "" {
					continue
				}
				if _, ok := menuRoles[it.Role]; !ok {
					t.Errorf("%s %s: role %q has no Wails role", goos, g.Title, it.Role)
				}
			}
		}
	}
}
