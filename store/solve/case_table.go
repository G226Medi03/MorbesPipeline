package solve

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/MorbesPipeline/solve/domain"
	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
	source "github.com/G226Medi03/SourceEntity"
)

type CaseTable struct{}

func (CaseTable) init(ctx context.Context, db sqlutil.Executor) error {
	legacy, err := has_legacy_case_schema(ctx, db)
	if err != nil {
		return err
	}
	if legacy {
		_, _ = db.ExecContext(ctx, `DELETE FROM solve_steps`)
		_, _ = db.ExecContext(ctx, `DELETE FROM solve_llm_calls`)
		if _, err := db.ExecContext(ctx, `DROP TABLE solve_cases`); err != nil {
			return fmt.Errorf("legacy solve_cases 삭제 실패: %w", err)
		}
	}
	query := `CREATE TABLE IF NOT EXISTS solve_cases (
		case_id INTEGER PRIMARY KEY AUTOINCREMENT,
		exam_question_id INTEGER NOT NULL,
		situation_json TEXT NOT NULL DEFAULT '',
		answer_json TEXT NOT NULL DEFAULT '',
		state TEXT NOT NULL DEFAULT 'pending',
		stage TEXT NOT NULL DEFAULT 'queued',
		question_type TEXT NOT NULL DEFAULT '',
		error TEXT NOT NULL DEFAULT '',
		source_id INTEGER NOT NULL CHECK (source_id > 0),
		situation_source_id INTEGER NOT NULL DEFAULT 0,
		replay_of INTEGER NOT NULL DEFAULT 0,
		replay_mode TEXT NOT NULL DEFAULT '',
		replay_from_node TEXT NOT NULL DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("solve_cases 테이블 생성 실패: %w", err)
	}
	if _, err := db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_solve_cases_exam_question ON solve_cases(exam_question_id);`); err != nil {
		return fmt.Errorf("solve_cases exam_question index 생성 실패: %w", err)
	}
	if err := ensure_question_type_column(ctx, db); err != nil {
		return err
	}
	if err := drop_legacy_processed_column(ctx, db); err != nil {
		return err
	}
	return nil
}

func ensure_question_type_column(ctx context.Context, db sqlutil.Executor) error {
	rows, err := db.QueryContext(ctx, `PRAGMA table_info(solve_cases)`)
	if err != nil {
		return fmt.Errorf("solve_cases schema 조회 실패: %w", err)
	}
	has_column := false
	for rows.Next() {
		var cid, notnull, pk int
		var name, column_type string
		var default_value sql.NullString
		if err := rows.Scan(&cid, &name, &column_type, &notnull, &default_value, &pk); err != nil {
			_ = rows.Close()
			return err
		}
		if name == "question_type" {
			has_column = true
		}
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	_ = rows.Close()
	if has_column {
		return nil
	}
	if _, err := db.ExecContext(ctx, `ALTER TABLE solve_cases ADD COLUMN question_type TEXT NOT NULL DEFAULT ''`); err != nil {
		return fmt.Errorf("solve_cases question_type 컬럼 추가 실패: %w", err)
	}
	return nil
}

func drop_legacy_processed_column(ctx context.Context, db sqlutil.Executor) error {
	rows, err := db.QueryContext(ctx, `PRAGMA table_info(solve_cases)`)
	if err != nil {
		return fmt.Errorf("solve_cases schema 조회 실패: %w", err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	has_processed := false
	for rows.Next() {
		var cid, notnull, pk int
		var name, column_type string
		var default_value sql.NullString
		if err := rows.Scan(&cid, &name, &column_type, &notnull, &default_value, &pk); err != nil {
			return err
		}
		if name == "processed" {
			has_processed = true
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if !has_processed {
		return nil
	}
	if _, err := db.ExecContext(ctx, `ALTER TABLE solve_cases DROP COLUMN processed`); err != nil {
		return fmt.Errorf("solve_cases processed 컬럼 제거 실패: %w", err)
	}
	return nil
}

func (CaseTable) CreateForExamQuestion(ctx context.Context, db sqlutil.Executor, examQuestionID entity.ExamQuestionID, sourceID source.SourceId) (domain.CaseID, error) {
	if sourceID.ToI64() <= 0 {
		return "", fmt.Errorf("solve case source가 필요합니다")
	}
	result, err := db.ExecContext(ctx,
		`INSERT INTO solve_cases (exam_question_id, source_id, state, stage) VALUES (?, ?, 'pending', 'queued')`,
		examQuestionID.ToI64(), sourceID.ToI64())
	if err != nil {
		return "", fmt.Errorf("solve case 생성 실패: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("solve case id 조회 실패: %w", err)
	}
	return domain.CaseID(strconv.FormatInt(id, 10)), nil
}

func (CaseTable) CreateReplay(ctx context.Context, db sqlutil.Executor, examQuestionID entity.ExamQuestionID, replayOf domain.CaseID, mode, fromNode string) (domain.CaseID, error) {
	result, err := db.ExecContext(ctx, `
		INSERT INTO solve_cases (
			exam_question_id, source_id, state, stage,
			replay_of, replay_mode, replay_from_node
		) VALUES (?, (SELECT source_id FROM solve_cases WHERE case_id = ?), 'pending', 'queued', ?, ?, ?)`,
		examQuestionID.ToI64(), replayOf, replayOf, mode, fromNode)
	if err != nil {
		return "", fmt.Errorf("solve replay case 생성 실패: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("solve replay case id 조회 실패: %w", err)
	}
	return domain.CaseID(strconv.FormatInt(id, 10)), nil
}

func (CaseTable) List(ctx context.Context, db sqlutil.Executor, limit int) ([]domain.Case, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := db.QueryContext(ctx, `
		SELECT case_id, exam_question_id, situation_json, answer_json,
		       state, stage, question_type, error, source_id, situation_source_id,
		       replay_of, replay_mode, replay_from_node, created_at, updated_at
		FROM solve_cases
		ORDER BY updated_at DESC, created_at DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("solve case 목록 조회 실패: %w", err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	result := make([]domain.Case, 0)
	for rows.Next() {
		item, err := scan_case(rows.Scan)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (CaseTable) Get(ctx context.Context, db sqlutil.Executor, id domain.CaseID) (domain.Case, bool, error) {
	row := db.QueryRowContext(ctx, `
		SELECT case_id, exam_question_id, situation_json, answer_json,
		       state, stage, question_type, error, source_id, situation_source_id,
		       replay_of, replay_mode, replay_from_node, created_at, updated_at
		FROM solve_cases WHERE case_id = ?`, id)
	item, err := scan_case(row.Scan)
	if err == sql.ErrNoRows {
		return domain.Case{}, false, nil
	}
	if err != nil {
		return domain.Case{}, false, fmt.Errorf("solve case 조회 실패: %w", err)
	}
	return item, true, nil
}

func scan_case(scan func(...any) error) (domain.Case, error) {
	var item domain.Case
	var id int64
	var replayOf int64
	if err := scan(&id, &item.ExamQuestionID, &item.SituationJSON,
		&item.AnswerJSON, &item.State, &item.Stage, &item.QuestionType, &item.Error, &item.SourceID,
		&item.SituationSourceID, &replayOf, &item.ReplayMode, &item.ReplayFromNode,
		&item.CreatedAt, &item.UpdatedAt); err != nil {
		return domain.Case{}, err
	}
	item.ID = domain.CaseID(strconv.FormatInt(id, 10))
	if replayOf > 0 {
		item.ReplayOf = domain.CaseID(strconv.FormatInt(replayOf, 10))
	}
	return item, nil
}

func (CaseTable) ResetForReplay(ctx context.Context, db sqlutil.Executor, id domain.CaseID, mode, fromNode string) error {
	if _, err := db.ExecContext(ctx, `
		UPDATE solve_cases
		SET state = 'pending', stage = 'queued', error = '',
		    situation_json = '', answer_json = '',
		    situation_source_id = 0,
		    replay_mode = ?, replay_from_node = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE case_id = ?`, mode, fromNode, id); err != nil {
		return fmt.Errorf("solve replay case 초기화 실패: %w", err)
	}
	return nil
}

func (CaseTable) MarkInterrupted(ctx context.Context, db sqlutil.Executor) error {
	if _, err := db.ExecContext(ctx, `
		UPDATE solve_cases
		SET state = 'failed', stage = 'interrupted',
		    error = '서버가 중단되어 실행이 중단되었습니다. 다시 실행해 주세요.',
		    updated_at = CURRENT_TIMESTAMP
		WHERE state IN ('running', 'pending')`); err != nil {
		return fmt.Errorf("interrupted case 처리 실패: %w", err)
	}
	return nil
}

func (CaseTable) SetRunning(ctx context.Context, db sqlutil.Executor, id domain.CaseID) error {
	return (CaseTable{}).set(ctx, db, id, `state = 'running', error = '', updated_at = CURRENT_TIMESTAMP`, nil)
}

func (CaseTable) SetStage(ctx context.Context, db sqlutil.Executor, id domain.CaseID, stage string) error {
	return (CaseTable{}).set(ctx, db, id, `stage = ?, updated_at = CURRENT_TIMESTAMP`, []any{stage})
}

func (CaseTable) SetQuestionType(ctx context.Context, db sqlutil.Executor, id domain.CaseID, questionType string) error {
	return (CaseTable{}).set(ctx, db, id, `question_type = ?, updated_at = CURRENT_TIMESTAMP`, []any{questionType})
}

func (CaseTable) SetSituation(ctx context.Context, db sqlutil.Executor, id domain.CaseID, situation string) error {
	return (CaseTable{}).set(ctx, db, id, `situation_json = ?, updated_at = CURRENT_TIMESTAMP`, []any{situation})
}

func (CaseTable) SetSource(ctx context.Context, db sqlutil.Executor, id domain.CaseID, sourceID source.SourceId) error {
	return (CaseTable{}).set(ctx, db, id, `source_id = ?, updated_at = CURRENT_TIMESTAMP`, []any{sourceID})
}

func (CaseTable) SetSituationSource(ctx context.Context, db sqlutil.Executor, id domain.CaseID, sourceID source.SourceId) error {
	return (CaseTable{}).set(ctx, db, id, `situation_source_id = ?, updated_at = CURRENT_TIMESTAMP`, []any{sourceID})
}

func (CaseTable) SetAnswer(ctx context.Context, db sqlutil.Executor, id domain.CaseID, answerJSON string) error {
	return (CaseTable{}).set(ctx, db, id, `answer_json = ?, state = 'done', stage = 'done', error = '', updated_at = CURRENT_TIMESTAMP`, []any{answerJSON})
}

func (CaseTable) SetFailed(ctx context.Context, db sqlutil.Executor, id domain.CaseID, message string) error {
	return (CaseTable{}).set(ctx, db, id, `state = 'failed', error = ?, updated_at = CURRENT_TIMESTAMP`, []any{message})
}

func (CaseTable) set(ctx context.Context, db sqlutil.Executor, id domain.CaseID, assignment string, args []any) error {
	query := `UPDATE solve_cases SET ` + assignment + ` WHERE case_id = ?`
	args = append(args, string(id))
	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("solve case 갱신 실패: %w", err)
	}
	return nil
}

type FailedCase struct {
	ID             domain.CaseID
	ExamQuestionID entity.ExamQuestionID
	State          string
	Stage          string
	Error          string
}

func (CaseTable) ListExamQuestionIDsWithCases(ctx context.Context, db sqlutil.Executor) ([]entity.ExamQuestionID, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT exam_question_id
		FROM solve_cases
		WHERE exam_question_id > 0
		ORDER BY exam_question_id ASC`)
	if err != nil {
		return nil, fmt.Errorf("case 보유 문제 목록 조회 실패: %w", err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	result := make([]entity.ExamQuestionID, 0)
	for rows.Next() {
		var questionID entity.ExamQuestionID
		if err := rows.Scan(&questionID); err != nil {
			return nil, err
		}
		result = append(result, questionID)
	}
	return result, rows.Err()
}

func (CaseTable) CountFailed(ctx context.Context, db sqlutil.Executor) (int, error) {
	var count int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM solve_cases c
		LEFT JOIN solve_answer_checks a ON a.case_id = c.case_id
		WHERE c.state IN ('done', 'failed', 'cancelled')
		  AND COALESCE(a.answer_satisfied, 0) = 0
		  AND COALESCE(a.contains_answer, 0) = 0`).Scan(&count); err != nil {
		return 0, fmt.Errorf("실패 case 수 조회 실패: %w", err)
	}
	return count, nil
}

func (CaseTable) ListFailed(ctx context.Context, db sqlutil.Executor, limit int) ([]FailedCase, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := db.QueryContext(ctx, `
		SELECT c.case_id, c.exam_question_id, c.state, c.stage, c.error
		FROM solve_cases c
		LEFT JOIN solve_answer_checks a ON a.case_id = c.case_id
		WHERE c.state IN ('done', 'failed', 'cancelled')
		  AND COALESCE(a.answer_satisfied, 0) = 0
		  AND COALESCE(a.contains_answer, 0) = 0
		ORDER BY c.exam_question_id ASC
		LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("실패 case 목록 조회 실패: %w", err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	result := make([]FailedCase, 0)
	for rows.Next() {
		var item FailedCase
		var caseID string
		if err := rows.Scan(&caseID, &item.ExamQuestionID, &item.State, &item.Stage, &item.Error); err != nil {
			return nil, err
		}
		item.ID = domain.CaseID(caseID)
		result = append(result, item)
	}
	return result, rows.Err()
}

type ExamQuestionCaseStats struct {
	LatestCaseID domain.CaseID
	LatestState  string
	Processed    bool
	Attempts     int
}

func (CaseTable) StatsByExamQuestionIDs(ctx context.Context, db sqlutil.Executor, ids []entity.ExamQuestionID) (map[entity.ExamQuestionID]ExamQuestionCaseStats, error) {
	result := make(map[entity.ExamQuestionID]ExamQuestionCaseStats, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	placeholders := make([]string, 0, len(ids))
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		placeholders = append(placeholders, "?")
		args = append(args, id.ToI64())
	}
	rows, err := db.QueryContext(ctx, `
		SELECT c.exam_question_id, c.case_id, c.state,
		       COALESCE((
		           SELECT CASE WHEN COALESCE(ac.answer_satisfied, 0) != 0 OR ac.contains_answer != 0 THEN 1 ELSE 0 END
		           FROM solve_answer_checks ac
		           WHERE ac.case_id = c.case_id
		       ), 0)
		FROM solve_cases c
		WHERE c.exam_question_id IN (`+strings.Join(placeholders, ",")+`)
		ORDER BY c.updated_at DESC, c.created_at DESC`, args...)
	if err != nil {
		return nil, fmt.Errorf("exam question case stats 조회 실패: %w", err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	for rows.Next() {
		var questionID entity.ExamQuestionID
		var caseID string
		var state string
		var processed int
		if err := rows.Scan(&questionID, &caseID, &state, &processed); err != nil {
			return nil, err
		}
		stats := result[questionID]
		if stats.LatestCaseID == "" {
			stats.LatestCaseID = domain.CaseID(caseID)
			stats.LatestState = state
		}
		stats.Attempts++
		if processed != 0 {
			stats.Processed = true
		}
		result[questionID] = stats
	}
	return result, rows.Err()
}

func (CaseTable) CountByExamQuestionID(ctx context.Context, db sqlutil.Executor, examQuestionID entity.ExamQuestionID) (int, error) {
	var count int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM solve_cases
		WHERE exam_question_id = ?`, examQuestionID.ToI64()).Scan(&count); err != nil {
		return 0, fmt.Errorf("solve_cases 문제별 개수 조회 실패: %w", err)
	}
	return count, nil
}

func (CaseTable) ListIDsByExamQuestionID(ctx context.Context, db sqlutil.Executor, examQuestionID entity.ExamQuestionID) ([]domain.CaseID, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT case_id
		FROM solve_cases
		WHERE exam_question_id = ?
		ORDER BY case_id`, examQuestionID.ToI64())
	if err != nil {
		return nil, fmt.Errorf("solve_cases 문제별 id 조회 실패: %w", err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	result := make([]domain.CaseID, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result = append(result, domain.CaseID(strconv.FormatInt(id, 10)))
	}
	return result, rows.Err()
}

func (CaseTable) DeleteByIDs(ctx context.Context, db sqlutil.Executor, ids []domain.CaseID) error {
	if len(ids) == 0 {
		return nil
	}
	placeholders := make([]string, 0, len(ids))
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		placeholders = append(placeholders, "?")
		args = append(args, string(id))
	}
	if _, err := db.ExecContext(ctx, `
		DELETE FROM solve_cases
		WHERE case_id IN (`+strings.Join(placeholders, ",")+`)`, args...); err != nil {
		return fmt.Errorf("solve_cases 삭제 실패: %w", err)
	}
	return nil
}

func has_legacy_case_schema(ctx context.Context, db sqlutil.Executor) (bool, error) {
	rows, err := db.QueryContext(ctx, `PRAGMA table_info(solve_cases)`)
	if err != nil {
		return false, fmt.Errorf("solve_cases schema 조회 실패: %w", err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	for rows.Next() {
		var cid, notnull, pk int
		var name, column_type string
		var default_value sql.NullString
		if err := rows.Scan(&cid, &name, &column_type, &notnull, &default_value, &pk); err != nil {
			return false, err
		}
		if name == "problem" || name == "answer" || name == "explanation" {
			return true, nil
		}
	}
	return false, rows.Err()
}
