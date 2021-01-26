// Copyright (c) 2019-2021, AT&T Intellectual Property. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0
//

package platform_test

import (
	"strings"
	"testing"

	"github.com/danos/config/platform"
	"github.com/danos/config/testutils/assert"
	"github.com/danos/utils/exec"
)

var expected = map[string]*platform.Definition{
	"tinybox": &platform.Definition{
		Yang: []string{"vyatta-deviations-foo-tiny-v1.yang",
			"vyatta-deviations-bar-tiny-v1.yang",
			"vyatta-deviations-bar-base-v1.yang",
			"vyatta-bar-base-yang-v1.yang"},
		Features: []string{"vyatta-baz-v1:enable-foo",
			"vyatta-op-baz-v1:show-foo",
			"vyatta-bar-base-v1:enable-bar",
			"vyatta-bar-v1:test-feature",
			"vyatta-bar-v1:extra-feature",
			"vyatta-bar-v1:software-features"},
		DisabledFeatures: []string{"vyatta-bar-v1:big-queues",
			"vyatta-op-bar-base-v1:fast-bar"},
	},
	"smallbox": &platform.Definition{
		Yang: []string{"vyatta-deviations-foo-small-v1.yang"},
		Features: []string{"vyatta-baz-v1:enable-foo",
			"vyatta-op-baz-v1:show-foo"},
		DisabledFeatures: []string{},
	},
	"mediumbox": &platform.Definition{
		Yang: []string{"vyatta-deviations-foo-medium-v1.yang",
			"vyatta-deviations-bar-base-v1.yang",
			"vyatta-deviations-bar-medium-v1.yang",
			"vyatta-bar-base-yang-v1.yang"},
		Features: []string{"vyatta-baz-v1:enable-foo",
			"vyatta-op-baz-v1:show-foo",
			"vyatta-bar-base-v1:enable-bar",
			"vyatta-bar-v1:test-feature",
			"vyatta-bar-v1:extra-feature",
			"vyatta-hw-cache-v1:enable"},
		DisabledFeatures: []string{"vyatta-op-bar-base-v1:fast-bar"},
	},
	"largebox": &platform.Definition{
		Yang: []string{"foo-large-v1.yang",
			"foo-large-extra-v1.yang"},
		Features: []string{"vyatta-baz-v1:test",
			"vyatta-op-baz-v1:test",
			"vyatta-op-baz-v1:large"},
		DisabledFeatures: []string{},
	},
}

func slicesMatch(slc1, slc2 []string) bool {
	inslice := func(v1 string, s1 []string) bool {
		for _, v := range s1 {
			if v == v1 {
				return true
			}
		}
		return false
	}

	if len(slc1) == 0 && len(slc2) == 0 {
		return true
	}
	for _, str := range slc1 {
		if !inslice(str, slc2) {
			return false
		}
	}
	for _, str := range slc2 {
		if !inslice(str, slc1) {
			return false
		}
	}
	return true
}

func TestLoadDefinitions(t *testing.T) {
	ps := platform.NewPlatform().PlatformBaseDir("testdata/good").
		LoadDefinitions()
	for platID, exp := range expected {

		got, ok := ps.Platforms[platID]
		if !ok {
			t.Fatalf("Expected platform %s not found\n", platID)
		}

		if !slicesMatch(exp.Yang, got.Yang) {
			t.Fatalf("Yang not as expected for platform %s\n Exp: %v\n Got: %v\n",
				platID, exp.Yang, got.Yang)
		}
		if !slicesMatch(exp.Features, got.Features) {
			t.Fatalf("Features not as expected for platform %s\n Exp: %v Got: %v\n",
				platID, exp.Features, got.Features)
		}
		if !slicesMatch(exp.DisabledFeatures, got.DisabledFeatures) {
			t.Fatalf("DisabledFeatures not as expected for platform %s Exp: %v Got: %v\n",
				platID, exp.DisabledFeatures, got.DisabledFeatures)
		}
	}

}

type ignoreTest struct {
	platforms string
	expected  []string
}

func runIgnoreTest(t *testing.T, test ignoreTest) {

	fn := func() ([]*exec.Output, []error, bool) {
		platform.NewPlatform().PlatformBaseDir(test.platforms).
			LoadDefinitions()

		return nil, []error{}, true
	}

	_, _, _, out := assert.RunTestAndCaptureStdout(fn)

	for _, exp := range test.expected {
		if !strings.Contains(out, exp) {
			t.Fatalf("Output not as expected for platforms %s\n Expected: %s\n Got: %s\n",
				test.platforms, exp, out)
		}
	}
}

func TestLoadDefinitionsIgnore(t *testing.T) {
	testcases := []ignoreTest{
		{
			platforms: "testdata/badmodule",
			expected: []string{"Ignoring invalid Yang file: 'vyatta-bar-base-yang-v1.yang  malformedbasemodule.yang'",
				"Ignoring invalid Yang file: 'vyatta-deviations-bar-tiny-v1.yang malformedplatformmodule.yang'"},
		},
		{
			platforms: "testdata/badfeature",
			expected: []string{"Ignoring invalid Yang feature: 'vyatta-bar-base-v1:enable-bar malformed:basefeature'",
				"Ignoring invalid Yang feature: 'vyatta-bar-v1:software-features malformed:platformfeature'"},
		},
		{
			platforms: "testdata/baddisabledfeature",
			expected: []string{"Ignoring invalid Yang feature: 'vyatta-op-bar-base-v1:fast-bar malformedbase:disabledfeature'",
				"Ignoring invalid Yang feature: 'vyatta-bar-v1:big-queues malformedplatformdisabledfeature'"},
		},
	}

	for _, test := range testcases {
		runIgnoreTest(t, test)
	}

}
