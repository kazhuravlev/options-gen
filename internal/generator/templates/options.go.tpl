// Code generated by options-gen. DO NOT EDIT.
package {{ .packageName }}

import (
    "fmt"
	errors461e464ebed9 "github.com/kazhuravlev/options-gen/pkg/errors"
	goplvalidator "github.com/go-playground/validator/v10"
	{{- range $import := .imports }}
	{{ $import -}}
	{{- end }}
)

{{ if .hasValidation }}
var _validator461e464ebed9 = goplvalidator.New()
{{- end }}

type opt{{ $.optionsStructName }}Setter{{ $.optionsTypeParamsSpec }} func(o *{{ .optionsStructInstanceType }})

func New{{ .optionsStructType }}(
	{{ range .options -}}
		{{ if .TagOption.IsRequired -}}
			{{ .Field }} {{ .Type }},
		{{ end }}
	{{- end -}}
	options ...opt{{ $.optionsStructName }}Setter{{ $.optionsTypeParams }},
) {{ .optionsStructInstanceType }} {
	o := {{ .optionsStructInstanceType }}{}
	{{ range .options }}{{ if .TagOption.IsRequired -}}
		o.{{ .Field }} = {{ .Field }}
	{{ end }}{{ end }}

	for _, opt := range options {
		opt(&o)
	}
	return o
}

{{ range .options }}
	{{ if not .TagOption.IsRequired }}
		func With{{ .Name }}{{ $.optionsTypeParamsSpec }}(opt {{ .Type }}) opt{{ $.optionsStructName }}Setter{{ $.optionsTypeParams }} {
			return func(o *{{ $.optionsStructInstanceType }}) {
				o.{{ .Field }} = opt
			}
		}
	{{ end }}
{{ end }}


func (o *{{ .optionsStructInstanceType }}) Validate() error {
	{{- if not .hasValidation -}}
		return nil
	{{- else }}
		errs := new(errors461e464ebed9.ValidationErrors)
		{{- range .options }}
			{{- if .TagOption.GoValidator }}
				errs.Add(errors461e464ebed9.NewValidationError("{{ .Name }}", _validate_{{ $.optionsStructName }}_{{ .Field }}{{ $.optionsTypeParams }}(o)))
			{{- end }}
		{{- end }}
		return errs.AsError()
	{{- end }}
}

{{ range .options }}
	{{- if .TagOption.GoValidator }}
		func _validate_{{ $.optionsStructName }}_{{ .Field }}{{ $.optionsTypeParamsSpec }}(o *{{ $.optionsStructInstanceType }}) error {
			if err := _validator461e464ebed9.Var(o.{{ .Field }}, "{{ .TagOption.GoValidator }}"); err != nil {
				return fmt.Errorf("field `{{ .Field }}` did not pass the test: %w", err)
			}
			return nil
		}
	{{- end }}
{{ end }}
