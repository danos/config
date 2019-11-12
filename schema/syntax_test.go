// Copyright (c) 2017,2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/danos/mgmterror/errtest"
)

func TestSyntaxPass(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`leaf syntaxLeaf {
             type string {
                 configd:syntax "/bin/true";
             }
         }`))

	expectValidationSuccess(
		t,
		schema_text,
		"syntaxLeaf", "1",
	)
}

func TestSyntaxFail(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`leaf syntaxLeaf {
             type string {
                 configd:syntax "echo \"Bugz\" && /bin/false";
             }
         }`))

	expectValidationError(
		t,
		schema_text,
		"syntaxLeaf", "1",
		errtest.NewSyntaxError(
			t, "/syntaxLeaf/1", "Bugz").RawErrorStrings()...,
	)
}

func TestInheritedFail(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`typedef syntaxCheck {
             type string {
                 configd:syntax "echo \"Bugz\" && /bin/false";
             }
         }
         leaf syntaxLeaf {
             type syntaxCheck;
         }
`))

	expectValidationError(
		t,
		schema_text,
		"syntaxLeaf", "1",
		errtest.NewSyntaxError(
			t, "/syntaxLeaf/1", "Bugz").RawErrorStrings()...,
	)
}

func TestUintSyntaxFail(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`leaf syntaxLeaf {
             type uint16 {
                 configd:syntax "echo \"Bugz\" && /bin/false";
             }
         }`))

	expectValidationError(
		t,
		schema_text,
		"syntaxLeaf", "1",
		errtest.NewSyntaxError(
			t, "/syntaxLeaf/1", "Bugz").RawErrorStrings()...,
	)
}

func TestDec64SyntaxFail(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`leaf syntaxLeaf {
             type decimal64 {
                 fraction-digits 2;
                 configd:syntax "echo \"Bugz\" && /bin/false";
             }
         }`))

	expectValidationError(
		t,
		schema_text,
		"syntaxLeaf", "1",
		errtest.NewSyntaxError(
			t, "/syntaxLeaf/1", "Bugz").RawErrorStrings()...,
	)
}
