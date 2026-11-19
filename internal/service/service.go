package service

import corestore "github.com/G226Medi03/MorbesPipeline/internal/store"

type EntityService struct {
	store *corestore.Store
}

func NewEntity(store *corestore.Store) *EntityService {
	return &EntityService{store: store}
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
	store  *corestore.Store
	entity *EntityService
}

func NewPostprocess(store *corestore.Store, entity *EntityService) *PostprocessService {
	return &PostprocessService{store: store, entity: entity}
}
