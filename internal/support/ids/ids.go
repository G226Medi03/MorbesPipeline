// Package ids defines the database row identifiers used by the pipeline.
package ids

type StatusID int64
type ActorID int64
type ActionID int64
type ProcedureID int64
type MedicalProcedureOptionID int64
type DeviceID int64
type ExamQuestionID int64
type LLMProblemResultID int64
type SolveStepID int64
type LLMCallID int64
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
