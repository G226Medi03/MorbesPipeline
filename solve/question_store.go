package csolve

import "github.com/G226Medi03/MorbesPipeline/solve/domain"

// The question port is owned by the domain package so consumers do not need to
// import the pipeline implementation.
type (
	QuestionInput = domain.QuestionInput
	QuestionStore = domain.QuestionStore
)
