package golden

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mavogel/kirokit/internal/render"
	"github.com/mavogel/kirokit/internal/sopswrap"
)

// Every team in the repository must install on Windows, where sops runs MCP
// commands with cmd.exe and refuses arguments it cannot pass.
func TestEveryTeamRendersForTheWindowsLaunchShell(t *testing.T) {
	root := filepath.Join("..", "..")
	entries, err := os.ReadDir(filepath.Join(root, "teams"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		t.Run(e.Name(), func(t *testing.T) {
			err := render.RenderTeamWith(filepath.Join(root, "teams"), e.Name(), filepath.Join(root, "global", "steerings"),
				filepath.Join(root, "env", "mcp-versions.yaml"), filepath.Join(t.TempDir(), "out"),
				render.Options{KiroHome: "C:/Users/a/.kiro", LaunchShell: sopswrap.Cmd})
			if err != nil {
				t.Fatalf("render for Windows: %v", err)
			}
		})
	}
}
