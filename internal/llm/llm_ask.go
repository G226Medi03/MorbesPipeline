package cservice

import (
	"context"
	"time"

	cllmcall "github.com/G226Medi03/MorbesPipeline/internal/external/llmclient/call"
	source "github.com/G226Medi03/MorbesPipeline/internal/external/sourceentity"
	"github.com/G226Medi03/MorbesPipeline/internal/persistence"
	"github.com/G226Medi03/MorbesPipeline/internal/support/logs"
)

// LLMCacheMode controls how an LLM request interacts with llm_responses.
type LLMCacheMode int

const (
	// LLMCacheUse reads and writes the shared LLM response cache.
	LLMCacheUse LLMCacheMode = iota
	// LLMCacheBypass skips cache reads. It is used by code replay so the replay
	// client is always asked for the stored raw output.
	LLMCacheBypass
	// LLMCacheRefresh skips cache reads, calls the live LLM, and refreshes the
	// existing cache row with the new output.
	LLMCacheRefresh
)

// AskLLM reads or writes the shared LLM response cache.
func AskLLM(
	ctx context.Context,
	llm *cllmcall.Client,
	st *store.Store,
	prompt string,
	referedSourceID source.SourceId,
) (*source.LlmResponse, error) {
	return ask_llm(ctx, llm, st, prompt, referedSourceID)
}

// AskLLMWithMode runs an LLM request with the requested cache mode.
func AskLLMWithMode(
	ctx context.Context,
	llm *cllmcall.Client,
	st *store.Store,
	prompt string,
	referedSourceID source.SourceId,
	cacheMode LLMCacheMode,
) (*source.LlmResponse, error) {
	return ask_llm_with_mode(ctx, llm, st, prompt, referedSourceID, cacheMode)
}
func ask_llm(
	ctx context.Context,
	llm *cllmcall.Client,
	st *store.Store,
	prompt string,
	refered_source_id source.SourceId,
) (*source.LlmResponse, error) {
	return ask_llm_with_mode(ctx, llm, st, prompt, refered_source_id, LLMCacheUse)
}

func ask_llm_with_mode(
	ctx context.Context,
	llm *cllmcall.Client,
	st *store.Store,
	prompt string,
	refered_source_id source.SourceId,
	cache_mode LLMCacheMode,
) (*source.LlmResponse, error) {
	if cache_mode == LLMCacheUse {
		logs.Info("LLM cache 조회 시작",
			logs.AttrOf("prompt_len", len(prompt)),
			logs.AttrOf("refered_source_id", refered_source_id))
		cached, found, ce := st.FindLlmResponse(llm.Model(), prompt, refered_source_id)
		if ce != nil {
			logs.Error("LLM cache 조회 실패", logs.AttrOf("error", ce))
			return nil, ce
		}
		if found {
			logs.Info("LLM cache hit",
				logs.AttrOf("prompt_len", len(prompt)),
				logs.AttrOf("source_id", cached.SourceID))
			return cached, nil
		}
		logs.Info("LLM cache miss",
			logs.AttrOf("prompt_len", len(prompt)),
			logs.AttrOf("refered_source_id", refered_source_id))
	} else {
		logs.Info("LLM cache 우회",
			logs.AttrOf("prompt_len", len(prompt)),
			logs.AttrOf("refered_source_id", refered_source_id))
	}
	started := time.Now()
	logs.Info("LLM 요청 시작",
		logs.AttrOf("prompt_len", len(prompt)),
		logs.AttrOf("refered_source_id", refered_source_id))
	output, ce := llm.Generate(ctx, prompt)
	if ce != nil {
		logs.Error("LLM 요청 실패",
			logs.AttrOf("prompt_len", len(prompt)),
			logs.AttrOf("duration_ms", time.Since(started).Milliseconds()),
			logs.AttrOf("error", ce))
		return nil, ce
	}
	logs.Info("LLM 요청 완료",
		logs.AttrOf("prompt_len", len(prompt)),
		logs.AttrOf("duration_ms", time.Since(started).Milliseconds()))
	logs.Info("LLM 응답 저장 시작",
		logs.AttrOf("prompt_len", len(prompt)),
		logs.AttrOf("output_len", len(output)))
	if cache_mode == LLMCacheRefresh {
		return st.RefreshLlmResponse(llm.Model(), prompt, refered_source_id, output)
	}
	response, err := st.GetOrInsertLlmResponse(llm.Model(), prompt, refered_source_id, output)
	if err == nil && response != nil && response.Output != output {
		copy := *response
		copy.Output = output
		response = &copy
	}
	return response, err
}
