package graph

import (
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	option "github.com/G226Medi03/MorbesPipeline/store/medicalprocedureoption"
	source "github.com/G226Medi03/SourceEntity"
)

type ActorRow struct {
	ID   entity.ActorID
	Kind string
	Name string
}

type StatusRow struct {
	ID   entity.StatusID
	Kind string
	Ptr  int64
}

type ActionRow struct {
	ID   entity.ActionID
	Kind string
	Ptr  int64
}

type ProcedureRow struct {
	ID         entity.ProcedureID
	Kind       string
	Ptr        int64
	TraitsJSON string
}

type OptionRow struct {
	ID                      option.MedicalProcedureOptionID
	ProcedureID             entity.ProcedureID
	Category                string
	TraitsJSON              string
	ContextTraitsJSON       string
	IndicationStatusIDs     []entity.StatusID
	ExpectedEffectStatusIDs []entity.StatusID
}

type StatusActionRow struct {
	StatusID entity.StatusID
	ActionID entity.ActionID
	SourceID source.SourceId
}

type ActionResultStatusRow struct {
	ActionID entity.ActionID
	StatusID entity.StatusID
	SourceID source.SourceId
}

type ActorActionRow struct {
	ActorID  entity.ActorID
	ActionID entity.ActionID
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
