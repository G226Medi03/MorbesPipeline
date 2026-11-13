package legacy

import (
	"errors"
	"fmt"
	"strings"

	"github.com/G226Medi03/SourceEntity"
)

var (
	ErrSourceRequired = errors.New("source가 필요합니다")
	ErrInvalidSource  = errors.New("유효하지 않은 source입니다")
)

func ValidateSources(sources []source.SourceRef) error {
	if len(sources) == 0 {
		return ErrSourceRequired
	}
	for _, ref := range sources {
		if ref.SourceID.ToI64() <= 0 {
			return fmt.Errorf("%w: sourceId=%d", ErrInvalidSource, ref.SourceID.ToI64())
		}
	}
	return nil
}

func validateName(label, name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%s name이 비어 있습니다", label)
	}
	return nil
}

func (value Status) Validate() error {
	if strings.TrimSpace(value.Variant) == "" {
		return fmt.Errorf("status variant가 비어 있습니다")
	}
	return ValidateSources(value.Sources)
}

func (value Action) Validate() error {
	if strings.TrimSpace(value.Variant) == "" {
		return fmt.Errorf("action variant가 비어 있습니다")
	}
	return ValidateSources(value.Sources)
}

func (value Procedure) Validate() error {
	if strings.TrimSpace(value.Variant) == "" {
		return fmt.Errorf("procedure variant가 비어 있습니다")
	}
	return ValidateSources(value.Sources)
}
