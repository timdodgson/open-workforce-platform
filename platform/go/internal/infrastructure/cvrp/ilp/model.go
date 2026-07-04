// Package ilp provides an ILP formulation for the Capacitated Vehicle Routing Problem.
//
// Formulation: Two-index vehicle flow model with Miller-Tucker-Zemlin (MTZ)
// subtour elimination constraints.
//
// Decision variables:
//   x_ij ∈ {0,1} — vehicle travels from node i to node j
//   u_i ∈ [demand_i, capacity] — cumulative load when arriving at customer i
//
// Objective: minimise Σ c_ij * x_ij
//
// Constraints:
//   (1) Each customer visited exactly once: Σ_j x_ij = 1 for all customers i
//   (2) Each customer departed exactly once: Σ_i x_ij = 1 for all customers j
//   (3) Depot flow balance: Σ_j x_0j = Σ_j x_j0 = K (number of vehicles)
//   (4) MTZ subtour elimination: u_j >= u_i + demand_j - capacity*(1 - x_ij)
//   (5) Load bounds: demand_i <= u_i <= capacity
//
// The number of vehicles K is determined by ceil(total_demand / capacity).
package ilp

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/cvrp"
)

// ModelInfo describes the generated LP model.
type ModelInfo struct {
	Variables   int
	Constraints int
	Nodes       int // depot + customers
	Vehicles    int
}

// BuildModel writes a CVRP LP model in CPLEX LP format.
// The depot is node 0, customers are nodes 1..n.
func BuildModel(ds *cvrp.Dataset, modelPath string, maxVehicles int) (ModelInfo, error) {
	n := len(ds.Customers) // number of customers
	if n == 0 {
		return ModelInfo{}, fmt.Errorf("no customers in dataset")
	}

	// Compute distance matrix (depot=0, customers=1..n).
	numNodes := n + 1
	dist := make([][]int, numNodes)
	nodeX := make([]float64, numNodes)
	nodeY := make([]float64, numNodes)
	nodeX[0] = ds.Depot.X
	nodeY[0] = ds.Depot.Y
	for i, c := range ds.Customers {
		nodeX[i+1] = c.X
		nodeY[i+1] = c.Y
	}
	for i := 0; i < numNodes; i++ {
		dist[i] = make([]int, numNodes)
		for j := 0; j < numNodes; j++ {
			if i != j {
				dist[i][j] = cvrp.DistanceRounded(nodeX[i], nodeY[i], nodeX[j], nodeY[j])
			}
		}
	}

	// Determine number of vehicles.
	totalDemand := 0
	for _, c := range ds.Customers {
		totalDemand += c.Demand
	}
	minVehicles := int(math.Ceil(float64(totalDemand) / float64(ds.Capacity)))
	K := minVehicles
	if maxVehicles > 0 && maxVehicles > K {
		K = maxVehicles // allow more vehicles for feasibility
	}

	// Build LP file.
	var b strings.Builder

	// Objective: minimise total distance.
	b.WriteString("Minimize\n obj: ")
	first := true
	for i := 0; i < numNodes; i++ {
		for j := 0; j < numNodes; j++ {
			if i == j {
				continue
			}
			if dist[i][j] == 0 {
				continue
			}
			if !first {
				b.WriteString(" + ")
			}
			b.WriteString(fmt.Sprintf("%d x_%d_%d", dist[i][j], i, j))
			first = false
		}
	}
	b.WriteString("\n\n")

	// Constraints.
	b.WriteString("Subject To\n")
	conCount := 0

	// (1) Each customer visited exactly once (in-degree = 1).
	for j := 1; j <= n; j++ {
		b.WriteString(fmt.Sprintf(" visit_in_%d: ", j))
		first := true
		for i := 0; i < numNodes; i++ {
			if i == j {
				continue
			}
			if !first {
				b.WriteString(" + ")
			}
			b.WriteString(fmt.Sprintf("x_%d_%d", i, j))
			first = false
		}
		b.WriteString(" = 1\n")
		conCount++
	}

	// (2) Each customer departed exactly once (out-degree = 1).
	for i := 1; i <= n; i++ {
		b.WriteString(fmt.Sprintf(" visit_out_%d: ", i))
		first := true
		for j := 0; j < numNodes; j++ {
			if i == j {
				continue
			}
			if !first {
				b.WriteString(" + ")
			}
			b.WriteString(fmt.Sprintf("x_%d_%d", i, j))
			first = false
		}
		b.WriteString(" = 1\n")
		conCount++
	}

	// (3) Depot out-degree = K.
	b.WriteString(fmt.Sprintf(" depot_out: "))
	first = true
	for j := 1; j <= n; j++ {
		if !first {
			b.WriteString(" + ")
		}
		b.WriteString(fmt.Sprintf("x_0_%d", j))
		first = false
	}
	b.WriteString(fmt.Sprintf(" = %d\n", K))
	conCount++

	// Depot in-degree = K.
	b.WriteString(fmt.Sprintf(" depot_in: "))
	first = true
	for i := 1; i <= n; i++ {
		if !first {
			b.WriteString(" + ")
		}
		b.WriteString(fmt.Sprintf("x_%d_0", i))
		first = false
	}
	b.WriteString(fmt.Sprintf(" = %d\n", K))
	conCount++

	// (4) MTZ subtour elimination: u_j - u_i >= demand_j - capacity*(1 - x_ij)
	// Rearranged: u_j - u_i + capacity*x_ij >= demand_j
	// Only for customer pairs (i >= 1, j >= 1, i != j).
	for i := 1; i <= n; i++ {
		for j := 1; j <= n; j++ {
			if i == j {
				continue
			}
			demandJ := ds.Customers[j-1].Demand
			b.WriteString(fmt.Sprintf(" mtz_%d_%d: u_%d - u_%d + %d x_%d_%d >= %d\n",
				i, j, j, i, ds.Capacity, i, j, demandJ))
			conCount++
		}
	}

	// (5) Load bounds: demand_i <= u_i <= capacity.
	b.WriteString("\nBounds\n")
	for i := 1; i <= n; i++ {
		demandI := ds.Customers[i-1].Demand
		b.WriteString(fmt.Sprintf(" %d <= u_%d <= %d\n", demandI, i, ds.Capacity))
	}

	// Binary variables.
	b.WriteString("\nBinary\n")
	for i := 0; i < numNodes; i++ {
		for j := 0; j < numNodes; j++ {
			if i == j {
				continue
			}
			b.WriteString(fmt.Sprintf(" x_%d_%d\n", i, j))
		}
	}

	// General (integer/continuous) for u variables — they're continuous in MTZ.
	b.WriteString("\nGenerals\n")
	for i := 1; i <= n; i++ {
		b.WriteString(fmt.Sprintf(" u_%d\n", i))
	}

	b.WriteString("\nEnd\n")

	// Write to file.
	if err := os.WriteFile(modelPath, []byte(b.String()), 0644); err != nil {
		return ModelInfo{}, fmt.Errorf("write model: %w", err)
	}

	numVars := numNodes*(numNodes-1) + n // x variables + u variables
	return ModelInfo{
		Variables:   numVars,
		Constraints: conCount,
		Nodes:       numNodes,
		Vehicles:    K,
	}, nil
}
