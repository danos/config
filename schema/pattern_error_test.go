// Copyright (c) 2017,2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2014-2017 by Brocade Communications Systems, Inc.
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

func TestStringPatternMismatch(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`leaf strLeaf {
             type string {
                 pattern "[a-z]+";
             }
         }`))

	expect := "Does not match pattern [a-z]+"

	expectValidationError(
		t,
		schema_text,
		"strLeaf", "1",
		expect,
	)
}

func TestStringPatternMismatchWithHelpOnly(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`typedef ipv4-prefix {
			type string {
				pattern '(([0-9]|[1-9][0-9]|1[0-9][0-9]|'
				+ '2[0-4][0-9]|25[0-5])\.){3}'
				+  '([0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-4][0-9]|25[0-5])'
				+ '/(([0-9])|([1-2][0-9])|(3[0-2]))';
				configd:pattern-help "<x.x.x.x/x>";
				configd:help "IPv4 Prefix";
			}
		}
		leaf-list address {
			type ipv4-prefix {
			    configd:normalize "normalize ipv4-prefix";
            }
			ordered-by user;
		}`))

	dontExpect := []string{
		"Does not match pattern",
	}
	dontExpectValidationError(
		t,
		schema_text,
		"address", "1",
		dontExpect...,
	)

	expect := errtest.NewInvalidPatternError(t,
		"/address/1.1.1.1%2F999", "<x.x.x.x/x>")
	expectValidationError(
		t,
		schema_text,
		"address", "1.1.1.1/999",
		expect.RawErrorStrings()...,
	)
}

func TestString1PatternMismatchWithErrMsg(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`leaf strLeaf {
             type string {
                 pattern "[a-z]+" {
					 error-message "Must be all lower case";
				 }
             }
         }`))

	expect := "Must be all lower case"

	expectValidationError(
		t,
		schema_text,
		"strLeaf", "1",
		expect,
	)
}

func TestString2PatternMismatchErrMsg(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`leaf strLeaf {
             type string {
                 pattern "[A-Xa-x]+" {
                         error-message "A to X \b \b only";
                 }
                 pattern "[a-z]+" {
                     error-message "lower-case \b \b only";
                 }
             }
         }`))

	expectValidationError(
		t,
		schema_text,
		"strLeaf", "1",
		"A to X \\b \\b only",
	)

	expectValidationError(
		t,
		schema_text,
		"strLeaf", "A",
		"lower-case \\b \\b only",
	)
}

func TestStringMulitplePatternMismatchErrMsg(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`typedef myStr {
             type string {
                 pattern "[A-Xa-x]+" {
                         error-message "A to X \b \b only";
                 }
             }
         }
         leaf strLeaf {
             type myStr {
                 pattern "[a-z]+" {
                     error-message "lower-case \b \b only";
                 }
             }
         }`))

	expectValidationError(
		t,
		schema_text,
		"strLeaf", "1",
		"A to X \\b \\b only",
	)

	expectValidationError(
		t,
		schema_text,
		"strLeaf", "A",
		"lower-case \\b \\b only",
	)
}

func TestStringPatternMismatchConfigdErrMsg(t *testing.T) {
	leafSchema := `
		leaf strLeaf {
		type string {
			pattern "[a-z]+" {
				configd:error-message "lower-case only";
			}
		}
	}`

	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate, leafSchema))

	expect := "lower-case only"

	expectValidationError(
		t,
		schema_text,
		"strLeaf", "1",
		expect,
	)
}

func TestStringLongPatternMismatchConfigdErrMsg(t *testing.T) {
	leafSchema := `
		leaf strLeaf {
		type string {
			pattern "[a-z][A-Za-z0-9]+" {
				configd:error-message "lower-case only";
			}
		}
	}`

	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate, leafSchema))

	expect := "lower-case only"

	expectValidationError(
		t,
		schema_text,
		"strLeaf", "1",
		expect,
	)
}

// For errors, error-message wins over pattern-help.
func TestStringPatternMismatchWithHelpAndErrMsg(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`leaf strLeaf {
             type string {
                 pattern "[a-z]+" {
                     error-message "err-msg text";
                 }
                 configd:pattern-help "lower-case-only";
             }
         }`))

	expect := "err-msg text"

	expectValidationError(
		t,
		schema_text,
		"strLeaf", "1",
		expect,
	)
}

// For errors, configd:error-message wins over pattern-help.
func TestStringPatternMismatchWithHelpAndCfgErrMsg(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`leaf strLeaf {
             type string {
                 pattern "[a-z]+" {
                     configd:error-message "cfgd-err-msg text";
                 }
                 configd:pattern-help "lower-case-only";
             }
         }`))

	expect := "cfgd-err-msg text"

	expectValidationError(
		t,
		schema_text,
		"strLeaf", "1",
		expect,
	)
}

// Configd error message wins over plain YANG one.
func TestStringPatternMismatch2ErrorMsgs(t *testing.T) {
	leafSchema := `
		leaf strLeaf {
		type string {
			pattern "[a-z][A-Za-z0-9]+" {
				error-message "YANG error";
				configd:error-message "configd error";
			}
		}
	}`

	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate, leafSchema))

	expect := "configd error"

	expectValidationError(
		t,
		schema_text,
		"strLeaf", "1",
		expect,
	)
}
