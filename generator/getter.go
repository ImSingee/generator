package generator

import (
	"bytes"
	"fmt"
	"github.com/ImSingee/god/utils"
	"github.com/spf13/viper"
)

var t = utils.GetTemplate("getter", `
// Code generated by god getter, DO NOT EDIT.

package {{ $.pkg }}

{{ $.struct.ImportedStatements }}

{{ range $_, $field := $.struct.Fields }}
{{ if $field.WillGenerateGetter }}
func ({{ $.struct.ShortName }} *{{ $.struct.Name }}) {{ $field.GetterName }}() {{ $field.Type }} {
	return {{ $.struct.ShortName }}.{{ $field.Name }}
}
{{ end }}
{{ end }}
`)

func GenerateGetter(s *utils.Struct) ([]byte, error) {
	packageName := viper.GetString("gopackage")

	w := bytes.NewBuffer(make([]byte, 0, 1024))

	err := t.Execute(w, map[string]interface{}{
		"pkg":    packageName,
		"struct": s,
	})

	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

func GenerateGetters(structs utils.Structs) (map[*utils.Struct][]byte, error) {
	results := make(map[*utils.Struct][]byte, len(structs))

	for _, s := range structs {
		result, err := GenerateGetter(s)

		if err != nil {
			return nil, fmt.Errorf("cannot generate getter for struct %s: %w", s.Name, err)
		}

		results[s] = result
	}

	return results, nil
}
