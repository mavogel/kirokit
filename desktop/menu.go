package main

import "log"

// The two addresses the Help menu opens; no user input reaches them.
const (
	docsURL   = "https://github.com/mavogel/kirokit/blob/main/docs/desktop-app.md"
	reportURL = "https://github.com/mavogel/kirokit/issues/new"
)

// menuItem is one entry of the native menu: a system role (Wails names and
// labels it), a kirokit action with its label, or a separator (zero value).
type menuItem struct {
	Label    string // empty for a role or a separator
	Role     string // "about", "services", "hide", "hideOthers", "showAll", "quit"
	Action   string // "settings", "docs", "report"
	Shortcut string // "CmdOrCtrl+," for Settings only
}

// menuGroup is one top-level menu: the system's own (Role "edit", "window")
// or a list of items.
type menuGroup struct {
	Title string
	Role  string
	Items []menuItem
}

// menuFor describes the native menu on goos. macOS has the app menu (About,
// Settings…, Quit); Windows and Linux show File, Edit, Window and Help at the
// top of the window, with About under Help.
func menuFor(goos string) []menuGroup {
	var (
		sep      = menuItem{}
		settings = menuItem{Label: "Settings…", Action: "settings", Shortcut: "CmdOrCtrl+,"}
		docs     = menuItem{Label: "kirokit documentation", Action: "docs"}
		report   = menuItem{Label: "Report a problem with kirokit…", Action: "report"}
		edit     = menuGroup{Title: "Edit", Role: "edit"}
		window   = menuGroup{Title: "Window", Role: "window"}
	)
	if goos == "darwin" {
		return []menuGroup{
			{Title: "kirokit", Items: []menuItem{
				{Role: "about"}, sep, settings, sep, {Role: "services"}, sep,
				{Role: "hide"}, {Role: "hideOthers"}, {Role: "showAll"}, sep, {Role: "quit"},
			}},
			edit,
			window,
			{Title: "Help", Items: []menuItem{docs, report}},
		}
	}
	return []menuGroup{
		{Title: "File", Items: []menuItem{settings, sep, {Role: "quit"}}},
		edit,
		window,
		{Title: "Help", Items: []menuItem{docs, report, sep, {Role: "about"}}},
	}
}

// menuAction returns what a kirokit menu entry does: Settings… asks the page to
// open its Settings page (menu:settings), the Help entries open their fixed
// address with open. An unknown action returns nil.
func (a *App) menuAction(action string, open func(string) error) func() {
	browse := func(url string) func() {
		return func() {
			if err := open(url); err != nil {
				log.Printf("menu: could not open %s: %v", url, err)
			}
		}
	}
	switch action {
	case "settings":
		return func() { a.emit("menu:settings", nil) }
	case "docs":
		return browse(docsURL)
	case "report":
		return browse(reportURL)
	}
	return nil
}
