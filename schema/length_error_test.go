// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"bytes"
	"fmt"
	"testing"
)

func TestNormalLengthOutOfBounds(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`leaf strLeaf {
             type string {
                 length 10..20;
             }
         }`))

	expect := "Must have length between 10 and 20 characters"

	expectValidationError(
		t,
		schema_text,
		"strLeaf", "1",
		expect,
	)
}

func TestOverriddenLengthOutOfBounds(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`leaf strLeaf {
             type string {
                 length 10..20 {
                     error-message "YANG override";
                 }
             }
         }`))

	expect := "YANG override"

	expectValidationError(
		t,
		schema_text,
		"strLeaf", "1",
		expect,
	)
}

func TestConfigdOverriddenLengthOutOfBounds(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`leaf strLeaf {
             type string {
                 length 10..20 {
                     configd:error-message "Configd override";
                 }
             }
         }`))

	expect := "Configd override"

	expectValidationError(
		t,
		schema_text,
		"strLeaf", "1",
		expect,
	)
}

func TestInheritedLengthOutOfBounds(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`typedef narrowInt {
             type string {
                 length 10..20 {
                     error-message "will not see me";
                     configd:error-message "Configd override";
                 }
             }
         }
         leaf strLeaf {
             type narrowInt;
         }
`))

	expect := "Configd override"

	expectValidationError(
		t,
		schema_text,
		"strLeaf", "1",
		expect,
	)
}

func TestOverrideInheritedLengthOutOfBounds(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`typedef narrowInt {
             type string {
                 length 10..20 {
                     error-message "will not see me";
                     configd:error-message "or me";
                 }
             }
         }
         leaf strLeaf {
             type narrowInt {
                 length 11..19 {
                     error-message "YANG override";
                 }
             }
         }
`))

	expect := "YANG override"

	expectValidationError(
		t,
		schema_text,
		"strLeaf", "1",
		expect,
	)
}

func TestConfigdOverrideInheritedLengthOutOfBounds(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`typedef narrowInt {
             type string {
                 length 10..20 {
                     error-message "will not see me";
                     configd:error-message "or me";
                 }
             }
         }
         leaf strLeaf {
             type narrowInt {
                 length 11..19 {
                     error-message "YANG override";
                     // Make sure echo is processing this
                     configd:error-message "Configd override";
                 }
             }
         }
`))

	expect := "Configd override"

	expectValidationError(
		t,
		schema_text,
		"strLeaf", "1",
		expect,
	)
}
