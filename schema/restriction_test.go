// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This file contains tests on restrictions available to types

package schema

import (
	"bytes"
	"fmt"
	"testing"
)

func buildSynRestrictionSchema(typeType, extra string) []byte {

	testLeaf := fmt.Sprintf(
		`identity foo;
        leaf testLeaf {
			type %s {
                %s
				configd:syntax ls;
			}
		}`, typeType, extra)

	return bytes.NewBufferString(
		fmt.Sprintf(schemaTemplate, testLeaf)).Bytes()
}

func checkInvalidSynRestriction(t *testing.T, typeType, extra string) {

	schema_text := buildSynRestrictionSchema(typeType, extra)
	_, err := GetConfigSchema(schema_text)

	expected := fmt.Sprintf(
		"type %s: configd:syntax restriction is not valid for this type",
		typeType)
	if typeType == "union" {
		expected = fmt.Sprintf(
			"type union: cannot restrict configd:syntax of a union type - " +
				"restrictions must be applied to members instead")
	}

	assertErrorContains(t, err, expected)
}

func checkValidSynRestriction(t *testing.T, typeType, extra string) {

	schema_text := buildSynRestrictionSchema(typeType, extra)
	_, err := GetConfigSchema(schema_text)

	if err != nil {
		t.Errorf("Unexpected error when testing syntax restriction with %s:\n  %s",
			typeType, err.Error())
	}
}

func TestRestrictions(t *testing.T) {
	checkInvalidSynRestriction(t, "boolean", "")
	checkInvalidSynRestriction(t, "empty", "")
	checkInvalidSynRestriction(t, "enumeration", "enum foo;")
	checkInvalidSynRestriction(t, "identityref", "base foo;")
	checkValidSynRestriction(t, "int16", "")
	checkValidSynRestriction(t, "uint64", "")
	checkValidSynRestriction(t, "decimal64", "fraction-digits 2;")
	checkInvalidSynRestriction(t, "bits", "")
	checkInvalidSynRestriction(t, "leafref", "path /foo/bar;")
	checkInvalidSynRestriction(t, "union", "type string;")
	checkValidSynRestriction(t, "string", "")
	checkInvalidSynRestriction(t, "instance-identifier", "")
}
