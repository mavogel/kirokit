package commands

import (
	"strings"
	"testing"
)

func TestParseKnowledgeBaseFlags_AllOrNone(t *testing.T) {
	tests := []struct {
		name                  string
		agent, source, mount  string
		wantNil, wantTogether bool
	}{
		{"none given", "", "", "", true, false},
		{"all given", "devops", "/kb", "/mnt", false, false},
		{"only agent", "devops", "", "", false, true},
		{"only source", "", "/kb", "", false, true},
		{"only mount", "", "", "/mnt", false, true},
		{"agent and source", "devops", "/kb", "", false, true},
		{"agent and mount", "devops", "", "/mnt", false, true},
		{"source and mount", "", "/kb", "/mnt", false, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kb, err := parseKnowledgeBaseFlags(tc.agent, tc.source, tc.mount)

			if tc.wantTogether {
				if err == nil || !strings.Contains(err.Error(), "must be given together") {
					t.Fatalf("err = %v, want a %q error", err, "must be given together")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantNil {
				if kb != nil {
					t.Errorf("kb = %+v, want nil when no flag is given", kb)
				}
				return
			}
			if kb == nil || kb.Agent != tc.agent || kb.Source != tc.source || kb.Mount != tc.mount {
				t.Errorf("kb = %+v, want agent=%q source=%q mount=%q", kb, tc.agent, tc.source, tc.mount)
			}
		})
	}
}
