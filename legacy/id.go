package legacy

// StatusID is the canonical identifier of a status row.
type StatusID int64

type ActorID int64
type ActionID int64
type ProcedureID int64
type MedicalProcedureOptionID int64

// DeviceID is kept for compatibility with legacy device references. New
// device values are stored as actors with kind=device.
type DeviceID int64

// ExamQuestionID is the PK of an exam question row.
type ExamQuestionID int64

// LLMProblemResultID is the PK of an llm_problem_results row.
type LLMProblemResultID int64

// SolveStepID is the PK of a solve_steps row.
type SolveStepID int64

// LLMCallID is the PK of a solve_llm_calls row.
type LLMCallID int64

// StatusEntityID is the PK of a symptom, disease, or index row. The status
// type travels beside this ID in status_other_name.
type StatusEntityID int64

func (id StatusID) ToI64() int64                 { return int64(id) }
func (id ActorID) ToI64() int64                  { return int64(id) }
func (id ActionID) ToI64() int64                 { return int64(id) }
func (id ProcedureID) ToI64() int64              { return int64(id) }
func (id MedicalProcedureOptionID) ToI64() int64 { return int64(id) }
func (id DeviceID) ToI64() int64                 { return int64(id) }
func (id ExamQuestionID) ToI64() int64           { return int64(id) }
func (id LLMProblemResultID) ToI64() int64       { return int64(id) }
func (id SolveStepID) ToI64() int64              { return int64(id) }
func (id LLMCallID) ToI64() int64                { return int64(id) }
func (id StatusEntityID) ToI64() int64           { return int64(id) }
