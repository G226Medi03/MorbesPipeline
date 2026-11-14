// Package call re-exports the llm-client call package.
package call

import ext "github.com/G226Medi03/llm-client/call"

type (
	Client       = ext.Client
	GenerateFunc = ext.GenerateFunc
)

var (
	FromLLMClient       = ext.FromLLMClient
	NewFunc             = ext.NewFunc
	LLM_MAX_CONCURRENT  = ext.LLM_MAX_CONCURRENT
	LLM_REQUEST_RETRIES = ext.LLM_REQUEST_RETRIES
	LLM_REQUEST_TIMEOUT = ext.LLM_REQUEST_TIMEOUT
)
