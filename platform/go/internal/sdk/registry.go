package sdk

import (
	"fmt"
	"sort"
	"sync"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation/searchdef"
)

var (
	searchMu sync.RWMutex
	searches = map[string]SearchRunner{}
)

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
	searchMu.Lock()
	defer searchMu.Unlock()
	if _, exists := searches[mode]; exists {
		return fmt.Errorf("sdk: search mode %q already registered", mode)
	}
	searches[mode] = runner
	return nil
}

// RegisteredSearchModes returns custom search mode names in sorted order.
func RegisteredSearchModes() []string {
	searchMu.RLock()
	defer searchMu.RUnlock()
	modes := make([]string, 0, len(searches))
	for mode := range searches {
		modes = append(modes, mode)
	}
	sort.Strings(modes)
	return modes
}

// ResolveSearchRunner returns the runner for a mode: custom registration first, else built-in.
func ResolveSearchRunner(mode string) SearchRunner {
	searchMu.RLock()
	runner, ok := searches[mode]
	searchMu.RUnlock()
	if ok {
		return runner
	}
	return optimisation.RunSearch
}

// RunSearch executes a registered or built-in search mode.
func RunSearch(problem searchdef.Problem, config optimisation.SearchConfig) optimisation.SearchResult {
	return ResolveSearchRunner(config.Mode)(problem, config)
}
