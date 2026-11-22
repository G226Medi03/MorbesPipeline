package resolve

import (
	"strings"

	"github.com/G226Medi03/MedicalEntityParser/llmdto"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	store "github.com/G226Medi03/MorbesPipeline/internal/repository/facade"
	source "github.com/G226Medi03/SourceEntity"
)

func Actor(draft llmdto.ActorItem) (repo.IEActor, error) {
	name := strings.TrimSpace(draft.Name)
	if name == "" {
		return nil, nil
	}
	if repo.ParseActorKind(string(draft.Kind)) == repo.ActorKindDevice {
		return repo.NewDevice(name), nil
	}
	return repo.NewUnknownActor(name), nil
}

func EnsureInfusionOption(st *store.StoreContainer, name string, sources []source.Source) error {
	name = strings.TrimSpace(name)
	if name == "" || len(sources) == 0 {
		return nil
	}
	procedure := repo.Infuse{Material: repo.NewUnknownActor(name), Route: repo.InfuseRouteUnknown}
	if err := st.SaveProcedure(procedure, nil, sources); err != nil {
		return err
	}
	_, err := st.SaveMedicalProcedureOptionByValues(procedure, "treatment", nil, nil, nil, nil, sources)
	return err
}
