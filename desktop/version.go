package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/mavogel/kirokit/internal/checkout"
)

// refResult is the payload of ref:done.
type refResult struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

// Checkout reads what the clone is on and the branches and tags it can switch
// to, from the clone's own refs: no network, so no sign-in prompt outside the
// kirokit tab. Each sync's fetch refreshes them. Before a clone it returns
// empty lists and no error.
func (a *App) Checkout() (checkout.Refs, error) {
	repo := a.repoDir()
	if !exists(filepath.Join(repo, ".git")) {
		return checkout.Refs{Branches: []string{}, Tags: []string{}}, nil
	}
	git := "git"
	if p, ok := a.loginEnv().Lookup("git"); ok {
		git = p
	}
	return checkout.Read(git, repo)
}

// SwitchRef runs `kirokit sync [installed team] --branch <ref>`, or
// `--latest-tag` for "" (Newest release), in the kirokit tab, and saves ref as
// the choice every later sync uses only once the command succeeded. A bad name
// or a command already running is the returned error; the rest ends with
// ref:done.
func (a *App) SwitchRef(ref string) error {
	ref = strings.TrimSpace(ref)
	if ref != "" {
		// The list the page offers comes from git, but this method is bound
		// and must not trust its caller.
		if err := checkout.CheckRefName(ref); err != nil {
			return fmt.Errorf("kirokit sync: --branch: %w", err)
		}
	}
	cli, err := BundledCLI()
	if err != nil {
		return err
	}
	if err := a.reserve(); err != nil {
		return err
	}
	go func() {
		defer a.release()
		a.emit("ref:done", a.switchRef(cli, ref))
	}()
	return nil
}

func (a *App) switchRef(cli, ref string) refResult {
	args := []string{"sync"}
	// Unreadable install records leave the team out: the checkout still switches.
	if records, _ := a.Installs(); len(records) > 0 {
		args = append(args, records[0].Team)
	}
	if ref == "" {
		args = append(args, "--latest-tag")
	} else {
		args = append(args, "--branch", ref)
	}
	res, err := a.runStep("kirokit", cli, args)
	if err != nil {
		return refResult{Error: err.Error()}
	}
	if res.Code != 0 {
		return refResult{Error: syncFailure(res)}
	}
	s := a.loadSettings()
	s.Ref = ref
	if err := a.writeSettings(s); err != nil {
		return refResult{Error: "The version was switched but could not be saved: " + err.Error()}
	}
	return refResult{OK: true}
}

// terminalEscape matches the control sequences a terminal adds around text.
var terminalEscape = regexp.MustCompile(`\x1b(?:\[[0-9;?]*[ -/]*[@-~]|\][^\x07\x1b]*(?:\x07|\x1b\\))`)

// syncFailure is the reason a failed sync gives: the last line of its output
// that is the CLI's own message, else a pointer to the kirokit tab.
func syncFailure(res runResult) string {
	lines := strings.Split(terminalEscape.ReplaceAllString(res.Tail, ""), "\n")
	for _, line := range slices.Backward(lines) {
		if line = strings.TrimSpace(line); strings.HasPrefix(line, "kirokit sync:") {
			return line
		}
	}
	return fmt.Sprintf("kirokit sync failed (exit %d). See the kirokit tab.", res.Code)
}
