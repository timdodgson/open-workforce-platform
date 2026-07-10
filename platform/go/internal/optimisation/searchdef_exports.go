package optimisation

import (
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation/assist"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation/searchdef"
)

// Core search engine types (searchdef).
type (
	Solution         = searchdef.Solution
	Move             = searchdef.Move
	MoveResult       = searchdef.MoveResult
	Problem          = searchdef.Problem
	DatasetLoader    = searchdef.DatasetLoader
	Confidence       = searchdef.Confidence
	SafetyStatus     = searchdef.SafetyStatus
	SearchAction     = searchdef.SearchAction
	SearchProgress   = searchdef.SearchProgress
	SearchRecommendation = searchdef.SearchRecommendation
	SearchAssist         = searchdef.SearchAssist
	SearchAssistConfig   = searchdef.SearchAssistConfig
	SearchAssistRecord   = searchdef.SearchAssistRecord
)

const (
	SafetyPassed   = searchdef.SafetyPassed
	SafetyRejected = searchdef.SafetyRejected

	SearchContinue     = searchdef.SearchContinue
	SearchEarlyStop    = searchdef.SearchEarlyStop
	SearchRestart      = searchdef.SearchRestart
	SearchAdjustTemp   = searchdef.SearchAdjustTemp
	SearchAdjustLAHC   = searchdef.SearchAdjustLAHC
	SearchAdjustTabu   = searchdef.SearchAdjustTabu
	SearchAdjustBudget = searchdef.SearchAdjustBudget
)

var DefaultSearchAssistConfig = searchdef.DefaultSearchAssistConfig

// SI v1 search assist (assist).
type (
	SearchHookRunner    = assist.SearchHookRunner
	SearchAssistRecorder = assist.SearchAssistRecorder
	RuleBasedSearchAssist = assist.RuleBasedSearchAssist
	AdaptiveSearchAssist  = assist.AdaptiveSearchAssist
)

var (
	NewSearchHookRunner      = assist.NewSearchHookRunner
	NewSearchAssistRecorder  = assist.NewSearchAssistRecorder
	NewRuleBasedSearchAssist = assist.NewRuleBasedSearchAssist
	NewAdaptiveSearchAssist  = assist.NewAdaptiveSearchAssist
	WriteSearchAssistCSV     = assist.WriteSearchAssistCSV
	EvaluateSearchSafety     = assist.EvaluateSearchSafety
)

// PortfolioAssist (assist/portfolio.go).
type (
	PortfolioAssistAction   = assist.PortfolioAssistAction
	PortfolioAssistRecord   = assist.PortfolioAssistRecord
	PortfolioAssistRecorder = assist.PortfolioAssistRecorder
	PortfolioAssistConfig   = assist.PortfolioAssistConfig
	RuleBasedPortfolioAdvisor = assist.RuleBasedPortfolioAdvisor
	StrategyAdvice          = assist.StrategyAdvice
	AdviceSource            = assist.AdviceSource
	LearnedAdviceResult     = assist.LearnedAdviceResult
)

const (
	PortfolioActionRun          = assist.PortfolioActionRun
	PortfolioActionSkip         = assist.PortfolioActionSkip
	PortfolioActionReduceBudget = assist.PortfolioActionReduceBudget
	PortfolioActionBoostBudget  = assist.PortfolioActionBoostBudget
)

var (
	NewPortfolioAssistRecorder   = assist.NewPortfolioAssistRecorder
	NewRuleBasedPortfolioAdvisor = assist.NewRuleBasedPortfolioAdvisor
	WritePortfolioAssistCSV      = assist.WritePortfolioAssistCSV
	EvaluatePortfolioSafety      = assist.EvaluatePortfolioSafety
)
