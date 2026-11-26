package tracing

import (
	"encoding/json"
	"strings"
)

// ParseCauseTraceJSON normalizes a persisted cause-trace step payload into the
// tracing result shape. It keeps only the fields owned by CauseTraceResult and
// the known candidate/path fields so older step payloads remain readable.
func ParseCauseTraceJSON(raw string, target *CauseTraceResult) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(trimmed), &fields); err != nil {
		return err
	}
	allowed := map[string]json.RawMessage{}
	for _, key := range []string{"observed", "missing"} {
		if value, ok := fields[key]; ok {
			allowed[key] = value
		}
	}
	if rawCandidates, ok := fields["candidates"]; ok {
		var candidates []map[string]json.RawMessage
		if err := json.Unmarshal(rawCandidates, &candidates); err != nil {
			return err
		}
		cleanCandidates := make([]map[string]json.RawMessage, 0, len(candidates))
		for _, candidate := range candidates {
			clean := map[string]json.RawMessage{}
			for _, key := range []string{"causeType", "causeKind", "causeLabel", "viaActionKind", "depth", "sourceIds"} {
				if value, exists := candidate[key]; exists {
					clean[key] = value
				}
			}
			if rawPath, exists := candidate["path"]; exists {
				var nodes []map[string]json.RawMessage
				if err := json.Unmarshal(rawPath, &nodes); err == nil {
					cleanNodes := make([]map[string]json.RawMessage, 0, len(nodes))
					for _, node := range nodes {
						cleanNode := map[string]json.RawMessage{}
						for _, key := range []string{"type", "label"} {
							if value, ok := node[key]; ok {
								cleanNode[key] = value
							}
						}
						cleanNodes = append(cleanNodes, cleanNode)
					}
					if encoded, err := json.Marshal(cleanNodes); err == nil {
						clean["path"] = encoded
					}
				}
			}
			cleanCandidates = append(cleanCandidates, clean)
		}
		if encoded, err := json.Marshal(cleanCandidates); err == nil {
			allowed["candidates"] = encoded
		}
	}
	normalized, err := json.Marshal(allowed)
	if err != nil {
		return err
	}
	return json.Unmarshal(normalized, target)
}
