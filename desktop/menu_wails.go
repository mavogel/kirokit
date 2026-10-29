package main

import (
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// menuRoles maps menuFor's role names to Wails' roles.
var menuRoles = map[string]application.Role{
	"about":      application.About,
	"services":   application.ServicesMenu,
	"hide":       application.Hide,
	"hideOthers": application.HideOthers,
	"showAll":    application.ShowAll,
	"quit":       application.Quit,
	"edit":       application.EditMenu,
	"window":     application.WindowMenu,
}

// buildMenu turns menuFor's description into a Wails menu; a kirokit action
// runs app.menuAction, the Help addresses open in the system browser.
func buildMenu(wapp *application.App, groups []menuGroup, app *App) *application.Menu {
	menu := application.NewMenu()
	for _, g := range groups {
		if g.Role != "" {
			addRole(menu, g.Role)
			continue
		}
		sub := menu.AddSubmenu(g.Title)
		for _, it := range g.Items {
			switch {
			case it.Role != "":
				addRole(sub, it.Role)
			case it.Action != "":
				run := app.menuAction(it.Action, wapp.Browser.OpenURL)
				item := sub.Add(it.Label).OnClick(func(*application.Context) { run() })
				if it.Shortcut != "" {
					item.SetAccelerator(it.Shortcut)
				}
			default:
				sub.AddSeparator()
			}
		}
	}
	return menu
}

func addRole(menu *application.Menu, name string) {
	role, ok := menuRoles[name]
	if !ok {
		log.Printf("menu: unknown role %q", name)
		return
	}
	menu.AddRole(role)
}
