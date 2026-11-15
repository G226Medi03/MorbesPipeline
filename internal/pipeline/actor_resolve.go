package csolve

import morbesllm "github.com/G226Medi03/MorbesLLmPipeline"

func resolve_actor_kind_from_name(name string) string {
	parsed, err := morbesllm.ParseActorName(name)
	if err != nil {
		return ""
	}
	return actor_kind(parsed)
}
