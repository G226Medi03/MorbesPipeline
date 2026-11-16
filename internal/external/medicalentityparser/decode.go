// Package medicalentityparser exposes the MedicalEntityParser surface used by
// the pipeline through the external adapter boundary.
package medicalentityparser

import "github.com/G226Medi03/MedicalEntityParser/llmparser"

// DecodeStatuses decodes one or more raw status objects.
var DecodeStatuses = llmparser.DecodeStatuses
