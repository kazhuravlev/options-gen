package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	optionsgen "github.com/kazhuravlev/options-gen/options-gen"
	"github.com/stretchr/testify/require"
)

const warnSrc = `package testpkg

type Options struct {
	Public string
}
`

func runGen(t *testing.T, extra ...string) (string, error) {
	t.Helper()

	dir := t.TempDir()
	in := filepath.Join(dir, "in.go")
	out := filepath.Join(dir, "out_generated.go")
	require.NoError(t, os.WriteFile(in, []byte(warnSrc), 0o600))

	args := append([]string{
		"-filename", in, "-out-filename", out, "-pkg", "testpkg", "-from-struct", "Options",
	}, extra...)

	var buf bytes.Buffer
	err := run(args, func(string) string { return "" }, &buf)

	return out, err
}

func TestRun_WarningsAreErrors(t *testing.T) {
	t.Parallel()

	t.Run("warnings fail and nothing is written", func(t *testing.T) {
		t.Parallel()

		out, err := runGen(t)
		require.ErrorIs(t, err, optionsgen.ErrGeneratorWarnings)
		require.ErrorContains(t, err, "Public")
		require.NoFileExists(t, out)
	})

	t.Run("errors are not silenced by mute-warnings", func(t *testing.T) {
		t.Parallel()

		_, err := runGen(t, "-mute-warnings")
		require.ErrorIs(t, err, optionsgen.ErrGeneratorWarnings)
	})

	t.Run("ignore-errors still generates", func(t *testing.T) {
		t.Parallel()

		out, err := runGen(t, "-ignore-errors")
		require.NoError(t, err)
		require.FileExists(t, out)
	})
}
