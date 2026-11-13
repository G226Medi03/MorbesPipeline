package store

import "fmt"

type kindPtrCheck struct {
	name  string
	query string
}

var kindPtrChecks = []kindPtrCheck{
	{
		"estatus",
		`SELECT COUNT(*) FROM estatus s WHERE s.ptr <= 0 OR NOT (
			(s.kind = 'increased_index' AND EXISTS (SELECT 1 FROM status_increased_index n WHERE n.id = s.ptr)) OR
			(s.kind = 'decreased_index' AND EXISTS (SELECT 1 FROM status_decreased_index n WHERE n.id = s.ptr)) OR
			(s.kind = 'have_disease' AND EXISTS (SELECT 1 FROM status_have_disease n WHERE n.id = s.ptr)) OR
			(s.kind = 'have_symptom' AND EXISTS (SELECT 1 FROM status_have_symptom n WHERE n.id = s.ptr)) OR
			(s.kind = 'damaged' AND EXISTS (SELECT 1 FROM status_damaged n WHERE n.id = s.ptr)) OR
			(s.kind = 'inflammation' AND EXISTS (SELECT 1 FROM status_inflammation n WHERE n.id = s.ptr)) OR
			(s.kind = 'received_medical_procedure' AND EXISTS (SELECT 1 FROM status_received_medical_procedure n WHERE n.id = s.ptr))
		)`,
	},
	{
		"eaction",
		`SELECT COUNT(*) FROM eaction a WHERE a.ptr <= 0 OR NOT (
			(a.kind = 'cause_status' AND EXISTS (SELECT 1 FROM action_cause_status n WHERE n.id = a.ptr)) OR
			(a.kind = 'block' AND EXISTS (SELECT 1 FROM action_block n WHERE n.id = a.ptr)) OR
			(a.kind = 'relax' AND EXISTS (SELECT 1 FROM action_relax n WHERE n.id = a.ptr))
		)`,
	},
	{
		"eactor",
		`SELECT COUNT(*) FROM eactor a WHERE a.ptr <= 0 OR NOT (
			(a.kind = 'device' AND EXISTS (SELECT 1 FROM actor_device n WHERE n.id = a.ptr AND n.name = a.name)) OR
			(a.kind = 'unknown_actor' AND EXISTS (SELECT 1 FROM actor_unknown n WHERE n.id = a.ptr AND n.name = a.name)) OR
			(a.kind = 'procedure' AND EXISTS (SELECT 1 FROM actor_procedure n WHERE n.id = a.ptr AND n.name = a.name))
		)`,
	},
	{
		"emedical_procedure",
		`SELECT
			(SELECT COUNT(*) FROM emedical_procedure WHERE ptr <= 0) +
			(SELECT COUNT(*) FROM emedical_procedure p WHERE p.kind IN ('Procedure','DiagnosticTest','Surgery','Unknown') AND NOT EXISTS (SELECT 1 FROM procedure_named n WHERE n.id = p.ptr)) +
			(SELECT COUNT(*) FROM emedical_procedure p WHERE p.kind = 'Instruction' AND NOT EXISTS (SELECT 1 FROM procedure_instruction n WHERE n.id = p.ptr)) +
			(SELECT COUNT(*) FROM emedical_procedure p WHERE p.kind = 'Education' AND NOT EXISTS (SELECT 1 FROM procedure_education n WHERE n.id = p.ptr)) +
			(SELECT COUNT(*) FROM emedical_procedure p WHERE p.kind = 'Infuse' AND NOT EXISTS (SELECT 1 FROM procedure_infuse n WHERE n.id = p.ptr)) +
			(SELECT COUNT(*) FROM emedical_procedure p WHERE p.kind = 'Transplant' AND NOT EXISTS (SELECT 1 FROM procedure_transplant n WHERE n.id = p.ptr)) +
			(SELECT COUNT(*) FROM emedical_procedure p WHERE p.kind = 'Cleaning' AND NOT EXISTS (SELECT 1 FROM procedure_cleaning n WHERE n.id = p.ptr)) +
			(SELECT COUNT(*) FROM emedical_procedure p WHERE p.kind = 'Insertion' AND NOT EXISTS (SELECT 1 FROM procedure_insertion n WHERE n.id = p.ptr))`,
	},
}

// validateKindPtr verifies that kind/ptr rows point at existing subtype rows.
// With no arguments it checks every kind/ptr table; arguments restrict the
// check to the named tables.
func (s *Store) validateKindPtr(tables ...string) error {
	selected := make(map[string]struct{}, len(tables))
	for _, table := range tables {
		selected[table] = struct{}{}
	}
	for _, check := range kindPtrChecks {
		if len(selected) > 0 {
			if _, ok := selected[check.name]; !ok {
				continue
			}
		}
		var count int
		if err := s.QueryRowContext(s.ctx, check.query).Scan(&count); err != nil {
			return fmt.Errorf("kind/ptr %s 검증 실패: %w", check.name, err)
		}
		if count != 0 {
			return fmt.Errorf("kind/ptr %s: 유효하지 않은 row %d개", check.name, count)
		}
	}
	return nil
}
