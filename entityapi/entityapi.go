// Package entityapi exposes the MedicalEntity types used by application code.
//
// MorbesGoDB consumes this package instead of importing MedicalEntity
// directly. Persistence repositories remain inside MorbesPipeline/store.
package entityapi

import (
	medicalcodec "github.com/G226Medi03/MedicalEntity/codec"
	"github.com/G226Medi03/MedicalEntity/ename/eaction"
	"github.com/G226Medi03/MedicalEntity/ename/eactor"
	"github.com/G226Medi03/MedicalEntity/ename/edisease"
	"github.com/G226Medi03/MedicalEntity/ename/eindex"
	"github.com/G226Medi03/MedicalEntity/ename/emedical_procedure"
	"github.com/G226Medi03/MedicalEntity/ename/estatus"
	"github.com/G226Medi03/MedicalEntity/ename/esymptom"
	medicalentity "github.com/G226Medi03/MedicalEntity/entity"
)

// Aggregate model and repository aliases.
type (
	Executor = medicalentity.Executor

	Action              = medicalentity.Action
	ActionID            = medicalentity.ActionID
	ActionRepository    = medicalentity.ActionRepository
	Actor               = medicalentity.Actor
	ActorID             = medicalentity.ActorID
	ActorRepository     = medicalentity.ActorRepository
	Disease             = medicalentity.Disease
	DiseaseID           = medicalentity.DiseaseID
	DiseaseRepository   = medicalentity.DiseaseRepository
	Index               = medicalentity.Index
	IndexID             = medicalentity.IndexID
	IndexRepository     = medicalentity.IndexRepository
	Procedure           = medicalentity.Procedure
	ProcedureID         = medicalentity.ProcedureID
	ProcedureRepository = medicalentity.ProcedureRepository
	Status              = medicalentity.Status
	StatusID            = medicalentity.StatusID
	StatusRepository    = medicalentity.StatusRepository
	Symptom             = medicalentity.Symptom
	SymptomID           = medicalentity.SymptomID
	SymptomRepository   = medicalentity.SymptomRepository
)

// Typed ename aliases.
type (
	IEAction          = eaction.IEAction
	IEActor           = eactor.IEActor
	IEStatus          = estatus.IEStatus
	IMedicalProcedure = emedical_procedure.IMedicalProcedure
	IESymptom         = esymptom.IESymptom
	IAspect           = esymptom.IAspect

	Block                   = eaction.Block
	CauseStatus             = eaction.CauseStatus
	Relax                   = eaction.Relax
	Device                  = eactor.Device
	ActorProcedure          = eactor.Procedure
	UnknownActor            = eactor.UnknownActor
	EDisease                = edisease.EDisease
	EIndex                  = eindex.EIndex
	Cleaning                = emedical_procedure.Cleaning
	DiagnosticTest          = emedical_procedure.DiagnosticTest
	Education               = emedical_procedure.Education
	Infuse                  = emedical_procedure.Infuse
	Insertion               = emedical_procedure.Insertion
	Instruction             = emedical_procedure.Instruction
	Surgery                 = emedical_procedure.Surgery
	Transplant              = emedical_procedure.Transplant
	UnknownMedicalProcedure = emedical_procedure.UnknownMedicalProcedure
	Damaged                 = estatus.Damaged
	DecreasedIndex          = estatus.DecreasedIndex
	HaveDisease             = estatus.HaveDisease
	HaveSymptom             = estatus.HaveSymptom
	IncreasedIndex          = estatus.IncreasedIndex
	Inflammation            = estatus.Inflammation
	ReceivedProcedure       = estatus.ReceivedMedicalProcedure
	EOtherSymptom           = esymptom.EOtherSymptom
	EPain                   = esymptom.EPain
)

// Shared values.
var ErrNotFound = medicalentity.ErrNotFound

// Constructors and schema helpers.
var (
	InitSchema             = medicalentity.InitSchema
	NewAction              = medicalentity.NewAction
	NewActionRepository    = medicalentity.NewActionRepository
	NewActorRepository     = medicalentity.NewActorRepository
	NewDiseaseRepository   = medicalentity.NewDiseaseRepository
	NewIndexRepository     = medicalentity.NewIndexRepository
	NewProcedureRepository = medicalentity.NewProcedureRepository
	NewStatus              = medicalentity.NewStatus
	NewStatusRepository    = medicalentity.NewStatusRepository
	NewSymptomRepository   = medicalentity.NewSymptomRepository
)

// Codec helpers.
var (
	MarshalAction    = medicalcodec.MarshalAction
	MarshalActor     = medicalcodec.MarshalActor
	MarshalProcedure = medicalcodec.MarshalProcedure
	MarshalStatus    = medicalcodec.MarshalStatus
	MarshalSymptom   = medicalcodec.MarshalSymptom

	UnmarshalAction           = medicalcodec.UnmarshalAction
	UnmarshalOrParseProcedure = medicalcodec.UnmarshalOrParseProcedure
	UnmarshalProcedure        = medicalcodec.UnmarshalProcedure
	UnmarshalStatus           = medicalcodec.UnmarshalStatus
	UnmarshalSymptom          = medicalcodec.UnmarshalSymptom
)
