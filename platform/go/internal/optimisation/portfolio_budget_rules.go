package optimisation

// Shared v1 portfolio budget heuristics used by both Search Intelligence v1
// (RuleBasedPortfolioAdvisor) and v2 (NewPortfolioBudgetRulePolicy).

const (
	portfolioDefaultBudgetMult = 1.0
	portfolioDefaultConfidence = 0.5
	portfolioSABudgetMult      = 1.1
	portfolioSAConfidence      = 0.55
	portfolioLAHCBudgetMult    = 0.9
	portfolioLAHCConfidence    = 0.5
	portfolioTabuBudgetMult    = 1.15
	portfolioTabuConfidence    = 0.55
	portfolioMinStrategiesLAHC = 3
)

// portfolioBudgetHeuristic returns the v1 rule-based budget advice for one strategy.
func portfolioBudgetHeuristic(strategy, domain string, strategyCount int) (budgetMult, confidence float64, reason string) {
	budgetMult = portfolioDefaultBudgetMult
	confidence = portfolioDefaultConfidence
	reason = "default_run"

	if strategy == "sa" {
		return portfolioSABudgetMult, portfolioSAConfidence, "sa_generally_strong"
	}
	if strategy == "lahc" && strategyCount >= portfolioMinStrategiesLAHC {
		return portfolioLAHCBudgetMult, portfolioLAHCConfidence, "lahc_slower_convergence_in_portfolio"
	}
	if strategy == "tabu" && (domain == "jss" || domain == "nrp") {
		return portfolioTabuBudgetMult, portfolioTabuConfidence, "tabu_strong_on_constrained"
	}
	return budgetMult, confidence, reason
}
