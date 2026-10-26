package render

import (
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/mavogel/kirokit/internal/manifest"
	"github.com/mavogel/kirokit/internal/sopswrap"
)

// FuncMap returns the text/template functions available to every team
// template, bound to the given render Context.
func FuncMap(ctx *Context) template.FuncMap {
	return template.FuncMap{
		"mcpVersion":     mcpVersionFunc(ctx),
		"globalSteering": globalSteeringFunc(ctx),
		"jsonString":     jsonStringFunc,
		"sopsCommand":    sopsCommandFunc,
		"sopsArgsJSON":   sopsArgsJSONFunc(ctx.LaunchShell),
		"owners":         ownersFunc(ctx),
		"feedbackRepo":   feedbackRepoFunc(ctx),
	}
}

// OwnerRow is one agent or skill of the team with its owner. Kind is
// "agent" or "skill"; Owner is "" when neither the item nor the team names
// one, and FromTeam is true when Owner is the team owner.
type OwnerRow struct {
	Kind     string
	Name     string
	Owner    string
	FromTeam bool
}

// ownersFunc lists every agent, then every skill, of the team by name, each
// with its owner.
func ownersFunc(ctx *Context) func() []OwnerRow {
	return func() []OwnerRow {
		rows := make([]OwnerRow, 0, len(ctx.Agents)+len(ctx.Skills))
		for _, group := range []struct {
			kind  string
			names []string
		}{{manifest.KindAgent, ctx.Agents}, {manifest.KindSkill, ctx.Skills}} {
			for _, name := range group.names {
				owner, fromTeam := ctx.Team.OwnerOf(group.kind, name)
				rows = append(rows, OwnerRow{Kind: group.kind, Name: name, Owner: owner, FromTeam: fromTeam})
			}
		}
		return rows
	}
}

// feedbackRepoFunc returns the owner/repo that receives the team's feedback
// issues, and fails like mcpVersion does when the manifest names none.
func feedbackRepoFunc(ctx *Context) func() (string, error) {
	return func() (string, error) {
		if ctx.Team.Feedback.GitHub == "" {
			return "", fmt.Errorf("feedbackRepo: %s's manifest.yaml sets no feedback.github", ctx.Team.Team)
		}
		return ctx.Team.Feedback.GitHub, nil
	}
}

func mcpVersionFunc(ctx *Context) func(string) (string, error) {
	return func(name string) (string, error) {
		v, ok := ctx.Env[name]
		if !ok {
			return "", fmt.Errorf("mcpVersion: no pinned version for %q in env/mcp-versions.yaml", name)
		}
		return v, nil
	}
}

func globalSteeringFunc(ctx *Context) func(string) (string, error) {
	return func(slug string) (string, error) {
		v, ok := ctx.Global[slug]
		if !ok {
			return "", fmt.Errorf("globalSteering: %q is not available; the team's manifest.yaml must opt into it under globalSteerings", slug)
		}
		return v, nil
	}
}

// jsonStringFunc marshals a Go string into a properly quoted+escaped JSON
// string literal, for embedding a resolved value inside a template-authored
// JSON file.
func jsonStringFunc(s string) (string, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return "", fmt.Errorf("jsonString: %w", err)
	}
	return string(b), nil
}

// sopsCommandFunc returns the JSON string literal "sops" (quotes included),
// ready to paste directly into a template's "command" field.
func sopsCommandFunc() string {
	return `"sops"`
}

// sopsArgsJSONFunc returns the already-JSON-array-formatted
// ["exec-env", "<secretsFilePath>", "<command for shell>"] string, ready to
// paste directly into a template's "args" field. An argument the shell cannot
// pass fails with a *sopswrap.UnsafeArgError.
func sopsArgsJSONFunc(shell sopswrap.Shell) func(secretsFilePath, realCommand string, realArgs ...string) (string, error) {
	return func(secretsFilePath, realCommand string, realArgs ...string) (string, error) {
		_, args, err := sopswrap.BuildWrapperFor(shell, secretsFilePath, realCommand, realArgs)
		if err != nil {
			return "", err
		}
		b, err := json.Marshal(args)
		if err != nil {
			return "", fmt.Errorf("sopsArgsJSON: %w", err)
		}
		return string(b), nil
	}
}
