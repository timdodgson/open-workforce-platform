package builtin

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/cvrp"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/jobshop"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/vrptw"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/sdk"
)

func init() {
	must(sdk.RegisterProblem(sdk.ProblemDescriptor{
		Name:    "cvrp",
		Command: "solve-cvrp",
		Usage:   "owp solve-cvrp --instance <path.vrp> [--mode sa|lahc|tabu|portfolio]",
		Defaults: sdk.ProblemDefaults{
			Mode: "sa", Iterations: 500000, Temperature: 100.0, Seed: 42,
		},
		Load: loadCVRP,
	}))
	must(sdk.RegisterProblem(sdk.ProblemDescriptor{
		Name:    "vrptw",
		Command: "solve-vrptw",
		Usage:   "owp solve-vrptw --instance <path.txt> [--mode sa|lahc|tabu|portfolio]",
		Defaults: sdk.ProblemDefaults{
			Mode: "sa", Iterations: 500000, Temperature: 100.0, Seed: 42,
		},
		Load: loadVRPTW,
	}))
	must(sdk.RegisterProblem(sdk.ProblemDescriptor{
		Name:    "jobshop",
		Command: "solve-jobshop",
		Usage:   "owp solve-jobshop --instance <path> [--mode sa|lahc|adaptive|portfolio]",
		Defaults: sdk.ProblemDefaults{
			Mode: "sa", Iterations: 500000, Temperature: 100.0, Seed: 42,
		},
		Load: loadJobShop,
	}))
}

func must(err error) {
	if err != nil {
		panic("sdk/builtin: " + err.Error())
	}
}

func loadCVRP(path string) (sdk.Problem, sdk.InstanceMeta, error) {
	ds, err := cvrp.LoadDataset(path)
	if err != nil {
		return nil, sdk.InstanceMeta{}, err
	}
	return cvrp.NewCVRPProblem(ds), sdk.InstanceMeta{
		Name: ds.Name,
		Fields: map[string]string{
			"customers": fmt.Sprintf("%d", len(ds.Customers)),
			"capacity":  fmt.Sprintf("%d", ds.Capacity),
		},
	}, nil
}

func loadVRPTW(path string) (sdk.Problem, sdk.InstanceMeta, error) {
	ds, err := vrptw.LoadDataset(path)
	if err != nil {
		return nil, sdk.InstanceMeta{}, err
	}
	return vrptw.NewVRPTWProblem(ds), sdk.InstanceMeta{
		Name: ds.Name,
		Fields: map[string]string{
			"customers": fmt.Sprintf("%d", len(ds.Customers)),
			"capacity":  fmt.Sprintf("%d", ds.Capacity),
			"vehicles":  fmt.Sprintf("%d", ds.Vehicles),
			"horizon":   fmt.Sprintf("[%d, %d]", ds.Depot.ReadyTime, ds.Depot.DueDate),
		},
	}, nil
}

func loadJobShop(path string) (sdk.Problem, sdk.InstanceMeta, error) {
	ds, err := jobshop.LoadDataset(path)
	if err != nil {
		return nil, sdk.InstanceMeta{}, err
	}
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	return jobshop.NewJSSProblem(ds), sdk.InstanceMeta{
		Name: name,
		Fields: map[string]string{
			"jobs":     fmt.Sprintf("%d", ds.Jobs),
			"machines": fmt.Sprintf("%d", ds.Machines),
		},
	}, nil
}
