package assist

// Shared v1 portfolio budget heuristics used by Search Intelligence v1
// (RuleBasedPortfolioAdvisor) and v2 (NewPortfolioBudgetRulePolicy).

const (
	PortfolioDefaultBudgetMult = 1.0
	PortfolioDefaultConfidence = 0.5
	PortfolioSABudgetMult      = 1.1
	PortfolioSAConfidence      = 0.55
	PortfolioLAHCBudgetMult    = 0.9
	PortfolioLAHCConfidence    = 0.5
	PortfolioTabuBudgetMult    = 1.15
	PortfolioTabuConfidence    = 0.55
	PortfolioMinStrategiesLAHC = 3
)

// PortfolioBudgetHeuristic returns the v1 rule-based budget advice for one strategy.
func PortfolioBudgetHeuristic(strategy, domain string, strategyCount int) (budgetMult, confidence float64, reason string) {
	budgetMult = PortfolioDefaultBudgetMult
	confidence = PortfolioDefaultConfidence
	reason = "default_run"

	if strategy == "sa" {
		return PortfolioSABudgetMult, PortfolioSAConfidence, "sa_generally_strong"
	}
	if strategy == "lahc" && strategyCount >= PortfolioMinStrategiesLAHC {
		return PortfolioLAHCBudgetMult, PortfolioLAHCConfidence, "lahc_slower_convergence_in_portfolio"
	}
	if strategy == "tabu" && (domain == "jss" || domain == "nrp") {
		return PortfolioTabuBudgetMult, PortfolioTabuConfidence, "tabu_strong_on_constrained"
	}
	return budgetMult, confidence, reason
}
