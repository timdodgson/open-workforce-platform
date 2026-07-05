// Package vrptw implements the Vehicle Routing Problem with Time Windows (VRPTW) domain.
//
// The VRPTW extends CVRP by adding time constraints: each customer has a
// time window [ReadyTime, DueDate] within which service must begin. A vehicle
// may arrive early (and wait) but may not arrive late. Each customer also has
// a service duration. The depot has its own time window constraining when
// vehicles must depart and return.
//
// Objective: minimise total travel distance (same as CVRP).
// Hard constraints: capacity + time windows (no late arrivals).
//
// This package implements the generic optimisation.Problem interface.
package vrptw

import "math"

// Depot represents the central depot with a scheduling horizon.
type Depot struct {
	ID        int
	X         float64
	Y         float64
	ReadyTime int // earliest departure time
	DueDate   int // latest return time (planning horizon)
	Service   int // service time at depot (usually 0)
}

// Customer represents a delivery location with demand and time window.
type Customer struct {
	ID        int
	X         float64
	Y         float64
	Demand    int
	ReadyTime int // earliest time service can begin
	DueDate   int // latest time service can begin
	Service   int // service duration at this customer
}

// Dataset holds all VRPTW instance data.
type Dataset struct {
	Name       string
	Dimension  int // total nodes (depot + customers)
	Capacity   int // vehicle capacity
	Vehicles   int // max number of vehicles (0 = unlimited)
	Depot      Depot
	Customers  []Customer
}

// DistanceRounded computes the rounded Euclidean distance (TSPLIB convention).
func DistanceRounded(x1, y1, x2, y2 float64) int {
	dx := x1 - x2
	dy := y1 - y2
	return int(math.Round(math.Sqrt(dx*dx + dy*dy)))
}

// DistanceExact computes the exact Euclidean distance (used for time calculations).
func DistanceExact(x1, y1, x2, y2 float64) float64 {
	dx := x1 - x2
	dy := y1 - y2
	return math.Sqrt(dx*dx + dy*dy)
}
