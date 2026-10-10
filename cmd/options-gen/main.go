package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kazhuravlev/options-gen/internal/version"
	optionsgen "github.com/kazhuravlev/options-gen/options-gen"
)

var (
	// errMissedRequiredOptions is returned by run when at least one of the required flags is empty.
	errMissedRequiredOptions = errors.New("missed required options")
	// errParseFlags wraps flag parsing errors; the FlagSet has already written them and the usage to out.
	errParseFlags = errors.New("parse flags")
)

func main() {
	if err := run(os.Args[1:], os.Getenv, os.Stdout); err != nil {
		if !errors.Is(err, errParseFlags) {
			fmt.Fprintln(os.Stdout, err.Error())
		}

		os.Exit(1)
	}
}

// run parses args, runs the generator and returns a non-nil error when the code was not generated,
// so that `go generate` fails together with the generator. Usage and flag parse errors are written to out.
// getenv provides the GOFILE/GOPACKAGE defaults that `go generate` sets for the tool.
func run(args []string, getenv func(string) string, out io.Writer) error {
	var (
		inFilename            string
		outFilename           string
		optionsStructName     string
		outPackageName        string
		outPrefix             string
		defaultsFrom          string
		muteWarnings          bool
		strict                bool
		withIsset             bool
		allVariadic           bool
		constructorTypeRender optionsgen.ConstructorTypeRender
		outSetterName         string
		exclude               string
	)

	envGoFile := getenv("GOFILE")
	envGoPackage := getenv("GOPACKAGE")

	// Without GOFILE there is nothing to derive the output filename from, so the flag stays required.
	var defaultOutFilename string
	if envGoFile != "" {
		defaultOutFilename = strings.Replace(filepath.Base(envGoFile), ".go", "_generated.go", 1)
	}

	flags := flag.NewFlagSet("options-gen", flag.ContinueOnError)
	flags.SetOutput(out)

	flags.StringVar(&inFilename,
		"filename", envGoFile,
		"input filename")
	flags.StringVar(&outPackageName,
		"pkg", envGoPackage,
		"output package name")
	flags.StringVar(&outFilename,
		"out-filename", defaultOutFilename,
		"output filename")
	flags.StringVar(&optionsStructName,
		"from-struct", "",
		"struct that contains options")
	flags.StringVar(&defaultsFrom,
		"defaults-from", "tag=default",
		"where to get defaults for options. none, tag=TagName, func=FuncName, var=VarName")
	flags.BoolVar(&muteWarnings,
		"mute-warnings", false,
		"mute all warnings")
	flags.BoolVar(&strict,
		"strict", false,
		"treat generator warnings as errors: fail without writing the output file")
	flags.StringVar(&outPrefix,
		"out-prefix", "",
		"prefix for generated structs and functions. It is like namespace that can be used in case "+
			"when you have a several options structs in one package")
	flags.BoolVar(&withIsset,
		"with-isset", false,
		"generate a function that helps check which fields have been set")
	flags.BoolVar(&allVariadic,
		"all-variadic", false,
		"generate variadic functions")
	flags.StringVar((*string)(&constructorTypeRender),
		"constructor", string(optionsgen.ConstructorPublicRender),
		"generate a function constructor. Possible values: "+strings.Join([]string{
			string(optionsgen.ConstructorPublicRender),
			string(optionsgen.ConstructorPrivateRender),
			string(optionsgen.ConstructorNoRender),
		}, ", ")+".")
	flags.StringVar(&outSetterName,
		"out-setter-name", "",
		"name for the option setter type (function alias). If not specified, the 'Opt[StructName]Setter' template is used.")
	flags.StringVar(&exclude, "exclude", "", "list of masks for field names excluded from generation, semicolon-separated")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}

		return fmt.Errorf("%w: %w", errParseFlags, err)
	}

	if isEmpty(inFilename, outFilename, outPackageName, optionsStructName, defaultsFrom) {
		flags.Usage()

		return errMissedRequiredOptions
	}

	defaults, err := parseDefaults(defaultsFrom)
	if err != nil {
		return fmt.Errorf("bad defaults spec: %w", err)
	}

	excludes, err := splitExcludes(exclude)
	if err != nil {
		return fmt.Errorf("parse excludes: %w", err)
	}

	errRun := optionsgen.Run(
		optionsgen.NewOptions(
			optionsgen.WithVersion(version.GetVersion()),
			optionsgen.WithInFilename(inFilename),
			optionsgen.WithOutFilename(outFilename),
			optionsgen.WithStructName(optionsStructName),
			optionsgen.WithPackageName(outPackageName),
			optionsgen.WithOutPrefix(outPrefix),
			optionsgen.WithDefaults(*defaults),
			optionsgen.WithShowWarnings(!muteWarnings),
			optionsgen.WithStrict(strict),
			optionsgen.WithWithIsset(withIsset),
			optionsgen.WithAllVariadic(allVariadic),
			optionsgen.WithConstructorTypeRender(constructorTypeRender),
			optionsgen.WithOutOptionTypeName(outSetterName),
			optionsgen.WithExclude(excludes...),
			optionsgen.WithWarningsHandler(func(msg string) {
				fmt.Fprintln(out, msg)
			}),
		),
	)
	if errRun != nil {
		return fmt.Errorf("cannot run options gen: %w", errRun)
	}

	return nil
}

func parseDefaults(in string) (*optionsgen.Defaults, error) {
	parts := strings.Split(in, "=")

	from := optionsgen.DefaultsFrom(parts[0])

	switch from {
	case optionsgen.DefaultsFromNone:
		return &optionsgen.Defaults{
			From:  from,
			Param: "",
		}, nil
	case optionsgen.DefaultsFromTag:
		return &optionsgen.Defaults{
			From:  from,
			Param: get1(parts),
		}, nil
	case optionsgen.DefaultsFromVar:
		return &optionsgen.Defaults{
			From:  from,
			Param: get1(parts),
		}, nil
	case optionsgen.DefaultsFromFunc:
		return &optionsgen.Defaults{
			From:  from,
			Param: get1(parts),
		}, nil
	}

	return nil, errors.New("bad syntax")
}

func get1(parts []string) string {
	if len(parts) == 2 { //nolint:mnd // expect exactly two part
		return parts[1]
	}

	return ""
}

func isEmpty(values ...string) bool {
	for i := range values {
		if values[i] == "" {
			return true
		}
	}

	return false
}

func splitExcludes(exclude string) ([]*regexp.Regexp, error) {
	if len(exclude) == 0 {
		return nil, nil
	}

	patterns := strings.Split(exclude, ";")
	result := make([]*regexp.Regexp, 0, len(patterns))

	for _, pattern := range patterns {
		reg, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("compile pattern '%s': %w", pattern, err)
		}

		result = append(result, reg)
	}

	return result, nil
}
