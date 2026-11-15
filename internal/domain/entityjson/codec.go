// Package entityjson encodes and decodes typed medical entity values.
package entityjson

import (
	"errors"
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

// DecodeAction decodes an action value.
func DecodeAction(data []byte) (repo.IEAction, error) {
	return decodeAction(data)
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

// DecodeStatus decodes a status value.
func DecodeStatus(data []byte) (repo.IEStatus, error) {
	return decodeStatus(data)
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
