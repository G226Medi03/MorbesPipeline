// Package entityjson encodes and decodes typed medical entity values that are
// not covered by the parser DTO facade.
package entityjson

import (
	"errors"
	"fmt"

	"github.com/G226Medi03/MedicalEntityParser/llmparser"
	repo "github.com/G226Medi03/MedicalEntityRepository/repo"
)

// Status errors returned by status validation.
var (
	ErrTransplantTargetMissing = errors.New("transplant target missing")
	ErrInfuseMaterialMissing   = errors.New("infuse material missing")
)

// EncodeAction encodes an action value.
func EncodeAction(value repo.IEAction) ([]byte, error) {
	return encodeAction(value)
}

// DecodeAction decodes one action value.
func DecodeAction(data []byte) (repo.IEAction, error) {
	actions, err := llmparser.DecodeActions(string(data))
	if err != nil {
		return nil, err
	}
	return singleAction(actions)
}

// EncodeActor encodes an actor value.
func EncodeActor(value repo.IEActor) ([]byte, error) {
	return encodeActor(value)
}

// DecodeActor decodes an actor value.
func DecodeActor(data []byte) (repo.IEActor, error) {
	return decodeActor(data)
}

// EncodeProcedure encodes a medical procedure value.
func EncodeProcedure(value repo.IMedicalProcedure) ([]byte, error) {
	return encodeProcedure(value)
}

// DecodeProcedure decodes a medical procedure value.
func DecodeProcedure(data []byte) (repo.IMedicalProcedure, error) {
	return decodeProcedure(data)
}

// DecodeOrParseProcedure accepts either JSON or a free-text procedure name.
func DecodeOrParseProcedure(data []byte) (repo.IMedicalProcedure, error) {
	return decodeOrParseProcedure(data)
}

// EncodeStatus encodes a status value.
func EncodeStatus(value repo.IEStatus) ([]byte, error) {
	return encodeStatus(value)
}

// DecodeStatus decodes one status value.
func DecodeStatus(data []byte) (repo.IEStatus, error) {
	statuses, err := llmparser.DecodeStatuses(string(data))
	if err != nil {
		return nil, err
	}
	return singleStatus(statuses)
}

// EncodeSymptom encodes a symptom value.
func EncodeSymptom(value repo.IESymptom) ([]byte, error) {
	return encodeSymptom(value)
}

// DecodeSymptom decodes a symptom value.
func DecodeSymptom(data []byte) (repo.IESymptom, error) {
	return decodeSymptom(data)
}

// EncodeAspect encodes a symptom aspect value.
func EncodeAspect(value repo.IPainAspect) ([]byte, error) {
	return encodeAspect(value)
}

// DecodeAspect decodes a symptom aspect value.
func DecodeAspect(data []byte) (repo.IPainAspect, error) {
	return decodeAspect(data)
}

func singleStatus(statuses map[string]repo.IEStatus) (repo.IEStatus, error) {
	if len(statuses) != 1 {
		return nil, fmt.Errorf("status response must contain exactly one status: got %d", len(statuses))
	}
	for _, status := range statuses {
		return status, nil
	}
	return nil, fmt.Errorf("status response is empty")
}

func singleAction(actions map[string]repo.IEAction) (repo.IEAction, error) {
	if len(actions) != 1 {
		return nil, fmt.Errorf("action response must contain exactly one action: got %d", len(actions))
	}
	for _, action := range actions {
		return action, nil
	}
	return nil, fmt.Errorf("action response is empty")
}
