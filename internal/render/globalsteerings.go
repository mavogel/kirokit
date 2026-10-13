package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

// globalSteeringsDir is the kit-owned directory (under $KIRO_HOME) that holds
// the global steerings a team opted into. It is deliberately not Kiro's
// auto-loaded steering/ directory: these files reach an agent only through
// its "resources", so the opt-in stays per team, and each team gets its own
// copy so uninstalling one team never deletes a file another still uses.
const globalSteeringsDir = "kirokit-steerings"

// workspaceResources are added to every rendered agent so it loads the steering
// files and skills of whichever workspace it is started in (relative to the
// current directory). Kiro's docs say a custom agent has to list them in its
// resources explicitly. Skills use the skill:// scheme, which loads only a
// skill's metadata at startup and the rest on demand; file:// would put every
// skill's full text into the context.
var workspaceResources = []string{
	"file://.kiro/steering/**/*.md",
	"skill://.kiro/skills/**/SKILL.md",
}

// globalSteeringRel is the output-tree path of one opted-in global steering.
func globalSteeringRel(team, slug string) string {
	return path.Join(globalSteeringsDir, team, slug+".md")
}

// plainName reports whether s is a single path element, so it is safe to
// join into the output tree.
func plainName(s string) bool {
	return s != "" && s != "." && s != ".." && filepath.Base(s) == s
}

// globalSteeringURIs returns the file:// URI of every global steering the
// team opted into, in manifest order, rooted at the context's KiroHome.
func globalSteeringURIs(ctx *Context) []string {
	uris := make([]string, 0, len(ctx.Team.GlobalSteerings))
	for _, slug := range ctx.Team.GlobalSteerings {
		uris = append(uris, "file://"+path.Join(ctx.KiroHome, globalSteeringRel(ctx.Team.Team, slug)))
	}
	return uris
}

// writeGlobalSteerings writes each opted-in global steering's content into
// root, at the path globalSteeringURIs points to.
func writeGlobalSteerings(root string, ctx *Context) error {
	for _, slug := range ctx.Team.GlobalSteerings {
		if !plainName(slug) || !plainName(ctx.Team.Team) {
			return fmt.Errorf("render: invalid team %q or global steering name %q", ctx.Team.Team, slug)
		}
		out := filepath.Join(root, filepath.FromSlash(globalSteeringRel(ctx.Team.Team, slug)))
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return fmt.Errorf("render: %w", err)
		}
		if err := os.WriteFile(out, []byte(ctx.Global[slug]), 0o600); err != nil {
			return fmt.Errorf("render: write %s: %w", globalSteeringRel(ctx.Team.Team, slug), err)
		}
	}
	return nil
}

// addResources appends every uri not already listed to the top-level
// "resources" array of an agent config, creating the array when it is
// missing. It inserts text in place, so the template author's key order and
// indentation survive; only the new entries are formatted, to match the
// surrounding style.
func addResources(data []byte, uris []string) ([]byte, error) {
	if len(uris) == 0 {
		return data, nil
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		return nil, fmt.Errorf("agent config must be a JSON object")
	}

	insertAt := dec.InputOffset() // after "{", then after the last member's value
	memberSep := ""               // whitespace that introduces a member, taken from the last one
	for dec.More() {
		keyStart := skipSeparators(data, int(dec.InputOffset()))
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		memberSep = lineIndent(data, keyStart, memberSep)

		if keyTok == "resources" {
			return addToResourcesArray(data, dec, uris)
		}
		var skip json.RawMessage
		if err := dec.Decode(&skip); err != nil {
			return nil, err
		}
		insertAt = dec.InputOffset()
	}

	member := `"resources": [` + quoteJoin(uris) + "]"
	if data[insertAt-1] != '{' {
		member = "," + orSpace(memberSep) + member
	}
	return splice(data, int(insertAt), member), nil
}

// addToResourcesArray handles the array that follows the "resources" key the
// decoder just consumed.
func addToResourcesArray(data []byte, dec *json.Decoder, uris []string) ([]byte, error) {
	if tok, err := dec.Token(); err != nil || tok != json.Delim('[') {
		return nil, fmt.Errorf(`"resources" must be a JSON array`)
	}
	openEnd := int(dec.InputOffset())

	var existing []string
	lastEnd, lastStart := -1, -1
	for dec.More() {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, err
		}
		lastEnd = int(dec.InputOffset())
		lastStart = lastEnd - len(raw)
		var s string
		if json.Unmarshal(raw, &s) == nil {
			existing = append(existing, s)
		}
	}

	var missing []string
	for _, u := range uris {
		if !slices.Contains(existing, u) && !slices.Contains(missing, u) {
			missing = append(missing, u)
		}
	}
	if len(missing) == 0 {
		return data, nil
	}

	if lastEnd < 0 { // empty array
		return splice(data, openEnd, quoteJoin(missing)), nil
	}
	sep := orSpace(lineIndent(data, lastStart, ""))
	return splice(data, lastEnd, ","+sep+strings.Join(quoteAll(missing), ","+sep)), nil
}

// skipSeparators returns the offset of the first byte at or after pos that is
// neither whitespace nor the comma between members, i.e. where the next
// member's key begins.
func skipSeparators(data []byte, pos int) int {
	for pos < len(data) && strings.IndexByte(" \t\r\n,", data[pos]) >= 0 {
		pos++
	}
	return pos
}

// lineIndent returns the newline-plus-indentation that precedes pos when the
// text before it breaks onto a new line, and fallback otherwise.
func lineIndent(data []byte, pos int, fallback string) string {
	start := pos
	for start > 0 && strings.IndexByte(" \t\r\n", data[start-1]) >= 0 {
		start--
	}
	if run := string(data[start:pos]); strings.Contains(run, "\n") {
		return run
	}
	return fallback
}

func orSpace(s string) string {
	if s == "" {
		return " "
	}
	return s
}

func quoteAll(ss []string) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		b, _ := json.Marshal(s) // a string always marshals
		out[i] = string(b)
	}
	return out
}

func quoteJoin(ss []string) string { return strings.Join(quoteAll(ss), ", ") }

func splice(data []byte, at int, insert string) []byte {
	out := make([]byte, 0, len(data)+len(insert))
	out = append(out, data[:at]...)
	out = append(out, insert...)
	return append(out, data[at:]...)
}
