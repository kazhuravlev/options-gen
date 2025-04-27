package optionsgen_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	optionsgen "github.com/kazhuravlev/options-gen/options-gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Parallel()

	t.Run("simple_cases", func(t *testing.T) {
		t.Parallel()

		var testDirs []string
		{
			const testdataDir = "./testdata"
			fileInfos, err := os.ReadDir(testdataDir)
			require.NoError(t, err)

			for _, file := range fileInfos {
				if !file.IsDir() {
					continue
				}

				testDirs = append(testDirs, filepath.Join(testdataDir, file.Name()))
			}
		}

		for _, dir := range testDirs {
			dir := dir
			t.Run(dir, func(t *testing.T) {
				outFilename := filepath.Join(dir, "options_generated.go")
				expFilename := filepath.Join(dir, "options_generated.go.expected")
				paramsFilename := filepath.Join(dir, ".params.json")
				params := readParams(paramsFilename)

				err := optionsgen.Run(optionsgen.NewOptions(
					optionsgen.WithVersion("qa-version"),
					optionsgen.WithInFilename(filepath.Join(dir, "options.go")),
					optionsgen.WithOutFilename(outFilename),
					optionsgen.WithStructName("Options"),
					optionsgen.WithPackageName("testcase"),
					optionsgen.WithOutPrefix(params.OutPrefix),
					optionsgen.WithDefaults(params.Defaults),
					optionsgen.WithShowWarnings(true),
					optionsgen.WithWithIsset(params.WithIsset),
					optionsgen.WithAllVariadic(params.AllVariadic),
					optionsgen.WithConstructorTypeRender(params.Constructor),
					optionsgen.WithOutOptionTypeName(params.OptionTypeName),
				))
				assert.NoError(t, err)

				helpEqualFiles(t, expFilename, outFilename)
			})
		}
	})

	t.Run("source_not_found", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		err := optionsgen.Run(optionsgen.NewOptions(
			optionsgen.WithVersion("qa-version"),
			optionsgen.WithInFilename(filepath.Join(dir, "options.go")),
			optionsgen.WithOutFilename(filepath.Join(dir, "options_generated.go")),
			optionsgen.WithStructName("Options"),
			optionsgen.WithPackageName("testcase"),
			optionsgen.WithOutPrefix("XXX"),
			optionsgen.WithDefaults(optionsgen.Defaults{From: optionsgen.DefaultsFromTag, Param: ""}),
			optionsgen.WithShowWarnings(true),
			optionsgen.WithWithIsset(false),
			optionsgen.WithAllVariadic(false),
			optionsgen.WithConstructorTypeRender(optionsgen.ConstructorPublicRender),
			optionsgen.WithOutOptionTypeName(""),
		))
		assert.ErrorIs(t, err, syscall.ENOENT)
	})
}

type Params struct {
	OutPrefix      string                           `json:"out_prefix"` //nolint:tagliatelle
	Defaults       optionsgen.Defaults              `json:"defaults"`
	Constructor    optionsgen.ConstructorTypeRender `json:"constructor"`
	WithIsset      bool                             `json:"with_isset"`       //nolint:tagliatelle
	AllVariadic    bool                             `json:"all_variadic"`     //nolint:tagliatelle
	OptionTypeName string                           `json:"option_type_name"` //nolint:tagliatelle
}

func readParams(filename string) Params {
	params := Params{
		OutPrefix: "",
		Defaults: optionsgen.Defaults{
			From:  optionsgen.DefaultsFromTag,
			Param: "",
		},
		Constructor:    optionsgen.ConstructorPublicRender,
		AllVariadic:    false,
		OptionTypeName: "",
	}

	bb, err := os.ReadFile(filename)
	if err != nil {
		return params
	}

	if err := json.Unmarshal(bb, &params); err != nil {
		return params
	}

	return params
}

func helpEqualFiles(t *testing.T, filename1, filename2 string) {
	t.Helper()

	f1Bytes, err := os.ReadFile(filename1)
	require.NoError(t, err)

	f2Bytes, err := os.ReadFile(filename2)
	require.NoError(t, err)

	assert.Equal(t, string(f1Bytes), string(f2Bytes))
}
