package store

import "github.com/G226Medi03/MedicalEntityRepository/repo"

// Entities exposes the name-based entity repository owned by the store.
func (s *StoreContainer) Entities() (*repo.Store, error) {
	return s.entities()
}
