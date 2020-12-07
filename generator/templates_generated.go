package generator

import (
	"time"

	"github.com/jessevdk/go-assets"
)

var _templatescaf48765b58cfe36de807e0b47fc0a26d8d48de0 = "package {{ .packageName }}\n\nimport (\n\t\"github.com/pkg/errors\"\n\t\"github.com/kazhuravlev/options-gen/generator/utils\"\n\t{{- range $import := .imports }}\n\t{{ $import -}}\n\t{{- end }}\n)\n\ntype optMeta struct {\n\tsetter    func(o *Options)\n\tvalidator func(o *Options) error\n}\n\n{{ range .options }}\n    func _{{ .Field }}OptValidator(o *Options) error {\n        {{ if .TagOption.IsNotEmpty -}}\n            if utils.IsNil(o.{{ .Field }}) {\n                return errors.Wrap(ErrInvalidOption, \"{{ .Name }} must be set (type {{ .Type }})\")\n            }\n        {{- end }}\n        return nil\n    }\n\n    {{ if not .TagOption.IsRequired }}\n        func With{{ .Name }}(opt {{ .Type }}) optMeta {\n             return optMeta{\n                 setter: func(o *Options) { o.{{ .Field }} = opt },\n                 validator: _{{ .Field }}OptValidator,\n             }\n        }\n    {{ end }}\n{{ end }}\n\n\nfunc NewOptions(\n    {{ range .options }}{{ if .TagOption.IsRequired -}}\n        {{ .Field }} {{ .Type }},\n    {{ end }}{{ end }}\n    options ...optMeta,\n) Options {\n    o := Options{}\n    {{ range .options }}{{ if .TagOption.IsRequired -}}\n        o.{{ .Field }} = {{ .Field }}\n    {{ end }}{{ end }}\n\n    for i := range options{\n        options[i].setter(&o)\n    }\n\n    return o\n}\n\nfunc (o *Options) Validate() error {\n    {{ range .options -}}\n        if err := _{{ .Field }}OptValidator(o); err != nil{\n            return errors.Wrap(err, \"invalid value for option With{{ .Name }}\")\n        }\n    {{ end }}\n\n    return nil\n}\n"

// templates returns go-assets FileSystem
var templates = assets.NewFileSystem(map[string][]string{"/": []string{"templates"}, "/templates": []string{"options.go.tpl"}}, map[string]*assets.File{
	"/templates": &assets.File{
		Path:     "/templates",
		FileMode: 0x800001fd,
		Mtime:    time.Unix(1607345555, 1607345555345911772),
		Data:     nil,
	}, "/templates/options.go.tpl": &assets.File{
		Path:     "/templates/options.go.tpl",
		FileMode: 0x1b4,
		Mtime:    time.Unix(1607345555, 1607345555333912007),
		Data:     []byte(_templatescaf48765b58cfe36de807e0b47fc0a26d8d48de0),
	}, "/": &assets.File{
		Path:     "/",
		FileMode: 0x800001fd,
		Mtime:    time.Unix(1607347684, 1607347684939906962),
		Data:     nil,
	}}, "")
