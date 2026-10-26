package commands

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// argvEchoOut, when set, turns this test binary into the fixture MCP server:
// it writes the arguments it was started with, and the TEST_AUTH secret sops
// decrypted into its environment, to that file as JSON and exits.
const argvEchoOut = "KIROKIT_ARGV_ECHO_OUT"

type argvEcho struct {
	Args     []string `json:"args"`
	TestAuth string   `json:"testAuth"`
}

func TestMain(m *testing.M) {
	if out := os.Getenv(argvEchoOut); out != "" {
		b, err := json.Marshal(argvEcho{Args: os.Args[1:], TestAuth: os.Getenv("TEST_AUTH")})
		if err == nil {
			err = os.WriteFile(out, b, 0o600)
		}
		if err != nil {
			os.Exit(3)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// sopsServerArgs are the arguments the fixture agent passes its MCP server:
// one of each character class a real launch uses (URL, @, =, ${VAR}, :).
var sopsServerArgs = []string{"https://example.com/mcp?team=x", "--header", "Authorization:${TEST_AUTH}", "user@example.com", "a=b", "--port", "8080"}

// sopsFixtureRepo writes a kirokit repository with one team whose agent starts
// argv-echo through sops exec-env, using the committed test-only secrets.
func sopsFixtureRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	quoted := make([]string, len(sopsServerArgs))
	for i, a := range sopsServerArgs {
		b, _ := json.Marshal(a) //nolint:errchkjson // a string always marshals
		quoted[i] = string(b)
	}
	secrets, err := os.ReadFile(filepath.Join(thisRepo(t), "testdata", "sops", "secrets.enc.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	writeTree(t, repo, map[string]string{
		"env/mcp-versions.yaml":                            "argv-echo: \"1.0.0\"\n",
		"global/steerings/.gitkeep":                        "",
		"teams/sops-team/manifest.yaml":                    "team: sops-team\nglobalSteerings: []\n",
		"teams/sops-team/agents/echo/env/secrets.enc.yaml": string(secrets),
		"teams/sops-team/agents/echo/agent.json.tmpl": `{"name": "echo", "mcpServers": {"echo": {"command": {{ sopsCommand }}, "args": {{ sopsArgsJSON (printf "%s/kirokit-secrets/sops-team/echo/secrets.enc.yaml" .KiroHome) "argv-echo" ` +
			strings.Join(quoted, " ") + ` }}}}}`,
	})
	return repo
}

// installArgvEcho copies this test binary onto a fresh PATH folder as
// argv-echo, the way npx or a server binary sits on a user's PATH.
func installArgvEcho(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	name := "argv-echo"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	src, err := os.Open(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close() //nolint:errcheck // read-only
	dst, err := os.OpenFile(filepath.Join(bin, name), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		t.Fatal(err)
	}
	if err := dst.Close(); err != nil {
		t.Fatal(err)
	}
	return bin
}

// launchEnv is the environment Kiro CLI would start the server with: the
// user's, with argv-echo on PATH and only the test key available to sops.
func launchEnv(bin, keyFile, out string) []string {
	var env []string
	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		switch strings.ToUpper(k) {
		case "PATH", "SOPS_AGE_KEY", "SOPS_AGE_KEY_FILE", "SOPS_AGE_KEY_CMD", "SOPS_AGE_SSH_PRIVATE_KEY_FILE", argvEchoOut:
			continue
		}
		env = append(env, kv)
	}
	return append(env, "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"), "SOPS_AGE_KEY_FILE="+keyFile, argvEchoOut+"="+out)
}

func TestInstall_SopsLaunchReachesTheServerUnchanged(t *testing.T) {
	if _, err := exec.LookPath("sops"); err != nil {
		if os.Getenv("KIROKIT_REQUIRE_SOPS") == "1" {
			t.Fatalf("KIROKIT_REQUIRE_SOPS=1 but sops is not on PATH: %v", err)
		}
		t.Skip("sops is not on PATH; set KIROKIT_REQUIRE_SOPS=1 to require it")
	}
	key := filepath.Join(thisRepo(t), "testdata", "sops", "age-test-key.txt")
	kiroHome := filepath.Join(t.TempDir(), "kiro")
	t.Setenv("KIRO_HOME", kiroHome)
	t.Setenv("KIROKIT_REPO", sopsFixtureRepo(t))
	chdir(t, t.TempDir())
	if out, err := runRoot(t, "install", "sops-team"); err != nil {
		t.Fatalf("install: %v\n%s", err, out)
	}

	data, err := os.ReadFile(filepath.Join(kiroHome, "agents", "echo.json"))
	if err != nil {
		t.Fatal(err)
	}
	var agent struct {
		MCPServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &agent); err != nil {
		t.Fatalf("installed echo.json: %v\n%s", err, data)
	}
	server := agent.MCPServers["echo"]
	bin := installArgvEcho(t)

	// Kiro CLI starts the configured command with its args as one process.
	out := filepath.Join(t.TempDir(), "argv.json")
	launch := exec.Command(server.Command, server.Args...)
	launch.Env = launchEnv(bin, key, out)
	if b, err := launch.CombinedOutput(); err != nil {
		t.Fatalf("%s %q: %v\n%s", server.Command, server.Args, err, b)
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("the server did not start: %v", err)
	}
	var got argvEcho
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("server output %s: %v", raw, err)
	}
	if !reflect.DeepEqual(got.Args, sopsServerArgs) {
		t.Errorf("server argv = %q\nwant         %q", got.Args, sopsServerArgs)
	}
	if got.TestAuth != "Bearer s3cret value" {
		t.Errorf("TEST_AUTH = %q, want the decrypted %q", got.TestAuth, "Bearer s3cret value")
	}

	// Without the key, sops refuses and the server never starts.
	empty := filepath.Join(t.TempDir(), "keys.txt")
	if err := os.WriteFile(empty, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	refused := filepath.Join(t.TempDir(), "argv.json")
	noKey := exec.Command(server.Command, server.Args...)
	noKey.Env = launchEnv(bin, empty, refused)
	if b, err := noKey.CombinedOutput(); err == nil {
		t.Errorf("launch without the key succeeded:\n%s", b)
	}
	if _, err := os.Stat(refused); !os.IsNotExist(err) {
		t.Errorf("the server started without its secret (err=%v)", err)
	}
}
