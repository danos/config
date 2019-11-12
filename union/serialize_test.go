// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package union

import (
	"strings"
	"testing"

	"github.com/danos/config/auth"
)

const test_schema = `container top {
		list outer {
			key outer_key;
			leaf outer_key {
				type string;
			}
			list inner {
				key inner_key;
				leaf inner_key {
					type string;
				}
				leaf my_leaf {
					type string;
				}
			}
		}
	}`

func verifyUnmarshal(
	t *testing.T,
	input, testPath, encoding string,
) Node {
	var root Node
	var err error
	switch encoding {
	case "xml":
		root, err = UnmarshalXML(newTestSchema(t, test_schema), []byte(input))
	case "json":
		root, err = UnmarshalJSON(newTestSchema(t, test_schema), []byte(input))
	case "netconf":
		t.Fatalf("'netconf' encoding only applies to Marshal not Unmarshal!")
	}
	if err != nil {
		t.Fatalf("Unexpected failure (unmarshal %s): %s\n", encoding, err)
	}

	testNode := root

	for _, elem := range strings.Split(testPath, "/") {
		if elem == "" {
			break
		}
		testNode = testNode.Child(elem)
		if testNode == nil {
			t.Fatalf("Invalid test path '%s', elem '%s'", testPath, elem)
		}
	}

	return testNode
}

func verifyUnmarshalAndRemarshalFromDescendantPass(
	t *testing.T,
	input, testPath string,
	inputEnc string, descendant []string, outputEnc, expected string,
	options ...UnionOption,
) {
	var err error

	testNode := verifyUnmarshal(t, input, testPath, inputEnc)

	// Do marshal from specified descendant
	if descendant != nil {
		testNode, err = testNode.descendant(descendant, []string{})
		if err != nil {
			t.Fatal(err)
		}
	}

	actual, err := testNode.Marshal("data", outputEnc, options...)
	if err != nil {
		t.Fatalf("Unexpected marshalling error: %s\n", err)
	}
	if string(actual) != expected {
		t.Errorf("Re-encoded %s does not match.\n   expect=%s\n   actual=%s",
			outputEnc, expected, string(actual))
	}
}

func verifyUnmarshalAndRemarshalPass(
	t *testing.T,
	input, testPath string,
	inputEnc, outputEnc, expected string,
) {
	verifyUnmarshalAndRemarshalFromDescendantPass(
		t, input, testPath, inputEnc, nil, outputEnc, expected, IncludeDefaults)
}

const xmlInput = `<data>` +
	`<top xmlns="urn:vyatta.com:test:union">` +
	`<outer xmlns="urn:vyatta.com:test:union">` +
	`<outer_key xmlns="urn:vyatta.com:test:union">outer_entry</outer_key>` +
	`<inner xmlns="urn:vyatta.com:test:union">` +
	`<inner_key xmlns="urn:vyatta.com:test:union">inner_entry</inner_key>` +
	`<my_leaf xmlns="urn:vyatta.com:test:union">some value</my_leaf>` +
	`</inner></outer></top></data>`

func TestXMLSerializationRoot(t *testing.T) {

	expected := `<data>` +
		`<top xmlns="urn:vyatta.com:test:union">` +
		`<outer xmlns="urn:vyatta.com:test:union">` +
		`<outer_key xmlns="urn:vyatta.com:test:union">outer_entry</outer_key>` +
		`<inner xmlns="urn:vyatta.com:test:union">` +
		`<inner_key xmlns="urn:vyatta.com:test:union">inner_entry</inner_key>` +
		`<my_leaf xmlns="urn:vyatta.com:test:union">some value</my_leaf>` +
		`</inner></outer></top></data>`

	verifyUnmarshalAndRemarshalPass(t, xmlInput, "", "xml", "xml", expected)
}

func TestXMLSerializationTop(t *testing.T) {

	expected := `<data>` +
		`<top xmlns="urn:vyatta.com:test:union">` +
		`<outer xmlns="urn:vyatta.com:test:union">` +
		`<outer_key xmlns="urn:vyatta.com:test:union">outer_entry</outer_key>` +
		`<inner xmlns="urn:vyatta.com:test:union">` +
		`<inner_key xmlns="urn:vyatta.com:test:union">inner_entry</inner_key>` +
		`<my_leaf xmlns="urn:vyatta.com:test:union">some value</my_leaf>` +
		`</inner></outer></top></data>`

	verifyUnmarshalAndRemarshalPass(t, xmlInput, "top", "xml", "xml", expected)
}

func TestXMLSerializationList(t *testing.T) {

	expected := `<data>` +
		`<outer xmlns="urn:vyatta.com:test:union">` +
		`<outer_key xmlns="urn:vyatta.com:test:union">outer_entry</outer_key>` +
		`<inner xmlns="urn:vyatta.com:test:union">` +
		`<inner_key xmlns="urn:vyatta.com:test:union">inner_entry</inner_key>` +
		`<my_leaf xmlns="urn:vyatta.com:test:union">some value</my_leaf>` +
		`</inner></outer></data>`

	verifyUnmarshalAndRemarshalPass(t, xmlInput, "top/outer", "xml", "xml",
		expected)
}

func TestXMLSerializationListInsideList(t *testing.T) {

	expected := `<data>` +
		`<inner xmlns="urn:vyatta.com:test:union">` +
		`<inner_key xmlns="urn:vyatta.com:test:union">inner_entry</inner_key>` +
		`<my_leaf xmlns="urn:vyatta.com:test:union">some value</my_leaf>` +
		`</inner></data>`

	verifyUnmarshalAndRemarshalPass(t,
		xmlInput, "top/outer/outer_entry/inner", "xml", "xml", expected)
}

func TestXMLSerializationListNonKeyEntryInsideOtherList(t *testing.T) {

	expected := `<data>` +
		`<my_leaf xmlns="urn:vyatta.com:test:union">some value</my_leaf>` +
		`</data>`

	verifyUnmarshalAndRemarshalPass(t,
		xmlInput, "top/outer/outer_entry/inner/inner_entry/my_leaf", "xml",
		"xml", expected)
}

type netconfTest struct {
	name,
	path,
	expected string
}

func TestNETCONFSerialization(t *testing.T) {

	// yang/data/encoding:UnmarshalXML() cannot handle multiple list
	// entries, so for now we use JSON input.
	netconfJSONInput := `
			{"top":
			{"outer":[
				{"outer_key":"outer2",
					"inner":[
						{"inner_key":"in2", "my_leaf":"aValue2"}
					]},
				{"outer_key":"outer1",
					"inner":[
						{"inner_key":"in1", "my_leaf":"aValue1"}
					]}
			]}}`

	expected1OuterEntry := `<data>` +
		`<top xmlns="urn:vyatta.com:test:union">` +
		`<outer xmlns="urn:vyatta.com:test:union">` +
		`<outer_key xmlns="urn:vyatta.com:test:union">outer1</outer_key>` +
		`<inner xmlns="urn:vyatta.com:test:union">` +
		`<inner_key xmlns="urn:vyatta.com:test:union">in1</inner_key>` +
		`<my_leaf xmlns="urn:vyatta.com:test:union">aValue1</my_leaf>` +
		`</inner>` +
		`</outer>` +
		`</top></data>`

	expected2OuterEntries := `<data>` +
		`<top xmlns="urn:vyatta.com:test:union">` +
		`<outer xmlns="urn:vyatta.com:test:union">` +
		`<outer_key xmlns="urn:vyatta.com:test:union">outer1</outer_key>` +
		`<inner xmlns="urn:vyatta.com:test:union">` +
		`<inner_key xmlns="urn:vyatta.com:test:union">in1</inner_key>` +
		`<my_leaf xmlns="urn:vyatta.com:test:union">aValue1</my_leaf>` +
		`</inner>` +
		`</outer>` +
		`<outer xmlns="urn:vyatta.com:test:union">` +
		`<outer_key xmlns="urn:vyatta.com:test:union">outer2</outer_key>` +
		`<inner xmlns="urn:vyatta.com:test:union">` +
		`<inner_key xmlns="urn:vyatta.com:test:union">in2</inner_key>` +
		`<my_leaf xmlns="urn:vyatta.com:test:union">aValue2</my_leaf>` +
		`</inner>` +
		`</outer>` +
		`</top></data>`

	tests := []netconfTest{
		{
			name:     "root",
			path:     "",
			expected: expected2OuterEntries,
		},
		{
			name:     "top",
			path:     "top",
			expected: expected2OuterEntries,
		},
		{
			name:     "list",
			path:     "top/outer",
			expected: expected2OuterEntries,
		},
		{
			name:     "list entry",
			path:     "top/outer/outer1",
			expected: expected1OuterEntry,
		},
		{
			name:     "list inside other list",
			path:     "top/outer/outer1/inner",
			expected: expected1OuterEntry,
		},
		{
			name:     "key leaf in list inside other list",
			path:     "top/outer/outer1/inner/in1",
			expected: expected1OuterEntry,
		},
		{
			name:     "non-key leaf in list inside other list",
			path:     "top/outer/outer1/inner/in1/my_leaf",
			expected: expected1OuterEntry,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			verifyUnmarshalAndRemarshalPass(t,
				netconfJSONInput, test.path, "json", "netconf", test.expected)
		})
	}
}

const xmlInputWithoutMyLeaf = `<data>` +
	`<top xmlns="urn:vyatta.com:test:union">` +
	`<outer xmlns="urn:vyatta.com:test:union">` +
	`<outer_key xmlns="urn:vyatta.com:test:union">outer_entry</outer_key>` +
	`<inner xmlns="urn:vyatta.com:test:union">` +
	`<inner_key xmlns="urn:vyatta.com:test:union">inner_entry</inner_key>` +
	`</inner></outer></top></data>`

func TestXMLSerializationAuthBasicAllow(t *testing.T) {
	auther := newTestAuther(auth.TestAutherAllowAll(), true)

	verifyUnmarshalAndRemarshalFromDescendantPass(
		t, xmlInput, "", "xml", nil, "xml", xmlInput, Authorizer(auther))
}

func TestXMLSerializationAuthBasicDeny(t *testing.T) {
	auther := newTestAuther(auth.TestAutherDenyAll(), true)

	verifyUnmarshalAndRemarshalFromDescendantPass(
		t, xmlInput, "", "xml", nil, "xml", "<data></data>", Authorizer(auther))
}

func TestXMLSerializationAuthFilter(t *testing.T) {
	auther := newTestAuther(
		auth.NewTestAuther(
			auth.NewTestRule(auth.Deny, auth.AllOps, "/top/outer/*/inner/*/my_leaf"),
			auth.NewTestRule(auth.Allow, auth.AllOps, "*"),
		), true)

	// Check filtering works as expected when marshalling from the root node
	verifyUnmarshalAndRemarshalFromDescendantPass(
		t, xmlInput, "", "xml", nil, "xml", xmlInputWithoutMyLeaf, Authorizer(auther))

	// And also from a descendant node
	verifyUnmarshalAndRemarshalFromDescendantPass(
		t, xmlInput, "", "xml", []string{"top"}, "xml", xmlInputWithoutMyLeaf, Authorizer(auther))
}
