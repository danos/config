// Copyright (c) 2018-2019, AT&T Intellectual Property. All rights reserved.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package union

import (
	"fmt"
	"testing"

	"github.com/danos/config/data"
)

const quoteSchema = `
	module test-union {
    namespace "urn:vyatta.com:test:union";
    prefix utest;
	organization "AT&T Inc.";
	contact
		"AT&T
		 Postal: 208 S. Akard Street
		         Dallas, TX 75202
		 Web: www.att.com";

    revision 2018-06-13 {
        description "Test schema for serialisation of quotes in strings";
    }

	container top {
		leaf description {
			type string;
		}
	}
}`

func getQuoteTree(t *testing.T) Node {
	st, err := getSchema([]byte(quoteSchema))
	if err != nil {
		t.Fatal(err)
	}

	return NewNode(data.New("root"), data.New("root"), st, nil, 0)
}

func runQuoteTest(t *testing.T, test quoteTest) {

	expTree := fmt.Sprintf("top {\n\tdescription %s\n}\n", test.output)

	t.Logf("\n%d %s:\t%s\t%s\t%s\n", test.id, test.name, test.asTyped,
		test.bashed, test.output)

	root := getQuoteTree(t)
	root.Data().AddChild(data.New("top"))
	root.Child("top").Data().AddChild(data.New("description"))
	root.Child("top").Child("description").Data().
		AddChild(data.New(test.bashed))
	testSerializeNonFatal(t, root, expTree)
}

func testSerializeNonFatal(t *testing.T, tree Node, expected string) {
	out := serializeTree(tree)

	if out != expected {
		t.Errorf("Bad serialization:\n%s\nExpected:\n%s\n", out, expected)
	}
}

type quoteTest struct {
	name    string
	id      int
	asTyped string // What user types
	bashed  string // What it looks like once bash has processed it.
	output  string // What ends up in config.boot
}

// ID matches up with acceptance test id.  Tests that fail acceptance for
// set or echo are not listed here as are not relevant.  Those that fail on
// commit are included here so we can detect any behaviour change.
var quoteTests = []quoteTest{

	{"EscIntSgl", 2,
		"foo\\'bar", "foo'bar", "foo'bar"}, // commit fail
	{"EscIntDbl", 3,
		"foo\\\"bar", "foo\"bar", "\"foo\\\"bar\""},
	{"PrIntSgl", 4,
		"foo'b'ar", "foobar", "foobar"},
	{"PrIntDbl", 5,
		"foo\"b\"ar", "foobar", "foobar"},
	{"PrEscIntSgl", 6,
		"foo\\'b\\'ar", "foo'b'ar", "foo'b'ar"},
	{"PrEscIntDbl", 7,
		"foo\\\"b\\\"ar", "foo\"b\"ar", "\"foo\\\"b\\\"ar\""},

	{"ExtDiffIntSgl", 8, // commit fail
		"\"foo'bar\"", "foo'bar", "foo'bar"},
	{"ExtDiffIntDbl", 9,
		"'foo\"bar'", "foo\"bar", "\"foo\\\"bar\""},
	{"ExtDiffEscIntSgl", 10, // commit fail
		"\"foo\\'bar\"", "foo\\'bar", "foo\\'bar"},
	{"ExtDiffEscIntDbl", 11,
		"'foo\\\"bar'", "foo\\\"bar", "\"foo\\\"bar\""},
	{"ExtDiffPrIntSgl", 12,
		"\"foo'b'ar\"", "foo'b'ar", "foo'b'ar"},
	{"ExtDiffPrIntDbl", 13,
		"'foo\"b\"ar'", "foo\"b\"ar", "\"foo\\\"b\\\"ar\""},
	{"ExtDiffPrEscIntSgl", 14, // commit fail
		"\"foo\\'b\\'ar\"", "foo\\'b\\'ar", "foo\\'b\\'ar"},
	{"ExtDiffPrEscIntDbl", 15,
		"'foo\\\"b\\\"ar'", "foo\\\"b\\\"ar", "\"foo\\\"b\\\"ar\""},

	{"ExtSameEscIntDbl", 27,
		"\"foo\\\"bar\"", "foo\"bar", "\"foo\\\"bar\""},
	{"ExtSamePrIntSgl", 28,
		"'foo'b'ar'", "foobar", "foobar"},
	{"ExtSamePrIntDbl", 29,
		"\"foo\"b\"ar\"", "foobar", "foobar"},
	{"ExtSamePrEscIntDbl", 31,
		"\"foo\\\"b\\\"ar\"", "foo\"b\"ar", "\"foo\\\"b\\\"ar\""},

	{"Spc1ExtDiffIntSgl", 40, // commit fail
		"\"fo o'bar\"", "fo o'bar", "\"fo o'bar\""},
	{"Spc1ExtDiffIntDbl", 41,
		"'fo o\"bar'", "fo o\"bar", "\"fo o\\\"bar\""},
	{"Spc1ExtDiffEscIntSgl", 42, // commit fail
		"\"fo o\\'bar\"", "fo o\\'bar", "\"fo o\\'bar\""},
	{"Spc1ExtDiffEscIntDbl", 43,
		"'fo o\\\"bar'", "fo o\\\"bar", "\"fo o\\\"bar\""},
	{"Spc1ExtDiffPrIntSgl", 44,
		"\"fo o'b'ar\"", "fo o'b'ar", "\"fo o'b'ar\""},
	{"Spc1ExtDiffPrIntDbl", 45,
		"'fo o\"b\"ar'", "fo o\"b\"ar", "\"fo o\\\"b\\\"ar\""},
	{"Spc1ExtDiffPrEscIntSgl", 46, // commit fail
		"\"fo o\\'b\\'ar\"", "fo o\\'b\\'ar", "\"fo o\\'b\\'ar\""},
	{"Spc1ExtDiffPrEscIntDbl", 47,
		"'fo o\\\"b\\\"ar'", "fo o\\\"b\\\"ar", "\"fo o\\\"b\\\"ar\""},

	{"Spc1ExtSameEscIntDbl", 59,
		"\"fo o\\\"bar\"", "fo o\"bar", "\"fo o\\\"bar\""},
	{"Spc1ExtSamePrIntSgl", 60,
		"'fo o'b'ar'", "fo obar", "\"fo obar\""},
	{"Spc1ExtSamePrIntDbl", 61,
		"\"fo o\"b\"ar\"", "fo obar", "\"fo obar\""},
	{"Spc1ExtSamePrEscIntDbl", 63,
		"\"fo o\\\"b\\\"ar\"", "fo o\"b\"ar", "\"fo o\\\"b\\\"ar\""},

	{"Spc2PrIntSgl", 68,
		"foo'b 'ar", "foob ar", "\"foob ar\""},
	{"Spc2PrIntDbl", 69,
		"foo\"b \"ar", "foob ar", "\"foob ar\""},

	{"Spc2ExtDiffPrIntSgl", 76,
		"\"foo'b 'ar\"", "foo'b 'ar", "\"foo'b 'ar\""},
	{"Spc2ExtDiffPrIntDbl", 77,
		"'foo\"b \"ar'", "foo\"b \"ar", "\"foo\\\"b \\\"ar\""},
	{"Spc2ExtDiffPrEscIntSgl", 78, // commit fail
		"\"foo\\'b \\'ar\"", "foo\\'b \\'ar", "\"foo\\'b \\'ar\""},
	{"Spc2ExtDiffPrEscIntDbl", 79,
		"'foo\\\"b \\\"ar'", "foo\\\"b \\\"ar", "\"foo\\\"b \\\"ar\""},

	{"Spc2ExtSamePrEscIntDbl", 95,
		"\"foo\\\"b \\\"ar\"", "foo\"b \"ar", "\"foo\\\"b \\\"ar\""},
}

func TestQuoteHandling(t *testing.T) {
	for _, test := range quoteTests {
		t.Run(test.name, func(t *testing.T) {
			runQuoteTest(t, test)
		})
	}
}

func dblQuoteTest(t *testing.T, in, out string) {
	res := escapeUnescapedDoubleQuotes(in)
	if res != out {
		t.Errorf("<%s> => <%s>, expected <%s>", in, res, out)
	}
}

func TestDoubleQuoteEscaping(t *testing.T) {
	dblQuoteTest(t, "foo", "foo")
	dblQuoteTest(t, "\"quoteAtStart", "\\\"quoteAtStart")
	dblQuoteTest(t, "\"\"twoQuotesAtStart", "\\\"\\\"twoQuotesAtStart")
	dblQuoteTest(t, "aaa\"QuoteInMiddle", "aaa\\\"QuoteInMiddle")
	dblQuoteTest(t, "\\\"escapedQuoteAtStart", "\\\"escapedQuoteAtStart")
	dblQuoteTest(t, "aaa\\\"escapedQuoteInMid", "aaa\\\"escapedQuoteInMid")
	dblQuoteTest(t, "aaa\"two\"unescQuotes", "aaa\\\"two\\\"unescQuotes")
	dblQuoteTest(t, "aaa\\\"two\\\"escQuotes", "aaa\\\"two\\\"escQuotes")
	dblQuoteTest(t,
		"aaa\\\"bbb\"ccc\\\"ddd\\\"eee\"fff\"ggg\"hhh",
		"aaa\\\"bbb\\\"ccc\\\"ddd\\\"eee\\\"fff\\\"ggg\\\"hhh")
	dblQuoteTest(t,
		"\"aaa\\\"bbb\"ccc\\\"ddd\\\"eee\"fff\"ggg\"hhh",
		"\\\"aaa\\\"bbb\\\"ccc\\\"ddd\\\"eee\\\"fff\\\"ggg\\\"hhh")
	dblQuoteTest(t,
		"\\\"aaa\\\"bbb\"ccc\\\"ddd\\\"eee\"fff\"ggg\"hhh",
		"\\\"aaa\\\"bbb\\\"ccc\\\"ddd\\\"eee\\\"fff\\\"ggg\\\"hhh")

}
