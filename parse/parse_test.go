// Copyright (c) 2020, AT&T Intellectual Property. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0 and BSD-3-Clause

package parse_test

import (
	"testing"

	"github.com/danos/config/parse"
)

// These tests match the test strings used in the 'quotes_pass' acceptance test.
// Note that a few of the test cases that pass in the acceptance tests will not
// be parsed here because this parser is not doing shell expansion and thus
// validly rejects some strings that would pass via shell expansion.
var inputStringsWithQuotesAndOtherEscapableChars = []string{
	`f02\'bar`,
	`f03\"bar`,
	`f04'b'ar`,
	`f07\"b\"ar`,
	`"f08'bar"`,
	`'f09"bar'`,
	`"f10\'bar"`,
	`'f11\"bar'`,
	`"f12'b'ar"`,
	`'f13"b"ar'`,
	`"f14\'b\'ar"`,
	`'f15\"b\"ar'`,
	`"f27\"bar"`,
	`'f28'b'ar'`,
	`"f4 0'bar"`,
	`'f4 1"bar'`,
	`"f4 2\'bar"`,
	`'f4 3\"bar'`,
	`"f4 4'b'ar"`,
	`'f4 5"b"ar'`,
	`"f4 6\'b\'ar"`,
	`'f4 7\"b\"ar'`,
	`"f5 9\"bar"`,
	`'f6 0'b'ar'`,
	`"f6 3\"b\"ar"`,
	`"f76'b 'ar"`,
	`'f77"b "ar'`,
	`"f78\'b \'ar"`,
	`'f79\"b \"ar'`,
	`"f95\"b \"ar"`,
}

func TestParser(t *testing.T) {
	for _, input := range inputStringsWithQuotesAndOtherEscapableChars {
		_, err := parse.Parse("filename", input)
		if err != nil {
			t.Errorf("Parse error for `%s`: %s\n", input, err)
		}
	}
}
