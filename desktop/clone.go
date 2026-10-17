package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// cloneResult is the payload of clone:done. A true OK with an Error is the
// "no release yet" case: the clone is kept and the message is shown.
type cloneResult struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

const noReleaseMessage = "This repository has no release yet (no v* tag). Ask your team lead to tag one."

// Clone fetches the repository at address into <app folder>/repo and checks
// out its newest release tag. A refused address or a command already running
// comes back as the error; the rest runs in the kirokit tab and ends with
// clone:done.
func (a *App) Clone(address string) error {
	address = strings.TrimSpace(address)
	if !repoURLPattern.MatchString(address) {
		return errors.New(repoURLMessage) //nolint:staticcheck // shown to the person beside the address field
	}
	cli, err := BundledCLI()
	if err != nil {
		return err
	}
	if err := a.reserve(); err != nil {
		return err
	}

	s := a.loadSettings()
	s.RepoURL = address
	if err := a.writeSettings(s); err != nil {
		a.release()
		return err
	}
	go func() {
		defer a.release()
		a.emit("clone:done", a.cloneAndSync(address, cli))
	}()
	return nil
}

// cloneAndSync keeps any old clone under another name, clones the address and
// syncs to the newest tag. A failure puts the old clone back.
func (a *App) cloneAndSync(address, cli string) cloneResult {
	repo := a.repoDir()
	var old string
	if _, err := os.Stat(repo); err == nil {
		stamp := fmt.Sprintf("%s.old-%d", repo, time.Now().Unix())
		old = stamp
		for n := 2; exists(old); n++ {
			old = stamp + "-" + strconv.Itoa(n)
		}
		if err := os.Rename(repo, old); err != nil {
			return cloneResult{Error: "Could not set the old copy aside: " + err.Error()}
		}
	}
	fail := func(message string) cloneResult {
		_ = os.RemoveAll(repo) //nolint:errcheck // a leftover partial clone is replaced by the next try
		if old != "" {
			_ = os.Rename(old, repo) //nolint:errcheck // reported through the missing clone
		}
		return cloneResult{Error: message}
	}

	git := "git"
	if p, ok := a.loginEnv().Lookup("git"); ok {
		git = p
	}
	res, err := a.runStep("git", git, []string{"clone", "--", address, repo})
	if err != nil {
		return fail(err.Error())
	}
	if res.Code != 0 {
		return fail(fmt.Sprintf("git clone failed (exit %d). Check the address and your access, then try again.", res.Code))
	}

	res, err = a.runStep("kirokit", cli, []string{"sync", "--latest-tag"})
	switch {
	case err != nil:
		return fail(err.Error())
	case res.Code == 0:
		return cloneResult{OK: true}
	case strings.Contains(res.Tail, "no v* tag found"):
		return cloneResult{OK: true, Error: noReleaseMessage}
	default:
		return fail(fmt.Sprintf("kirokit sync failed (exit %d). See the kirokit tab.", res.Code))
	}
}

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}
