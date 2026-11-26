package solve

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/G226Medi03/MorbesPipeline/internal/entity/solve"
	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
	source "github.com/G226Medi03/SourceEntity"
)

const (
	answerCheckTableLLM    = "solve_llm_answer_checks"
	answerCheckTableRule   = "solve_rule_answer_checks"
	answerCheckTableLegacy = "solve_answer_checks"
)

type AnswerCheckTable struct{}

func (AnswerCheckTable) init(ctx context.Context, db sqlutil.Executor) error {
	for _, table := range []string{answerCheckTableLLM, answerCheckTableRule} {
		if err := createAnswerCheckTable(ctx, db, table); err != nil {
			return err
		}
		if err := ensureAnswerCheckColumns(ctx, db, table); err != nil {
			return err
		}
	}
	return migrateLegacyAnswerChecks(ctx, db)
}

func createAnswerCheckTable(ctx context.Context, db sqlutil.Executor, table string) error {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		case_id INTEGER PRIMARY KEY,
		resp_id INTEGER NOT NULL DEFAULT 0,
		question_type TEXT NOT NULL DEFAULT '',
		contains_answer INTEGER NOT NULL DEFAULT 0 CHECK (contains_answer IN (0, 1)),
		answer_present INTEGER,
		answer_satisfied INTEGER,
		effect_matched INTEGER,
		matched_type TEXT NOT NULL DEFAULT 'none',
		analysis_json TEXT NOT NULL DEFAULT '{}',
		error TEXT NOT NULL DEFAULT '',
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (case_id) REFERENCES solve_cases(case_id) ON DELETE CASCADE
	);`, table)
	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("%s 테이블 생성 실패: %w", table, err)
	}
	return nil
}

func ensureAnswerCheckColumns(ctx context.Context, db sqlutil.Executor, table string) error {
	rows, err := db.QueryContext(ctx, `PRAGMA table_info(`+table+`)`)
	if err != nil {
		return fmt.Errorf("%s schema 조회 실패: %w", table, err)
	}
	columns := map[string]bool{}
	for rows.Next() {
		var cid, notnull, pk int
		var name, column_type string
		var default_value sql.NullString
		if err := rows.Scan(&cid, &name, &column_type, &notnull, &default_value, &pk); err != nil {
			_ = rows.Close()
			return err
		}
		columns[name] = true
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, column := range []struct {
		name       string
		definition string
	}{
		{"question_type", "TEXT NOT NULL DEFAULT ''"},
		{"answer_present", "INTEGER"},
		{"answer_satisfied", "INTEGER"},
		{"effect_matched", "INTEGER"},
		{"analysis_json", "TEXT NOT NULL DEFAULT '{}'"},
	} {
		if columns[column.name] {
			continue
		}
		if _, err := db.ExecContext(ctx, fmt.Sprintf(
			`ALTER TABLE %s ADD COLUMN %s %s`,
			table,
			column.name,
			column.definition,
		)); err != nil {
			return fmt.Errorf("%s %s 컬럼 추가 실패: %w", table, column.name, err)
		}
	}
	return nil
}

// migrateLegacyAnswerChecks moves the pre-split shared table into the LLM
// table. Existing rows were produced by the LLM/current path. The migration is
// idempotent and drops the legacy table only after the copy succeeds.
func migrateLegacyAnswerChecks(ctx context.Context, db sqlutil.Executor) error {
	found, err := tableExists(ctx, db, answerCheckTableLegacy)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	if err := ensureAnswerCheckColumns(ctx, db, answerCheckTableLegacy); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, fmt.Sprintf(`
		INSERT OR IGNORE INTO %s (
			case_id, resp_id, question_type, contains_answer, answer_present, answer_satisfied, effect_matched,
			matched_type, analysis_json, error, updated_at
		)
		SELECT
			case_id,
			COALESCE(resp_id, 0),
			COALESCE(question_type, ''),
			COALESCE(contains_answer, 0),
			answer_present,
			answer_satisfied,
			effect_matched,
			COALESCE(matched_type, 'none'),
			COALESCE(analysis_json, '{}'),
			COALESCE(error, ''),
			COALESCE(updated_at, CURRENT_TIMESTAMP)
		FROM %s`, answerCheckTableLLM, answerCheckTableLegacy)); err != nil {
		return fmt.Errorf("legacy answer check 이관 실패: %w", err)
	}
	if _, err := db.ExecContext(ctx, `DROP TABLE `+answerCheckTableLegacy); err != nil {
		return fmt.Errorf("legacy answer check 테이블 제거 실패: %w", err)
	}
	return nil
}

func tableExists(ctx context.Context, db sqlutil.Executor, table string) (bool, error) {
	var name string
	err := db.QueryRowContext(ctx, `SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&name)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("%s 존재 여부 확인 실패: %w", table, err)
	}
	return true, nil
}

func answerCheckTableName(kind solve.AnswerCheckKind) (string, error) {
	switch kind {
	case "", solve.AnswerCheckKindLLM:
		return answerCheckTableLLM, nil
	case solve.AnswerCheckKindRule:
		return answerCheckTableRule, nil
	default:
		return "", fmt.Errorf("알 수 없는 answer check kind 입니다: %s", kind)
	}
}

func (AnswerCheckTable) Upsert(ctx context.Context, db sqlutil.Executor, value solve.AnswerCheck) error {
	return (AnswerCheckTable{}).UpsertKind(ctx, db, solve.AnswerCheckKindLLM, value)
}

func (AnswerCheckTable) UpsertKind(ctx context.Context, db sqlutil.Executor, kind solve.AnswerCheckKind, value solve.AnswerCheck) error {
	table, err := answerCheckTableName(kind)
	if err != nil {
		return err
	}
	contains := 0
	if value.ContainsAnswer {
		contains = 1
	}
	answer_present := 0
	if value.AnswerPresent {
		answer_present = 1
	}
	answer_satisfied := 0
	if value.AnswerSatisfied {
		answer_satisfied = 1
	}
	effect_matched := 0
	if value.EffectMatched {
		effect_matched = 1
	}
	if _, err := db.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s (
			case_id, resp_id, question_type, contains_answer, answer_present, answer_satisfied, effect_matched,
			matched_type, analysis_json, error
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(case_id) DO UPDATE SET
			resp_id = excluded.resp_id,
			question_type = excluded.question_type,
			contains_answer = excluded.contains_answer,
			answer_present = excluded.answer_present,
			answer_satisfied = excluded.answer_satisfied,
			effect_matched = excluded.effect_matched,
			matched_type = excluded.matched_type,
			analysis_json = excluded.analysis_json,
			error = excluded.error,
			updated_at = CURRENT_TIMESTAMP`, table),
		value.CaseID,
		value.RespID.ToI64(),
		value.QuestionType,
		contains,
		answer_present,
		answer_satisfied,
		effect_matched,
		value.MatchedType,
		value.AnalysisJSON,
		value.Error,
	); err != nil {
		return fmt.Errorf("%s 저장 실패: %w", table, err)
	}
	return nil
}

func (AnswerCheckTable) Get(ctx context.Context, db sqlutil.Executor, caseID solve.CaseID) (solve.AnswerCheck, bool, error) {
	return (AnswerCheckTable{}).GetKind(ctx, db, solve.AnswerCheckKindLLM, caseID)
}

func (AnswerCheckTable) GetKind(ctx context.Context, db sqlutil.Executor, kind solve.AnswerCheckKind, caseID solve.CaseID) (solve.AnswerCheck, bool, error) {
	table, err := answerCheckTableName(kind)
	if err != nil {
		return solve.AnswerCheck{}, false, err
	}
	var value solve.AnswerCheck
	var caseIDValue string
	var respID int64
	var contains int
	var answer_present, answer_satisfied, effect_matched sql.NullInt64
	err = db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT case_id, resp_id, question_type, contains_answer, answer_present, answer_satisfied, effect_matched,
		       matched_type, analysis_json, error, updated_at
		FROM %s WHERE case_id = ?`, table), string(caseID)).
		Scan(&caseIDValue, &respID, &value.QuestionType, &contains, &answer_present, &answer_satisfied, &effect_matched,
			&value.MatchedType, &value.AnalysisJSON, &value.Error, &value.UpdatedAt)
	if err == sql.ErrNoRows {
		return solve.AnswerCheck{}, false, nil
	}
	if err != nil {
		return solve.AnswerCheck{}, false, fmt.Errorf("%s 조회 실패: %w", table, err)
	}
	value.CaseID = solve.CaseID(caseIDValue)
	value.RespID = source.NewRespId(respID)
	value.ContainsAnswer = contains != 0
	value.AnswerPresent = answer_present.Valid && answer_present.Int64 != 0
	value.AnswerSatisfied = answer_satisfied.Valid && answer_satisfied.Int64 != 0
	value.EffectMatched = effect_matched.Valid && effect_matched.Int64 != 0
	return value, true, nil
}

// Delete removes both independent answer-check results for a case.
func (AnswerCheckTable) Delete(ctx context.Context, db sqlutil.Executor, caseID solve.CaseID) error {
	for _, kind := range []solve.AnswerCheckKind{solve.AnswerCheckKindLLM, solve.AnswerCheckKindRule} {
		if err := (AnswerCheckTable{}).DeleteKind(ctx, db, kind, caseID); err != nil {
			return err
		}
	}
	return nil
}

func (AnswerCheckTable) DeleteKind(ctx context.Context, db sqlutil.Executor, kind solve.AnswerCheckKind, caseID solve.CaseID) error {
	table, err := answerCheckTableName(kind)
	if err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM `+table+` WHERE case_id = ?`, string(caseID)); err != nil {
		return fmt.Errorf("%s 삭제 실패: %w", table, err)
	}
	return nil
}

func (AnswerCheckTable) DeleteByCaseIDs(ctx context.Context, db sqlutil.Executor, caseIDs []solve.CaseID) error {
	if len(caseIDs) == 0 {
		return nil
	}
	placeholders := make([]string, 0, len(caseIDs))
	args := make([]any, 0, len(caseIDs))
	for _, id := range caseIDs {
		placeholders = append(placeholders, "?")
		args = append(args, string(id))
	}
	for _, table := range []string{answerCheckTableLLM, answerCheckTableRule} {
		if _, err := db.ExecContext(ctx, `
			DELETE FROM `+table+`
			WHERE case_id IN (`+strings.Join(placeholders, ",")+`)`, args...); err != nil {
			return fmt.Errorf("%s case별 삭제 실패: %w", table, err)
		}
	}
	return nil
}

func (AnswerCheckTable) DeleteByExamQuestionID(ctx context.Context, db sqlutil.Executor, ExamQuestionID int64) error {
	for _, table := range []string{answerCheckTableLLM, answerCheckTableRule} {
		if _, err := db.ExecContext(ctx, `
			DELETE FROM `+table+`
			WHERE case_id IN (SELECT case_id FROM solve_cases WHERE exam_question_id = ?)`, ExamQuestionID); err != nil {
			return fmt.Errorf("%s exam question 삭제 실패: %w", table, err)
		}
	}
	return nil
}
