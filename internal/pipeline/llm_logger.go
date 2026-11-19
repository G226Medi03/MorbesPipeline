package csolve

import (
	"fmt"

	clog "github.com/G226Medi03/log-sqlite"
)

type llmLogger struct{}

func (llmLogger) Info(msg string, attrs ...any) {
	clog.Info(msg, llmLogAttrs(attrs)...)
}

func (llmLogger) Error(msg string, attrs ...any) {
	clog.Error(msg, llmLogAttrs(attrs)...)
}

func llmLogAttrs(attrs []any) []clog.Attr {
	if len(attrs) == 0 {
		return nil
	}
	result := make([]clog.Attr, 0, (len(attrs)+1)/2)
	for index := 0; index < len(attrs); index += 2 {
		key := fmt.Sprint(attrs[index])
		var value any
		if index+1 < len(attrs) {
			value = attrs[index+1]
		}
		result = append(result, clog.AttrOf(key, value))
	}
	return result
}
