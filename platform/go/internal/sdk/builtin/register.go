package builtin

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/cvrp"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/jobshop"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/vrptw"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/sdk"
)

func init() {
	must(sdk.RegisterProblem(sdk.ProblemDescriptor{
		Name:    "cvrp",
		Usage:   "owp solve cvrp --instance <path.vrp> [--mode sa|lahc|tabu|ga|portfolio]",
		Defaults: sdk.ProblemDefaults{
			Mode: "sa", Iterations: 500000, Temperature: 100.0, Seed: 42,
		},
		Load: loadCVRP,
	}))
	must(sdk.RegisterProblem(sdk.ProblemDescriptor{
		Name:    "vrptw",
		Usage:   "owp solve vrptw --instance <path.txt> [--mode sa|lahc|tabu|ga|portfolio]",
		Defaults: sdk.ProblemDefaults{
			Mode: "sa", Iterations: 500000, Temperature: 100.0, Seed: 42,
		},
		Load: loadVRPTW,
	}))
	must(sdk.RegisterProblem(sdk.ProblemDescriptor{
		Name:    "jobshop",
		Usage:   "owp solve jobshop --instance <path> [--mode sa|lahc|tabu|ga|adaptive|portfolio]",
		Defaults: sdk.ProblemDefaults{
			Mode: "sa", Iterations: 500000, Temperature: 100.0, Seed: 42,
		},
		Load: loadJobShop,
	}))
	must(sdk.RegisterProblem(sdk.ProblemDescriptor{
		Name:           "nrp",
		Usage:          "owp solve nrp --instance <name|dir> [--mode sa|lahc|tabu|ga|portfolio]",
		Title:          "NRP Solver",
		PolicyDomain:   "nrp",
		ObjectiveLabel: "Penalty",
		Defaults: sdk.ProblemDefaults{
			Mode: "sa", Iterations: 500000, Temperature: 100.0, Seed: 42,
		},
		Load: loadNRP,
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
		Name:         ds.Name,
		InstancePath: path,
		Data:         ds,
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
		Name:         ds.Name,
		InstancePath: path,
		Data:         ds,
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
		Name:         name,
		InstancePath: path,
		Data:         ds,
		Fields: map[string]string{
			"jobs":     fmt.Sprintf("%d", ds.Jobs),
			"machines": fmt.Sprintf("%d", ds.Machines),
		},
	}, nil
}

// nrpLoadData holds week-scoped NRP instance data for solve finalize hooks.
type nrpLoadData struct {
	Bundle   inrc2.InstanceBundle
	Week     inrc2.WeekData
	WeekFile string
}

func loadNRP(instanceName string) (sdk.Problem, sdk.InstanceMeta, error) {
	bundle, err := inrc2.LoadInstanceBundle(instanceName)
	if err != nil {
		return nil, sdk.InstanceMeta{}, err
	}
	wd, err := inrc2.LoadWeekData(bundle.WeekFiles[0])
	if err != nil {
		return nil, sdk.InstanceMeta{}, err
	}
	name := filepath.Base(bundle.Dir)
	problem := inrc2.NewNRPProblem(inrc2.NRPProblemConfig{
		Scenario: bundle.Scenario,
		WeekData: wd,
		History:  bundle.History,
	})
	return problem, sdk.InstanceMeta{
		Name:         name,
		InstancePath: instanceName,
		Data: nrpLoadData{
			Bundle: bundle, Week: wd, WeekFile: bundle.WeekFiles[0],
		},
		Fields: map[string]string{
			"nurses": fmt.Sprintf("%d", len(bundle.Scenario.Nurses)),
			"week":   filepath.Base(bundle.WeekFiles[0]),
		},
	}, nil
}
