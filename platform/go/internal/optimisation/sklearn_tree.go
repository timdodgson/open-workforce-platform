package optimisation

// SklearnTree is a sklearn DecisionTreeClassifier exported to JSON.
type SklearnTree struct {
	FeatureNames  []string    `json:"feature_names"`
	ChildrenLeft  []int       `json:"children_left"`
	ChildrenRight []int       `json:"children_right"`
	Feature       []int       `json:"feature"`
	Threshold     []float64   `json:"threshold"`
	Value         [][]float64 `json:"value"`
}

// PositiveClassProbability returns P(class=1) for a feature vector aligned to FeatureNames.
func (t *SklearnTree) PositiveClassProbability(features []float64) float64 {
	if t == nil || len(t.ChildrenLeft) == 0 {
		return 0.5
	}
	node := 0
	for {
		left := t.ChildrenLeft[node]
		right := t.ChildrenRight[node]
		if left == right {
			return leafPositiveProb(t.Value[node])
		}
		featIdx := t.Feature[node]
		if featIdx < 0 || featIdx >= len(features) {
			return leafPositiveProb(t.Value[node])
		}
		if features[featIdx] <= t.Threshold[node] {
			node = left
		} else {
			node = right
		}
	}
}

func leafPositiveProb(values []float64) float64 {
	if len(values) == 0 {
		return 0.5
	}
	if len(values) == 1 {
		return values[0]
	}
	total := 0.0
	for _, v := range values {
		total += v
	}
	if total <= 0 {
		return 0.5
	}
	return values[1] / total
}
