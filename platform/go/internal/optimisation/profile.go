package optimisation

// AlgorithmProfile provides configuration values for optimisation algorithms.
//
// Algorithms read their parameters from this profile rather than using
// hard-coded constants. This allows different search strategies without
// modifying algorithm source code.
type AlgorithmProfile struct {
	// Hill climbing / simulated annealing iteration limit.
	MaxIterations int
	// Tabu list size.
	TabuListSize int
	// LNS iterations.
	LNSIterations int
	// LNS destroy size (assignments removed per iteration).
	DestroySize int
}

// DefaultProfile returns the current default configuration.
func DefaultProfile() AlgorithmProfile {
	return AlgorithmProfile{
		MaxIterations: 100,
		TabuListSize:  10,
		LNSIterations: 50,
		DestroySize:   2,
	}
}

// FastProfile returns a configuration optimised for speed.
func FastProfile() AlgorithmProfile {
	return AlgorithmProfile{
		MaxIterations: 25,
		TabuListSize:  5,
		LNSIterations: 10,
		DestroySize:   2,
	}
}

// QualityProfile returns a configuration that searches longer for better solutions.
func QualityProfile() AlgorithmProfile {
	return AlgorithmProfile{
		MaxIterations: 200,
		TabuListSize:  20,
		LNSIterations: 100,
		DestroySize:   3,
	}
}

// ResearchProfile returns a configuration with large search limits for experimentation.
func ResearchProfile() AlgorithmProfile {
	return AlgorithmProfile{
		MaxIterations: 500,
		TabuListSize:  50,
		LNSIterations: 250,
		DestroySize:   4,
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
