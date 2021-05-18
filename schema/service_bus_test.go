// Copyright (c) 2017,2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0
//
// This test was written when the 'service-bus' extension was going to
// be used to extend the Model object to allow for components.  This
// was then changed to extend the ModelSet object instead.  Kept as
// an example of YANG using the service-bus extension, but may not be
// of much use ...

package schema

import (
	"testing"

	"github.com/danos/vci/conf"
)

const yangdExtensionsSchema = `
module brocade-service-api-v1 {
	namespace "urn:brocade.com:mgmt:yangd:1";
	prefix brocade-service-api-v1;

	organization "Brocade Communications Systems, Inc.";
	contact
		"Brocade Communications Systems, Inc.
		 Postal: 130 Holger Way
			 San Jose, CA 95134
		 E-mail: support@Brocade.com
		 Web: www.brocade.com";

	revision 2016-04-22 {
		description "Initial revision.";
	}

	extension service-bus {
		argument text;
	}
}
`

const serviceBusModuleSchema = `
module test-configd-compile {
	namespace "urn:vyatta.com:test:configd-compile";
	prefix test;
	import brocade-service-api-v1 {
		prefix yangd;
	}

	organization "Brocade Communications Systems, Inc.";
	revision 2014-12-29 {
		description "Test schema for configd";
	}

	yangd:service-bus "net.vyatta.test.service.example";

	leaf test {
		type string;
	}
}
`

const testComp = `[Vyatta Component]
Name=net.vyatta.test.service.example
Description=Super Example Project
ExecName=/opt/vyatta/sbin/example-service
ConfigFile=/etc/vyatta/example.conf

[Model net.vyatta.test.service.example]
Modules=test-configd-compile
ModelSets=vyatta-v1`

func TestYangdServiceBus(t *testing.T) {

	ext_text := []byte(yangdExtensionsSchema)
	sch_text := []byte(serviceBusModuleSchema)

	compCfg, err := conf.ParseConfiguration([]byte(testComp))
	if err != nil {
		t.Fatalf("Unexpected component config parse failure:\n  %s\n\n", err.Error())
	}

	ms, err := GetConfigSchemaWithComponents(
		[]*conf.ServiceConfig{compCfg},
		ext_text, sch_text)
	if err != nil {
		t.Fatalf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	busMap := ms.(*modelSet).compMappings.components
	if len(busMap) != 1 {
		t.Fatalf("Unexpected number of buses found: %d\n", len(busMap))
	}

	expectBus := "net.vyatta.test.service.example"
	serv, ok := busMap[expectBus]
	if !ok {
		t.Fatalf("Expected service bus not found:\n")
	}

	if len(serv.modMap) != 1 {
		t.Fatalf("Unexpected number of modules found: %d\n", len(serv.modMap))
	}

	expectNs := "urn:vyatta.com:test:configd-compile"
	_, ok = serv.modMap[expectNs]
	if !ok {
		t.Fatalf("Expected namespace not found:\n")
	}
}
