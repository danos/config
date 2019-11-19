// Copyright (c) 2019, AT&T Intellectual Property.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package union

import (
	"testing"
)

const unionSchemaTemplate = `
	module test-yang-compile {
	namespace "urn:vyatta.com:test:yang-compile";
	prefix test;
	organization "AT&T";
	revision 2019-10-07 {
		description "Test JSON union encoding";
	}

	typedef anintunion {
		type union {
			type int8;
			type int32;
			type int64;
		}
	}

	typedef unioni8i32i64 {
		type anintunion;
	}
	typedef unioni64i32i8 {
		type union {
			type int64;
			type int32;
			type int8;
		}
	}
	typedef unioni8bs {
		type union {
			type uint8;
			type boolean;
			type string;
		}
	}
	typedef unionsbi {
		type union {
			type string;
			type boolean;
			type uint16;
		}
	}
	typedef unionstrlenib {
		type union {
			type string {
				length 5..10;
			}
			type boolean;
			type int32;
		}
	}
	typedef unioninunion {
		type union {
			type union {
				type anintunion;
			}
			type string;
		}
	}
	container testcontainer {
		leaf testunioni8i32i64 {
			type unioni8i32i64;
		}
		leaf testunioni64i32i8 {
			type unioni64i32i8;
		}
		leaf testunioni8bs {
			type unioni8bs;
		}
		leaf testunionsbi {
			type unionsbi;
		}
		leaf testunionstrlenib {
			type unionstrlenib;
		}
		leaf testunioninunion {
			type unioninunion;
		}
	}
}`

// Multiple integers in a union, uint8/16/32 unquoted
// uint64 quoted for RFC7951
// String with a min length is ignored for later integer
func TestMatchesIntegerOutput(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"testunioni8i32i64":32,
			"testunioni8bs":32,
			"testunioni64i32i8":32,
			"testunioninunion":32,
			"testunionstrlenib":32
		}
	}`)

	var outputJSON = `
		{
		"testcontainer":{
			"testunioni8bs":32,
			"testunioni8i32i64":32,
			"testunioni64i32i8":32,
			"testunioninunion":32,
			"testunionstrlenib":32
		}
	}`

	var outputRFC7951 = `
		{
		"test-yang-compile:testcontainer":{
			"testunioni8bs":32,
			"testunioni8i32i64":32,
			"testunioni64i32i8":"32",
			"testunioninunion":32,
			"testunionstrlenib":32
		}
	}`

	expectedPassTestWithTemplate(t, inputMessage, outputJSON, unionSchemaTemplate)
	expectedPassRFC7951WithTemplate(t, inputMessage, outputRFC7951, unionSchemaTemplate)
}

// an int64 value quoted for rfc7951 due to size, in preference to earlier int types in union
func TestMatchesInteger64Output(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"testunioni8i32i64":10000000000000,
			"testunioni64i32i8":32,
			"testunioninunion":10000000000000
		}
	}`)

	var outputJSON = `
		{
		"testcontainer":{
			"testunioni8i32i64":10000000000000,
			"testunioni64i32i8":32,
			"testunioninunion":10000000000000
		}
	}`

	var outputRFC7951 = `
		{
		"test-yang-compile:testcontainer":{
			"testunioni8i32i64":"10000000000000",
			"testunioni64i32i8":"32",
			"testunioninunion":"10000000000000"
		}
	}`

	expectedPassTestWithTemplate(t, inputMessage, outputJSON, unionSchemaTemplate)
	expectedPassRFC7951WithTemplate(t, inputMessage, outputRFC7951, unionSchemaTemplate)
}

// Values, which may appear as numbers, but correctly match a string within a union
// the encoded values are quoted
func TestMatchesStringOutput(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"testunioni8bs":"12345",
			"testunioninunion":"stringval",
			"testunionsbi":"12345",
			"testunionstrlenib":"12345"
		}
	}`)

	var outputJSON = `
		{
		"testcontainer":{
			"testunioni8bs":"12345",
			"testunioninunion":"stringval",
			"testunionsbi":"12345",
			"testunionstrlenib":"12345"
		}
	}`

	var outputRFC7951 = `
		{
		"test-yang-compile:testcontainer":{
			"testunioni8bs":"12345",
			"testunioninunion":"stringval",
			"testunionsbi":"12345",
			"testunionstrlenib":"12345"
		}
	}`

	expectedPassTestWithTemplate(t, inputMessage, outputJSON, unionSchemaTemplate)
	expectedPassRFC7951WithTemplate(t, inputMessage, outputRFC7951, unionSchemaTemplate)
}

// Values which match a string in a union because they are not a number
// the encoded values are quoted
func TestMatchesStringValueOutput(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"testunioni8bs":"foo",
			"testunioninunion":"bar",
			"testunionsbi":"baz",
			"testunionstrlenib":"foobar"
		}
	}`)

	var outputJSON = `
		{
		"testcontainer":{
			"testunioni8bs":"foo",
			"testunioninunion":"bar",
			"testunionsbi":"baz",
			"testunionstrlenib":"foobar"
		}
	}`

	var outputRFC7951 = `
		{
		"test-yang-compile:testcontainer":{
			"testunioni8bs":"foo",
			"testunioninunion":"bar",
			"testunionsbi":"baz",
			"testunionstrlenib":"foobar"
		}
	}`

	expectedPassTestWithTemplate(t, inputMessage, outputJSON, unionSchemaTemplate)
	expectedPassRFC7951WithTemplate(t, inputMessage, outputRFC7951, unionSchemaTemplate)
}

// Boolean values match a boolean within a union
// encoded value is not quoted
func TestMatchesBooleanOutput(t *testing.T) {
	var inputMessage = []byte(`
		{
		"testcontainer":{
			"testunioni8bs":false,
			"testunionstrlenib":true
		}
	}`)

	var outputJSON = `
		{
		"testcontainer":{
			"testunioni8bs":false,
			"testunionstrlenib":true
		}
	}`

	var outputRFC7951 = `
		{
		"test-yang-compile:testcontainer":{
			"testunioni8bs":false,
			"testunionstrlenib":true
		}
	}`

	expectedPassTestWithTemplate(t, inputMessage, outputJSON, unionSchemaTemplate)
	expectedPassRFC7951WithTemplate(t, inputMessage, outputRFC7951, unionSchemaTemplate)
}
