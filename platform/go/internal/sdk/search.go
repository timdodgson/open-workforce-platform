package sdk

import (
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation/searchdef"
)

// SearchRunner runs a single search mode for a problem.
// Register custom modes via RegisterSearch; built-in modes delegate to optimisation.RunSearch.
type SearchRunner func(problem searchdef.Problem, config optimisation.SearchConfig) optimisation.SearchResult

// BuiltInSearchModes lists modes always available via optimisation.RunSearch.
var BuiltInSearchModes = []string{"sa", "lahc", "tabu", "ga", "portfolio", "adaptive"}
