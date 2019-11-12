// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0
//
// Tests for different subsets of schema validation (state / cfg / all / none)

package union

import (
	"bytes"
	"strings"
	"testing"

	"github.com/danos/mgmterror/errtest"
	"github.com/danos/yang/data/encoding"
	"github.com/danos/yang/schema"
)

const validationSchemaTemplate = `
	module test-yang-compile {
	namespace "urn:vyatta.com:test:yang-compile";
	prefix test;
	organization "AT&T Inc.";
	revision 2019-03-12 {
		description "Test schema for configd";
	}

	container refCont {
		leaf refLeaf {
			type string;
		}
	}
	container configCont {
		leaf cfgMustLeaf {
			type string;
			must "not(contains(., 'invalidCfg'))";
		}
		leaf cfgMandatoryLeaf {
			type string;
			mandatory true;
		}
		leaf cfgLeafrefLeaf {
			type leafref {
				path "/refCont/refLeaf";
			}
		}
		list cfgUniqueList {
			key listKey;
			leaf listKey {
				type string;
			}
			unique "leaf1 leaf2";
			leaf leaf1 {
				type string;
			}
			leaf leaf2 {
				type string;
			}
		}
	}
	container stateCont {
		config false;
		leaf stateMustLeaf {
			type string;
			must "not(contains(., 'invalidState'))";
		}
		leaf stateMandatoryLeaf {
			type string;
			mandatory true;
		}
		leaf stateLeafrefLeaf {
			type leafref {
				path "/refCont/refLeaf";
			}
		}
		list stateUniqueList {
			key listKey;
			leaf listKey {
				type string;
			}
			unique "leaf1 leaf2";
			leaf leaf1 {
				type string;
			}
			leaf leaf2 {
				type string;
			}
		}
	}

}`

// Designed to be as 'invalid' as possible, triggering all the different
// validations:
//   - leafref
//   - mandatory (what about NP conts ...?) TODO
//   - must
//   - unique
//
// We don't check:
//   - list keys (not done in validation here)
//   - min/max elements, range etc etc.
//
var inputMsg = []byte(`
	{
		"refCont":{
			"refLeaf":"reference"
		},
		"configCont":{
			"cfgMustLeaf":"invalidCfg",
			"cfgLeafrefLeaf":"invalidReference",
			"cfgUniqueList": [
				{"listKey":"key1", "leaf1":"1", "leaf2":"2"},
				{"listKey":"key2", "leaf1":"1", "leaf2":"2"}
			]
		},
		"stateCont":{
			"stateMustLeaf":"invalidState",
			"stateLeafrefLeaf":"invalidReference",
			"stateUniqueList": [
				{"listKey":"key1", "leaf1":"1", "leaf2":"2"},
				{"listKey":"key2", "leaf1":"1", "leaf2":"2"}
			]
		}
	}`)

func checkValidationResults(
	t *testing.T,
	valType schema.ValidationType,
	expErrMsgs []string,
	unexpErrMsgs []string,
) {

	sch := bytes.NewBufferString(validationSchemaTemplate)
	compiledSchema, err := getFullSchema(sch.Bytes())
	if err != nil {
		t.Fatalf("Failed to compile schema: %s", err)
		return
	}

	_, err = NewUnmarshaller(encoding.RFC7951).
		SetValidation(valType).
		Unmarshal(compiledSchema, inputMsg)
	if err == nil {
		if len(expErrMsgs) == 0 {
			return
		}
		t.Fatalf("Test passed, but should have failed with:\n%s\n", expErrMsgs)
	}

	for _, expErrMsg := range expErrMsgs {
		if len(expErrMsg) == 0 {
			t.Errorf("Expected error is empty string.\n")
			return
		}

		if !strings.Contains(err.Error(), expErrMsg) {
			t.Errorf("Test failed with: \n%s\nShould have failed with:\n%s\n",
				err, expErrMsg)
			return
		}
	}

	for _, unexpErrMsg := range unexpErrMsgs {
		if len(unexpErrMsg) == 0 {
			t.Errorf("Expected error is empty string.\n")
			return
		}

		if strings.Contains(err.Error(), unexpErrMsg) {
			t.Errorf(
				"Test failed with: \n%s\nShould NOT have failed with:\n%s\n",
				err, unexpErrMsg)
			return
		}
	}
}

func TestValidation(t *testing.T) {

	// Mandatory
	cfgMandatoryErr := errtest.NewMissingMandatoryNodeError(t,
		"/configCont/cfgMandatoryLeaf").RawErrorStrings()
	stateMandatoryErr := errtest.NewMissingMandatoryNodeError(t,
		"/stateCont/stateMandatoryLeaf").RawErrorStrings()

	// Leafref
	cfgLeafrefErr := errtest.NewLeafrefError(t,
		"/configCont/cfgLeafrefLeaf",
		"/refCont/refLeaf").RawErrorStrings()
	stateLeafrefErr := errtest.NewLeafrefError(t,
		"/stateCont/stateLeafrefLeaf",
		"/refCont/refLeaf").RawErrorStrings()

	// Must
	cfgMustErr := errtest.NewMustDefaultError(t,
		"/configCont/cfgMustLeaf",
		"not(contains(., 'invalidCfg'))").RawErrorStrings()
	stateMustErr := errtest.NewMustDefaultError(t,
		"/stateCont/stateMustLeaf",
		"not(contains(., 'invalidState'))").RawErrorStrings()

	// Unique
	cfgUniqueErr := errtest.NewNonUniquePathsError(t,
		"/configCont/cfgUniqueList",
		[]string{
			"listKey/key1",
			"listKey/key2"},
		[]string{"leaf1 1", "leaf2 2"}).RawErrorStrings()
	stateUniqueErr := errtest.NewNonUniquePathsError(t,
		"/stateCont/stateUniqueList",
		[]string{
			"listKey/key1",
			"listKey/key2"},
		[]string{"leaf1 1", "leaf2 2"}).RawErrorStrings()

	var cfgErrMsgs []string
	cfgErrMsgs = append(cfgErrMsgs, cfgLeafrefErr...)
	cfgErrMsgs = append(cfgErrMsgs, cfgMandatoryErr...)
	cfgErrMsgs = append(cfgErrMsgs, cfgMustErr...)
	cfgErrMsgs = append(cfgErrMsgs, cfgUniqueErr...)

	var stateErrMsgs []string
	stateErrMsgs = append(stateErrMsgs, stateLeafrefErr...)
	stateErrMsgs = append(stateErrMsgs, stateMandatoryErr...)
	stateErrMsgs = append(stateErrMsgs, stateMustErr...)
	stateErrMsgs = append(stateErrMsgs, stateUniqueErr...)

	var allErrMsgs = append(cfgErrMsgs, stateErrMsgs...)

	type validationTest struct {
		name             string
		valType          schema.ValidationType
		expErrorMessages []string
		unexpPaths       []string
	}

	tests := []validationTest{
		{
			name:             "Validate All",
			valType:          schema.ValidateAll,
			expErrorMessages: allErrMsgs,
			unexpPaths:       []string{},
		},
		{
			name:             "Validate Config only",
			valType:          schema.ValidateConfig,
			expErrorMessages: cfgErrMsgs,
			unexpPaths:       []string{"/stateCont"},
		},
		{
			name:             "Validate State only",
			valType:          schema.ValidateState,
			expErrorMessages: stateErrMsgs,
			unexpPaths:       []string{"/configCont"},
		},
		{
			name:             "Validate None",
			valType:          schema.DontValidate,
			expErrorMessages: nil,
			unexpPaths:       []string{"/configCont", "/stateCont"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			checkValidationResults(t, test.valType, test.expErrorMessages,
				test.unexpPaths)
		})
	}
}
