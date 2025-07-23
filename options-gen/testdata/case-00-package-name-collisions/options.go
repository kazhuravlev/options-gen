package testcase

import (
	// Main point of this test case:
	//  - some package imported with an alias (api imported as http)
	//  - we have a package with name of alias (net/http)
	http "github.com/kazhuravlev/options-gen/options-gen/testdata/case-00-package-name-collisions/pkg/net/api"
)

type Options struct {
	F1 http.Client `option:"mandatory"`
}
