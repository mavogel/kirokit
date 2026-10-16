package kiroconfig

import "testing"

func TestValidateAgentJSON_AcceptsWellFormed(t *testing.T) {
	data := []byte(`{
		"name": "devops",
		"mcpServers": {
			"jira-mcp": {"command": "sops", "args": ["exec-env", "path", "cmd"]}
		},
		"resources": ["file://README.md"]
	}`)
	if err := ValidateAgentJSON(data); err != nil {
		t.Fatalf("ValidateAgentJSON returned error for well-formed input: %v", err)
	}
}

func TestValidateAgentJSON_RejectsMcpServerMissingCommandAndURL(t *testing.T) {
	data := []byte(`{
		"mcpServers": {
			"jira-mcp": {"timeout": 60000}
		}
	}`)
	if err := ValidateAgentJSON(data); err == nil {
		t.Fatal("ValidateAgentJSON returned no error for an mcpServers entry with neither command nor url")
	}
}

func TestValidateAgentJSON_AcceptsHTTPForm(t *testing.T) {
	data := []byte(`{
		"mcpServers": {
			"api": {"url": "https://example.com"}
		}
	}`)
	if err := ValidateAgentJSON(data); err != nil {
		t.Fatalf("ValidateAgentJSON returned error for HTTP-form mcpServers entry: %v", err)
	}
}

func TestValidateAgentJSON_RejectsInvalidJSON(t *testing.T) {
	if err := ValidateAgentJSON([]byte(`not json`)); err == nil {
		t.Fatal("ValidateAgentJSON returned no error for invalid JSON")
	}
}

func TestValidateHookJSON_AcceptsWellFormed(t *testing.T) {
	data := []byte(`{
		"version": "v1",
		"hooks": [
			{
				"name": "Lint on save",
				"trigger": "FileSave",
				"matcher": "\\.go$",
				"action": {"type": "command", "command": "gofmt -l ."}
			}
		]
	}`)
	if err := ValidateHookJSON(data); err != nil {
		t.Fatalf("ValidateHookJSON returned error for well-formed input: %v", err)
	}
}

func TestValidateHookJSON_RejectsUnrecognizedTrigger(t *testing.T) {
	data := []byte(`{
		"version": "v1",
		"hooks": [
			{"name": "Bad", "trigger": "NotARealTrigger", "action": {"type": "command", "command": "x"}}
		]
	}`)
	if err := ValidateHookJSON(data); err == nil {
		t.Fatal("ValidateHookJSON returned no error for an unrecognized trigger")
	}
}
