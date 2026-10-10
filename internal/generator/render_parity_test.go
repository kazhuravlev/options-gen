//nolint:testpackage
package generator

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// The tests in this file pin formatGeneratedSource, the single-pass formatting
// of the rendered template, to optimizeGeneratedSource: the previous pipeline
// that formats the source four times and is kept as the reference.

const (
	parityTestdataDir  = "../../options-gen/testdata"
	parityOptionsType  = "OptOptionsSetter"
	parityDefaultsNone = "none"
	parityDefaultsTag  = "tag"
	parityDefaultsVar  = "var"
	parityDefaultsFunc = "func"
)

// requireRenderParity renders opts through both pipelines, checks that they
// agree byte for byte and returns the rendered file.
func requireRenderParity(t *testing.T, opts Options) []byte {
	t.Helper()

	raw, err := renderTemplate(opts)
	require.NoError(t, err)

	want, err := optimizeGeneratedSource(raw)
	require.NoError(t, err)

	got, err := formatGeneratedSource(raw)
	require.NoError(t, err)
	require.Equal(t, string(want), string(got))

	rendered, err := Render(opts)
	require.NoError(t, err)
	require.Equal(t, string(want), string(rendered))

	return rendered
}

// testdataParams mirrors the .params.json files of options-gen/testdata.
type testdataParams struct {
	OutPrefix string `json:"out_prefix"` //nolint:tagliatelle
	Defaults  struct {
		From  string `json:"from"`
		Param string `json:"param"`
	} `json:"defaults"`
	Constructor    string `json:"constructor"`
	WithIsset      bool   `json:"with_isset"`       //nolint:tagliatelle
	AllVariadic    bool   `json:"all_variadic"`     //nolint:tagliatelle
	OptionTypeName string `json:"option_type_name"` //nolint:tagliatelle
}

// testdataRenderOptions builds the Render options for a testdata case the way
// options-gen.Run does.
func testdataRenderOptions(t *testing.T, dir string) Options {
	t.Helper()

	var params testdataParams
	params.Defaults.From = parityDefaultsTag
	params.Constructor = "public"

	if data, err := os.ReadFile(filepath.Join(dir, ".params.json")); err == nil {
		require.NoError(t, json.Unmarshal(data, &params))
	} else {
		require.ErrorIs(t, err, os.ErrNotExist)
	}

	var tagName, varName, funcName string
	switch params.Defaults.From {
	case parityDefaultsTag:
		tagName = cmp.Or(params.Defaults.Param, "default")
	case parityDefaultsVar:
		varName = cmp.Or(params.Defaults.Param, "defaultOptions")
	case parityDefaultsFunc:
		funcName = cmp.Or(params.Defaults.Param, "getDefaultOptions")
	}

	spec, err := GetOptionSpec(filepath.Join(dir, "options.go"), "Options", tagName, params.AllVariadic, nil)
	require.NoError(t, err)

	return NewOptions(
		WithVersion("qa-version"),
		WithPackageName("testcase"),
		WithOptionsStructName("Options"),
		WithFileImports(spec.Imports),
		WithSpec(&spec.Spec),
		WithTagName(tagName),
		WithVarName(varName),
		WithFuncName(funcName),
		WithPrefix(params.OutPrefix),
		WithWithIsset(params.WithIsset),
		WithConstructorTypeRender(params.Constructor),
		WithOptionTypeName(cmp.Or(params.OptionTypeName, parityOptionsType)),
	)
}

func TestRenderParity_Testdata(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(parityTestdataDir)
	require.NoError(t, err)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dir := filepath.Join(parityTestdataDir, entry.Name())
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()

			rendered := requireRenderParity(t, testdataRenderOptions(t, dir))

			expected, err := os.ReadFile(filepath.Join(dir, "options_generated.go.expected"))
			require.NoError(t, err)
			require.Equal(t, string(expected), string(rendered))
		})
	}
}

func makeTagOption(required bool, validator, defaultValue string) TagOption {
	return TagOption{
		IsRequired:    required,
		GoValidator:   validator,
		Default:       defaultValue,
		Variadic:      false,
		VariadicIsSet: false,
		Skip:          false,
		Name:          "",
	}
}

func synthOption(name, field, typ, docstring string, tag TagOption) OptionMeta {
	return OptionMeta{Name: name, Docstring: docstring, Field: field, Type: typ, TagOption: tag}
}

// synthImports covers every import shape the pruning and layout have to
// handle: standard library and other packages, aliases, dot and blank
// imports, duplicates, the same path with and without an alias, an import that
// nothing uses and the legacy appengine group of goimports.
func synthImports() []Import {
	alias := func(name string) *string { return &name }

	return []Import{
		{Path: `"net/http"`, Alias: alias("http")},
		{Path: `"github.com/go-playground/validator/v10"`, Alias: nil},
		{Path: `"io"`, Alias: nil},
		{Path: `"fmt"`, Alias: nil},
		{Path: `"appengine/datastore"`, Alias: nil},
		{Path: `"time"`, Alias: nil},
		{Path: `"github.com/kazhuravlev/options-gen/internal/generator/testdata"`, Alias: alias("td")},
		{Path: `"fmt"`, Alias: alias("fmtalias")},
		{Path: `"math"`, Alias: alias(".")},
		{Path: `"fmt"`, Alias: nil},
		{Path: `"embed"`, Alias: alias("_")},
	}
}

// synthSpec builds a spec that exercises every option shape: mandatory and
// optional options, defaults of every supported type, variadic options, custom
// option names, validators, type parameters and docstrings of all kinds.
func synthSpec(validators, generics bool) *OptionSpec {
	validator := func(rule string) string {
		if validators {
			return rule
		}

		return ""
	}

	// Mandatory options are rendered only in the constructor, so without one
	// the datastore import must go away.
	key := synthOption("Key", "key", "*datastore.Key", "", makeTagOption(true, "", ""))
	stringer := synthOption("Stringer", "stringer", "fmt.Stringer", "// Stringer is mandatory.",
		makeTagOption(true, validator("required"), ""))

	timeout := synthOption("Timeout", "timeout", "time.Duration",
		"// Timeout bounds the call.\n// \n// It is a time.Duration.", makeTagOption(false, validator("gte=0"), "1s"))
	retries := synthOption("Retries", "retries", "int", "", makeTagOption(false, "", "3"))
	title := synthOption("Name", "name", "string", "", makeTagOption(false, "", "default name"))
	title.TagOption.Name = "Title"
	enabled := synthOption("Enabled", "enabled", "bool", "", makeTagOption(false, "", "true"))

	embedded := synthOption("Embedded", "embedded", "*td.StructForEmbed",
		"// \t\tJust\n// \t\ta\n// \t\tcomment", makeTagOption(false, "", ""))
	errs := synthOption("Errors", "errors", "validator.FieldError", "", makeTagOption(false, validator("dive"), ""))
	errs.TagOption.Variadic = true
	errs.TagOption.VariadicIsSet = true
	states := synthOption("States", "states", "map[string]fmtalias.State", "", makeTagOption(false, "", ""))
	handler := synthOption("Handler", "handler", "func(http.ResponseWriter) error", "", makeTagOption(false, "", ""))
	events := synthOption("Events", "events", "chan<- int", "", makeTagOption(false, "", ""))
	doc := synthOption("Doc", "doc", "string",
		"// Doc holds:\n//   - one\n//   - two\n//\n//\tcode := 1\n//\n// Deprecated: use something else.",
		makeTagOption(false, validator("max=10"), ""))

	spec := &OptionSpec{
		TypeParamsSpec: "",
		TypeParams:     "",
		Options: []OptionMeta{
			key, stringer, timeout, retries, title, enabled, embedded, errs, states, handler, events, doc,
		},
	}

	if generics {
		spec.TypeParamsSpec = "[T fmt.Stringer, K comparable]"
		spec.TypeParams = "[T, K]"
		spec.Options = append(spec.Options,
			synthOption("Values", "values", "[]T", "// Values are generic.", makeTagOption(false, "", "")),
			synthOption("Lookup", "lookup", "map[K]T", "", makeTagOption(false, validator("required"), "")),
		)
	}

	return spec
}

func TestRenderParity_Synthetic(t *testing.T) {
	t.Parallel()

	type renderVariant struct {
		defaults    string
		constructor string
		withIsset   bool
		validators  bool
		generics    bool
	}

	var variants []renderVariant
	for _, defaults := range []string{parityDefaultsNone, parityDefaultsTag, parityDefaultsVar, parityDefaultsFunc} {
		for _, constructor := range []string{"public", "private", "no"} {
			for _, flags := range []int{0b000, 0b001, 0b010, 0b011, 0b100, 0b101, 0b110, 0b111} {
				variants = append(variants, renderVariant{
					defaults:    defaults,
					constructor: constructor,
					withIsset:   flags&0b001 != 0,
					validators:  flags&0b010 != 0,
					generics:    flags&0b100 != 0,
				})
			}
		}
	}

	for _, variant := range variants {
		name := fmt.Sprintf("defaults=%s/constructor=%s/isset=%t/validators=%t/generics=%t",
			variant.defaults, variant.constructor, variant.withIsset, variant.validators, variant.generics)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var tagName, varName, funcName string
			switch variant.defaults {
			case parityDefaultsTag:
				tagName = "default"
			case parityDefaultsVar:
				varName = "defaultOptions"
			case parityDefaultsFunc:
				funcName = "getDefaultOptions"
			}

			requireRenderParity(t, NewOptions(
				WithVersion("parity"),
				WithPackageName("testcase"),
				WithOptionsStructName("Options"),
				WithFileImports(synthImports()),
				WithSpec(synthSpec(variant.validators, variant.generics)),
				WithTagName(tagName),
				WithVarName(varName),
				WithFuncName(funcName),
				WithPrefix("Some"),
				WithWithIsset(variant.withIsset),
				WithConstructorTypeRender(variant.constructor),
				WithOptionTypeName(parityOptionsType),
			))
		})
	}
}

func TestFormatGeneratedSource_ImportLayout(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		src  string
		want string // Expected output, "" to only check parity with the reference.
	}{
		{
			name: "all imports unused",
			src:  "package p\n\nimport (\n\t\"fmt\"\n\t\"io\"\n)\n\nvar x = 1\n",
			want: "package p\n\nvar x = 1\n",
		},
		{
			name: "no import declaration",
			src:  "package p\n\nvar x = 1\n",
			want: "package p\n\nvar x = 1\n",
		},
		{
			name: "single import kept parenthesized",
			src:  "package p\n\nimport (\n\t\"fmt\"\n)\n\nvar _ = fmt.Sprintf\n",
			want: "package p\n\nimport (\n\t\"fmt\"\n)\n\nvar _ = fmt.Sprintf\n",
		},
		{
			name: "dot and blank imports are kept, unused named ones dropped",
			src: "package p\n\nimport (\n\t\"fmt\"\n\t\"io\"\n\t. \"math\"\n\t_ \"net/http/pprof\"\n\talias \"strings\"\n)\n\n" +
				"var _ = fmt.Sprintf\nvar _ = alias.Builder{}\nvar _ = Pi\n",
			want: "package p\n\nimport (\n\t\"fmt\"\n\n\t. \"math\"\n\t_ \"net/http/pprof\"\n\talias \"strings\"\n)\n\n" +
				"var _ = fmt.Sprintf\nvar _ = alias.Builder{}\nvar _ = Pi\n",
		},
		{
			name: "groups, duplicates and aliases",
			src: "package p\n\nimport (\n\tf \"fmt\"\n\t\"golang.org/x/tools/imports\"\n\t\"os\"\n\t\"fmt\"\n\t\"fmt\"\n" +
				"\t\"appengine/datastore\"\n\t\"github.com/stretchr/testify/require\"\n\thttp \"net/http\"\n)\n\n" +
				"var _ = fmt.Sprintf\nvar _ = f.Sprintf\nvar _ = os.Exit\nvar _ = imports.Process\n" +
				"var _ = datastore.Key{}\nvar _ = require.Equal\nvar _ = http.Get\n",
			want: "package p\n\nimport (\n\t\"fmt\"\n\tf \"fmt\"\n\thttp \"net/http\"\n\t\"os\"\n\n" +
				"\t\"github.com/stretchr/testify/require\"\n\t\"golang.org/x/tools/imports\"\n\n\t\"appengine/datastore\"\n)\n\n" +
				"var _ = fmt.Sprintf\nvar _ = f.Sprintf\nvar _ = os.Exit\nvar _ = imports.Process\n" +
				"var _ = datastore.Key{}\nvar _ = require.Equal\nvar _ = http.Get\n",
		},
		{
			name: "major version suffix and selector chains",
			src: "package p\n\nimport (\n\t\"github.com/go-playground/validator/v10\"\n\t\"example.com/cfg\"\n\t\"o\"\n)\n\n" +
				"var _ = validator.New\nvar _ = cfg.Default.Timeout\n\nfunc f(o *int) { _ = o }\n",
			want: "package p\n\nimport (\n\t\"example.com/cfg\"\n\t\"github.com/go-playground/validator/v10\"\n)\n\n" +
				"var _ = validator.New\nvar _ = cfg.Default.Timeout\n\nfunc f(o *int) { _ = o }\n",
		},
		{
			name: "unused first import leaves no blank line",
			src:  "package p\n\nimport (\n\t\"io\"\n\t\"fmt\"\n)\n\nvar _ = fmt.Sprintf\n",
			want: "package p\n\nimport (\n\t\"fmt\"\n)\n\nvar _ = fmt.Sprintf\n",
		},
		{
			name: "unused last import leaves no blank line",
			src:  "package p\n\nimport (\n\t\"fmt\"\n\t\"io\"\n)\n\nvar _ = fmt.Sprintf\n",
			want: "package p\n\nimport (\n\t\"fmt\"\n)\n\nvar _ = fmt.Sprintf\n",
		},
		{
			name: "unused import in the middle splits the block into runs",
			src: "package p\n\nimport (\n\t\"os\"\n\t\"io\"\n\t\"bufio\"\n\t\"fmt\"\n" +
				"\t\"github.com/stretchr/testify/require\"\n\t\"errors\"\n)\n\n" +
				"var _ = fmt.Sprintf\nvar _ = os.Exit\nvar _ = require.Equal\nvar _ = errors.New\n",
			want: "package p\n\nimport (\n\t\"os\"\n\n\t\"errors\"\n\t\"fmt\"\n\n" +
				"\t\"github.com/stretchr/testify/require\"\n)\n\n" +
				"var _ = fmt.Sprintf\nvar _ = os.Exit\nvar _ = require.Equal\nvar _ = errors.New\n",
		},
		{
			name: "several import declarations are merged",
			src: "package p\n\nimport \"os\"\nimport (\n\t\"io\"\n\t\"fmt\"\n)\n\n" +
				"var _ = fmt.Sprintf\nvar _ = os.Exit\nvar _ = io.EOF\n",
			want: "package p\n\nimport (\n\t\"fmt\"\n\t\"io\"\n\t\"os\"\n)\n\n" +
				"var _ = fmt.Sprintf\nvar _ = os.Exit\nvar _ = io.EOF\n",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			want, err := optimizeGeneratedSource([]byte(testCase.src))
			require.NoError(t, err)

			got, err := formatGeneratedSource([]byte(testCase.src))
			require.NoError(t, err)
			require.Equal(t, string(want), string(got))

			if testCase.want != "" {
				require.Equal(t, testCase.want, string(got))
			}
		})
	}
}
