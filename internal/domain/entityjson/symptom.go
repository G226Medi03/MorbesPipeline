package entityjson

import (
	"encoding/json"
	"fmt"

	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
	"strings"
)

type symptomDocument struct {
	Kind string `json:"kind"`
	Name string `json:"name,omitempty"`
	// Code is accepted as a legacy field name for stored payloads.
	Code    string            `json:"code,omitempty"`
	Site    json.RawMessage   `json:"site,omitempty"`
	Aspects []json.RawMessage `json:"aspects,omitempty"`
}

type aspectDocument struct {
	Kind    string `json:"kind"`
	Value   string `json:"value,omitempty"`
	Measure string `json:"measure,omitempty"`
	Name    string `json:"name,omitempty"`
}

// encodeSymptom encodes a symptom interface value.
func encodeSymptom(value repo.IESymptom) ([]byte, error) {
	if value == nil {
		return nil, fmt.Errorf("symptom is nil")
	}
	switch typed := value.(type) {
	case repo.EOtherSymptom:
		return json.Marshal(symptomDocument{Kind: typed.Kind(), Name: typed.Name})
	case repo.EPain:
		site, err := encodeActor(typed.Site)
		if err != nil {
			return nil, err
		}
		aspects := make([]json.RawMessage, 0, len(typed.Aspects))
		for _, aspect := range typed.Aspects {
			encoded, err := encodeAspect(aspect)
			if err != nil {
				return nil, err
			}
			aspects = append(aspects, encoded)
		}
		return json.Marshal(symptomDocument{Kind: typed.Kind(), Site: site, Aspects: aspects})
	default:
		return nil, fmt.Errorf("지원하지 않는 symptom 타입입니다: %T", value)
	}
}

// decodeSymptom decodes a kind-tagged symptom value.
func decodeSymptom(data []byte) (repo.IESymptom, error) {
	var raw symptomDocument
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("symptom JSON 파싱 실패: %w", err)
	}
	switch repo.SymptomKind(canonicalSymptomKind(raw.Kind)) {
	case repo.SymptomKindPain:
		var site repo.IEActor = repo.NewUnknownActor("")
		if len(raw.Site) > 0 {
			decoded, err := decodeActor(raw.Site)
			if err != nil {
				return nil, err
			}
			site = decoded
		}
		aspects := make([]repo.IPainAspect, 0, len(raw.Aspects))
		for _, encoded := range raw.Aspects {
			aspect, err := decodeAspect(encoded)
			if err != nil {
				return nil, err
			}
			aspects = append(aspects, aspect)
		}
		return repo.EPain{Site: site, Aspects: aspects}, nil
	default:
		return repo.EOtherSymptom{Name: firstNonEmptySymptom(raw.Name, raw.Code)}, nil
	}
}

// encodeAspect encodes an aspect interface value.
func encodeAspect(value repo.IPainAspect) ([]byte, error) {
	if value == nil {
		return nil, fmt.Errorf("aspect is nil")
	}
	switch typed := value.(type) {
	case repo.SymptomDuration:
		return json.Marshal(aspectDocument{Kind: typed.Kind(), Value: typed.Value, Measure: typed.Measure})
	case repo.Intensity:
		return json.Marshal(aspectDocument{Kind: typed.Kind(), Value: typed.Value, Measure: typed.Measure})
	case repo.Quality:
		return json.Marshal(aspectDocument{Kind: typed.Kind(), Name: typed.Name})
	case repo.Radiation:
		return json.Marshal(aspectDocument{Kind: typed.Kind(), Name: typed.Name})
	default:
		return nil, fmt.Errorf("지원하지 않는 aspect 타입입니다: %T", value)
	}
}

// decodeAspect decodes a kind-tagged aspect value.
func decodeAspect(data []byte) (repo.IPainAspect, error) {
	var raw aspectDocument
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("aspect JSON 파싱 실패: %w", err)
	}
	switch repo.AspectKind(strings.ToLower(strings.TrimSpace(raw.Kind))) {
	case repo.AspectKindDuration:
		return repo.SymptomDuration{Value: raw.Value, Measure: raw.Measure}, nil
	case repo.AspectKindIntensity:
		return repo.Intensity{Value: raw.Value, Measure: raw.Measure}, nil
	case repo.AspectKindQuality:
		return repo.Quality{Name: raw.Name}, nil
	case repo.AspectKindRadiation:
		return repo.Radiation{Name: raw.Name}, nil
	default:
		return repo.Quality{Name: raw.Name}, nil
	}
}

// canonicalSymptomKind normalizes symptom kind aliases.
func canonicalSymptomKind(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "pain":
		return "pain"
	case "other_symptom", "symptom_code", "code", "symptom":
		return "other_symptom"
	default:
		return "other_symptom"
	}
}

func firstNonEmptySymptom(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
