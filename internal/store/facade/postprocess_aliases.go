package facade

import postprocessrepo "github.com/G226Medi03/MorbesPipeline/internal/store/internal/repository/postprocess"

type (
	Operation             = postprocessrepo.Operation
	EntityType            = postprocessrepo.EntityType
	PostprocessOperation  = postprocessrepo.Operation
	PostprocessEntityType = postprocessrepo.EntityType
)

const (
	PostprocessorTransplantRejection  = postprocessrepo.PostprocessorTransplantRejection
	PostprocessorIncreasedIndexAction = postprocessrepo.PostprocessorIncreasedIndexAction
	PostprocessorCauseStatusResult    = postprocessrepo.PostprocessorCauseStatusResult

	OperationObserved    = postprocessrepo.OperationObserved
	OperationCreated     = postprocessrepo.OperationCreated
	OperationUpdated     = postprocessrepo.OperationUpdated
	OperationRemoved     = postprocessrepo.OperationRemoved
	OperationLinkAdded   = postprocessrepo.OperationLinkAdded
	OperationLinkRemoved = postprocessrepo.OperationLinkRemoved

	PostprocessOperationObserved    = postprocessrepo.OperationObserved
	PostprocessOperationCreated     = postprocessrepo.OperationCreated
	PostprocessOperationUpdated     = postprocessrepo.OperationUpdated
	PostprocessOperationRemoved     = postprocessrepo.OperationRemoved
	PostprocessOperationLinkAdded   = postprocessrepo.OperationLinkAdded
	PostprocessOperationLinkRemoved = postprocessrepo.OperationLinkRemoved

	EntityStatus    = postprocessrepo.EntityStatus
	EntityAction    = postprocessrepo.EntityAction
	EntityActor     = postprocessrepo.EntityActor
	EntityProcedure = postprocessrepo.EntityProcedure
	EntityOption    = postprocessrepo.EntityOption

	PostprocessEntityStatus    = postprocessrepo.EntityStatus
	PostprocessEntityAction    = postprocessrepo.EntityAction
	PostprocessEntityActor     = postprocessrepo.EntityActor
	PostprocessEntityProcedure = postprocessrepo.EntityProcedure
	PostprocessEntityOption    = postprocessrepo.EntityOption
)
