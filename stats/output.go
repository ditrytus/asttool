package stats

import (
	"fmt"
	"strings"
)

func FormatOutputStats(v Visitor) string {
	s := v.Stats()
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Funcs: %d\n", s.FuncCount))
	b.WriteString(fmt.Sprintf("Types: %d\n", s.TypeCount))
	b.WriteString(fmt.Sprintf("Consts: %d\n", s.ConstCount))
	b.WriteString(fmt.Sprintf("Vars: %d\n", s.VarCount))
	return b.String()
}
