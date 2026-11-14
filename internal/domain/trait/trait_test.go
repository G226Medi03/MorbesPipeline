package trait_test

import (
	"testing"

	"github.com/G226Medi03/MorbesPipeline/internal/domain/trait"
)

func TestNormalizeTraits(t *testing.T) {
	values := trait.NormalizeTraits([]string{
		"급성",
		"INVASIVE",
		"non-invasive",
		"acute",
		"unknown",
	})
	want := []trait.Trait{
		trait.TraitAcute,
		trait.TraitInvasive,
		trait.TraitNonInvasive,
	}
	if len(values) != len(want) {
		t.Fatalf("values = %#v, want %#v", values, want)
	}
	for index := range want {
		if values[index] != want[index] {
			t.Fatalf("values[%d] = %q, want %q", index, values[index], want[index])
		}
	}
}

func TestPriorityTraitsKeepCanonicalOrder(t *testing.T) {
	values := trait.PriorityTraits([]trait.Trait{
		trait.TraitRescue,
		trait.TraitDefinitive,
		trait.TraitAcute,
		trait.TraitInvasive,
	})
	want := []trait.Trait{
		trait.TraitAcute,
		trait.TraitDefinitive,
		trait.TraitRescue,
	}
	if len(values) != len(want) {
		t.Fatalf("values = %#v, want %#v", values, want)
	}
	for index := range want {
		if values[index] != want[index] {
			t.Fatalf("values[%d] = %q, want %q", index, values[index], want[index])
		}
	}
}

func TestComparePriorityUsesPresenceOrder(t *testing.T) {
	acute_and_definitive := []trait.Trait{trait.TraitAcute, trait.TraitDefinitive}
	acute := []trait.Trait{trait.TraitAcute}
	definitive := []trait.Trait{trait.TraitDefinitive}
	none := []trait.Trait{trait.TraitInvasive}

	if cmp := trait.ComparePriority(acute_and_definitive, acute); cmp >= 0 {
		t.Fatalf("acute+definitive vs acute = %d, want -1", cmp)
	}
	if cmp := trait.ComparePriority(acute, definitive); cmp >= 0 {
		t.Fatalf("acute vs definitive = %d, want -1", cmp)
	}
	if cmp := trait.ComparePriority(definitive, none); cmp >= 0 {
		t.Fatalf("definitive vs none = %d, want -1", cmp)
	}
	if cmp := trait.ComparePriority(none, definitive); cmp <= 0 {
		t.Fatalf("none vs definitive = %d, want 1", cmp)
	}
}

func TestTraitScopes(t *testing.T) {
	if !trait.IsPriorityTrait(trait.TraitAcute) {
		t.Fatal("acute should be a priority trait")
	}
	if trait.IsPriorityTrait(trait.TraitInvasive) {
		t.Fatal("invasive should not be a priority trait")
	}
	if !trait.IsContextTrait(trait.TraitAcute) {
		t.Fatal("acute should be a context trait")
	}
	if !trait.IsIntrinsicTrait(trait.TraitInvasive) {
		t.Fatal("invasive should be an intrinsic trait")
	}
}

func TestDefaultIntrinsicTraits(t *testing.T) {
	values := trait.DefaultIntrinsicTraits(trait.KindSurgery)
	if !trait.ContainsTrait(values, trait.TraitSurgical) ||
		!trait.ContainsTrait(values, trait.TraitInvasive) ||
		!trait.ContainsTrait(values, trait.TraitDefinitive) {
		t.Fatalf("surgery traits = %#v", values)
	}
	if values := trait.DefaultIntrinsicTraits(trait.KindUnknown); len(values) != 0 {
		t.Fatalf("unknown traits = %#v", values)
	}
}
