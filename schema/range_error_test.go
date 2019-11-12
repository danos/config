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

func TestDefaultYangRangeErrorMessage(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`leaf intLeaf {
             type int16 {
                 range 10..20;
             }
         }`))

	expect := "Must have value between 10 and 20"

	expectValidationError(
		t,
		schema_text,
		"intLeaf", "1",
		expect,
	)
}

func TestSpecifiedYangRangeErrorMessage(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`leaf intLeaf {
             type int16 {
                 range 10..20 {
                     error-message "YANG override";
                 }
             }
         }`))

	expect := "YANG override"

	expectValidationError(
		t,
		schema_text,
		"intLeaf", "1",
		expect,
	)
}

func TestSpecifiedConfigdErrorMessage(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`leaf intLeaf {
             type int16 {
                 range 10..20 {
                     configd:error-message "Configd override";
                 }
             }
         }`))

	expect := "Configd override"

	expectValidationError(
		t,
		schema_text,
		"intLeaf", "1",
		expect,
	)
}

func TestInheritedConfigdErrorMessage(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`typedef narrowInt {
             type int16 {
                 range 10..20 {
                     error-message "will not see me";
                     configd:error-message "Configd override";
                 }
             }
         }
         leaf intLeaf {
             type narrowInt;
         }`))

	expect := "Configd override"

	expectValidationError(
		t,
		schema_text,
		"intLeaf", "1",
		expect,
	)
}

func TestYangOverrideConfigdErrorMessage(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`typedef narrowInt {
             type int16 {
                 range 10..20 {
                     error-message "will not see me";
                     configd:error-message "or me";
                 }
             }
         }
         leaf intLeaf {
             type narrowInt {
                 range 11..19 {
                     error-message "YANG override";
                 }
             }
         }`))

	expect := "YANG override"

	expectValidationError(
		t,
		schema_text,
		"intLeaf", "1",
		expect,
	)
}

func TestConfigdOverrideConfigdErrorMessage(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`typedef narrowInt {
             type int16 {
                 range 10..20 {
                     error-message "will not see me";
                     configd:error-message "or me";
                 }
             }
         }
         leaf intLeaf {
             type narrowInt {
                 range 11..19 {
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
		"intLeaf", "1",
		expect,
	)
}

func TestUintHasConfigdErrorMessage(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`typedef narrowInt {
             type uint16 {
                 range 10..20 {
                     error-message "will not see me";
                     configd:error-message "or me";
                 }
             }
         }
         leaf intLeaf {
             type narrowInt {
                 range 11..19 {
                     error-message "YANG override";
                     // Make sure echo is processing this
                     configd:error-message "Configd override";
                 }
             }
         }
`))

	// Oddly the backspaces are present in the string and need checked
	expect := "Configd override"

	expectValidationError(
		t,
		schema_text,
		"intLeaf", "1",
		expect,
	)
}

func TestDec64HasConfigdErrorMessage(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`typedef narrowInt {
             type decimal64 {
                 fraction-digits 2;
                 range 1.00..2.00 {
                     error-message "will not see me";
                     configd:error-message "or me";
                 }
             }
         }
         leaf intLeaf {
             type narrowInt {
                 range 1.1..1.90 {
                     error-message "YANG override";
                     // Make sure echo is processing this
                     configd:error-message "Configd override";
                 }
             }
         }
`))

	// Oddly the backspaces are present in the string and need checked
	expect := "Configd override"

	expectValidationError(
		t,
		schema_text,
		"intLeaf", "0.88",
		expect,
	)
}
