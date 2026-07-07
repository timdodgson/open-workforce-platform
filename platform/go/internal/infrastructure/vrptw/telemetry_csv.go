package vrptw

import (
	"fmt"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// DiscoveriesCSVHeader is the canonical header for VRPTW discoveries.csv.
const DiscoveriesCSVHeader = "elapsed_ms,candidate,old_best,new_best,improvement\n"

// BuildDiscoveriesCSV returns discoveries.csv bytes for VRPTW runs.
func BuildDiscoveriesCSV(discoveries []optimisation.Discovery) []byte {
	if len(discoveries) == 0 {
		return nil
	}
	var b strings.Builder
	b.WriteString(DiscoveriesCSVHeader)
	for _, d := range discoveries {
		b.WriteString(fmt.Sprintf("%d,%d,%d,%d,%d\n", d.ElapsedMs, d.Candidate, d.OldBest, d.NewBest, d.Improvement))
	}
	return []byte(b.String())
}
