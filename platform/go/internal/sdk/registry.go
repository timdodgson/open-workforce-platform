package sdk

import (
	"fmt"
	"sort"
	"sync"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation/searchdef"
)

var (
	registryMu sync.RWMutex
	problems   = map[string]ProblemDescriptor{}
	searches   = map[string]SearchRunner{}
)

// RegisterProblem registers a domain loader. Name must be unique.
func RegisterProblem(desc ProblemDescriptor) error {
	if desc.Name == "" {
		return fmt.Errorf("sdk: problem name is required")
	}
	if desc.Load == nil {
		return fmt.Errorf("sdk: problem %q: Load is required", desc.Name)
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := problems[desc.Name]; exists {
		return fmt.Errorf("sdk: problem %q already registered", desc.Name)
	}
	problems[desc.Name] = desc
	return nil
}

// RegisterSearch registers a custom search mode. Mode must be unique among registered modes.
// Built-in modes (sa, lahc, tabu, portfolio, adaptive) are always resolved via optimisation.RunSearch
// unless explicitly overridden here.
func RegisterSearch(mode string, runner SearchRunner) error {
	if mode == "" {
		return fmt.Errorf("sdk: search mode is required")
	}
	if runner == nil {
		return fmt.Errorf("sdk: search mode %q: runner is required", mode)
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := searches[mode]; exists {
		return fmt.Errorf("sdk: search mode %q already registered", mode)
	}
	searches[mode] = runner
	return nil
}

// GetProblem returns a registered problem descriptor by domain name.
func GetProblem(name string) (ProblemDescriptor, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	desc, ok := problems[name]
	return desc, ok
}

// Problems returns registered problem names in sorted order.
func Problems() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	names := make([]string, 0, len(problems))
	for name := range problems {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// RegisteredSearchModes returns custom search mode names in sorted order.
func RegisteredSearchModes() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	modes := make([]string, 0, len(searches))
	for mode := range searches {
		modes = append(modes, mode)
	}
	sort.Strings(modes)
	return modes
}

// ResolveSearchRunner returns the runner for a mode: custom registration first, else built-in.
func ResolveSearchRunner(mode string) SearchRunner {
	registryMu.RLock()
	runner, ok := searches[mode]
	registryMu.RUnlock()
	if ok {
		return runner
	}
	return optimisation.RunSearch
}

// RunSearch executes a registered or built-in search mode.
func RunSearch(problem searchdef.Problem, config optimisation.SearchConfig) optimisation.SearchResult {
	return ResolveSearchRunner(config.Mode)(problem, config)
}
