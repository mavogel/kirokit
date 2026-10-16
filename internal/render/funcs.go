package render

import (
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/mavogel/kiro-cli-kit/internal/sopswrap"
)

// FuncMap returns the text/template functions available to every team
// template, bound to the given render Context.
func FuncMap(ctx *Context) template.FuncMap {
	return template.FuncMap{
		"mcpVersion":     mcpVersionFunc(ctx),
		"globalSteering": globalSteeringFunc(ctx),
		"jsonString":     jsonStringFunc,
		"sopsCommand":    sopsCommandFunc,
		"sopsArgsJSON":   sopsArgsJSONFunc,
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
// ["exec-env", "<secretsFilePath>", "<shell-quoted command>"] string,
// ready to paste directly into a template's "args" field.
func sopsArgsJSONFunc(secretsFilePath, realCommand string, realArgs ...string) (string, error) {
	_, args := sopswrap.BuildWrapper(secretsFilePath, realCommand, realArgs)
	b, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("sopsArgsJSON: %w", err)
	}
	return string(b), nil
}
