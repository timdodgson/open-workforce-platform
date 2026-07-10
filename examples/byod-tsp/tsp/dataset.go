package tsp

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
)

type City struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Dataset struct {
	Name   string `json:"name"`
	Cities []City `json:"cities"`
}

func LoadDataset(path string) (*Dataset, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var ds Dataset
	if err := json.Unmarshal(data, &ds); err != nil {
		return nil, err
	}
	if ds.Name == "" {
		ds.Name = path
	}
	if len(ds.Cities) < 2 {
		return nil, fmt.Errorf("tsp: need at least 2 cities in %s", path)
	}
	return &ds, nil
}

func (ds *Dataset) Distance(i, j int) int {
	dx := float64(ds.Cities[i].X - ds.Cities[j].X)
	dy := float64(ds.Cities[i].Y - ds.Cities[j].Y)
	return int(math.Round(math.Hypot(dx, dy)))
}
