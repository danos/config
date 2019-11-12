// Copyright (c) 2017,2019, AT&T Intellectual Property. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/danos/mgmterror/errtest"
)

const (
	invalidLength    = "Invalid length"
	isNotValidSuffix = " is not valid"
	noInfoExpected   = ""
)

func testValidationError(
	t *testing.T,
	expectToFind bool,
	schema_text *bytes.Buffer,
	nodePath string,
	expMsg,
	expPath,
	expInfoVal string,
) {
	ms, err := GetConfigSchema(schema_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	pathSlice := strings.Split(nodePath, "/")
	err = ms.Validate(nil, []string{}, pathSlice)
	if err == nil {
		t.Fatalf("Unexpected success\n")
	}

	errtest.CheckMsg(t, err, expMsg)
	errtest.CheckPath(t, err, expPath)
	errtest.CheckInfo(t, err, expInfoVal)
}

// PathError format:
//
// e.Message = Path is invalid
// e.Path = path/to/invalid/node OR path/to/level/above/invalid/node?
// e.Info = invalid node?
//
// Type (unless it's top level invalid) is UnknownElementApplicationError
//
// Need to find out how path errors are being done now as UEAE for other places
// stores path in 2 parts (path to above, invalid node).
/*
* Pre VCI-collapse output:

 [unterfaces]: Configuration path: [] is not valid
 Path is invalid

 [interfaces lopbock]: Configuration path: [interfaces] is not valid
 Path is invalid

 [interfaces loopback lo999999]: Configuration path: interfaces loopback [lo999999] is not valid
 name must be lo or loN, N in [1-99999]
 Value validation failed

 [interfaces loopback lo2 description]: Configuration path: interfaces loopback lo2 [description] is not valid
 Node requires a value

 [service ssh port 0]: Configuration path: service ssh port [0] is not valid
 Port number must be in range 1 to 65535
 Value validation failed

 [service ssh port stringy]: Configuration path: service ssh port [stringy] is not valid
 Port number must be in range 1 to 65535 Value validation failed
*/

func TestPathTopLevelInvalid(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`container topCont {
             presence "test container needs presence";
         }`))

	testValidationError(
		t, true,
		schema_text,
		"bottomCont",
		"[bottomCont]"+isNotValidSuffix,
		"",
		"bottomCont",
	)
}

func TestPathContainerInvalidChild(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`container topCont {
			leaf subLeaf {
				type string;
			}
         }`))

	testValidationError(
		t, true,
		schema_text,
		"topCont/anotherLeaf",
		"topCont [anotherLeaf]"+isNotValidSuffix,
		"/topCont",
		"anotherLeaf",
	)
}

func TestPathListInvalidLeaf(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`container topCont {
			list aList {
				key name;
				leaf name {
					type string;
				}
             }
         }`))

	testValidationError(
		t, true,
		schema_text,
		"topCont/aList/entry1/nonExistentLeaf/leafValue",
		"topCont aList entry1 [nonExistentLeaf]"+isNotValidSuffix,
		"/topCont/aList/entry1",
		"nonExistentLeaf",
	)
}

func TestPathInvalidLeafUnexpectedExtraValue(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`container topCont {
			leaf name {
				type string;
			}
         }`))

	testValidationError(
		t, true,
		schema_text,
		"topCont/name/leafName/extra",
		"topCont name leafName [extra]"+isNotValidSuffix,
		"/topCont/name/leafName",
		"extra",
	)
}

func TestPathInvalidEmptyLeafValue(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`container topCont {
			leaf emptyLeaf {
				type empty;
			}
         }`))

	testValidationError(
		t, true,
		schema_text,
		"topCont/emptyLeaf/notEmpty",
		"topCont emptyLeaf [notEmpty]"+isNotValidSuffix,
		"/topCont/emptyLeaf",
		"notEmpty",
	)
}

func TestPathInvalidLeafListUnexpectedExtraValue(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`container topCont {
			leaf-list name {
				type string;
			}
         }`))

	testValidationError(
		t, true,
		schema_text,
		"topCont/name/leafName/extra",
		"topCont name leafName [extra]"+isNotValidSuffix,
		"/topCont/name/leafName",
		"extra",
	)
}

func TestPathTopLevelInvalidLength(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`leaf strLeaf {
             type string {
                 length 10..20;
             }
         }`))

	testValidationError(
		t, true,
		schema_text,
		"strLeaf/1",
		"Must have length between 10 and 20 characters",
		"/strLeaf/1",
		invalidLength,
	)
}

func TestPathLowerLevelInvalidLength(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`container top {
			container lower {
				leaf bottom {
					type string {
						length 10..20;
					}
				}
             }
         }`))

	testValidationError(
		t, true,
		schema_text,
		"top/lower/bottom/999",
		"Must have length between 10 and 20 characters",
		"/top/lower/bottom/999",
		invalidLength,
	)
}

func TestPathLowerLevelInvalidValue(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`container top {
			container lower {
				leaf bottom {
					type int8 {
						range 1..10;
					}
				}
             }
         }`))

	testValidationError(
		t, true,
		schema_text,
		"top/lower/bottom/999",
		"Must have value between 1 and 10",
		"/top/lower/bottom/999",
		noInfoExpected,
	)
}
