// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

//
//  Helper Functions
//
func buildSchema(t *testing.T, schema_snippet string) ModelSet {

	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate, schema_snippet))
	st, err := GetConfigSchema(schema_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected error when parsing RPC schema: %s", err)
	}

	return st
}

func getSchemaBuildError(t *testing.T, schema_snippet string) error {

	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate, schema_snippet))
	_, err := GetConfigSchema(schema_text.Bytes())
	if err == nil {
		t.Fatalf("Unexpected success when parsing schema")
	}

	return err
}

// This returns a standard checker function that can be used from NodeChecker
func checkNormalize(expected_script string) checkFn {
	return func(t *testing.T, actual Node) {
		actual_script := actual.Type().(Type).ConfigdExt().Normalize
		if expected_script != actual_script {
			t.Errorf("Node normalization script does not match\n"+
				"  expect = %s\n"+
				"  actual = %s",
				expected_script, actual_script)
		}
	}
}

func assertNormalizeMatches(
	t *testing.T, st ModelSet, node_name, normalize_script string) {

	checklist := []checkFn{
		CheckName(node_name),
		checkNormalize(normalize_script),
	}
	expected := NodeChecker{node_name, checklist}
	actual := st.SchemaChild(node_name)

	expected.check(t, actual)
}

//
//  Test Cases
//
func TestNormalizeSuccessSimple(t *testing.T) {
	schema_snippet := `
		leaf test-leaf {
			type string {
			    configd:normalize "normalize from-leaf";
			}
		}`

	st := buildSchema(t, schema_snippet)
	assertNormalizeMatches(t, st, "test-leaf", "normalize from-leaf")
}

func TestNormalizeSuccessTypedef(t *testing.T) {
	schema_snippet := `
		typedef testType {
			type string {
			    configd:normalize "normalize from-typedef";
			}
		}

		leaf test-leaf {
			type testType;
		}`

	st := buildSchema(t, schema_snippet)
	assertNormalizeMatches(t, st, "test-leaf", "normalize from-typedef")
}

func TestNormalizeSuccessTypedefOvereridden(t *testing.T) {
	schema_snippet := `
		typedef testType {
			type string {
			    configd:normalize "normalize from-typedef";
			}
		}

		leaf test-leaf {
			type testType {
				configd:normalize "normalize from-leaf";
			}
		}`

	st := buildSchema(t, schema_snippet)
	assertNormalizeMatches(t, st, "test-leaf", "normalize from-leaf")
}

func TestNormalizeSuccessUnion(t *testing.T) {
	schema_snippet := `
		typedef testType {
			type string {
			    configd:normalize "normalize from-typedef";
			}
		}

		typedef testUnion {
			type union {
				type string {
					configd:normalize "normalize unionString";
				}
				type testType {
					configd:normalize "normalize override from-typedef";
				}
				configd:normalize "normalize from-union";
			}
		}

		leaf-list test-leaf-list {
			type testUnion;
		}`

	st := buildSchema(t, schema_snippet)
	assertNormalizeMatches(t, st, "test-leaf-list", "normalize from-union")
}

func TestNormalizeSuccessUnionOverride(t *testing.T) {
	schema_snippet := `
		typedef testType {
			type string {
			    configd:normalize "normalize from-typedef";
			}
		}

		typedef testUnion {
			type union {
				type string {
					configd:normalize "normalize unionString";
				}
				type testType {
					configd:normalize "normalize override from-typedef";
				}
				configd:normalize "normalize from-union";
			}
		}

		leaf-list test-leaf-list {
			type testUnion {
				configd:normalize "normalize from-leaf";
			}
		}`

	st := buildSchema(t, schema_snippet)
	assertNormalizeMatches(t, st, "test-leaf-list", "normalize from-leaf")
}

func TestNormalizeSuccessList(t *testing.T) {
	schema_snippet := `
		list test-list {
			key name;
			leaf name {
				type string {
				    configd:normalize "normalize from-leaf";
				}
			}
		}`

	st := buildSchema(t, schema_snippet)
	assertNormalizeMatches(t, st, "test-list", "normalize from-leaf")
}

func TestNormalizeFailsOnContainer(t *testing.T) {
	schema_snippet := `
		container test {
			configd:normalize "normalize from-leaf";
		}`

	expected := "container test: cardinality mismatch: invalid substatement 'configd:normalize'"
	err := getSchemaBuildError(t, schema_snippet)

	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("Unexpected error:\n  expect: %s\n  actual: %s",
			expected, err.Error())
	}
}

func TestNormalizeCardinality(t *testing.T) {

	schema_snippet := `
	    typedef foo {
            type int16 {
			    configd:normalize "/usr/bin/normalize";
			    configd:normalize "/usr/bin/normalize2";
		    }
	    }`

	expected := "cardinality mismatch: only one 'configd:normalize' statement is allowed"
	err := getSchemaBuildError(t, schema_snippet)

	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("Unexpected error:\n  expect: %s\n  actual: %s",
			expected, err.Error())
	}
}
