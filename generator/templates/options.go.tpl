package {{ .packageName }}

import (
	"github.com/pkg/errors"
	"github.com/kazhuravlev/options-gen/generator/utils"
)

type optMeta struct {
	setter    func(o *Options)
	validator func(o *Options) error
}

{{ range .options }}
    func _{{ .Field }}OptValidator(o *Options) error {
        {{ if .TagOption.IsNotEmpty -}}
            if utils.IsNil(o.{{ .Field }}) {
                return errors.Wrap(ErrInvalidOption, "{{ .Name }} must be set (type {{ .Type }})")
            }
        {{- end }}
        return nil
    }

    {{ if not .TagOption.IsRequired }}
        func With{{ .Name }}(opt {{ .Type }}) optMeta {
             return optMeta{
                 setter: func(o *Options) { o.{{ .Field }} = opt },
                 validator: _{{ .Field }}OptValidator,
             }
        }
    {{ end }}
{{ end }}


func NewOptions(
    {{ range .options }}{{ if .TagOption.IsRequired -}}
        {{ .Field }} {{ .Type }},
    {{ end }}{{ end }}
    options ...optMeta,
) Options {
    o := Options{}
    {{ range .options }}{{ if .TagOption.IsRequired -}}
        o.{{ .Field }} = {{ .Field }}
    {{ end }}{{ end }}

    for i := range options{
        options[i].setter(&o)
    }

    return o
}

func (o *Options) Validate() error {
    {{ range .options -}}
        if err := _{{ .Field }}OptValidator(o); err != nil{
            return errors.Wrap(err, "invalid value for option With{{ .Name }}")
        }
    {{ end }}

    return nil
}
