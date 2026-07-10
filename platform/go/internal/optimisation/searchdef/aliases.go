// Package searchdef re-exports the stable engine contract from owp-sdk.
package searchdef

import "github.com/timdodgson/open-workforce-platform/owp-sdk/searchdef"

type (
	Solution             = searchdef.Solution
	Move                 = searchdef.Move
	MoveResult           = searchdef.MoveResult
	Problem              = searchdef.Problem
	DatasetLoader        = searchdef.DatasetLoader
	Confidence           = searchdef.Confidence
	SafetyStatus         = searchdef.SafetyStatus
	SearchAction         = searchdef.SearchAction
	SearchProgress       = searchdef.SearchProgress
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
