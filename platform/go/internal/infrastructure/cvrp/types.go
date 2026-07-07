// Package cvrp implements the Capacitated Vehicle Routing Problem (CVRP) domain.
//
// The CVRP asks: given a depot, a fleet of vehicles with uniform capacity,
// and a set of customers with known demand, find the minimum-cost set of
// routes that starts and ends at the depot, serves every customer exactly
// once, and never exceeds vehicle capacity.
//
// This package models the problem domain and implements the generic
// optimisation.Problem interface. It does not contain optimisation algorithms.
package cvrp

import "math"

// Depot represents the central depot where all routes start and end.
type Depot struct {
	ID int
	X  float64
	Y  float64
}

// Customer represents a delivery location with a demand quantity.
type Customer struct {
	ID     int
	X      float64
	Y      float64
	Demand int
}

// Vehicle represents a vehicle in the fleet.
// All vehicles are assumed identical with the same capacity.
type Vehicle struct {
	Capacity int
}

// Route represents a single vehicle route: depot → customers → depot.
// Customers are visited in the order they appear in the slice.
type Route struct {
	Customers []int   // customer IDs in visit order
	Load      int     // total demand served on this route
	Distance  float64 // total Euclidean distance including depot legs
}

// Solution represents a complete CVRP solution: a set of routes.
type Solution struct {
	Routes   []Route
	Cost     float64 // total distance across all routes
	Feasible bool    // true if no route exceeds vehicle capacity
}

// Dataset holds all problem instance data needed to solve a CVRP.
type Dataset struct {
	Name      string
	Dimension int // number of nodes (depot + customers)
	Capacity  int // vehicle capacity
	Depot     Depot
	Customers []Customer
}

// Distance computes the Euclidean distance between two points.
func Distance(x1, y1, x2, y2 float64) float64 {
	dx := x1 - x2
	dy := y1 - y2
	return math.Sqrt(dx*dx + dy*dy)
}

// DistanceRounded computes the rounded Euclidean distance (TSPLIB convention).
func DistanceRounded(x1, y1, x2, y2 float64) int {
	dx := x1 - x2
	dy := y1 - y2
	return int(math.Round(math.Sqrt(dx*dx + dy*dy)))
}
