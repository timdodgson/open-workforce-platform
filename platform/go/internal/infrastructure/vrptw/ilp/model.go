// Package ilp provides an ILP formulation for the Vehicle Routing Problem with Time Windows.
//
// Extends the CVRP two-index flow + MTZ model with continuous service-start times t_i
// and time-window constraints.
package ilp

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/vrptw"
)

// ModelInfo describes the generated LP model.
type ModelInfo struct {
	Variables   int
	Constraints int
	Nodes       int
	Vehicles    int
}

// BuildModel writes a VRPTW LP model in CPLEX LP format.
func BuildModel(ds *vrptw.Dataset, modelPath string, maxVehicles int) (ModelInfo, error) {
	n := len(ds.Customers)
	if n == 0 {
		return ModelInfo{}, fmt.Errorf("no customers in dataset")
	}

	numNodes := n + 1
	dist := make([][]int, numNodes)
	nodeX := make([]float64, numNodes)
	nodeY := make([]float64, numNodes)
	service := make([]int, numNodes)
	ready := make([]int, numNodes)
	due := make([]int, numNodes)

	nodeX[0] = ds.Depot.X
	nodeY[0] = ds.Depot.Y
	service[0] = ds.Depot.Service
	ready[0] = ds.Depot.ReadyTime
	due[0] = ds.Depot.DueDate

	for i, c := range ds.Customers {
		nodeX[i+1] = c.X
		nodeY[i+1] = c.Y
		service[i+1] = c.Service
		ready[i+1] = c.ReadyTime
		due[i+1] = c.DueDate
	}

	for i := 0; i < numNodes; i++ {
		dist[i] = make([]int, numNodes)
		for j := 0; j < numNodes; j++ {
			if i != j {
				dist[i][j] = vrptw.DistanceRounded(nodeX[i], nodeY[i], nodeX[j], nodeY[j])
			}
		}
	}

	totalDemand := 0
	for _, c := range ds.Customers {
		totalDemand += c.Demand
	}
	minVehicles := int(math.Ceil(float64(totalDemand) / float64(ds.Capacity)))
	K := n
	if ds.Vehicles > 0 {
		K = ds.Vehicles
	}
	if maxVehicles > 0 {
		K = maxVehicles
	}
	_ = minVehicles

	bigM := due[0] + 1
	for i := 1; i < numNodes; i++ {
		if due[i]+service[i] > bigM {
			bigM = due[i] + service[i]
		}
	}
	bigM += dist[0][1] * n // slack for travel

	var b strings.Builder

	// Objective: minimise total distance.
	b.WriteString("Minimize\n obj: ")
	first := true
	for i := 0; i < numNodes; i++ {
		for j := 0; j < numNodes; j++ {
			if i == j || dist[i][j] == 0 {
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

	b.WriteString("Subject To\n")
	conCount := 0

	// Visit in/out (same as CVRP).
	for j := 1; j <= n; j++ {
		b.WriteString(fmt.Sprintf(" visit_in_%d: ", j))
		first = true
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

	for i := 1; i <= n; i++ {
		b.WriteString(fmt.Sprintf(" visit_out_%d: ", i))
		first = true
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

	b.WriteString(" depot_out: ")
	first = true
	for j := 1; j <= n; j++ {
		if !first {
			b.WriteString(" + ")
		}
		b.WriteString(fmt.Sprintf("x_0_%d", j))
		first = false
	}
	b.WriteString(fmt.Sprintf(" <= %d\n", K))
	conCount++

	b.WriteString(" depot_in: ")
	first = true
	for i := 1; i <= n; i++ {
		if !first {
			b.WriteString(" + ")
		}
		b.WriteString(fmt.Sprintf("x_%d_0", i))
		first = false
	}
	b.WriteString(fmt.Sprintf(" <= %d\n", K))
	conCount++

	b.WriteString(" depot_balance: ")
	first = true
	for j := 1; j <= n; j++ {
		if !first {
			b.WriteString(" + ")
		}
		b.WriteString(fmt.Sprintf("x_0_%d", j))
		first = false
	}
	for i := 1; i <= n; i++ {
		b.WriteString(fmt.Sprintf(" - x_%d_0", i))
	}
	b.WriteString(" = 0\n")
	conCount++

	// MTZ capacity.
	for i := 1; i <= n; i++ {
		for j := 1; j <= n; j++ {
			if i == j {
				continue
			}
			demandJ := ds.Customers[j-1].Demand
			rhs := demandJ - ds.Capacity
			b.WriteString(fmt.Sprintf(" mtz_%d_%d: u_%d - u_%d - %d x_%d_%d >= %d\n",
				i, j, j, i, ds.Capacity, i, j, rhs))
			conCount++
		}
	}

	// Time propagation: t_j >= t_i + service_i + dist_ij when x_ij = 1.
	for i := 0; i < numNodes; i++ {
		for j := 0; j < numNodes; j++ {
			if i == j {
				continue
			}
			rhs := service[i] + dist[i][j] - bigM
			b.WriteString(fmt.Sprintf(" time_%d_%d: t_%d - t_%d - %d x_%d_%d >= %d\n",
				i, j, j, i, bigM, i, j, rhs))
			conCount++
		}
	}

	// Return to depot within horizon when arc i->0 is used.
	for i := 1; i <= n; i++ {
		rhs := due[0] - service[i] - dist[i][0] + bigM
		b.WriteString(fmt.Sprintf(" return_%d: t_%d - %d x_%d_0 <= %d\n",
			i, i, bigM, i, rhs))
		conCount++
	}

	b.WriteString("\nBounds\n")
	for i := 0; i < numNodes; i++ {
		b.WriteString(fmt.Sprintf(" %d <= t_%d <= %d\n", ready[i], i, due[i]))
	}
	for i := 1; i <= n; i++ {
		demandI := ds.Customers[i-1].Demand
		b.WriteString(fmt.Sprintf(" %d <= u_%d <= %d\n", demandI, i, ds.Capacity))
	}

	b.WriteString("\nBinary\n")
	for i := 0; i < numNodes; i++ {
		for j := 0; j < numNodes; j++ {
			if i == j {
				continue
			}
			b.WriteString(fmt.Sprintf(" x_%d_%d\n", i, j))
		}
	}

	b.WriteString("\nGenerals\n")
	for i := 0; i < numNodes; i++ {
		b.WriteString(fmt.Sprintf(" t_%d\n", i))
	}
	for i := 1; i <= n; i++ {
		b.WriteString(fmt.Sprintf(" u_%d\n", i))
	}

	b.WriteString("\nEnd\n")

	if err := os.WriteFile(modelPath, []byte(b.String()), 0644); err != nil {
		return ModelInfo{}, fmt.Errorf("write model: %w", err)
	}

	numVars := numNodes*(numNodes-1) + n + numNodes // x + u + t
	return ModelInfo{
		Variables:   numVars,
		Constraints: conCount,
		Nodes:       numNodes,
		Vehicles:    minVehicles,
	}, nil
}
