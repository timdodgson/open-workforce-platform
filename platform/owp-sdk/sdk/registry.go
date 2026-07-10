package sdk

import (
	"fmt"
	"sort"
	"sync"
)

var (
	registryMu sync.RWMutex
	problems   = map[string]ProblemDescriptor{}
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
