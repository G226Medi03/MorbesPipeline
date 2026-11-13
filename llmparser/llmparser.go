// Package llmparser re-exports the parser helper used by the pipeline.
package llmparser

import (
	"encoding/json"

	parser "github.com/G226Medi03/MedicalEntityParser/llmparser"
)

// ExtractJSONElements extracts JSON elements from an LLM response.
func ExtractJSONElements(raw string) ([]json.RawMessage, error) {
	return parser.ExtractJSONElements(raw)
}
