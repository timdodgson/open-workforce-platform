package inrc2

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
