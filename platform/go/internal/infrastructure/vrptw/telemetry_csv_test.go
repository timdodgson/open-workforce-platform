package vrptw

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func TestDiscoveriesCSVHeaderUnchanged(t *testing.T) {
	data := BuildDiscoveriesCSV([]optimisation.Discovery{
		{ElapsedMs: 1, Candidate: 2, OldBest: 3, NewBest: 2, Improvement: 1},
	})
	if string(data) != "elapsed_ms,candidate,old_best,new_best,improvement\n1,2,3,2,1\n" {
		t.Fatalf("unexpected VRPTW discoveries CSV:\n%s", data)
	}
}
