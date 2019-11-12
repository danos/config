// Copyright (c) 2017-2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package union

import (
	"bytes"
	"strings"
	"testing"

	"github.com/danos/config/data"
	"github.com/danos/mgmterror/errtest"
	"github.com/danos/yang/data/encoding"
)

const jsonSchemaTemplate = `
	module test-yang-compile {
	namespace "urn:vyatta.com:test:yang-compile";
	prefix test;
	organization "Brocade Communications Systems, Inc.";
	revision 2014-12-29 {
		description "Test schema for configd";
	}

	grouping myGroupNested {
		container groupContainer {
			presence "To limit scope of mandatory";
			leaf groupLeaf {
				type string;
				mandatory "true";
			}
			list groupList {
				key "name";
				leaf name {
					type string;
				}
				leaf value {
					type int32;
				}
				list innerGroupList {
					key "innerName";
					leaf innerName {
                        type string;
                    }
				}
			}
			leaf-list groupLeafList {
				type string;
			}
		}
	}

	grouping myGroupMandNode {
		container groupMandContainer {
			presence "To limit scope of mandatory";
			leaf groupMandLeaf {
				type string;
				mandatory "true";
			}
			leaf groupOptLeaf {
				type string;
			}
		}
	}

	grouping myGroupMinMaxElemLeafList {
		container groupMinMaxElemLeafList {
			presence "To limit scope of mandatory";
			leaf-list testMinMaxElemLeafList {
				type string;
				min-elements 2;
				max-elements 3;
				ordered-by user;
			}
		}
	}

	grouping myGroupNonStringLists {
		container groupNonStringLists {
			list intKeyList {
				key id;
				leaf id {
					type int8;
				}
				leaf name {
					type string;
				}
			}
			list dec64KeyList {
				key id;
				leaf id {
					type decimal64 {
                        fraction-digits 10;
                    }
				}
				leaf name {
					type string;
				}
			}
			list boolKeyList {
				key id;
				leaf id {
					type boolean;
				}
				leaf name {
					type string;
				}
			}
		}
	}

	container presencecontainer {
		presence "For testing presence";
		leaf config {
			type string;
		}
		container statePresenceContainer {
			presence "State presence container";
			leaf state {
				type string;
			}
		}
		leaf state {
			type string;
			config false;
		}
	}

	container testcontainer {
		leaf testboolean {
			type boolean;
	        default false;
		}
		leaf teststring {
			type string;
		}
		leaf testdec64 {
			type decimal64 {
				fraction-digits 6;
			}
		}
		leaf testempty {
			type empty;
		}
		leaf testint32 {
			type int32;
		    default "33";
		}
		leaf-list testleaflist {
			type string;
			ordered-by user;
		}
		leaf-list nonStringleaflist {
			type int8;
			ordered-by user;
		}
		list testlist {
			key name;
			unique "bar2 bar3";
			leaf name {
				type string;
			}
			leaf bar {
				type empty;
			}
			leaf bar2 {
				type uint8;
			    default 66;
			}
			leaf bar3 {
				type string;
			}
		}
		uses myGroupNested;
		uses myGroupMandNode;
		uses myGroupMinMaxElemLeafList;
		uses myGroupNonStringLists;
	}
	list testlist {
		key name;
		leaf name {
			type string;
		}
		leaf foo {
			type string;
		}
	}
}`

// Useful when unsure what raw JSON should look like ... eg:
//	testPath := []string{"testcontainer", "groupContainer", "groupLeaf", "foo"}
//	setPathAndDumpJSON(t, testPath)
//  t.Errorf("Dummy fail call so printfs are dumped to console.")
func setPathAndDumpJSON(t *testing.T, path []string) {
	sch := bytes.NewBufferString(jsonSchemaTemplate)
	compiledSchema, err := getSchema(sch.Bytes())
	if err != nil {
		return
	}

	// Create data tree which we can put decoded message into.
	can, run := data.New("root"), data.New("root")
	ut := NewNode(can, run, compiledSchema, nil, 0)
	if ut == nil {
		return
	}

	ut.Set(nil, path)
	outb := ut.ToJSON()

	outs := bytes.NewBuffer(outb).String()
	t.Logf("Output JSON is:\n%s\n", outs)
}

// It's useful to be able to pretty print JSON below for readability, but
// we need to uglify it (remove whitespace/WS) before comparing to machine
// generated JSON.
func stripWS(pretty string) string {
	r := strings.NewReplacer(" ", "", "\n", "", "\t", "")
	return r.Replace(pretty)
}

// Common part of running tests, including compiling schema.
func runTest(inputMsg []byte) (ut Node, err error) {
	sch := bytes.NewBufferString(jsonSchemaTemplate)
	compiledSchema, err := getSchema(sch.Bytes())
	if err != nil {
		return nil, err
	}
	return UnmarshalJSON(compiledSchema, inputMsg)
}

func assertTreeMatchesJson(t *testing.T, root Node, expectedJSON string) {

	outb := root.ToJSON(IncludeDefaults)
	outs := bytes.NewBuffer(outb).String()
	expectedJSON = stripWS(expectedJSON)
	if outs != expectedJSON {
		t.Fatalf("Failed to get expected JSON\nExp: %s\nGot: %s\n",
			expectedJSON, outb)
	}
}

func assertTreeMatchesRFC7951(t *testing.T, root Node, expectedRFC7951 string) {

	outb := root.ToRFC7951(IncludeDefaults)
	outs := bytes.NewBuffer(outb).String()
	expectedRFC7951 = stripWS(expectedRFC7951)
	if outs != expectedRFC7951 {
		t.Fatalf("Failed to get expected RFC7951\nExp: %s\nGot: %s\n",
			expectedRFC7951, outb)
	}
}

// Run a test we expect to pass by decoding the input message and verifying
// that the re-encoded (as JSON) matches expectations.
func expectedPassTest(t *testing.T, inputMsg []byte, expJSON string,
) {
	sch := bytes.NewBufferString(jsonSchemaTemplate)
	compiledSchema, err := getFullSchema(sch.Bytes())
	if err != nil {
		t.Fatalf("Failed to compile schema")
		return
	}
	ut, err := UnmarshalJSON(compiledSchema, inputMsg)
	if err != nil {
		t.Errorf("Failed to decode JSON: %s\n", err)
		return
	}

	assertTreeMatchesJson(t, ut, expJSON)

	ut, err = NewUnmarshaller(encoding.JSON).
		Unmarshal(compiledSchema, inputMsg)
	if err != nil {
		t.Errorf("Failed to decode JSON: %s\n", err)
		return
	}

	assertTreeMatchesJson(t, ut, expJSON)
}

func expectedPassRFC7951(
	t *testing.T,
	inputMsg []byte,
	expRFC7951 string,
) {
	sch := bytes.NewBufferString(jsonSchemaTemplate)
	compiledSchema, err := getSchema(sch.Bytes())
	if err != nil {
		t.Fatalf("Failed to compile schema")
		return
	}
	ut, err := UnmarshalRFC7951(compiledSchema, inputMsg)
	if err != nil {
		t.Errorf("Failed to decode JSON: %s\n", err)
		return
	}

	assertTreeMatchesRFC7951(t, ut, expRFC7951)

	ut, err = NewUnmarshaller(encoding.RFC7951).
		Unmarshal(compiledSchema, inputMsg)
	if err != nil {
		t.Errorf("Failed to decode JSON: %s\n", err)
		return
	}

	assertTreeMatchesRFC7951(t, ut, expRFC7951)
}

// Run a test we expect to fail by decoding the input message and verifying
// that we get the expected error message.
func expectedFailTest(t *testing.T, inputMsg []byte, errMsgs []string) {
	_, err := runTest(inputMsg)
	if err == nil {
		t.Errorf("Test passed but should have failed with: \n%s\n", errMsgs)
		return
	}

	if len(errMsgs) == 0 {
		t.Errorf("Must specify at least one expected error.\n")
		return
	}

	for _, errMsg := range errMsgs {
		if len(errMsg) == 0 {
			t.Errorf("Expected error is empty string.\n")
			return
		}

		if !strings.Contains(err.Error(), errMsg) {
			t.Errorf("Test failed with: \n%s\nShould have failed with:\n%s\n",
				err, errMsg)
			return
		}
	}
}

// Good first test - all we do is verify empty JSON input gives expected
// default output.
func TestDefaultOuptut(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"teststring":"foo"
		}
	}`)

	var outputJSON = `
		{
		"testcontainer":{
            "testboolean":false,
            "testint32":33,
			"teststring":"foo"
		}
	}`

	var outputRFC7951 = `
		{
		"test-yang-compile:testcontainer":{
            "testboolean":false,
            "testint32":33,
			"teststring":"foo"
		}
	}`

	expectedPassTest(t, inputMessage, outputJSON)
	expectedPassRFC7951(t, inputMessage, outputRFC7951)
}

func TestLeafStringDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"teststring":"foo"
		}
	}`)

	var outputJSON = `
		{
		"testcontainer":{
            "testboolean":false,
            "testint32":33,
			"teststring":"foo"
		}
	}`

	var outputRFC7951 = `
		{
		"test-yang-compile:testcontainer":{
            "testboolean":false,
            "testint32":33,
			"teststring":"foo"
		}
	}`

	expectedPassTest(t, inputMessage, outputJSON)
	expectedPassRFC7951(t, inputMessage, outputRFC7951)
}

func TestLeafZeroLengthStringDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"teststring":""
		}
	}`)

	var outputJSON = `
		{
		"testcontainer":{
            "testboolean":false,
            "testint32":33,
			"teststring":""
		}
	}`

	var outputRFC7951 = `
		{
		"test-yang-compile:testcontainer":{
            "testboolean":false,
            "testint32":33,
			"teststring":""
		}
	}`

	expectedPassTest(t, inputMessage, outputJSON)
	expectedPassRFC7951(t, inputMessage, outputRFC7951)
}

// Doubles as boolean test
func TestLeafStringDefaultOverrideDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"testboolean":true
		}
	}`)

	var outputJSON = `
		{
		"testcontainer":{
            "testboolean":true,
            "testint32":33
		}
	}`

	var outputRFC7951 = `
		{
		"test-yang-compile:testcontainer":{
            "testboolean":true,
            "testint32":33
		}
	}`

	expectedPassTest(t, inputMessage, outputJSON)
	expectedPassRFC7951(t, inputMessage, outputRFC7951)
}

func TestBooleanUnquotedDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"testboolean":true
		}
	}`)

	var outputJSON = `
		{
		"testcontainer":{
            "testboolean":true,
            "testint32":33
		}
	}`

	var outputRFC7951 = `
		{
		"test-yang-compile:testcontainer":{
            "testboolean":true,
            "testint32":33
		}
	}`

	expectedPassTest(t, inputMessage, outputJSON)
	expectedPassRFC7951(t, inputMessage, outputRFC7951)
}

func TestLeafBooleanInvalidDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"testboolean":"66.66"
		}
	}`)

	errorMessages := []string{"Must have one of the following values",
		"true", "false"}

	expectedFailTest(t, inputMessage, errorMessages)
}

func TestEmptyLeafDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"testempty":null
		}
	}`)

	var outputMessage = `
		{
		"testcontainer":{
            "testboolean":false,
			"testempty":null,
            "testint32":33
		}
	}`
	var outputMessageRFC7951 = `
		{
		"test-yang-compile:testcontainer":{
            "testboolean":false,
			"testempty":[null],
            "testint32":33
		}
	}`

	expectedPassTest(t, inputMessage, outputMessage)
	expectedPassRFC7951(t, inputMessage, outputMessageRFC7951)
}

func TestLeafIntDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"testint32":666
		}
	}`)

	var outputMessage = `
		{
		"testcontainer":{
            "testboolean":false,
            "testint32":666
		}
	}`

	expectedPassTest(t, inputMessage, outputMessage)
}

func TestLeafIntUnquotedDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"testint32":666
		}
	}`)

	var outputMessage = `
		{
		"testcontainer":{
            "testboolean":false,
            "testint32":666
		}
	}`

	expectedPassTest(t, inputMessage, outputMessage)
}

func TestLeafIntInvalidDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"testint32":"66.66"
		}
	}`)

	errorMessages := errtest.NewInvalidTypeError(t,
		"/testcontainer/testint32/66.66", "an int32").RawErrorStrings()
	expectedFailTest(t, inputMessage, errorMessages)
}

// Check it's no different when we use a string not a dec64
func TestLeafIntInvalid2Decode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"testint32":"foo"
		}
	}`)

	errorMessages := errtest.NewInvalidTypeError(t,
		"/testcontainer/testint32/foo", "an int32").RawErrorStrings()
	expectedFailTest(t, inputMessage, errorMessages)
}

func TestLeafDec64Decode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"testdec64":"77.77"
		}
	}`)

	var outputMessage = `
		{
		"testcontainer":{
            "testboolean":false,
            "testdec64":"77.77",
            "testint32":33
		}
	}`

	expectedPassTest(t, inputMessage, outputMessage)
}

func TestLeafHugeDec64Decode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"testdec64":"1234567890000"
		}
	}`)

	var outputMessage = `
		{
		"testcontainer":{
            "testboolean":false,
            "testdec64":"1234567890000",
            "testint32":33
		}
	}`

	expectedPassTest(t, inputMessage, outputMessage)
}

func TestLeafDec64UnquotedDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"testdec64":"77.77"
		}
	}`)

	var outputMessage = `
		{
		"testcontainer":{
            "testboolean":false,
            "testdec64":"77.77",
            "testint32":33
		}
	}`

	expectedPassTest(t, inputMessage, outputMessage)
}

func TestListDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"testlist":[
				{"name":"nm1","bar":null,"bar2":33},
				{"name":"nm2","bar2":44}
			]
		}
	}`)

	var outputMessage = `
		{
		"testcontainer":{
            "testboolean":false,
            "testint32":33,
			"testlist":[
				{"name":"nm1","bar":null,"bar2":33},
				{"name":"nm2","bar2":44}
			]
		}
	}`

	expectedPassTest(t, inputMessage, outputMessage)
}

const simpleListSchema = `
	module test-yang-compile {
	namespace "urn:vyatta.com:test:yang-compile";
	prefix test;
	organization "Brocade Communications Systems, Inc.";
	revision 2014-12-29 {
		description "Test schema for configd";
	}

	list testlist {
		key name;
		leaf name {
			type string;
		}
		leaf foo {
			type string;
		}
	}
}`

func getNewRootForSchema(t *testing.T, schemaInput string) Node {

	sch := bytes.NewBufferString(simpleListSchema)
	compiledSchema, err := getSchema(sch.Bytes())
	if err != nil {
		t.FailNow()
	}

	root := NewNode(data.New("root"), data.New("root"), compiledSchema, nil, 0)
	if root == nil {
		t.Fatalf("Invalid schema provided")
	}

	return root
}

func decodeJsonIntoNode(t *testing.T, targetNode Node, encodedData []byte) {

	err := unmarshalJSONIntoNode(targetNode, encodedData)
	if err != nil {
		t.Fatalf("Failed to decode JSON: %s\n", err)
	}
}

func TestListDecodeIntoNode(t *testing.T) {

	// Create a new root and list schema
	root := getNewRootForSchema(t, simpleListSchema)
	root.addChild(data.New("testlist"))
	targetNode := root.Child("testlist")

	// Decode a set of data into the list
	var inputMessage = []byte(`
		[{"name":"nm1","foo":"bar"},
		 {"name":"nm2","foo":"bar"}]`)
	decodeJsonIntoNode(t, targetNode, inputMessage)

	// Verify changes are reflected in new tree
	var expected = `
		{
		"testlist":[
			{"name":"nm1","foo":"bar"},
			{"name":"nm2","foo":"bar"}]
	}`
	assertTreeMatchesJson(t, root, expected)
}

const keyOnlyListSchema = `
	module test-yang-compile {
	namespace "urn:vyatta.com:test:yang-compile";
	prefix test;
	organization "Brocade Communications Systems, Inc.";
	revision 2017-03-02 {
		description "Test schema for configd";
	}

	list testlist {
		key name;
		leaf name {
			type string;
		}
	}
}`

func TestListKeyOnlyDecodeIntoNode(t *testing.T) {

	// Create a new root and list schema
	root := getNewRootForSchema(t, simpleListSchema)
	root.addChild(data.New("testlist"))
	targetNode := root.Child("testlist")

	// Decode a set of data into the list
	var inputMessage = []byte(`
		[{"name":"nm1"},
		 {"name":"nm2"}]`)
	decodeJsonIntoNode(t, targetNode, inputMessage)

	// Verify changes are reflected in new tree
	var expected = `
		{
		"testlist":[
			{"name":"nm1"},
			{"name":"nm2"}]
	}`
	assertTreeMatchesJson(t, root, expected)
}

// This checks that we can run more than one state script and have
// the first one create the list and the second one add to it.
func TestListOverriddeEntryInNode(t *testing.T) {

	// Create a new root and list schema
	root := getNewRootForSchema(t, simpleListSchema)
	root.addChild(data.New("testlist"))
	targetNode := root.Child("testlist")

	// Decode a set of data into the list
	var firstInput = []byte(`
		[{"name":"nm1","foo":"bar"},
		 {"name":"nm2","foo":"bar"}]`)
	decodeJsonIntoNode(t, targetNode, firstInput)

	// Decode a second set of data into the list
	var secondInput = []byte(`
		[{"name":"nm1","foo":"wibble"}]`)
	decodeJsonIntoNode(t, targetNode, secondInput)

	// Verify changes are reflected in new tree
	var expected = `
		{
		"testlist":[
			{"name":"nm1","foo":"wibble"},
			{"name":"nm2","foo":"bar"}]
	}`
	assertTreeMatchesJson(t, root, expected)
}

func TestInvalidKeyListDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"testlist":[
				{"nom":"nm1","bar":null,"bar2":33},
				{"nom":"nm2","bar2":44}
			]
		}
	}`)

	errMessages := errtest.NewMissingKeyError(t, "/name").RawErrorStrings()
	expectedFailTest(t, inputMessage, errMessages)
}

func TestListNoKeyDecode(t *testing.T) {
	t.Skipf("Skipping ListNoKeyDecode Test")
}

func TestListNonStringKeyDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"groupNonStringLists":{
				"intKeyList":[
					{"id":11,"name":"nm1"},
					{"id":22,"name":"nm2"}
				],
				"dec64KeyList":[
					{"id":"66.6","name":"nm1"},
					{"id":"66.666","name":"nm2"}
				],
				"boolKeyList":[
					{"id":true,"name":"nm1"},
					{"id":false,"name":"nm2"}
				]
			}
		}
	}`)

	var outputMessage = `
		{
		"testcontainer":{
			"groupNonStringLists":{
				"boolKeyList":[
					{"id":false,"name":"nm2"},
					{"id":true,"name":"nm1"}
				],
				"dec64KeyList":[
					{"id":"66.6","name":"nm1"},
					{"id":"66.666","name":"nm2"}
				],
				"intKeyList":[
					{"id":11,"name":"nm1"},
					{"id":22,"name":"nm2"}
				]
			},
            "testboolean":false,
            "testint32":33
		}
	}`

	expectedPassTest(t, inputMessage, outputMessage)
}

func TestListMultiKeyDecode(t *testing.T) {
	t.Skipf("Skipping ListMultiKeyDecode Test")
}

func TestLeafListDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"testleaflist":["foo","foo2","foo3"]
		}
	}`)

	var outputMessage = `
		{
		"testcontainer":{
            "testboolean":false,
            "testint32":33,
			"testleaflist":["foo","foo2","foo3"]
		}
	}`

	expectedPassTest(t, inputMessage, outputMessage)
}

func TestNonStringLeafListDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"nonStringleaflist":[1,2,3]
		}
	}`)

	var outputMessage = `
		{
		"testcontainer":{
			"nonStringleaflist":[1,2,3],
            "testboolean":false,
            "testint32":33
		}
	}`

	expectedPassTest(t, inputMessage, outputMessage)
}

func TestSimpleNestedDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"groupContainer":{
				"groupList":[
					{"name":"item1","value":11},
					{"name":"item2","value":22}
				],
				"groupLeafList":["foo","foo2","foo3"],
				"groupLeaf":"foo"
			}
		}
	}`)

	var outputMessage = `
		{
		"testcontainer":{
			"groupContainer":{
				"groupLeaf":"foo",
				"groupLeafList":["foo","foo2","foo3"],
				"groupList":[
					{"name":"item1","value":11},
					{"name":"item2","value":22}
				]
			},
            "testboolean":false,
            "testint32":33
		}
	}`

	expectedPassTest(t, inputMessage, outputMessage)
}

func TestDoublyNestedListDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"groupContainer":{
				"groupList":[
					{"name":"item1",
					 "innerGroupList":[
						{"innerName":"innerItem1"}]},
					{"name":"item2","value":22}
				],
				"groupLeaf":"foo"
			}
		}
	}`)

	var outputMessage = `
		{
		"testcontainer":{
			"groupContainer":{
				"groupLeaf":"foo",
				"groupList":[
					{"name":"item1",
					 "innerGroupList":[
						{"innerName":"innerItem1"}]},
					{"name":"item2","value":22}
				]
			},
            "testboolean":false,
            "testint32":33
		}
	}`

	expectedPassTest(t, inputMessage, outputMessage)
}

func TestEmptyPresenceContainerDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
			"presencecontainer":{}
		}`)

	var outputJSON = `
		{
		"presencecontainer":{},
		"testcontainer":{
			"testboolean":false,
			"testint32":33
		}
	}`

	var outputRFC7951 = `
		{
		"test-yang-compile:presencecontainer":{},
		"test-yang-compile:testcontainer":{
			"testboolean":false,
			"testint32":33
		}
	}`

	expectedPassTest(t, inputMessage, outputJSON)
	expectedPassRFC7951(t, inputMessage, outputRFC7951)
}

//
// Tests for overall (multi-line) validation eg cardinality, mandatory,
// unique, min/max elements.
//

// Make sure we are doing validation across whole schema not just per-line
// validation.
func TestMissingMandatoryNodesDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"groupContainer":{
				"groupList":[
					{"name":"item1","value":11},
					{"name":"item2","value":22}
				],
				"groupLeafList":["foo","foo2","foo3"]
			},
			"groupMandContainer":{
				"groupOptLeaf":"foo"
			}
		}
	}`)

	var errorMessages []string
	errorMessages = append(errorMessages,
		errtest.NewMissingMandatoryNodeError(t,
			"/testcontainer/groupContainer/groupLeaf").
			RawErrorStrings()...)
	errorMessages = append(errorMessages,
		errtest.NewMissingMandatoryNodeError(t,
			"/testcontainer/groupMandContainer/groupMandLeaf").
			RawErrorStrings()...)

	expectedFailTest(t, inputMessage, errorMessages)
}

func TestTooFewElementsDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"groupMinMaxElemLeafList":{
				"testMinMaxElemLeafList":["fooTooFew"]
			}
		}
	}`)

	errorMessages := errtest.NewInvalidNumElementsError(t,
		"/testcontainer/groupMinMaxElemLeafList/testMinMaxElemLeafList",
		2, 3).RawErrorStrings()
	expectedFailTest(t, inputMessage, errorMessages)
}

func TestTooManyElementsDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"groupMinMaxElemLeafList":{
				"testMinMaxElemLeafList":["fooTooFew"]
			}
		}
	}`)

	errorMessages := errtest.NewInvalidNumElementsError(t,
		"/testcontainer/groupMinMaxElemLeafList/testMinMaxElemLeafList",
		2, 3).RawErrorStrings()

	expectedFailTest(t, inputMessage, errorMessages)
}

func TestDuplicateNodeDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"testlist":[
				{"name":"nm2","bar2":44,"bar3":"wibble"},
				{"name":"nm2","bar2":55,"bar3":"wibble"}
			]
		}
	}`)

	errorMessages := []string{
		"Node exists",
	}

	expectedFailTest(t, inputMessage, errorMessages)
}

func TestUniqueDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"testlist":[
				{"name":"nm1","bar":null,"bar2":33},
				{"name":"nm2","bar2":44,"bar3":"wibble"},
				{"name":"nm3","bar2":44,"bar3":"wibble"}
			]
		}
	}`)

	errorMessages := errtest.NewNonUniquePathsError(t,
		"/testcontainer/testlist",
		[]string{"name/nm2", "name/nm3"},
		[]string{"bar2/44", "bar3/wibble"}).
		RawErrorStrings()
	expectedFailTest(t, inputMessage, errorMessages)
}

//
// Tests for invalid incoming JSON / valid JSON that doesn't match the schema
//

// Observed by accident that initial implementation crashed on decoding when
// there was a trailing comma, so let's test for it!
func TestTrailingCommaDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"groupContainer":{
				"groupLeaf":"foo",
			}
		}
	}`)

	errMsgs := []string{"invalid character '}' looking for beginning of"}

	expectedFailTest(t, inputMessage, errMsgs)
}

// Leaf with leaf-list type content
func TestBadlyFormattedJsonDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"groupContainer":{
				"groupLeaf":["foo","foo2","foo3"]
			}
		}
	}`)

	errMsgs := []string{
		"More than one entry for non-list",
		": groupLeaf"}

	expectedFailTest(t, inputMessage, errMsgs)
}

func TestNonExistentLeafDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"teststringy":"foo"
		}
	}`)

	errMsgs := errtest.NewSchemaMismatchError(t,
		"/testcontainer/teststringy").RawErrorStrings()
	expectedFailTest(t, inputMessage, errMsgs)
}

func TestNonExistentLeafListJsonDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"groupContainer":{
				"undefinedLeafList":["foo","foo2","foo3"]
			}
		}
	}`)

	errMsgs := errtest.NewSchemaMismatchError(t,
		"/groupContainer/undefinedLeafList").RawErrorStrings()
	expectedFailTest(t, inputMessage, errMsgs)
}

func TestNonExistentListJsonDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"undefinedTestlist":[
				{"name":"nm1","bar":null,"bar2":33},
				{"name":"nm2","bar2":44}
			]
		}
	}`)
	errMsgs := errtest.NewSchemaMismatchError(t,
		"/testcontainer/undefinedTestlist").RawErrorStrings()
	expectedFailTest(t, inputMessage, errMsgs)
}

func TestNonExistentListLeafJsonDecode(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"undefinedTestlist":[
				{"name":"nm1","bar":null,"bar2":33},
				{"name":"nm2","barNone":44}
			]
		}
	}`)
	errMsgs := errtest.NewSchemaMismatchError(t,
		"/testcontainer/undefinedTestlist").RawErrorStrings()
	expectedFailTest(t, inputMessage, errMsgs)
}
