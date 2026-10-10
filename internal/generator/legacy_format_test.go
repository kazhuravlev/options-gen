//nolint:testpackage
package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"

	"golang.org/x/tools/imports"
)

// This file keeps the previous formatting pipeline of Render as the reference
// implementation for the parity tests of formatGeneratedSource. It parsed and
// printed the rendered source four times: go/format once and
// golang.org/x/tools/imports (FormatOnly) twice more on top of its own parse.
// It is test-only code; production code must not call it.

const generatedFormatTabWidth = 8

func optimizeGeneratedSource(src []byte) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse generated source: %w", err)
	}

	pruneUnusedImports(file)
	ast.SortImports(fset, file)

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return nil, fmt.Errorf("format generated source: %w", err)
	}

	formatted, err := imports.Process("", buf.Bytes(), &imports.Options{
		Fragment:   false,
		AllErrors:  false,
		Comments:   true,
		TabIndent:  true,
		TabWidth:   generatedFormatTabWidth,
		FormatOnly: true,
	})
	if err != nil {
		return nil, fmt.Errorf("sort generated imports: %w", err)
	}

	return formatted, nil
}

func pruneUnusedImports(file *ast.File) {
	usedSelectors := usedSelectorBases(file)

	importDecls := file.Decls[:0]
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			importDecls = append(importDecls, decl)

			continue
		}

		importSpecs := genDecl.Specs[:0]
		for _, spec := range genDecl.Specs {
			if isImportUsed(spec.(*ast.ImportSpec), usedSelectors) {
				importSpecs = append(importSpecs, spec)
			}
		}

		if len(importSpecs) == 0 {
			continue
		}

		genDecl.Specs = importSpecs
		importDecls = append(importDecls, genDecl)
	}

	file.Decls = importDecls
}
