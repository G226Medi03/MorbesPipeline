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

func problem_only_instruction(stage string) string {
	base := "\n\n[중요] [정답]과 [해설]은 사용하지 말고 [문제] 본문만 근거로 판단하라."
	switch stage {
	case PromptMedicalProcedureOption:
		return base + " 문제의 모든 선택지(①~⑤)를 빠짐없이 후보로 추출하라. 정답 번호를 기준으로 후보를 고르지 마라."
	case PromptStatus, PromptAction, PromptRelation, PromptPatientDescription:
		return base + " 문제의 선택지에 나열된 치료/약물/검사는 환자 소견으로 추출하지 마라."
	default:
		return base
	}
}
