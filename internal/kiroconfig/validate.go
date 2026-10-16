// Package kiroconfig implements minimal structural validation for the
// standalone JSON shapes kirokit renders into a Kiro CLI v3 config tree:
// agent configs (agents/<name>.json) and standalone hook files
// (hooks/<name>.json). Validation is intentionally shallow — structural
// shape only — since golden files remain the primary correctness signal.
package kiroconfig

import (
	"encoding/json"
	"fmt"
)

// validHookTriggers are the verified Kiro CLI v3 standalone hook trigger
// names (distinct from the agent-inline hooks field's camelCase events).
var validHookTriggers = map[string]bool{
	"PromptSubmit":      true,
	"AgentStop":         true,
	"SessionStart":      true,
	"AgentSpawn":        true,
	"PreToolUse":        true,
	"PostToolUse":       true,
	"FileCreate":        true,
	"FileSave":          true,
	"FileDelete":        true,
	"PreTaskExecution":  true,
	"PostTaskExecution": true,
}

// ValidateAgentJSON checks that data is a well-formed Kiro CLI v3 agent
// config: valid JSON, and if present, every mcpServers entry has either a
// stdio "command" or an HTTP "url", and "resources" (if present) is an
// array.
func ValidateAgentJSON(data []byte) error {
	var agent struct {
		MCPServers map[string]json.RawMessage `json:"mcpServers"`
		Resources  json.RawMessage            `json:"resources"`
	}
	if err := json.Unmarshal(data, &agent); err != nil {
		return fmt.Errorf("kiroconfig: invalid agent JSON: %w", err)
	}

	for name, raw := range agent.MCPServers {
		var server struct {
			Command *string `json:"command"`
			URL     *string `json:"url"`
		}
		if err := json.Unmarshal(raw, &server); err != nil {
			return fmt.Errorf("kiroconfig: mcpServers[%q]: invalid entry: %w", name, err)
		}
		if server.Command == nil && server.URL == nil {
			return fmt.Errorf("kiroconfig: mcpServers[%q]: must have either %q or %q", name, "command", "url")
		}
	}

	if len(agent.Resources) > 0 {
		var resources []json.RawMessage
		if err := json.Unmarshal(agent.Resources, &resources); err != nil {
			return fmt.Errorf("kiroconfig: %q must be a JSON array: %w", "resources", err)
		}
	}

	return nil
}

// ValidateHookJSON checks that data is a well-formed standalone Kiro CLI v3
// hook file: a "version" string, a "hooks" array, and every hook entry has
// a "name", a recognized "trigger", and an "action" with a "type".
func ValidateHookJSON(data []byte) error {
	var file struct {
		Version string `json:"version"`
		Hooks   []struct {
			Name    string `json:"name"`
			Trigger string `json:"trigger"`
			Action  struct {
				Type string `json:"type"`
			} `json:"action"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return fmt.Errorf("kiroconfig: invalid hook JSON: %w", err)
	}
	if file.Version == "" {
		return fmt.Errorf("kiroconfig: hook file missing required field %q", "version")
	}

	for i, h := range file.Hooks {
		if h.Name == "" {
			return fmt.Errorf("kiroconfig: hooks[%d]: missing required field %q", i, "name")
		}
		if !validHookTriggers[h.Trigger] {
			return fmt.Errorf("kiroconfig: hooks[%d]: unrecognized trigger %q", i, h.Trigger)
		}
		if h.Action.Type != "command" && h.Action.Type != "agent" {
			return fmt.Errorf("kiroconfig: hooks[%d]: action.type must be %q or %q, got %q", i, "command", "agent", h.Action.Type)
		}
	}

	return nil
}
