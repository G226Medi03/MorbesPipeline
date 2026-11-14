package trait

import "strings"

// Trait is a semantic property of a medical procedure or of a procedure
// option in a case context.
type Trait string

// ProcedureType is the canonical procedure kind used for trait defaults.
type ProcedureType string

const (
	TraitResuscitative Trait = "resuscitative"
	TraitAcute         Trait = "acute"
	TraitDefinitive    Trait = "definitive"
	TraitFirstLine     Trait = "first_line"
	TraitRescue        Trait = "rescue"

	TraitInvasive      Trait = "invasive"
	TraitNonInvasive   Trait = "non_invasive"
	TraitSurgical      Trait = "surgical"
	TraitPharmacologic Trait = "pharmacologic"
	TraitDiagnostic    Trait = "diagnostic"
	TraitMonitoring    Trait = "monitoring"
	TraitSupportive    Trait = "supportive"
	TraitChronic       Trait = "chronic"
	TraitMaintenance   Trait = "maintenance"
	TraitProphylactic  Trait = "prophylactic"
	TraitEducational   Trait = "educational"
	TraitNursing       Trait = "nursing"

	KindProcedure      ProcedureType = "unknown"
	KindSurgery        ProcedureType = "surgery"
	KindTransplant     ProcedureType = "transplant"
	KindInfuse         ProcedureType = "infuse"
	KindCleaning       ProcedureType = "cleaning"
	KindDiagnosticTest ProcedureType = "diagnostic_test"
	KindEducation      ProcedureType = "education"
	KindInstruction    ProcedureType = "instruction"
	KindInsertion      ProcedureType = "insertion"
	KindUnknown        ProcedureType = "unknown"
)

var intrinsicTraitSet = map[Trait]struct{}{
	TraitInvasive:      {},
	TraitNonInvasive:   {},
	TraitSurgical:      {},
	TraitPharmacologic: {},
	TraitDiagnostic:    {},
	TraitMonitoring:    {},
	TraitSupportive:    {},
	TraitDefinitive:    {},
	TraitResuscitative: {},
	TraitEducational:   {},
	TraitNursing:       {},
}

var priorityTraitOrder = []Trait{
	TraitResuscitative,
	TraitAcute,
	TraitDefinitive,
	TraitFirstLine,
	TraitRescue,
}

var contextTraitSet = map[Trait]struct{}{
	TraitAcute:        {},
	TraitChronic:      {},
	TraitFirstLine:    {},
	TraitRescue:       {},
	TraitMaintenance:  {},
	TraitProphylactic: {},
}

var traitAliases = map[string]Trait{
	"resuscitative": TraitResuscitative,
	"resuscitation": TraitResuscitative,
	"resuscitate":   TraitResuscitative,
	"소생":            TraitResuscitative,

	"acute":     TraitAcute,
	"urgent":    TraitAcute,
	"emergency": TraitAcute,
	"급성":        TraitAcute,
	"응급":        TraitAcute,

	"definitive": TraitDefinitive,
	"definite":   TraitDefinitive,
	"curative":   TraitDefinitive,
	"근치":         TraitDefinitive,

	"first_line": TraitFirstLine,
	"first":      TraitFirstLine,
	"primary":    TraitFirstLine,
	"1차":         TraitFirstLine,
	"일차":         TraitFirstLine,

	"rescue":  TraitRescue,
	"salvage": TraitRescue,
	"구조":      TraitRescue,
	"구제":      TraitRescue,

	"invasive": TraitInvasive,
	"침습":       TraitInvasive,
	"침습적":      TraitInvasive,

	"non_invasive": TraitNonInvasive,
	"noninvasive":  TraitNonInvasive,
	"비침습":          TraitNonInvasive,
	"비침습적":         TraitNonInvasive,

	"surgical": TraitSurgical,
	"surgery":  TraitSurgical,
	"수술":       TraitSurgical,

	"pharmacologic":   TraitPharmacologic,
	"pharmacological": TraitPharmacologic,
	"drug":            TraitPharmacologic,
	"약물":              TraitPharmacologic,

	"diagnostic": TraitDiagnostic,
	"diagnosis":  TraitDiagnostic,
	"진단":         TraitDiagnostic,

	"monitoring": TraitMonitoring,
	"monitor":    TraitMonitoring,
	"모니터링":       TraitMonitoring,
	"감시":         TraitMonitoring,

	"supportive": TraitSupportive,
	"보조":         TraitSupportive,
	"대증":         TraitSupportive,

	"chronic": TraitChronic,
	"만성":      TraitChronic,

	"maintenance": TraitMaintenance,
	"유지":          TraitMaintenance,

	"prophylactic": TraitProphylactic,
	"preventive":   TraitProphylactic,
	"예방":           TraitProphylactic,

	"educational": TraitEducational,
	"education":   TraitEducational,
	"교육":          TraitEducational,

	"nursing": TraitNursing,
	"간호":      TraitNursing,
}

// NormalizeTrait converts a raw trait string to its canonical form.
func NormalizeTrait(raw string) (Trait, bool) {
	key := strings.ToLower(strings.TrimSpace(raw))
	key = strings.ReplaceAll(key, "-", "_")
	key = strings.ReplaceAll(key, " ", "_")
	for strings.Contains(key, "__") {
		key = strings.ReplaceAll(key, "__", "_")
	}
	if value, ok := traitAliases[key]; ok {
		return value, true
	}
	return "", false
}

// NormalizeTraits converts raw values, removes duplicates, and drops unknown
// traits. The returned order follows the raw input order.
func NormalizeTraits(values []string) []Trait {
	result := make([]Trait, 0, len(values))
	seen := make(map[Trait]struct{}, len(values))
	for _, raw := range values {
		value, ok := NormalizeTrait(raw)
		if !ok {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func ContainsTrait(values []Trait, target Trait) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func IsPriorityTrait(value Trait) bool {
	for _, trait := range priorityTraitOrder {
		if trait == value {
			return true
		}
	}
	return false
}

func IsIntrinsicTrait(value Trait) bool {
	_, ok := intrinsicTraitSet[value]
	return ok
}

func IsContextTrait(value Trait) bool {
	_, ok := contextTraitSet[value]
	return ok
}

// IntrinsicTraits returns traits that describe the procedure itself. They are
// persisted with the canonical procedure.
func IntrinsicTraits(values []Trait) []Trait {
	result := make([]Trait, 0, len(values))
	for _, value := range values {
		if IsIntrinsicTrait(value) {
			result = append(result, value)
		}
	}
	return result
}

// ContextTraits returns traits that depend on the case context. They are
// persisted with the medical procedure option.
func ContextTraits(values []Trait) []Trait {
	result := make([]Trait, 0, len(values))
	for _, value := range values {
		if IsContextTrait(value) {
			result = append(result, value)
		}
	}
	return result
}

// PriorityTraits returns the priority traits present in values, ordered from
// highest to lowest priority.
func PriorityTraits(values []Trait) []Trait {
	result := make([]Trait, 0, len(priorityTraitOrder))
	for _, trait := range priorityTraitOrder {
		if ContainsTrait(values, trait) {
			result = append(result, trait)
		}
	}
	return result
}

// DescriptiveTraits returns all non-priority traits, ordered by their first
// appearance in values.
func DescriptiveTraits(values []Trait) []Trait {
	result := make([]Trait, 0, len(values))
	seen := make(map[Trait]struct{}, len(values))
	for _, value := range values {
		if IsPriorityTrait(value) {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

// ComparePriority compares two trait sets by presence of priority traits.
// It returns -1 when left has higher priority, 1 when right has higher
// priority, and 0 when both have identical priority traits.
func ComparePriority(left, right []Trait) int {
	for _, trait := range priorityTraitOrder {
		leftHas := ContainsTrait(left, trait)
		rightHas := ContainsTrait(right, trait)
		if leftHas == rightHas {
			continue
		}
		if leftHas {
			return -1
		}
		return 1
	}
	return 0
}

func PriorityReason(values []Trait) string {
	traits := PriorityTraits(values)
	if len(traits) == 0 {
		return ""
	}
	parts := make([]string, 0, len(traits))
	for _, trait := range traits {
		parts = append(parts, string(trait))
	}
	return strings.Join(parts, " > ")
}

func CategoryRank(category string) int {
	switch strings.TrimSpace(category) {
	case "treatment":
		return 0
	case "diagnostic_test":
		return 1
	case "monitoring":
		return 2
	case "nursing_intervention":
		return 3
	case "education":
		return 4
	default:
		return 99
	}
}

// DefaultIntrinsicTraits derives intrinsic traits when extraction did not
// provide any. Context traits intentionally have no defaults.
func DefaultIntrinsicTraits(kind ProcedureType) []Trait {
	switch kind {
	case KindSurgery, KindTransplant:
		return []Trait{TraitSurgical, TraitInvasive, TraitDefinitive}
	case KindInfuse:
		return []Trait{TraitPharmacologic}
	case KindDiagnosticTest:
		return []Trait{TraitDiagnostic}
	case KindCleaning:
		return []Trait{TraitNursing, TraitSupportive}
	case KindInsertion:
		return []Trait{TraitInvasive, TraitNursing}
	case KindEducation:
		return []Trait{TraitEducational, TraitNursing}
	case KindInstruction:
		return []Trait{TraitEducational, TraitNursing, TraitSupportive}
	default:
		return nil
	}
}

// MergeTraits merges trait lists without duplicates, preserving the order of
// first appearance.
func MergeTraits(groups ...[]Trait) []Trait {
	result := make([]Trait, 0)
	seen := make(map[Trait]struct{})
	for _, group := range groups {
		for _, value := range group {
			if _, exists := seen[value]; exists {
				continue
			}
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}
	return result
}
