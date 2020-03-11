// Copyright (c) 2017, 2019 AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package load_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/danos/config/load"
	"github.com/danos/config/testutils"
	"github.com/danos/config/testutils/assert"
	"github.com/danos/mgmterror/errtest"
)

const schemaTemplate = `
module test-load {
	namespace "urn:vyatta.com:test:load";
	prefix test;
	organization "Brocade Communications Systems, Inc.";
	contact
		"Brocade Communications Systems, Inc.
		 Postal: 130 Holger Way
		         San Jose, CA 95134
		 E-mail: support@Brocade.com
		 Web: www.brocade.com";
	revision 2014-12-29 {
		description "Test schema for load";
	}
	%s
}
`

func getSchema(schema string) *bytes.Buffer {
	return bytes.NewBufferString(fmt.Sprintf(schemaTemplate, schema))
}

var mergeOrLoadSchema = `
leaf testbool {
	type boolean;
}
leaf teststring {
	type string;
}
leaf testint {
	type uint8 {
		range 1..64;
	}
	must "../testbool = true()";
}
container testcont {
	leaf testleaf {
		type string;
	}
}`

func loadTestConfig(t *testing.T, schema, config string) []error {
	ms, _, err := testutils.NewModelSetSpec(t).
		SetSchemas(getSchema(schema).Bytes()).
		GenerateModelSets()
	if err != nil {
		t.Fatalf("Unable to create modelSetSpec: %s\n", err.Error())
		return nil
	}

	_, err, warnings := load.LoadString("loadtest", config, ms)
	if err != nil {
		t.Fatalf("Unexpected error loading config: %s\n", err.Error())
		return nil
	}
	return warnings
}

// Assumes actual and expected warnings are in matching order.  As we run
// through config in deterministic (not 'map') order, this should be ok!
func checkWarnings(
	t *testing.T,
	actWarnings []error,
	expWarnings ...*assert.ExpectedMessages) {

	if len(actWarnings) == 0 {
		t.Fatalf("Expected warnings.")
		return
	}
	if len(actWarnings) != len(expWarnings) {
		t.Fatalf("Expected %d warnings, but got %d\n",
			len(expWarnings), len(actWarnings))
		return
	}

	for ix, actWarn := range actWarnings {
		expWarnings[ix].ContainedIn(t, actWarn.Error())
	}
}

func TestLoadNoError(t *testing.T) {

	const testConfig = `
		teststring stuff
	`

	warnings := loadTestConfig(t, mergeOrLoadSchema, testConfig)

	if len(warnings) > 0 {
		t.Fatalf("Invalid paths:\n%v\n", warnings)
	}
}

func TestLoadNonExistentPath(t *testing.T) {

	const testConfig = `
		teststringy stuff
	`

	warnings := loadTestConfig(t, mergeOrLoadSchema, testConfig)

	expErrors := assert.NewExpectedMessages(
		errtest.NewInvalidNodeError(t, "/teststringy").
			RawErrorStrings()...)

	checkWarnings(t, warnings, expErrors)
}

func TestLoadInvalidType(t *testing.T) {

	const testConfig = `
		testint not_a_number
	`

	warnings := loadTestConfig(t, mergeOrLoadSchema, testConfig)

	expErrors := assert.NewExpectedMessages(
		errtest.NewInvalidTypeError(t, "/testint/not_a_number",
			"an uint8").RawErrorStrings()...)

	checkWarnings(t, warnings, expErrors)
}

func TestLoadValueOutOfRange(t *testing.T) {

	const testConfig = `
		testint 99
	`

	warnings := loadTestConfig(t, mergeOrLoadSchema, testConfig)

	expErrors := assert.NewExpectedMessages(
		errtest.NewInvalidRangeError(t, "/testint/99", 1, 64).
			RawErrorStrings()...)

	checkWarnings(t, warnings, expErrors)
}

func TestLoadMultipleWarnings(t *testing.T) {

	const testConfig = `
		testint 99
		teststringy stuff
	`

	warnings := loadTestConfig(t, mergeOrLoadSchema, testConfig)

	expIntErrors := assert.NewExpectedMessages(
		errtest.NewInvalidRangeError(t, "/testint/99", 1, 64).
			RawErrorStrings()...)
	expStringErrors := assert.NewExpectedMessages(
		errtest.NewInvalidNodeError(t, "/teststringy").
			RawErrorStrings()...)

	checkWarnings(t, warnings, expIntErrors, expStringErrors)
}

var testNoTrailingNewlineConfigs = []string{
	"testint 33",
	"teststring stuff",
	"testcont {\ntestleaf aValue\n}",
}

func TestLoadNoTrailingNewline(t *testing.T) {

	for _, testConfig := range testNoTrailingNewlineConfigs {
		warnings := loadTestConfig(t, mergeOrLoadSchema, testConfig)

		if warnings != nil {
			t.Fatalf("Unexpected warning(s) for '%s': %v\n",
				testConfig, warnings)
		}
	}
}
