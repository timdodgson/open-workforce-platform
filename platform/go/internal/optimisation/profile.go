package optimisation

// AlgorithmProfile provides per-algorithm configuration values.
//
// Each algorithm reads only its own section. Profiles provide defaults;
// explicit CLI flags override individual values.
type AlgorithmProfile struct {
	// Hill Climbing.
	HCMaxIterations int

	// Simulated Annealing.
	SAMaxIterations      int
	SAInitialTemperature float64
	SACoolingRate        float64
	SAMinTemperature     float64

	// Tabu Search.
	TabuMaxIterations     int
	TabuListSize          int
	TabuAspirationEnabled bool

	// Large Neighbourhood Search.
	LNSIterations     int
	LNSDestroySize    int
	LNSRepairStrategy string
}

// DefaultProfile returns sensible defaults for quick runs.
func DefaultProfile() AlgorithmProfile {
	return AlgorithmProfile{
		HCMaxIterations:      100,
		SAMaxIterations:      200,
		SAInitialTemperature: 500.0,
		SACoolingRate:        0.97,
		SAMinTemperature:     1.0,
		TabuMaxIterations:    100,
		TabuListSize:         10,
		TabuAspirationEnabled: true,
		LNSIterations:        50,
		LNSDestroySize:       2,
		LNSRepairStrategy:    "greedy",
	}
}

// FastProfile returns a configuration optimised for speed.
func FastProfile() AlgorithmProfile {
	return AlgorithmProfile{
		HCMaxIterations:      25,
		SAMaxIterations:      50,
		SAInitialTemperature: 300.0,
		SACoolingRate:        0.95,
		SAMinTemperature:     5.0,
		TabuMaxIterations:    25,
		TabuListSize:         5,
		TabuAspirationEnabled: true,
		LNSIterations:        10,
		LNSDestroySize:       2,
		LNSRepairStrategy:    "greedy",
	}
}

// QualityProfile returns a configuration that searches longer for better solutions.
func QualityProfile() AlgorithmProfile {
	return AlgorithmProfile{
		HCMaxIterations:      200,
		SAMaxIterations:      1000,
		SAInitialTemperature: 1000.0,
		SACoolingRate:        0.995,
		SAMinTemperature:     0.1,
		TabuMaxIterations:    200,
		TabuListSize:         20,
		TabuAspirationEnabled: true,
		LNSIterations:        100,
		LNSDestroySize:       3,
		LNSRepairStrategy:    "greedy",
	}
}

// ResearchProfile returns a configuration with large search limits for experimentation.
func ResearchProfile() AlgorithmProfile {
	return AlgorithmProfile{
		HCMaxIterations:      500,
		SAMaxIterations:      5000,
		SAInitialTemperature: 2000.0,
		SACoolingRate:        0.998,
		SAMinTemperature:     0.01,
		TabuMaxIterations:    500,
		TabuListSize:         50,
		TabuAspirationEnabled: true,
		LNSIterations:        250,
		LNSDestroySize:       4,
		LNSRepairStrategy:    "greedy",
	}
}

// GetProfile returns a named algorithm profile.
func GetProfile(name string) (AlgorithmProfile, bool) {
	switch name {
	case "default", "":
		return DefaultProfile(), true
	case "fast":
		return FastProfile(), true
	case "quality":
		return QualityProfile(), true
	case "research":
		return ResearchProfile(), true
	default:
		return AlgorithmProfile{}, false
	}
}
