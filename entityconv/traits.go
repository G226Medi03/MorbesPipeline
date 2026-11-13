// Package entityconv contains conversions between trait domain values and
// persisted entity values.
package entityconv

import (
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/MorbesPipeline/trait"
)

// EntityTraits converts trait domain values to persisted entity traits.
func EntityTraits(values []trait.Trait) []entity.Trait {
	if len(values) == 0 {
		return nil
	}
	result := make([]entity.Trait, 0, len(values))
	for _, value := range values {
		result = append(result, entity.Trait(value))
	}
	return result
}
