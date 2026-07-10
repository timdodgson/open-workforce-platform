package optimisation

// mockModel is a test double for PolicyModel used by shadow and integration tests.
type mockModel struct {
	action     string
	confidence float64
}

func (m *mockModel) Predict(_ FeatureVector) ModelPrediction {
	return ModelPrediction{
		Action:     m.action,
		Confidence: m.confidence,
		Reason:     "mock_prediction",
	}
}
