// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package union

import (
	"bytes"
	"testing"
)

const animalRemoteSchema = `
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
}`

const animalSchema = `
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
    identity hedgehog {
        base remote:animal;
    }
    leaf animal {
        type identityref {
            base remote:animal;
        }
    }
    leaf union-with-idref {
	type union {
            type string {
		    length 0..4;
	    }
            type identityref {
                base remote:animal;
            }
        }
    }
}`

func TestPreferSimpleFormRfc7951(t *testing.T) {
	// Accept namespace qualified identifier value
	input := `{"test-union:animal":"test-union:hedgehog"}`
	// Always outputs in simple form
	expected := `{"test-union:animal":"hedgehog"}`

	schm := newTestSchemas(t, bytes.NewBufferString(animalSchema).Bytes(),
		bytes.NewBufferString(animalRemoteSchema).Bytes())

	root, err := UnmarshalRFC7951(schm, []byte(input))
	if err != nil {
		t.Errorf("Unexpected failure: %s\n", err)
	}

	actual := root.ToRFC7951(IncludeDefaults)
	if string(actual) != expected {
		t.Errorf("Re-encoded JSON does not match.\n   expect=%s\n   actual=%s",
			expected, actual)
	}
}

func TestNoSimpleFormRfc7951InUnion(t *testing.T) {
	// Accept namespace qualified identifier value within a union type
	input := `{"test-union:union-with-idref":"test-union:hedgehog"}`
	// output in simple form
	expected := `{"test-union:union-with-idref":"hedgehog"}`

	schm := newTestSchemas(t, bytes.NewBufferString(animalSchema).Bytes(),
		bytes.NewBufferString(animalRemoteSchema).Bytes())

	root, err := UnmarshalRFC7951(schm, []byte(input))
	if err != nil {
		t.Errorf("Unexpected failure: %s\n", err)
	}

	actual := root.ToRFC7951(IncludeDefaults)
	if string(actual) != expected {
		t.Errorf("Re-encoded JSON does not match.\n   expect=%s\n   actual=%s",
			expected, actual)
	}
}
