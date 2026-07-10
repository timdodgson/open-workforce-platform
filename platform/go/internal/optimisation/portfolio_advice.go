package optimisation

// portfolio_advice.go resolves portfolio budget advice for RunPortfolioWithAssist.

type portfolioAdviceBundle struct {
	advice  []StrategyAdvice
	sources []AdviceSource
}

func resolvePortfolioAdvice(
	config SearchConfig,
	assistConfig PortfolioAssistConfig,
	strategies []string,
) portfolioAdviceBundle {
	if config.PolicyMode != "" {
		modelPath := ""
		if config.PolicyMode != "rules" {
			modelPath = ResolveBudgetPolicyPath(config.PolicyDir, assistConfig.ModelPath, config.PolicyMode)
		}
		policy := NewPortfolioBudgetPolicy(PortfolioBudgetPolicyConfig{
			Domain:    assistConfig.Domain,
			ModelPath: modelPath,
		})
		allocations := AllocateBudgetsViaPolicy(
			policy, strategies, config.Iterations,
			assistConfig.Domain, assistConfig.Instance, "", NewFeatureExtractor(),
		)
		advice := make([]StrategyAdvice, len(allocations))
		sources := make([]AdviceSource, len(strategies))
		for i, alloc := range allocations {
			advice[i] = StrategyAdvice{
				Strategy:   alloc.Strategy,
				Action:     PortfolioActionRun,
				BudgetMult: alloc.BudgetMult,
				Confidence: alloc.Decision.Confidence,
				Reasons:    []string{alloc.Decision.Reason},
			}
			sources[i] = AdviceSource{
				Strategy:       alloc.Strategy,
				FallbackReason: alloc.Decision.FallbackReason,
				RuleBased:      advice[i],
			}
		}
		return portfolioAdviceBundle{advice: advice, sources: sources}
	}

	var learnedAdvisor *LearnedPortfolioAdvisor
	if assistConfig.ModelPath != "" {
		model, err := LoadPortfolioBudgetModel(assistConfig.ModelPath)
		if err == nil {
			learnedAdvisor = NewLearnedPortfolioAdvisor(model)
		}
	}

	if learnedAdvisor != nil {
		result := learnedAdvisor.Advise(strategies, config.Iterations, assistConfig.Domain, assistConfig.Instance)
		return portfolioAdviceBundle{advice: result.Advice, sources: result.Source}
	}

	ruleAdvisor := NewRuleBasedPortfolioAdvisor()
	advice := ruleAdvisor.Advise(strategies, config.Iterations, assistConfig.Domain)
	sources := make([]AdviceSource, len(strategies))
	for i, strat := range strategies {
		sources[i] = AdviceSource{
			Strategy:       strat,
			FallbackReason: "no_model_path",
			RuleBased:      advice[i],
		}
	}
	return portfolioAdviceBundle{advice: advice, sources: sources}
}

func resolvePortfolioAssistMode(assistConfig PortfolioAssistConfig, config SearchConfig) string {
	mode := assistConfig.Mode
	if mode == "" && config.AssistMode != "" {
		mode = config.AssistMode
	}
	if mode == "" && config.PolicyMode != "" {
		mode = "shadow"
	}
	return mode
}
