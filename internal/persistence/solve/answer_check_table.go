package solve

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/G226Medi03/MorbesPipeline/internal/domain/solve"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
	source "github.com/G226Medi03/SourceEntity"
)

type AnswerCheckTable struct{}

func (AnswerCheckTable) init(ctx context.Context, db sqlutil.Executor) error {
	query := `CREATE TABLE IF NOT EXISTS solve_answer_checks (
		case_id INTEGER PRIMARY KEY,
		resp_id INTEGER NOT NULL DEFAULT 0,
		question_type TEXT NOT NULL DEFAULT '',
		contains_answer INTEGER NOT NULL DEFAULT 0 CHECK (contains_answer IN (0, 1)),
		answer_present INTEGER,
		answer_satisfied INTEGER,
		effect_matched INTEGER,
		matched_type TEXT NOT NULL DEFAULT 'none',
		matched_key TEXT NOT NULL DEFAULT '',
		analysis_json TEXT NOT NULL DEFAULT '{}',
		error TEXT NOT NULL DEFAULT '',
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (case_id) REFERENCES solve_cases(case_id) ON DELETE CASCADE
	);`
	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("solve_answer_checks 테이블 생성 실패: %w", err)
	}
	if err := ensure_analysis_column(ctx, db); err != nil {
		return err
	}
	if err := ensure_answer_check_columns(ctx, db); err != nil {
		return err
	}
	return nil
}

func ensure_answer_check_columns(ctx context.Context, db sqlutil.Executor) error {
	rows, err := db.QueryContext(ctx, `PRAGMA table_info(solve_answer_checks)`)
	if err != nil {
		return fmt.Errorf("solve_answer_checks schema 조회 실패: %w", err)
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
	if !columns["question_type"] {
		if _, err := db.ExecContext(ctx, `ALTER TABLE solve_answer_checks ADD COLUMN question_type TEXT NOT NULL DEFAULT ''`); err != nil {
			return fmt.Errorf("solve_answer_checks question_type 컬럼 추가 실패: %w", err)
		}
	}
	if !columns["answer_present"] {
		if _, err := db.ExecContext(ctx, `ALTER TABLE solve_answer_checks ADD COLUMN answer_present INTEGER`); err != nil {
			return fmt.Errorf("solve_answer_checks answer_present 컬럼 추가 실패: %w", err)
		}
	}
	if !columns["answer_satisfied"] {
		if _, err := db.ExecContext(ctx, `ALTER TABLE solve_answer_checks ADD COLUMN answer_satisfied INTEGER`); err != nil {
			return fmt.Errorf("solve_answer_checks answer_satisfied 컬럼 추가 실패: %w", err)
		}
	}
	if !columns["effect_matched"] {
		if _, err := db.ExecContext(ctx, `ALTER TABLE solve_answer_checks ADD COLUMN effect_matched INTEGER`); err != nil {
			return fmt.Errorf("solve_answer_checks effect_matched 컬럼 추가 실패: %w", err)
		}
	}
	return nil
}

func ensure_analysis_column(ctx context.Context, db sqlutil.Executor) error {
	rows, err := db.QueryContext(ctx, `PRAGMA table_info(solve_answer_checks)`)
	if err != nil {
		return fmt.Errorf("solve_answer_checks schema 조회 실패: %w", err)
	}
	has_analysis := false
	for rows.Next() {
		var cid, notnull, pk int
		var name, column_type string
		var default_value sql.NullString
		if err := rows.Scan(&cid, &name, &column_type, &notnull, &default_value, &pk); err != nil {
			_ = rows.Close()
			return err
		}
		if name == "analysis_json" {
			has_analysis = true
		}
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if has_analysis {
		return nil
	}
	if _, err := db.ExecContext(ctx, `ALTER TABLE solve_answer_checks ADD COLUMN analysis_json TEXT NOT NULL DEFAULT '{}'`); err != nil {
		return fmt.Errorf("solve_answer_checks analysis_json 컬럼 추가 실패: %w", err)
	}
	return nil
}

func (AnswerCheckTable) Upsert(ctx context.Context, db sqlutil.Executor, value domain.AnswerCheck) error {
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
	if _, err := db.ExecContext(ctx, `
		INSERT INTO solve_answer_checks (
			case_id, resp_id, question_type, contains_answer, answer_present, answer_satisfied, effect_matched,
			matched_type, matched_key, analysis_json, error
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(case_id) DO UPDATE SET
			resp_id = excluded.resp_id,
			question_type = excluded.question_type,
			contains_answer = excluded.contains_answer,
			answer_present = excluded.answer_present,
			answer_satisfied = excluded.answer_satisfied,
			effect_matched = excluded.effect_matched,
			matched_type = excluded.matched_type,
			matched_key = excluded.matched_key,
			analysis_json = excluded.analysis_json,
			error = excluded.error,
			updated_at = CURRENT_TIMESTAMP`,
		value.CaseID,
		value.RespID.ToI64(),
		value.QuestionType,
		contains,
		answer_present,
		answer_satisfied,
		effect_matched,
		value.MatchedType,
		value.MatchedKey,
		value.AnalysisJSON,
		value.Error,
	); err != nil {
		return fmt.Errorf("answer check 저장 실패: %w", err)
	}
	return nil
}

func (AnswerCheckTable) Get(ctx context.Context, db sqlutil.Executor, caseID domain.CaseID) (domain.AnswerCheck, bool, error) {
	var value domain.AnswerCheck
	var caseIDValue string
	var respID int64
	var contains int
	var answer_present, answer_satisfied, effect_matched sql.NullInt64
	err := db.QueryRowContext(ctx, `
		SELECT case_id, resp_id, question_type, contains_answer, answer_present, answer_satisfied, effect_matched,
		       matched_type, matched_key, analysis_json, error, updated_at
		FROM solve_answer_checks WHERE case_id = ?`, string(caseID)).
		Scan(&caseIDValue, &respID, &value.QuestionType, &contains, &answer_present, &answer_satisfied, &effect_matched,
			&value.MatchedType, &value.MatchedKey, &value.AnalysisJSON, &value.Error, &value.UpdatedAt)
	if err == sql.ErrNoRows {
		return domain.AnswerCheck{}, false, nil
	}
	if err != nil {
		return domain.AnswerCheck{}, false, fmt.Errorf("answer check 조회 실패: %w", err)
	}
	value.CaseID = domain.CaseID(caseIDValue)
	value.RespID = source.NewRespId(respID)
	value.ContainsAnswer = contains != 0
	value.AnswerPresent = answer_present.Valid && answer_present.Int64 != 0
	value.AnswerSatisfied = answer_satisfied.Valid && answer_satisfied.Int64 != 0
	// Legacy rows predate answer_satisfied. Keep their old contains_answer
	// behavior until they are rechecked.
	if !answer_satisfied.Valid {
		value.AnswerSatisfied = value.ContainsAnswer
	}
	value.EffectMatched = effect_matched.Valid && effect_matched.Int64 != 0
	return value, true, nil
}

func (AnswerCheckTable) Delete(ctx context.Context, db sqlutil.Executor, caseID domain.CaseID) error {
	if _, err := db.ExecContext(ctx, `DELETE FROM solve_answer_checks WHERE case_id = ?`, string(caseID)); err != nil {
		return fmt.Errorf("answer check 삭제 실패: %w", err)
	}
	return nil
}

func (AnswerCheckTable) DeleteByCaseIDs(ctx context.Context, db sqlutil.Executor, caseIDs []domain.CaseID) error {
	if len(caseIDs) == 0 {
		return nil
	}
	placeholders := make([]string, 0, len(caseIDs))
	args := make([]any, 0, len(caseIDs))
	for _, id := range caseIDs {
		placeholders = append(placeholders, "?")
		args = append(args, string(id))
	}
	if _, err := db.ExecContext(ctx, `
		DELETE FROM solve_answer_checks
		WHERE case_id IN (`+strings.Join(placeholders, ",")+`)`, args...); err != nil {
		return fmt.Errorf("solve_answer_checks case별 삭제 실패: %w", err)
	}
	return nil
}

func (AnswerCheckTable) DeleteByExamQuestionID(ctx context.Context, db sqlutil.Executor, examQuestionID ids.ExamQuestionID) error {
	if _, err := db.ExecContext(ctx, `
		DELETE FROM solve_answer_checks
		WHERE case_id IN (SELECT case_id FROM solve_cases WHERE exam_question_id = ?)`, examQuestionID.ToI64()); err != nil {
		return fmt.Errorf("exam question answer check 삭제 실패: %w", err)
	}
	return nil
}
