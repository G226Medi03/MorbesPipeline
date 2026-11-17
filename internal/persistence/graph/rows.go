package graph

import (
	option "github.com/G226Medi03/MorbesPipeline/internal/domain/procedureoption"
	"github.com/G226Medi03/MorbesPipeline/internal/support/ids"
	source "github.com/G226Medi03/SourceEntity"
)

type ActorRow struct {
	ID   ids.ActorID
	Kind string
	Name string
}

type StatusRow struct {
	ID   ids.StatusID
	Kind string
	Ptr  int64
}

type ActionRow struct {
	ID   ids.ActionID
	Kind string
	Ptr  int64
}

type ProcedureRow struct {
	ID         ids.ProcedureID
	Kind       string
	Ptr        int64
	TraitsJSON string
}

type OptionRow struct {
	ID                      option.MedicalProcedureOptionID
	ProcedureID             ids.ProcedureID
	Category                string
	TraitsJSON              string
	ContextTraitsJSON       string
	IndicationStatusIDs     []ids.StatusID
	ExpectedEffectStatusIDs []ids.StatusID
}

type StatusActionRow struct {
	StatusID ids.StatusID
	ActionID ids.ActionID
	SourceID source.SourceId
}

type ActionResultStatusRow struct {
	ActionID ids.ActionID
	StatusID ids.StatusID
	SourceID source.SourceId
}

type ActorActionRow struct {
	ActorID  ids.ActorID
	ActionID ids.ActionID
	SourceID source.SourceId
}

type SourceRow struct {
	EntityID int64
	SourceID source.SourceId
}

type Rows struct {
	Actors     []ActorRow
	Statuses   []StatusRow
	Actions    []ActionRow
	Procedures []ProcedureRow
	Options    []OptionRow

	StatusActions        []StatusActionRow
	ActionResultStatuses []ActionResultStatusRow
	ActorActions         []ActorActionRow

	StatusSources    []SourceRow
	ActionSources    []SourceRow
	ActorSources     []SourceRow
	ProcedureSources []SourceRow
	OptionSources    []SourceRow
}
