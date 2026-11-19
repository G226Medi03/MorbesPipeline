package csolve

import "github.com/G226Medi03/MorbesPipeline/internal/pipeline/input"

type ProblemChoice = input.ProblemChoice

func problem_stem(problem string) string { return input.Stem(problem) }

func problem_choices(problem string) []ProblemChoice { return input.Choices(problem) }
