package generator

type OptionSpec struct {
	TypeParamsSpec string // [KeyT int | string, TT any]
	TypeParams     string // [KeyT, TT]
	Options        []OptionMeta
}

func (s OptionSpec) HasValidation() bool {
	for _, o := range s.Options {
		if o.TagOption.GoValidator != "" {
			return true
		}
	}

	return false
}

type OptionMeta struct {
	Name      string
	Docstring string // contains a comment with `//`. Can be empty or contain a multi-line string.
	Field     string
	Type      string
	TagOption TagOption
}

type TagOption struct {
	IsRequired    bool
	GoValidator   string
	Default       string
	Variadic      bool
	VariadicIsSet bool
	Skip          bool
}
