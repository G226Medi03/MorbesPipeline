package store

import corestore "github.com/G226Medi03/MorbesPipeline/internal/repository/store"

type MedicalEntityService struct {
	store *corestore.Store
}

func NewMedicalEntity(store *corestore.Store) *MedicalEntityService {
	return &MedicalEntityService{store: store}
}

type SolveService struct {
	store *corestore.Store
}

func NewSolve(store *corestore.Store) *SolveService {
	return &SolveService{store: store}
}

type SourceService struct {
	store *corestore.Store
}

func NewSource(store *corestore.Store) *SourceService {
	return &SourceService{store: store}
}

type BenchmarkService struct {
	store *corestore.Store
}

func NewBenchmark(store *corestore.Store) *BenchmarkService {
	return &BenchmarkService{store: store}
}

type PostprocessService struct {
	store *corestore.Store
}

func NewPostprocess(store *corestore.Store) *PostprocessService {
	return &PostprocessService{store: store}
}
