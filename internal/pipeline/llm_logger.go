package csolve

import (
	"fmt"

	"github.com/G226Medi03/MorbesPipeline/internal/support/logs"
)

type llmLogger struct{}

func (llmLogger) Info(msg string, attrs ...any) {
	logs.Info(msg, llmLogAttrs(attrs)...)
}

func (llmLogger) Error(msg string, attrs ...any) {
	logs.Error(msg, llmLogAttrs(attrs)...)
}

func llmLogAttrs(attrs []any) []logs.Attr {
	if len(attrs) == 0 {
		return nil
	}
	result := make([]logs.Attr, 0, (len(attrs)+1)/2)
	for index := 0; index < len(attrs); index += 2 {
		key := fmt.Sprint(attrs[index])
		var value any
		if index+1 < len(attrs) {
			value = attrs[index+1]
		}
		result = append(result, logs.AttrOf(key, value))
	}
	return result
}
