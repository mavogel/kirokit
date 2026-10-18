package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Record describes one team's install record under a Kiro home.
type Record struct {
	Team     string
	Files    int
	Modified time.Time
}

// Installed lists the teams with an install record under kiroHome, sorted by
// name. A missing records folder means no team is installed.
func Installed(kiroHome string) ([]Record, error) {
	dir := filepath.Dir(installManifestPath(kiroHome, "x"))
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("install: read %s: %w", dir, err)
	}
	var records []Record
	for _, e := range entries {
		team, ok := strings.CutSuffix(e.Name(), ".json")
		if e.IsDir() || !ok {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("install: read %s: %w", path, err)
		}
		var m manifest
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("install: parse %s: %w", path, err)
		}
		info, err := e.Info()
		if err != nil {
			return nil, fmt.Errorf("install: stat %s: %w", path, err)
		}
		records = append(records, Record{Team: team, Files: len(m.Files), Modified: info.ModTime()})
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Team < records[j].Team })
	return records, nil
}

// UninstallOthers uninstalls every installed team except keep and returns
// them in order. It stops at the first failure and returns the teams removed
// so far with the error.
func UninstallOthers(kiroHome, keep string) ([]string, error) {
	records, err := Installed(kiroHome)
	if err != nil {
		return nil, err
	}
	var removed []string
	for _, r := range records {
		if r.Team == keep {
			continue
		}
		if err := Uninstall(kiroHome, r.Team); err != nil {
			return removed, err
		}
		removed = append(removed, r.Team)
	}
	return removed, nil
}
