// bandit_policy.go — Step 5 contextual bandit lookup for portfolio/worker budgets.
package optimisation

import "math"

// BanditArmStat holds offline mean reward for one arm in a context.
type BanditArmStat struct {
	Arm        string  `json:"arm"`
	Value      float64 `json:"value"`
	MeanReward float64 `json:"mean_reward"`
	Samples    int     `json:"samples"`
}

// BanditEntry is a per-context bandit policy exported from Python training.
type BanditEntry struct {
	Domain        string          `json:"domain"`
	Instance      string          `json:"instance,omitempty"`
	Strategy      string          `json:"strategy,omitempty"`
	Context       string          `json:"context,omitempty"`
	Arms          []BanditArmStat `json:"arms"`
	BestArm       string          `json:"best_arm"`
	BestValue     float64         `json:"best_value"`
	Confidence    float64         `json:"confidence"`
	EpisodeRegret float64         `json:"episode_regret"`
	Samples       int             `json:"samples"`
}

// BanditPolicy is the exported offline contextual bandit block.
type BanditPolicy struct {
	Status         string        `json:"status"`
	Version        string        `json:"version"`
	Entries        []BanditEntry `json:"entries"`
	EpisodeRegret  float64       `json:"episode_regret"`
	PromotionReady bool          `json:"promotion_ready"`
	Samples        int           `json:"samples"`
}

const minBanditConfidence = 0.55

// FindPortfolioBanditEntry prefers instance-specific context.
func FindPortfolioBanditEntry(bandit *BanditPolicy, domain, instance, strategy string) *BanditEntry {
	if bandit == nil || bandit.Status != "trained" {
		return nil
	}
	var instanceMatch, domainMatch *BanditEntry
	for i := range bandit.Entries {
		e := &bandit.Entries[i]
		if e.Domain != domain || e.Strategy != strategy {
			continue
		}
		if e.Instance != "" && e.Instance == instance {
			return e
		}
		if e.Instance == "" {
			domainMatch = e
		}
		if e.Instance != "" {
			instanceMatch = e
		}
	}
	if domainMatch != nil {
		return domainMatch
	}
	return instanceMatch
}

// PortfolioBanditMultiplier returns learned arm value when confident.
func PortfolioBanditMultiplier(bandit *BanditPolicy, domain, instance, strategy string) (float64, bool, string) {
	entry := FindPortfolioBanditEntry(bandit, domain, instance, strategy)
	if entry == nil || entry.Samples < 3 || entry.Confidence < minBanditConfidence {
		return 1.0, false, "bandit:no_entry"
	}
	mult := entry.BestValue
	if mult > 2.0 {
		mult = 2.0
	}
	if mult < 0.25 {
		mult = 0.25
	}
	return mult, true, "bandit:" + entry.BestArm
}

// FindWorkerBanditEntry matches NRP worker context key.
func FindWorkerBanditEntry(bandit *BanditPolicy, context string) *BanditEntry {
	if bandit == nil || bandit.Status != "trained" {
		return nil
	}
	var best *BanditEntry
	for i := range bandit.Entries {
		e := &bandit.Entries[i]
		if e.Context == context {
			return e
		}
		if best == nil || e.Samples > best.Samples {
			best = e
		}
	}
	return best
}

// WorkerBanditDecision maps best arm to spawn recommendation.
func WorkerBanditDecision(bandit *BanditPolicy, context string) (string, float64, bool) {
	entry := FindWorkerBanditEntry(bandit, context)
	if entry == nil || entry.Samples < 3 || entry.Confidence < minBanditConfidence {
		return "", 0, false
	}
	switch entry.BestArm {
	case "skip":
		return "skip", 0, true
	case "boost":
		return "boost", math.Max(1.0, entry.BestValue), true
	default:
		return "default", 1.0, true
	}
}

// WorkerContextKey builds the training context slug from worker features.
func WorkerContextKey(week, depth int, distanceFromBest float64) string {
	dist := "near"
	if distanceFromBest > 50 {
		dist = "far"
	}
	return "week=" + itoa(week) + "|depth=" + itoa(depth) + "|dist=" + dist
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [12]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
