// Package csituation parses and converts structured medical situations.
package csituation

type Situation struct {
	Findings Findings `json:"findings"`
}
