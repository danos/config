// Copyright (c) 2019, AT&T Intellectual Property.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package yangconfig_test

import (
	"bytes"
	"github.com/danos/config/yangconfig"
	"strings"
	"testing"
)

const testconfig = `{
  "yang": [
    "/usr/share/configd/yang",
    "/run/vyatta-platform/activeyang"
  ],
  "features": [
    {
      "location": "/config/features",
      "enabled": true
    },
    {
      "location": "/run/vyatta-platform/features",
      "enabled": true
    },
    {
      "location": "/run/vyatta-platform/features-disabled",
      "enabled": false
    },
    {
      "location": "/opt/vyatta/etc/features",
      "enabled": true
    }
  ]
}`

const resultantconfig = `{
  "yang": [
    "/home/vyatta/testyang",
    "/usr/share/configd/yang",
    "/run/vyatta-platform/activeyang",
    "/home/vyatta/test-opd-yang",
    "/tmp/yang"
  ],
  "features": [
    {
      "location": "/config/features",
      "enabled": true
    },
    {
      "location": "/run/vyatta-platform/features",
      "enabled": true
    },
    {
      "location": "/run/vyatta-platform/features-disabled",
      "enabled": false
    },
    {
      "location": "/opt/vyatta/etc/features",
      "enabled": true
    },
    {
      "location": "/home/vyatta/testfeatures",
      "enabled": true
    },
    {
      "location": "/home/vyatta/testdisabledfeatures",
      "enabled": false
    }
  ]
}
`

// Test that yang config loads correctly, additional locations can added with order preserved
// Also ensure that directory names are cleaned up, and duplicates removed.
func TestLoadConfig(t *testing.T) {
	var b bytes.Buffer
	r := strings.NewReader(testconfig)
	cfg := yangconfig.NewConfig().IncludeYangDirs("/home/vyatta/testyang").Load(r).
		IncludeYangDirs("/home/vyatta/test-opd-yang", "/usr/share/configd/yang//").
		IncludeFeatures("/home/vyatta/testfeatures").
		IncludeDisabledFeatures("/home/vyatta/testdisabledfeatures").
		IncludeYangDirs("/tmp/yang")

	cfg.Save(&b)
	if resultantconfig != b.String() {
		t.Fatalf("Yangconfig is not as expected.\n Expected:\n%s\nGot:\n%s\n",
			resultantconfig, b.String())
	}
}
