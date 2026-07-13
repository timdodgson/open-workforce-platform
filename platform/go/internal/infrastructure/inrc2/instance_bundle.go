package inrc2

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// InstanceBundle holds loaded INRC-II instance data from a directory.
type InstanceBundle struct {
	Dir          string
	ScenarioFile string
	WeekFiles    []string
	HistFiles    []string
	Scenario     Scenario
	History      History
}

// DefaultInstanceSearchPaths are checked when resolving a short instance name.
var DefaultInstanceSearchPaths = []string{
	"../../examples/inrc2/testdatasets_json",
	"../../examples/inrc2/datasets_json",
}

// ResolveInstanceDir locates an INRC-II instance directory by path or short name.
func ResolveInstanceDir(instanceName string, searchPaths []string) (string, error) {
	if instanceName == "" {
		return "", fmt.Errorf("instance name is required")
	}
	if _, err := os.Stat(instanceName); err == nil {
		return instanceName, nil
	}
	if searchPaths == nil {
		searchPaths = DefaultInstanceSearchPaths
	}
	for _, base := range searchPaths {
		candidate := filepath.Join(base, instanceName)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("instance not found: %s", instanceName)
}

// ScanInstanceDir lists Sc-, WD-, and H0- files in an instance directory.
func ScanInstanceDir(dir string) (scenarioFile string, weekFiles, histFiles []string, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", nil, nil, fmt.Errorf("read instance dir: %w", err)
	}
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "Sc-") && strings.HasSuffix(name, ".json") {
			scenarioFile = filepath.Join(dir, name)
		} else if strings.HasPrefix(name, "WD-") && strings.HasSuffix(name, ".json") {
			weekFiles = append(weekFiles, filepath.Join(dir, name))
		} else if strings.HasPrefix(name, "H0-") && strings.HasSuffix(name, ".json") {
			histFiles = append(histFiles, filepath.Join(dir, name))
		}
	}
	sort.Strings(weekFiles)
	sort.Strings(histFiles)
	if scenarioFile == "" || len(weekFiles) == 0 || len(histFiles) == 0 {
		return "", nil, nil, fmt.Errorf("incomplete instance data in %s (need Sc-, WD-, H0- files)", dir)
	}
	return scenarioFile, weekFiles, histFiles, nil
}

// LoadInstanceBundle resolves and loads a complete INRC-II instance by name or path.
func LoadInstanceBundle(instanceName string) (InstanceBundle, error) {
	dir, err := ResolveInstanceDir(instanceName, nil)
	if err != nil {
		return InstanceBundle{}, err
	}
	scenarioFile, weekFiles, histFiles, err := ScanInstanceDir(dir)
	if err != nil {
		return InstanceBundle{}, err
	}
	sc, err := LoadScenario(scenarioFile)
	if err != nil {
		return InstanceBundle{}, fmt.Errorf("load scenario: %w", err)
	}
	hist, err := LoadHistory(histFiles[0])
	if err != nil {
		return InstanceBundle{}, fmt.Errorf("load history: %w", err)
	}
	return InstanceBundle{
		Dir:          dir,
		ScenarioFile: scenarioFile,
		WeekFiles:    weekFiles,
		HistFiles:    histFiles,
		Scenario:     sc,
		History:      hist,
	}, nil
}

// SelectHistory loads history file H0-<id>-<historyIndex>.json from the bundle.
// historyIndex < 0 leaves the currently loaded History unchanged.
func (b *InstanceBundle) SelectHistory(historyIndex int) error {
	if historyIndex < 0 {
		return nil
	}
	suffix := fmt.Sprintf("-%d.json", historyIndex)
	for _, path := range b.HistFiles {
		if strings.HasSuffix(filepath.Base(path), suffix) {
			hist, err := LoadHistory(path)
			if err != nil {
				return fmt.Errorf("load history %d: %w", historyIndex, err)
			}
			b.History = hist
			return nil
		}
	}
	return fmt.Errorf("history index %d not found in %s", historyIndex, b.Dir)
}

// SelectWeeks reorders WeekFiles to the given WD indices (e.g. 6,2,9,1 → competition weeks 6-2-9-1).
// Empty weekIndices leaves WeekFiles unchanged.
func (b *InstanceBundle) SelectWeeks(weekIndices []int) error {
	if len(weekIndices) == 0 {
		return nil
	}
	byIndex := make(map[int]string, len(b.WeekFiles))
	for _, path := range b.WeekFiles {
		base := filepath.Base(path)
		// WD-<scenario>-<n>.json
		trimmed := strings.TrimSuffix(base, ".json")
		i := strings.LastIndex(trimmed, "-")
		if i < 0 {
			continue
		}
		n, err := strconv.Atoi(trimmed[i+1:])
		if err != nil {
			continue
		}
		byIndex[n] = path
	}
	selected := make([]string, 0, len(weekIndices))
	for _, idx := range weekIndices {
		path, ok := byIndex[idx]
		if !ok {
			return fmt.Errorf("week index %d not found in %s", idx, b.Dir)
		}
		selected = append(selected, path)
	}
	b.WeekFiles = selected
	return nil
}

// ParseWeekSequence parses competition week strings like "6-2-9-1".
func ParseWeekSequence(s string) ([]int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, "-")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			return nil, fmt.Errorf("empty week token in %q", s)
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid week index %q in %q", p, s)
		}
		out = append(out, n)
	}
	return out, nil
}
