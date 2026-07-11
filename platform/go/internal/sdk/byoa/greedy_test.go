package byoa_test

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/sdk"
	_ "github.com/timdodgson/open-workforce-platform/platform/go/internal/sdk/byoa"
)

func TestGreedyModeRegistered(t *testing.T) {
	modes := sdk.RegisteredSearchModes()
	found := false
	for _, m := range modes {
		if m == "greedy" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected greedy in custom modes, got %v", modes)
	}
}
