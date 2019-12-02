// Copyright (c) 2017-2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package union

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/danos/config/schema"
	"github.com/danos/config/testutils/assert"
	"github.com/danos/mgmterror/errtest"
)

const schemaTemplate = `
module test-union {
    namespace "urn:vyatta.com:test:union";
    prefix utest;
    organization "AT&T Inc.";
    contact
        "AT&T
         Postal: 208 S. Akard Street
                 Dallas, TX 75202
         Web: www.attcom";
    revision 2019-11-29 {
        description "Test schema for unions";
    }

    %s
}
`

func newTestSchema(t *testing.T, inputStuff string) schema.ModelSet {
	input_schema := fmt.Sprintf(schemaTemplate, inputStuff)
	st, err := getSchema([]byte(input_schema))
	if err != nil {
		t.Fatal(err)
	}

	return st
}

func newTestSchemas(t *testing.T, buf ...[]byte) schema.ModelSet {
	st, err := getSchema(buf...)
	if err != nil {
		t.Fatal(err)
	}

	return st

}
func TestSetLeaf(t *testing.T) {

	// Note: natural order
	test_schema :=
		`leaf leave_as_default {
			type string;
			default "default";
		}
		leaf optional_set {
			type string;
		}
		leaf optional_not_set {
			type string;
		}
		leaf override_default {
			type string;
			default "default";
		}
		leaf set_to_default {
			type string;
			default "default";
		}`
	input := `<data>` +
		`<optional_set xmlns="urn:vyatta.com:test:union">set</optional_set>` +
		`<override_default xmlns="urn:vyatta.com:test:union">override</override_default>` +
		`<set_to_default xmlns="urn:vyatta.com:test:union">default</set_to_default>` +
		`</data>`
	expected := `<data>` +
		`<leave_as_default xmlns="urn:vyatta.com:test:union">default</leave_as_default>` +
		`<optional_set xmlns="urn:vyatta.com:test:union">set</optional_set>` +
		`<override_default xmlns="urn:vyatta.com:test:union">override</override_default>` +
		`<set_to_default xmlns="urn:vyatta.com:test:union">default</set_to_default>` +
		`</data>`

	root, err := UnmarshalXML(newTestSchema(t, test_schema), []byte(input))
	if err != nil {
		t.Errorf("Unexpected failure: %s\n", err)
	}

	actual := root.ToXML("data", IncludeDefaults)
	if string(actual) != expected {
		t.Errorf("Re-encoded XML does not match.\n   expect=%s\n   actual=%s",
			expected, actual)
	}
}

func TestSetLeafIdentityNamespacePrefixes(t *testing.T) {

	remoteSchema :=
		`
module test-remote {
    namespace "urn:vyatta.com:test:remote";
    prefix remote;
    organization "AT&T Inc.";
    contact
        "AT&T
         Postal: 208 S. Akard Street
                 Dallas, TX 75202
         Web: www.attcom";
    revision 2019-11-29 {
        description "Test schema for unions";
    }
	identity animal;
		identity cat {
			base animal;
		}
		identity dog {
			base animal;
		}
}
		`
	test_schema :=
		`
module test-union {
    namespace "urn:vyatta.com:test:union";
    prefix union;
    import test-remote {prefix remote;}
    organization "AT&T Inc.";
    contact
        "AT&T
         Postal: 208 S. Akard Street
                 Dallas, TX 75202
         Web: www.attcom";
    revision 2019-11-29 {
        description "Test schema for unions";
    }
    leaf an {
        type identityref {
             base remote:animal;
        }
    }
}`
	// Accept a namespace prefix that does not match module name
	input := `<data>` +
		`<an xmlns="urn:vyatta.com:test:union" xmlns:a="urn:vyatta.com:test:remote">a:cat</an>` +
		`</data>`
	// encode again using module name as namespace prefix
	expected := `<data>` +
		`<an xmlns="urn:vyatta.com:test:union" xmlns:test-remote="urn:vyatta.com:test:remote">test-remote:cat</an>` +
		`</data>`

	schm := newTestSchemas(t, bytes.NewBufferString(test_schema).Bytes(),
		bytes.NewBufferString(remoteSchema).Bytes())

	root, err := UnmarshalXML(schm, []byte(input))
	if err != nil {
		t.Errorf("Unexpected failure: %s\n", err)
	}

	actual := root.ToXML("data", IncludeDefaults)
	if string(actual) != expected {
		t.Errorf("Re-encoded XML does not match.\n   expect=%s\n   actual=%s",
			expected, actual)
	}
}

func TestSetIdentityrefUnionAndListKey(t *testing.T) {
	remoteSchema :=
		`
module test-remote {
    namespace "urn:vyatta.com:test:remote";
    prefix remote;
    organization "AT&T Inc.";
    contact
        "AT&T
         Postal: 208 S. Akard Street
                 Dallas, TX 75202
         Web: www.attcom";
    revision 2019-11-29 {
        description "Test schema for unions";
    }
	identity mineral;
	identity quartz {
		base mineral;
	}
	identity saphire {
		base mineral;
	}
	identity animal;
		identity cat {
			base animal;
		}
		identity mouse {
			base animal;
		}
}
		`
	test_schema :=
		`
module test-union {
    namespace "urn:vyatta.com:test:union";
    prefix union;
    import test-remote {prefix remote;}
    organization "AT&T Inc.";
    contact
        "AT&T
         Postal: 208 S. Akard Street
                 Dallas, TX 75202
         Web: www.attcom";
    revision 2019-11-29 {
        description "Test schema for unions";
    }
    list alist {
        key key;

        leaf key {
            type identityref {
                base remote:animal;
            }
        }
        leaf name {
            type string;
        }
    }
    leaf mineral {
        type union {
            type uint8;
            type union {
                type identityref {
                    base remote:mineral;
                }
            }
        }
    }
}`

	// test identityref within a union
	// and identityref as a list key
	input := `<data>` +
		`<mineral xmlns="urn:vyatta.com:test:union" xmlns:t="urn:vyatta.com:test:remote">t:quartz</mineral>` +
		`<alist xmlns="urn:vyatta.com:test:union">` +
		`<key xmlns="urn:vyatta.com:test:union" xmlns:foo="urn:vyatta.com:test:remote">foo:cat</key>` +
		`<name xmlns="urn:vyatta.com:test:union">tom</name>` +
		`</alist>` +
		`<alist xmlns="urn:vyatta.com:test:union">` +
		`<key xmlns="urn:vyatta.com:test:union" xmlns:i="urn:vyatta.com:test:remote">i:mouse</key>` +
		`<name xmlns="urn:vyatta.com:test:union">jerry</name></alist>` +
		`</data>`
	expected := `<data>` +
		`<alist xmlns="urn:vyatta.com:test:union">` +
		`<key xmlns="urn:vyatta.com:test:union" xmlns:test-remote="urn:vyatta.com:test:remote">test-remote:cat</key>` +
		`<name xmlns="urn:vyatta.com:test:union">tom</name>` +
		`</alist>` +
		`<alist xmlns="urn:vyatta.com:test:union">` +
		`<key xmlns="urn:vyatta.com:test:union" xmlns:test-remote="urn:vyatta.com:test:remote">test-remote:mouse</key>` +
		`<name xmlns="urn:vyatta.com:test:union">jerry</name>` +
		`</alist>` +
		`<mineral xmlns="urn:vyatta.com:test:union" xmlns:test-remote="urn:vyatta.com:test:remote">test-remote:quartz</mineral>` +
		`</data>`

	schm := newTestSchemas(t, bytes.NewBufferString(test_schema).Bytes(),
		bytes.NewBufferString(remoteSchema).Bytes())

	root, err := UnmarshalXML(schm, []byte(input))
	if err != nil {
		t.Errorf("Unexpected failure: %s\n", err)
	}

	actual := root.ToXML("data", IncludeDefaults)
	if string(actual) != expected {
		t.Errorf("Re-encoded XML does not match.\n   expect=%s\n   actual=%s",
			expected, actual)
	}
}

func TestZeroLengthStringLeaf(t *testing.T) {

	// Note: natural order
	test_schema :=
		`leaf optional_set {
			type string;
		}`
	input := `<data>` +
		`<optional_set xmlns="urn:vyatta.com:test:union"></optional_set>` +
		`</data>`
	expected := `<data>` +
		`<optional_set xmlns="urn:vyatta.com:test:union"></optional_set>` +
		`</data>`

	root, err := UnmarshalXML(newTestSchema(t, test_schema), []byte(input))
	if err != nil {
		t.Errorf("Unexpected failure: %s\n", err)
	}

	actual := root.ToXML("data", IncludeDefaults)
	if string(actual) != expected {
		t.Errorf("Re-encoded XML does not match.\n   expect=%s\n   actual=%s",
			expected, string(actual))
	}
}

func TestSetEmptyLeaf(t *testing.T) {

	// Note: natural order
	test_schema :=
		`leaf optional_set {
			type empty;
		}
		leaf optional_not_set {
			type empty;
		}`
	input := `<data>` +
		`<optional_set xmlns="urn:vyatta.com:test:union" />` +
		`</data>`
	expected := `<data>` +
		`<optional_set xmlns="urn:vyatta.com:test:union"></optional_set>` +
		`</data>`

	root, err := UnmarshalXML(newTestSchema(t, test_schema), []byte(input))
	if err != nil {
		t.Errorf("Unexpected failure: %s\n", err)
	}

	actual := root.ToXML("data", IncludeDefaults)
	if string(actual) != expected {
		t.Errorf("Re-encoded XML does not match.\n   expect=%s\n   actual=%s",
			expected, string(actual))
	}
}

func TestSetEmptyChanges(t *testing.T) {

	// Note: natural order
	test_schema :=
		`leaf optional_not_set {
			type empty;
		}`
	input := `<data>` +
		`</data>`
	expected := `<data>` +
		`</data>`

	root, err := UnmarshalXML(newTestSchema(t, test_schema), []byte(input))
	if err != nil {
		t.Errorf("Unexpected failure: %s\n", err)
	}

	actual := root.ToXML("data", IncludeDefaults)
	if string(actual) != expected {
		t.Errorf("Re-encoded XML does not match.\n   expect=%s\n   actual=%s",
			expected, string(actual))
	}
}

func TestSetEmptyLeafWithValue(t *testing.T) {

	// Note: natural order
	test_schema :=
		`leaf my_leaf {
			type empty;
		}`
	input :=
		`<data>` +
			`<my_leaf xmlns="urn:vyatta.com:test:union">invalid value</my_leaf>` +
			`</data>`
	expected := "Value found for empty leaf"

	_, actual := UnmarshalXML(newTestSchema(t, test_schema), []byte(input))
	if actual == nil {
		t.Errorf("Unexpected success\n")
		return
	}
	if !strings.Contains(actual.Error(), expected) {
		t.Errorf("Unexpected error\n   expect=%s\n   actual=%s\n", expected, actual)
	}
}

func TestInvalidXML(t *testing.T) {

	test_schema :=
		`leaf right_name {
			type string;
		}`
	input := `<data>` +
		`<right_name xmlns="urn:vyatta.com:test:union">override</wrong_name>` +
		`</data>`
	expected := "XML syntax error on line 1: element <right_name> closed by </wrong_name>"

	_, actual := UnmarshalXML(newTestSchema(t, test_schema), []byte(input))
	if actual == nil {
		t.Errorf("Unexpected success\n")
		return
	}
	if actual.Error() != expected {
		t.Errorf("Unexpected error\n   expect=%s\n   actual=%s\n", expected, actual)
	}
}

func TestInvalidNode(t *testing.T) {

	test_schema :=
		`leaf right_name {
			type string;
		}`
	input := `<data>` +
		`<wrong_name xmlns="urn:vyatta.com:test:union">override</wrong_name>` +
		`</data>`
	expected := assert.NewExpectedMessages(
		errtest.NewInvalidNodeError(t, "/wrong_name").RawErrorStrings()...)

	root, err := UnmarshalXML(newTestSchema(t, test_schema), []byte(input))
	if err == nil {
		t.Errorf("Unexpected success\n%s",
			root.ToXML("data", IncludeDefaults))
		return
	}

	expected.ContainedIn(t, err.Error())
}

func TestInvalidValue(t *testing.T) {

	test_schema :=
		`leaf right_name {
			type string {
				length 1..10;
			}
		}`
	input := `<data>` +
		`<right_name xmlns="urn:vyatta.com:test:union">far-too-long-name</right_name>` +
		`</data>`
	expected := assert.NewExpectedMessages(
		errtest.NewInvalidLengthError(t,
			"/right_name/far-too-long-name", 1, 10).RawErrorStrings()...)

	root, err := UnmarshalXML(newTestSchema(t, test_schema), []byte(input))
	if err == nil {
		t.Errorf("Unexpected success\n%s",
			root.ToXML("data", IncludeDefaults))
		return
	}
	expected.ContainedIn(t, err.Error())
}

func TestSetContainer(t *testing.T) {

	// Note: natural order
	test_schema :=
		`container outer {
			leaf leave_as_default {
				type string;
				default "default";
			}
			leaf optional_set {
				type string;
			}
			leaf-list my_list {
				type string;
			}
		}`
	input := `<data>` +
		`<outer xmlns="urn:vyatta.com:test:union">` +
		`<optional_set xmlns="urn:vyatta.com:test:union">set</optional_set>` +
		`<my_list xmlns="urn:vyatta.com:test:union">one</my_list>` +
		`<my_list xmlns="urn:vyatta.com:test:union">two</my_list>` +
		`<my_list xmlns="urn:vyatta.com:test:union">three</my_list>` +
		`</outer>` +
		`</data>`
	expected := `<data><outer xmlns="urn:vyatta.com:test:union">` +
		`<leave_as_default xmlns="urn:vyatta.com:test:union">default</leave_as_default>` +
		`<my_list xmlns="urn:vyatta.com:test:union">one</my_list>` +
		`<my_list xmlns="urn:vyatta.com:test:union">three</my_list>` +
		`<my_list xmlns="urn:vyatta.com:test:union">two</my_list>` +
		`<optional_set xmlns="urn:vyatta.com:test:union">set</optional_set>` +
		`</outer></data>`

	root, err := UnmarshalXML(newTestSchema(t, test_schema), []byte(input))
	if err != nil {
		t.Errorf("Unexpected failure: %s\n", err)
		t.FailNow()
	}

	actual := root.ToXML("data", IncludeDefaults)
	if string(actual) != expected {
		t.Errorf("Re-encoded XML does not match.\n   expect=%s\n   actual=%s",
			expected, string(actual))
	}
}

func TestSetList(t *testing.T) {

	test_schema :=
		`container testcontainer {
		list testlist {
			key key;
			leaf key {
				type string;
			}
			leaf default {
				type string;
				default "default";
			}
			leaf optional {
				type string;
			}
		}
	}`

	input := `<data><testcontainer>` +
		`<testlist xmlns="urn:vyatta.com:test:union">` +
		`<key xmlns="urn:vyatta.com:test:union">new_entry</key>` +
		`<optional xmlns="urn:vyatta.com:test:union">custom value</optional>` +
		`</testlist></testcontainer></data>`
	expected := `<data><testcontainer xmlns="urn:vyatta.com:test:union">` +
		`<testlist xmlns="urn:vyatta.com:test:union">` +
		`<key xmlns="urn:vyatta.com:test:union">new_entry</key>` +
		`<default xmlns="urn:vyatta.com:test:union">default</default>` +
		`<optional xmlns="urn:vyatta.com:test:union">custom value</optional>` +
		`</testlist></testcontainer></data>`

	root, err := UnmarshalXML(newTestSchema(t, test_schema), []byte(input))
	if err != nil {
		t.Errorf("Unexpected failure: %s\n", err)
		t.Fail()
		return
	}

	actual := root.ToXML("data", IncludeDefaults)
	if string(actual) != expected {
		t.Errorf("Re-encoded XML does not match.\n   expect=%s\n   actual=%s",
			expected, string(actual))
	}
}

func TestSetListMultipleElements(t *testing.T) {

	test_schema :=
		`container testcontainer {
		list testlist {
			key key;
			leaf key {
				type string;
			}
			leaf default {
				type string;
				default "default";
			}
			leaf optional {
				type string;
			}
		}
	}`

	input := `<data><testcontainer>` +
		`<testlist>` +
		`<key>new_entry</key>` +
		`<optional>custom value</optional>` +
		`</testlist>` +
		`<testlist>` +
		`<key>new_entry2</key>` +
		`<optional>custom value</optional>` +
		`</testlist></testcontainer></data>`
	expected := `<data>` +
		`<testcontainer xmlns="urn:vyatta.com:test:union">` +
		`<testlist xmlns="urn:vyatta.com:test:union">` +
		`<key xmlns="urn:vyatta.com:test:union">new_entry</key>` +
		`<default xmlns="urn:vyatta.com:test:union">default</default>` +
		`<optional xmlns="urn:vyatta.com:test:union">custom value</optional>` +
		`</testlist>` +
		`<testlist xmlns="urn:vyatta.com:test:union">` +
		`<key xmlns="urn:vyatta.com:test:union">new_entry2</key>` +
		`<default xmlns="urn:vyatta.com:test:union">default</default>` +
		`<optional xmlns="urn:vyatta.com:test:union">custom value</optional>` +
		`</testlist>` +
		`</testcontainer></data>`

	root, err := UnmarshalXML(newTestSchema(t, test_schema), []byte(input))
	if err != nil {
		t.Errorf("Unexpected failure: %s\n", err)
		t.Fail()
		return
	}

	actual := root.ToXML("data", IncludeDefaults)
	if string(actual) != expected {
		t.Errorf("Re-encoded XML does not match.\n   expect=%s\n   actual=%s",
			expected, string(actual))
	}
}

func TestSetNestedList(t *testing.T) {

	test_schema :=
		`list outer {
			key outer_key;
			leaf outer_key {
				type string;
			}
			list inner {
				key inner_key;
				leaf inner_key {
					type string;
				}
				leaf my_leaf {
					type string;
				}
			}
		}`

	input := `<data>` +
		`<outer xmlns="urn:vyatta.com:test:union">` +
		`<outer_key xmlns="urn:vyatta.com:test:union">outer_entry</outer_key>` +
		`<inner xmlns="urn:vyatta.com:test:union">` +
		`<inner_key xmlns="urn:vyatta.com:test:union">inner_entry</inner_key>` +
		`<my_leaf xmlns="urn:vyatta.com:test:union">some value</my_leaf>` +
		`</inner></outer></data>`
	expected := `<data>` +
		`<outer xmlns="urn:vyatta.com:test:union">` +
		`<outer_key xmlns="urn:vyatta.com:test:union">outer_entry</outer_key>` +
		`<inner xmlns="urn:vyatta.com:test:union">` +
		`<inner_key xmlns="urn:vyatta.com:test:union">inner_entry</inner_key>` +
		`<my_leaf xmlns="urn:vyatta.com:test:union">some value</my_leaf>` +
		`</inner></outer></data>`

	root, err := UnmarshalXML(newTestSchema(t, test_schema), []byte(input))
	if err != nil {
		t.Errorf("Unexpected failure: %s\n", err)
		t.Fail()
		return
	}

	actual := root.ToXML("data", IncludeDefaults)
	if string(actual) != expected {
		t.Errorf("Re-encoded XML does not match.\n   expect=%s\n   actual=%s",
			expected, string(actual))
	}
}

func TestRpcAddsInputDefaults(t *testing.T) {
	test_schema := `rpc ping {
		description "Generates Ping and return response";
		input {
			leaf host {
				type string;
				mandatory true;
			}
			leaf count {
				type uint32;
				default 3;
				description "Number of ping echo request message to send";
			}
			leaf ttl {
				type uint8;
				default "255";
				description "IP Time To Live";
			}
		}
		output {
			leaf tx-packet-count {
				type uint32;
				description "Transmitted Packet count";
			}
			leaf rx-packet-count {
				type uint32;
				description "Received packet count";
			}
			leaf min-delay {
				type uint32;
				units "milliseconds";
				description "Minimum packet delay";
			}
			leaf average-delay {
				type uint32;
				units "milliseconds";
				description "Average packet delay";
			}
			leaf max-delay {
				type uint32;
				units "millisecond";
				description "Minimum packet delay";
			}
		}
		configd:call-rpc "/opt/vyatta/bin/yangop-ping.pl";
	    }`

	input := `<ping>` +
		`<host xmlns="urn:vyatta.com:test:union">localhost</host>` +
		`</ping>`
	expected := `<rpc-reply>` +
		`<count xmlns="urn:vyatta.com:test:union">3</count>` +
		`<host xmlns="urn:vyatta.com:test:union">localhost</host>` +
		`<ttl xmlns="urn:vyatta.com:test:union">255</ttl>` +
		`</rpc-reply>`

	schTree := newTestSchema(t, test_schema)
	rpc := schTree.Rpcs()["urn:vyatta.com:test:union"]["ping"]
	if rpc == nil {
		t.Errorf("Could not find RPC\n")
		return
	}

	if rpc.Input() == nil {
		t.Errorf("Input RPC schema is missing\n")
		return
	}

	root, err := UnmarshalXML(rpc.Input().(schema.Tree), []byte(input))
	if err != nil {
		t.Errorf("Unexpected failure: %s\n", err)
		return
	}

	actual := root.ToXML("rpc-reply", IncludeDefaults)
	if string(actual) != expected {
		t.Errorf("Re-encoded XML does not match.\n   expect=%s\n   actual=%s",
			expected, string(actual))
	}
}
