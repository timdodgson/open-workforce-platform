package ilp

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/cvrp"
	nrpilp "github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/ilp"
)

type routeJSON struct {
	Customers []int `json:"customers"`
	Load      int   `json:"load"`
	Distance  int   `json:"distance"`
}

type routingSolutionJSON struct {
	Routes    []routeJSON `json:"routes"`
	TotalCost int         `json:"totalCost"`
	Vehicles  int         `json:"vehicles"`
	Feasible  bool        `json:"feasible"`
}

// ExtractSolution maps HiGHS x_ij and u_i variables to dashboard solution.json.
func ExtractSolution(ds *cvrp.Dataset, output nrpilp.SolverOutput) ([]byte, error) {
	if len(output.SolutionValues) == 0 {
		return nil, fmt.Errorf("no solution values in solver output")
	}

	n := len(ds.Customers)
	numNodes := n + 1
	dist := buildDistanceMatrix(ds, numNodes)

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

	var routes []routeJSON
	totalCost := 0
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
		custIDs := make([]int, len(custNodes))
		load := 0
		for i, node := range custNodes {
			cust := ds.Customers[node-1]
			custIDs[i] = cust.ID
			load += cust.Demand
		}

		routeDist := 0
		for i := 0; i < len(nodePath)-1; i++ {
			routeDist += dist[nodePath[i]][nodePath[i+1]]
		}
		if load > ds.Capacity {
			feasible = false
		}

		routes = append(routes, routeJSON{
			Customers: custIDs,
			Load:      load,
			Distance:  routeDist,
		})
		totalCost += routeDist
	}

	if len(routes) == 0 {
		return nil, fmt.Errorf("no routes found in ILP solution")
	}

	out := routingSolutionJSON{
		Routes:    routes,
		TotalCost: totalCost,
		Vehicles:  len(routes),
		Feasible:  feasible,
	}
	if int(math.Round(output.Objective)) > 0 {
		out.TotalCost = int(math.Round(output.Objective))
	}
	return json.Marshal(out)
}

func buildDistanceMatrix(ds *cvrp.Dataset, numNodes int) [][]int {
	nodeX := make([]float64, numNodes)
	nodeY := make([]float64, numNodes)
	nodeX[0] = ds.Depot.X
	nodeY[0] = ds.Depot.Y
	for i, c := range ds.Customers {
		nodeX[i+1] = c.X
		nodeY[i+1] = c.Y
	}
	dist := make([][]int, numNodes)
	for i := 0; i < numNodes; i++ {
		dist[i] = make([]int, numNodes)
		for j := 0; j < numNodes; j++ {
			if i != j {
				dist[i][j] = cvrp.DistanceRounded(nodeX[i], nodeY[i], nodeX[j], nodeY[j])
			}
		}
	}
	return dist
}
