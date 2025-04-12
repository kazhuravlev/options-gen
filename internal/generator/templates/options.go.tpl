// Code generated by options-gen. DO NOT EDIT.

package {{ .packageName }}{{$hasGoValidator := false}}{{ range .options }}{{- if .TagOption.GoValidator }}{{$hasGoValidator = true}}{{break}}{{end}}{{end}}

import (
	{{if $hasGoValidator}}fmt461e464ebed9 "fmt"
	{{ if .hasValidation }}errors461e464ebed9 "github.com/kazhuravlev/options-gen/pkg/errors"
{{end}}{{end}}{{ if .hasValidation }}validator461e464ebed9 "github.com/kazhuravlev/options-gen/pkg/validator"{{ end }}
	{{- range $import := .imports }}
	{{ $import -}}
	{{- end }}
)

{{ if .withIsset }}
type opt{{$.optionsPrefix}}Field int8
const(
	{{ range $i, $field := .options }}
		Field{{$.optionsPrefix}}{{ $field.Field }} opt{{$.optionsPrefix}}Field = {{ $i }}
	{{- end -}}
)

var opt{{$.optionsPrefix}}IsSet = [{{ .optionsLen }}]bool{}
{{ end }}

type Opt{{ $.optionsStructName }}Setter{{ $.optionsTypeParamsSpec }} func(o *{{ .optionsStructInstanceType }})

{{if ne .constructorTypeRender "no" }}
func {{if eq .constructorTypeRender "public" }}New{{else}}new{{end}}{{ .optionsStructType }}(
	{{ range .options -}}
		{{ if .TagOption.IsRequired -}}
			{{ .Field }} {{ .Type }},
		{{ end }}
	{{- end -}}
	options ...Opt{{ $.optionsStructName }}Setter{{ $.optionsTypeParams }},
) {{ .optionsStructInstanceType }} {
	o := {{ .optionsStructInstanceType }}{}
	{{ if .withIsset }}
		var empty [{{ .optionsLen }}]bool
		opt{{$.optionsPrefix}}IsSet = empty
	{{ end }}

	{{ if .defaultsVarName }}
		// Setting defaults from variable
		{{ range .options -}}
			o.{{ .Field }} = {{ $.defaultsVarName }}.{{ .Field }}
      {{ if $.withIsset -}}
				opt{{$.optionsPrefix}}IsSet[Field{{$.optionsPrefix}}{{ .Field }}] = true
      {{- end }}
    {{ end }}
	{{ end }}

	{{ if .defaultsFuncName }}
		// Setting defaults from func
		defaultOpts := {{ $.defaultsFuncName }}{{ $.optionsTypeParams }}()
		{{ range .options -}}
			o.{{ .Field }} = defaultOpts.{{ .Field }}
      {{ if $.withIsset -}}
				opt{{$.optionsPrefix}}IsSet[Field{{$.optionsPrefix}}{{ .Field }}] = true
      {{- end }}
    {{ end }}
	{{ end }}

	{{ if .defaultsTagName }}
		// Setting defaults from field tag (if present)
        {{ range .options -}}
            {{ if .TagOption.Default -}}
                {{ if eq .Type "time.Duration" }}o.{{ .Field }}, _ = time.ParseDuration("{{ .TagOption.Default }}")
                {{- else if eq .Type "string" }}o.{{ .Field }} = "{{ .TagOption.Default }}"
                {{- else }}o.{{ .Field }} = {{ .TagOption.Default }}{{ end }}
                {{ if $.withIsset -}}
	                opt{{$.optionsPrefix}}IsSet[Field{{$.optionsPrefix}}{{ .Field }}] = true
                {{- end }}
            {{ end -}}
        {{ end }}
	{{ end }}

	{{ range .options }}
	    {{- if .TagOption.IsRequired -}}
	        o.{{ .Field }} = {{ .Field }}
          {{ if $.withIsset -}}
		        opt{{$.optionsPrefix}}IsSet[Field{{$.optionsPrefix}}{{ .Field }}] = true
          {{- end }}
      {{ end -}}
	{{ end }}

	for _, opt := range options {
		opt(&o)
	}
	return o
}
{{end}}

{{ range .options }}
	{{ if not .TagOption.IsRequired }}
		{{- if ne .Docstring "" -}}
			{{ .Docstring }}
		{{- end }}
		func With{{$.optionsPrefix}}{{ .Name }}{{ $.optionsTypeParamsSpec }}(opt {{if .TagOption.Variadic}}...{{end}}{{ .Type }}) Opt{{ $.optionsStructName }}Setter{{ $.optionsTypeParams }} {
			return func(o *{{ $.optionsStructInstanceType }}) {
				{{if .TagOption.Variadic}}o.{{ .Field }} = append(o.{{ .Field }}, opt...){{else}}o.{{ .Field }} = opt{{end}}
				{{ if $.withIsset -}}
				opt{{$.optionsPrefix}}IsSet[Field{{$.optionsPrefix}}{{ .Field }}] = true{{- end }}
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
				errs.Add(errors461e464ebed9.NewValidationError("{{ .Field }}", _validate_{{ $.optionsStructName }}_{{ .Field }}{{ $.optionsTypeParams }}(o)))
			{{- end }}
		{{- end }}
		return errs.AsError()
	{{- end }}
}

{{ if .withIsset }}
	func (o *{{ .optionsStructInstanceType }}) IsSet(field opt{{$.optionsPrefix}}Field) bool {
	return opt{{$.optionsPrefix}}IsSet[field]
	}
{{ end }}

{{ range .options }}
	{{- if .TagOption.GoValidator }}
		func _validate_{{ $.optionsStructName }}_{{ .Field }}{{ $.optionsTypeParamsSpec }}(o *{{ $.optionsStructInstanceType }}) error {
			if err := validator461e464ebed9.GetValidatorFor(o).Var(o.{{ .Field }}, "{{ .TagOption.GoValidator }}"); err != nil {
				return fmt461e464ebed9.Errorf("field `{{ .Field }}` did not pass the test: %w", err)
			}
			return nil
		}
	{{- end }}
{{ end }}
