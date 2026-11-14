package csolve

import "github.com/G226Medi03/MedicalEntityParser/parser"

func resolve_actor_kind_from_name(name string) string {
	parsed, err := parser.ParseActor(name)
	if err != nil {
		return ""
	}
	return actor_kind(parsed)
}
