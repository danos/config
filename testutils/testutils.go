// Copyright (c) 2018-2019, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// Utilities used by configd unit tests.

package testutils

import (
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/danos/config/schema"
	"github.com/danos/yang/compile"
	"github.com/danos/yang/parse"
)

// Create ModelSet structure from multiple buffers, each buffer
// represents a single yang module.
func getSchema(getFullSchema, skipUnknown bool, bufs ...[]byte) (schema.ModelSet, error) {
	const name = "schema"
	modules := make(map[string]*parse.Tree)
	for index, b := range bufs {
		t, err := schema.Parse(name+strconv.Itoa(index), string(b))
		if err != nil {
			return nil, err
		}
		mod := t.Root.Argument().String()
		modules[mod] = t
	}
	st, err := schema.CompileModules(modules, "", skipUnknown,
		compile.Include(compile.IsConfig, compile.IncludeState(getFullSchema)), nil)
	if err != nil {
		return nil, err
	}
	return st, nil
}

func GetConfigSchema(buf ...[]byte) (schema.ModelSet, error) {
	return getSchema(false, false, buf...)
}

func GetConfigSchemaSkipUnknown(buf ...[]byte) (schema.ModelSet, error) {
	return getSchema(false, true, buf...)
}

func GetFullSchema(buf ...[]byte) (schema.ModelSet, error) {
	return getSchema(true, false, buf...)
}

// Given a file containing a list of feature capabilities, create a directory
// structure as required by the compiler to determine enabled features
// A new temporary directory will be created on each invocation
func CreateFeaturesChecker(caps string) (compile.FeaturesChecker, error) {
	features := make([]string, 0)
	// For development, access a file with list of features
	text, err := ioutil.ReadFile(caps)
	if err != nil && err != io.EOF {
		return nil, err
	}
	lines := strings.Split(string(text), "\n")
	for _, capLine := range lines {

		// Break each line up into WS seperated tokens
		tokens := strings.Fields(capLine)
		if len(tokens) == 0 {
			// empty line
			continue
		}

		// Only consider the first token
		if strings.ContainsAny(tokens[0], "#;*/\\") {
			// Ignore if contains junk/commented out
			continue
		}

		cPos := strings.Index(tokens[0], ":")

		if (cPos < 1) || (cPos == (len(tokens[0]) - 1)) {
			// Either the : is absent, or is the first or last
			// character in the token, so ignore
			continue
		}
		features = append(features, tokens[0])
	}
	return compile.FeaturesFromNames(true, features...), nil
}
