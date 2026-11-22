package postprocess

import repo "github.com/G226Medi03/MedicalEntityRepository/repo"

func equalStatus(a, b repo.IEStatus) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Equal(b)
}

func equalAction(a, b repo.IEAction) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Equal(b)
}
