package testcase

import "github.com/kazhuravlev/options-gen/options-gen/testdata/case-18-embed-inline-no-panic/embedpkg"

type EmbedStruct struct {
	embedField string
}

type Options struct {
	EmbedStruct
	*embedpkg.Struct
	inline struct {
		inlineField string
	}
	name string
}
