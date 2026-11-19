package llm

import (
	"fmt"

	clog "github.com/G226Medi03/log-sqlite"
)

type Logger struct{}

func (Logger) Info(msg string, attrs ...any)  { clog.Info(msg, logAttrs(attrs)...) }
func (Logger) Error(msg string, attrs ...any) { clog.Error(msg, logAttrs(attrs)...) }

func logAttrs(attrs []any) []clog.Attr {
	if len(attrs) == 0 {
		return nil
	}
	result := make([]clog.Attr, 0, (len(attrs)+1)/2)
	for index := 0; index < len(attrs); index += 2 {
		value := any(nil)
		if index+1 < len(attrs) {
			value = attrs[index+1]
		}
		result = append(result, clog.AttrOf(fmt.Sprint(attrs[index]), value))
	}
	return result
}
