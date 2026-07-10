// Package byodtsp registers a minimal TSP domain with the owp-sdk (BYOD demo).
package byodtsp

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/timdodgson/open-workforce-platform/examples/byod-tsp/tsp"
	"github.com/timdodgson/open-workforce-platform/owp-sdk/sdk"
)

func init() {
	if err := sdk.RegisterProblem(sdk.ProblemDescriptor{
		Name:    "tsp",
		Usage:   "owp solve tsp --instance <path.json> [--mode sa|lahc|tabu|portfolio]",
		Defaults: sdk.ProblemDefaults{
			Mode: "sa", Iterations: 100000, Temperature: 50.0, Seed: 42,
		},
		Load: loadTSP,
	}); err != nil {
		panic("byod-tsp: " + err.Error())
	}
}

func loadTSP(path string) (sdk.Problem, sdk.InstanceMeta, error) {
	ds, err := tsp.LoadDataset(path)
	if err != nil {
		return nil, sdk.InstanceMeta{}, err
	}
	name := ds.Name
	if name == path {
		name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	return tsp.NewProblem(ds), sdk.InstanceMeta{
		Name:         name,
		InstancePath: path,
		Data:         ds,
		Fields: map[string]string{
			"cities": fmt.Sprintf("%d", len(ds.Cities)),
		},
	}, nil
}
