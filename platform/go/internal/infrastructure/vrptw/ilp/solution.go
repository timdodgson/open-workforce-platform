package ilp

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/vrptw"
	nrpilp "github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/ilp"
)

type vrptwRouteJSON struct {
	Customers []int `json:"customers"`
	Load      int   `json:"load"`
	Distance  int   `json:"distance"`
	Feasible  bool  `json:"feasible"`
}

type vrptwSolutionJSON struct {
	Routes               []vrptwRouteJSON `json:"routes"`
	TotalCost            int              `json:"totalCost"`
	Vehicles             int              `json:"vehicles"`
	Feasible             bool             `json:"feasible"`
	TimeWindowViolations int              `json:"timeWindowViolations"`
}

// ExtractSolution maps HiGHS x_ij variables to dashboard solution.json.
func ExtractSolution(ds *vrptw.Dataset, output nrpilp.SolverOutput) ([]byte, error) {
	if len(output.SolutionValues) == 0 {
		return nil, fmt.Errorf("no solution values in solver output")
	}

	n := len(ds.Customers)
	numNodes := n + 1
	dist, timeMat := buildVRPTWMatrices(ds, numNodes)

	next := make(map[int]int)
	for i := 0; i < numNodes; i++ {
		for j := 0; j < numNodes; j++ {
			if i == j {
				continue
			}
			if output.SolutionValues[fmt.Sprintf("x_%d_%d", i, j)] >= 0.5 {
				next[i] = j
			}
		}
	}

	var routes []vrptwRouteJSON
	totalCost := 0
	twViolations := 0
	feasible := true

	for j := 1; j <= n; j++ {
		if output.SolutionValues[fmt.Sprintf("x_0_%d", j)] < 0.5 {
			continue
		}
		nodePath := []int{0}
		cur := j
		visited := 0
		for cur != 0 || len(nodePath) == 1 {
			nodePath = append(nodePath, cur)
			if cur == 0 {
				break
			}
			visited++
			if visited > numNodes+1 {
				return nil, fmt.Errorf("cycle detected leaving depot to node %d", j)
			}
			nxt, ok := next[cur]
			if !ok {
				return nil, fmt.Errorf("dead end at node %d", cur)
			}
			cur = nxt
		}

		custNodes := nodePath[1 : len(nodePath)-1]
		custIdxs := make([]int, len(custNodes))
		custIDs := make([]int, len(custNodes))
		load := 0
		for i, node := range custNodes {
			cust := ds.Customers[node-1]
			custIdxs[i] = node - 1
			custIDs[i] = cust.ID
			load += cust.Demand
		}

		routeDist := 0
		for i := 0; i < len(nodePath)-1; i++ {
			routeDist += dist[nodePath[i]][nodePath[i+1]]
		}

		routeFeasible, routeTW := checkVRPTWRoute(ds, dist, timeMat, custIdxs)
		if load > ds.Capacity {
			routeFeasible = false
		}
		if !routeFeasible {
			feasible = false
		}
		twViolations += routeTW

		routes = append(routes, vrptwRouteJSON{
			Customers: custIDs,
			Load:      load,
			Distance:  routeDist,
			Feasible:  routeFeasible,
		})
		totalCost += routeDist
	}

	if len(routes) == 0 {
		return nil, fmt.Errorf("no routes found in ILP solution")
	}

	out := vrptwSolutionJSON{
		Routes:               routes,
		TotalCost:            totalCost,
		Vehicles:             len(routes),
		Feasible:             feasible && twViolations == 0,
		TimeWindowViolations: twViolations,
	}
	if int(math.Round(output.Objective)) > 0 {
		out.TotalCost = int(math.Round(output.Objective))
	}
	return json.Marshal(out)
}

func buildVRPTWMatrices(ds *vrptw.Dataset, numNodes int) ([][]int, [][]int) {
	nodeX := make([]float64, numNodes)
	nodeY := make([]float64, numNodes)
	nodeX[0] = ds.Depot.X
	nodeY[0] = ds.Depot.Y
	for i, c := range ds.Customers {
		nodeX[i+1] = c.X
		nodeY[i+1] = c.Y
	}
	dist := make([][]int, numNodes)
	timeMat := make([][]int, numNodes)
	for i := 0; i < numNodes; i++ {
		dist[i] = make([]int, numNodes)
		timeMat[i] = make([]int, numNodes)
		for j := 0; j < numNodes; j++ {
			if i != j {
				dist[i][j] = vrptw.DistanceRounded(nodeX[i], nodeY[i], nodeX[j], nodeY[j])
				timeMat[i][j] = dist[i][j]
			}
		}
	}
	return dist, timeMat
}

func checkVRPTWRoute(ds *vrptw.Dataset, dist, timeMat [][]int, route []int) (bool, int) {
	if len(route) == 0 {
		return true, 0
	}
	violations := 0
	currentTime := float64(ds.Depot.ReadyTime)
	currentTime += float64(timeMat[0][route[0]+1])

	for i, custIdx := range route {
		c := ds.Customers[custIdx]
		if currentTime < float64(c.ReadyTime) {
			currentTime = float64(c.ReadyTime)
		}
		if currentTime > float64(c.DueDate) {
			violations++
		}
		currentTime += float64(c.Service)
		if i < len(route)-1 {
			currentTime += float64(timeMat[custIdx+1][route[i+1]+1])
		} else {
			currentTime += float64(timeMat[custIdx+1][0])
		}
	}
	if currentTime > float64(ds.Depot.DueDate) {
		violations++
	}
	return violations == 0, violations
}
