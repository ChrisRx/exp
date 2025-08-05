package assert

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"text/template"
)

type Message struct {
	Header   string
	Expected any
	Actual   any
	Elements []any
	Diff     []byte
}

func Fatal(tb testing.TB, msg Message) {
	var b bytes.Buffer
	if err := failureMessage.Execute(&b, msg); err != nil {
		panic(err)
	}
	tb.Helper()
	tb.Fatal(b.String())
}

var failureMessage = template.Must(template.New("").Funcs(map[string]any{
	"indent": func(spaces int, v any) string {
		pad := strings.Repeat(" ", spaces)
		return pad + strings.ReplaceAll(fmt.Sprint(v), "\n", "\n"+pad)
	},
	"string": func(v []byte) string {
		return string(v)
	},
}).Parse(`
{{- .Header }}
{{- with .Expected }}
expected:
{{ . | indent 4 }}
{{- end -}}
{{- with .Actual }}
actual:
{{ . | indent 4 }}
{{- end -}}
{{- range .Elements }}
	{{ . }}
{{- end -}}
{{- with .Diff }}
{{ . | string | indent 4 }}
{{- end -}}
`))
