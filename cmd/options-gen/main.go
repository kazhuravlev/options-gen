package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"

	optionsgen "github.com/kazhuravlev/options-gen/options-gen"
)

const versionUnknown = "unknown-local"

var Version = versionUnknown

func main() {
	// In case if not - someone (task examples:update) explicitly set the value of Version.
	if Version == versionUnknown {
		if bi, ok := debug.ReadBuildInfo(); ok {
			if bi.Main.Version != "" {
				Version = bi.Main.Version
			}
		}
	}

	var (
		inFilename            string
		outFilename           string
		optionsStructName     string
		outPackageName        string
		outPrefix             string
		defaultsFrom          string
		muteWarnings          bool
		withIsset             bool
		allVariadic           bool
		constructorTypeRender optionsgen.ConstructorTypeRender
		outSetterName         string
		exclude               string
	)

	envGoFile := os.Getenv("GOFILE")
	envGoPackage := os.Getenv("GOPACKAGE")

	defaultOutFilename := strings.Replace(filepath.Base(envGoFile), ".go", "_generated.go", 1)

	flag.StringVar(&inFilename,
		"filename", envGoFile,
		"input filename")
	flag.StringVar(&outPackageName,
		"pkg", envGoPackage,
		"output package name")
	flag.StringVar(&outFilename,
		"out-filename", defaultOutFilename,
		"output filename")
	flag.StringVar(&optionsStructName,
		"from-struct", "",
		"struct that contains options")
	flag.StringVar(&defaultsFrom,
		"defaults-from", "tag=default",
		"where to get defaults for options. none, tag=TagName, func=FuncName, var=VarName")
	flag.BoolVar(&muteWarnings,
		"mute-warnings", false,
		"mute all warnings")
	flag.StringVar(&outPrefix,
		"out-prefix", "",
		"prefix for generated structs and functions. It is like namespace that can be used in case "+
			"when you have a several options structs in one package")
	flag.BoolVar(&withIsset,
		"with-isset", false,
		"generate a function that helps check which fields have been set")
	flag.BoolVar(&allVariadic,
		"all-variadic", false,
		"generate variadic functions")
	flag.StringVar((*string)(&constructorTypeRender),
		"constructor", string(optionsgen.ConstructorPublicRender),
		"generate a function constructor. Possible values: "+strings.Join([]string{
			string(optionsgen.ConstructorPublicRender),
			string(optionsgen.ConstructorPrivateRender),
			string(optionsgen.ConstructorNoRender),
		}, ", ")+".")
	flag.StringVar(&outSetterName,
		"out-setter-name", "",
		"name for the option setter type (function alias). If not specified, the 'Opt[StructName]Setter' template is used.")
	flag.StringVar(&exclude, "exclude", "", "list of masks for field names excluded from generation, semicolon-separated")
	flag.Parse()

	if isEmpty(inFilename, outFilename, outPackageName, optionsStructName, defaultsFrom) {
		flag.Usage()
		//nolint:forbidigo
		fmt.Println("missed required options")

		return
	}

	defaults, err := parseDefaults(defaultsFrom)
	if err != nil {
		//nolint:forbidigo
		fmt.Println("bad defaults spec", err.Error())

		return
	}

	excludes, err := splitExcludes(exclude)
	if err != nil {
		//nolint:forbidigo
		fmt.Println("parse excludes", err.Error())

		return
	}

	errRun := optionsgen.Run(
		optionsgen.NewOptions(
			optionsgen.WithVersion(Version),
			optionsgen.WithInFilename(inFilename),
			optionsgen.WithOutFilename(outFilename),
			optionsgen.WithStructName(optionsStructName),
			optionsgen.WithPackageName(outPackageName),
			optionsgen.WithOutPrefix(outPrefix),
			optionsgen.WithDefaults(*defaults),
			optionsgen.WithShowWarnings(!muteWarnings),
			optionsgen.WithWithIsset(withIsset),
			optionsgen.WithAllVariadic(allVariadic),
			optionsgen.WithConstructorTypeRender(constructorTypeRender),
			optionsgen.WithOutOptionTypeName(outSetterName),
			optionsgen.WithExclude(excludes...),
		),
	)
	if errRun != nil {
		//nolint:forbidigo
		fmt.Println("cannot run options gen", errRun.Error())

		return
	}
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
