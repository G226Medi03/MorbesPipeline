package benchmarkresult

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	sqlutil "github.com/G226Medi03/SQLCore/sqlite"
	source "github.com/G226Medi03/SourceEntity"
)

type Filter struct {
	Correct     *bool
	ResponseIDs []source.RespId
	Order       string
	Limit       int
	Offset      int
}

type Row struct {
	ID             int64
	ExamQuestionID int64
	LLMResponseID  source.RespId
	Correct        bool
	CreatedAt      string
}

type Table struct{}

func (Table) Init(ctx context.Context, db sqlutil.Executor) error {
	rows, err := db.QueryContext(ctx, `PRAGMA table_info(llm_problem_results)`)
	if err != nil {
		return fmt.Errorf("llm_problem_results schema 조회 실패: %w", err)
	}
	legacy := false
	for rows.Next() {
		var cid, notnull, pk int
		var name, columnType string
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &columnType, &notnull, &defaultValue, &pk); err != nil {
			_ = rows.Close()
			return err
		}
		if name == "model" || name == "predicted" || name == "updated_at" {
			legacy = true
		}
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	_ = rows.Close()
	if legacy {
		if _, err := db.ExecContext(ctx, `DROP TABLE llm_problem_results`); err != nil {
			return err
		}
	}
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS llm_problem_results (
			llm_problem_result_id INTEGER PRIMARY KEY AUTOINCREMENT,
			exam_question_id  INTEGER NOT NULL,
			llm_response_id   INTEGER NOT NULL,
			correct           INTEGER NOT NULL DEFAULT 0 CHECK (correct IN (0, 1)),
			created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (llm_response_id)
		);`); err != nil {
		return fmt.Errorf("llm_problem_results 테이블 생성 실패: %w", err)
	}
	for _, query := range []string{
		`CREATE INDEX IF NOT EXISTS idx_llm_problem_results_question
		 ON llm_problem_results(exam_question_id);`,
		`CREATE INDEX IF NOT EXISTS idx_llm_problem_results_correct
		 ON llm_problem_results(correct);`,
	} {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func (Table) GetByID(ctx context.Context, db sqlutil.Executor, id int64) (Row, bool, error) {
	var row Row
	var correct int
	err := db.QueryRowContext(ctx, `
		SELECT llm_problem_result_id, exam_question_id, llm_response_id, correct, created_at
		FROM llm_problem_results
		WHERE llm_problem_result_id = ?`, id).
		Scan(&row.ID, &row.ExamQuestionID, &row.LLMResponseID, &correct, &row.CreatedAt)
	if err == sql.ErrNoRows {
		return Row{}, false, nil
	}
	if err != nil {
		return Row{}, false, fmt.Errorf("llm_problem_results 조회 실패: %w", err)
	}
	row.Correct = correct != 0
	return row, true, nil
}

func (Table) List(ctx context.Context, db sqlutil.Executor, filter Filter) ([]Row, error) {
	where, args, err := filterWhere(filter)
	if err != nil {
		return nil, err
	}
	order := "DESC"
	if strings.EqualFold(strings.TrimSpace(filter.Order), "asc") {
		order = "ASC"
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	args = append(args, limit, offset)
	rows, err := db.QueryContext(ctx, `
		SELECT llm_problem_result_id, exam_question_id, llm_response_id, correct, created_at
		FROM llm_problem_results
		`+where+`
		ORDER BY created_at `+order+`, llm_problem_result_id `+order+`
		LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return nil, fmt.Errorf("llm_problem_results 조회 실패: %w", err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	result := make([]Row, 0, limit)
	for rows.Next() {
		var item Row
		var correct int
		if err := rows.Scan(&item.ID, &item.ExamQuestionID, &item.LLMResponseID, &correct, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.Correct = correct != 0
		result = append(result, item)
	}
	return result, rows.Err()
}

func (Table) Count(ctx context.Context, db sqlutil.Executor, filter Filter) (int, error) {
	where, args, err := filterWhere(filter)
	if err != nil {
		return 0, err
	}
	var count int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM llm_problem_results
		`+where, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("llm_problem_results 개수 조회 실패: %w", err)
	}
	return count, nil
}

func (Table) CountByExamQuestionID(ctx context.Context, db sqlutil.Executor, ExamQuestionID int64) (int, error) {
	var count int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM llm_problem_results
		WHERE exam_question_id = ?`, ExamQuestionID).Scan(&count); err != nil {
		return 0, fmt.Errorf("llm_problem_results 문제별 개수 조회 실패: %w", err)
	}
	return count, nil
}

func (Table) ListResponseIDsByExamQuestionID(ctx context.Context, db sqlutil.Executor, ExamQuestionID int64) ([]source.RespId, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT llm_response_id
		FROM llm_problem_results
		WHERE exam_question_id = ?
		ORDER BY llm_response_id`, ExamQuestionID)
	if err != nil {
		return nil, fmt.Errorf("llm_problem_results response id 조회 실패: %w", err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	result := make([]source.RespId, 0)
	for rows.Next() {
		var id source.RespId
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, rows.Err()
}

func (Table) DeleteByExamQuestionID(ctx context.Context, db sqlutil.Executor, ExamQuestionID int64) error {
	if _, err := db.ExecContext(ctx, `
		DELETE FROM llm_problem_results
		WHERE exam_question_id = ?`, ExamQuestionID); err != nil {
		return fmt.Errorf("llm_problem_results 문제별 삭제 실패: %w", err)
	}
	return nil
}

func (Table) ListResponseIDs(ctx context.Context, db sqlutil.Executor) ([]source.RespId, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT llm_response_id
		FROM llm_problem_results
		ORDER BY llm_response_id`)
	if err != nil {
		return nil, fmt.Errorf("llm_problem_results response id 조회 실패: %w", err)
	}
	defer rows.Close() //nolint:errcheck // deferred close
	result := make([]source.RespId, 0)
	for rows.Next() {
		var id source.RespId
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, rows.Err()
}

func filterWhere(filter Filter) (string, []any, error) {
	conditions := make([]string, 0, 2)
	args := make([]any, 0, 2)
	if filter.Correct != nil {
		value := 0
		if *filter.Correct {
			value = 1
		}
		conditions = append(conditions, "correct = ?")
		args = append(args, value)
	}
	if len(filter.ResponseIDs) > 0 {
		placeholders := make([]string, 0, len(filter.ResponseIDs))
		for _, id := range filter.ResponseIDs {
			placeholders = append(placeholders, "?")
			args = append(args, id)
		}
		conditions = append(conditions, "llm_response_id IN ("+strings.Join(placeholders, ",")+")")
	}
	if len(conditions) == 0 {
		return "", args, nil
	}
	return "WHERE " + strings.Join(conditions, " AND "), args, nil
}
