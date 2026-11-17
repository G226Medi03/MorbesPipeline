package csolve

import (
	"regexp"
	"strconv"
	"strings"
)

var problemChoiceLinePattern = regexp.MustCompile(`(?m)^\s*(?:([0-9]+)[.)]|([①-⑳]))\s*(.+?)\s*$`)

type ProblemChoice struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
}

// problem_stem returns the problem text without numbered/encircled choices.
// Findings extraction must only see the clinical stem; the choices belong to
// the candidate-treatment stage.
func problem_stem(problem string) string {
	lines := strings.Split(strings.ReplaceAll(problem, "\r\n", "\n"), "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		if problemChoiceLinePattern.MatchString(line) {
			continue
		}
		kept = append(kept, line)
	}
	return strings.TrimSpace(strings.Join(kept, "\n"))
}

func problem_choices(problem string) []ProblemChoice {
	result := make([]ProblemChoice, 0)
	seen := map[int]bool{}
	for _, match := range problemChoiceLinePattern.FindAllStringSubmatch(problem, -1) {
		index := 0
		if match[1] != "" {
			index, _ = strconv.Atoi(match[1])
		} else if match[2] != "" {
			runes := []rune(match[2])
			if len(runes) > 0 && runes[0] >= '①' && runes[0] <= '⑳' {
				index = int(runes[0]-'①') + 1
			}
		}
		text := strings.TrimSpace(match[3])
		if index <= 0 || text == "" || seen[index] {
			continue
		}
		seen[index] = true
		result = append(result, ProblemChoice{Index: index, Text: text})
	}
	return result
}
