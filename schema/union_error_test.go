// Copyright (c) 2019, AT&T Intellectual Property
// All rights reserved.
// Copyright (c) 2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"bytes"
	"fmt"
	"testing"
)

func TestEnumerationsInError(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`leaf enumLeaf {
             type union {
                 type enumeration {
			 enum one;
			 enum two;
		 }
		 type uint32;
		 type enumeration {
			enum three;
			enum four;
		 }
             }
         }`))

	expect := []string{"Must have one of the following values:",
		"between 0 and 4294967295",
		"one",
		"two",
		"three",
		"four"}

	expectValidationError(
		t,
		schema_text,
		"enumLeaf", "zero",
		expect...,
	)
}

func TestIdentityrefInError(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`identity greek;
		 identity alpha {
			 base greek;
		 }
		 identity beta {
			base greek;
		 }
		 typedef allgreek {
			type identityref {
				base greek;
			}
		 }
		 leaf tdef {
			type allgreek;
		 }

		leaf idrefLeaf {
             type union {
                 type enumeration {
			 enum one;
			 enum two;
		 }
		 type uint32;
		 type enumeration {
			enum three;
			enum four;
		 }
		 type identityref {
			base greek;
		 }
             }
         }`))

	expect := []string{"Must have one of the following values:",
		"between 0 and 4294967295",
		"alpha",
		"beta",
		"one",
		"two",
		"three",
		"four"}

	expectValidationError(
		t,
		schema_text,
		"idrefLeaf", "zero",
		expect...,
	)
}

func TestPatternHelpInError(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`leaf unionLeaf {
		type union {
			type string {
				pattern 'ab[0-9]*';
				configd:pattern-help "<abN>";
			}
			type string {
				pattern 'de[0-9]*';
				configd:pattern-help "<deN>";

			}
		 	type uint32 {
				range 1000..4000;
			}
		 	type enumeration {
				enum three;
				enum four;
		 	}
                 }
         }`))

	expect := []string{"Must have one of the following values:",
		"between 1000 and 4000",
		"three",
		"four",
		"<abN>",
		"<deN>"}

	expectValidationError(
		t,
		schema_text,
		"unionLeaf", "aa88",
		expect...,
	)
}

func TestOpdPatternHelpInError(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`opd:option unionCommand {
		type union {
			type string {
				pattern 'ab[0-9]*';
				opd:pattern-help "<abN>";
			}
			type string {
				pattern 'de[0-9]*';
				opd:pattern-help "<deN>";

			}
			type uint32 {
				range 1000..4000;
			}
			type enumeration {
				enum three;
				enum four;
			}
                 }
         }`))

	expect := []string{"Must have one of the following values:",
		"between 1000 and 4000",
		"three",
		"four",
		"<abN>",
		"<deN>"}

	expectValidationError(
		t,
		schema_text,
		"unionCommand", "aa88",
		expect...,
	)
}
