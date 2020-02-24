// Copyright (c) 2017,2020, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This file contains tests relating to the ModelSet extension, which
// deals with construction of a service bus map (mapping VCI components to
// YANG modules and validating provided parameters)
//
// Until multiple modelsets are supported, only the vyatta-v1 modelset is
// supported.

package schema

import (
	"os"
	"testing"

	"github.com/danos/config/testutils/assert"
	"github.com/danos/utils/exec"
	"github.com/danos/vci/conf"
	"github.com/danos/yang/testutils"
)

// TEST DATA
//
// To test validation of DotComponent files, we need both YANG modules and
// DotComponent files, so we can check YANG module references in the latter.
//
// For valid DotComponent files, there are 4 different DotComponent files
// that exercise different permutations of valid options in terms of numbers
// of models, modelsets, etc.

// 1 model, 1 module
const firstComp = `[Vyatta Component]
Name=net.vyatta.test.first
Description=First Component
ExecName=/opt/vyatta/sbin/first
ConfigFile=/etc/vyatta/first.conf

[Model net.vyatta.test.first]
Modules=vyatta-test-first-v1
ModelSets=vyatta-v1`

// 1 model, 2 modules
const secondComp = `[Vyatta Component]
Name=net.vyatta.test.second
Description=Second Component
ExecName=/opt/vyatta/sbin/second
ConfigFile=/etc/vyatta/second.conf

[Model net.vyatta.test.second]
Modules=vyatta-test-second-a-v1,vyatta-test-second-b-v1
ModelSets=vyatta-v1`

// 2 models, 1 module each
const thirdComp = `[Vyatta Component]
Name=net.vyatta.test.third
Description=Third Component
ExecName=/opt/vyatta/sbin/third
ConfigFile=/etc/vyatta/third.conf

[Model net.vyatta.test.third.a]
Modules=vyatta-test-third-a-v1
ModelSets=open-v1

[Model net.vyatta.test.third.b]
Modules=vyatta-test-third-b-v1
ModelSets=vyatta-v1`

// 1 model, 2 model sets
const fourthComp = `[Vyatta Component]
Name=net.vyatta.test.fourth
Description=Fourth Component
ExecName=/opt/vyatta/sbin/fourth
ConfigFile=/etc/vyatta/fourth.conf

[Model net.vyatta.test.fourth]
Modules=vyatta-test-fourth-v1
ModelSets=vyatta-v1,open-v1`

const NsPfx = "urn:vyatta.com:test:"

func TestMultipleComponentRegistration(t *testing.T) {
	serviceMap, _ := getTestServiceMap(t, "testdata/yang",
		firstComp,
		secondComp,
		thirdComp,
		fourthComp)

	checkNumberOfServices(t, serviceMap, 4)

	checkServiceNamespaces(t, serviceMap,
		"net.vyatta.test.first",
		[]string{NsPfx + "vyatta-test-first:1"},
		[]string{})

	checkServiceNamespaces(t, serviceMap,
		"net.vyatta.test.second",
		[]string{NsPfx + "vyatta-test-second-a:1",
			NsPfx + "vyatta-test-second-b:1"},
		[]string{})

	checkServiceNamespaces(t, serviceMap,
		"net.vyatta.test.third.a",
		[]string{},
		[]string{})
	checkServiceNamespaces(t, serviceMap,
		"net.vyatta.test.third.b",
		[]string{NsPfx + "vyatta-test-third-b:1"},
		[]string{})

	checkServiceNamespaces(t, serviceMap,
		"net.vyatta.test.fourth",
		[]string{NsPfx + "vyatta-test-fourth:1"},
		[]string{})
}

const noModuleComp = `[Vyatta Component]
Name=net.vyatta.test.noModule
Description=NoModule Component
ExecName=/opt/vyatta/sbin/noModule
ConfigFile=/etc/vyatta/noModule.conf

[Model net.vyatta.test.noModule]
Modules=vyatta-test-noModule-v1
ModelSets=vyatta-v1`

func TestComponentWithNonExistentModule(t *testing.T) {

	fn := func() ([]*exec.Output, []error, bool) {
		_, err := getModelSet(t, "testdata/yang", noModuleComp)
		if err == nil {
			return nil, nil, true
		}
		return nil, []error{err}, true
	}

	_, errs, _, debug := assert.RunTestAndCaptureStdout(fn)
	if len(errs) != 0 {
		t.Fatalf("Unexpected error checking for non-existent module: %s",
			errs[0].Error())
		return
	}
	expMsgs := assert.NewExpectedMessages(
		"net.vyatta.test.noModule",
		"vyatta-test-noModule-v1 (sub)module not present in image")
	expMsgs.ContainedIn(t, debug)
}

const orderFirstComp = `[Vyatta Component]
Name=net.vyatta.test.first
Description=First Component
ExecName=/opt/vyatta/sbin/first
ConfigFile=/etc/vyatta/first.conf

[Model net.vyatta.test.first]
Modules=vyatta-test-first-v1
ModelSets=vyatta-v1`

const orderSecondAComp = `[Vyatta Component]
Name=net.vyatta.test.second-a
Description=Second Component (A)
ExecName=/opt/vyatta/sbin/second-a
ConfigFile=/etc/vyatta/second-a.conf
Before=net.vyatta.test.third
After=net.vyatta.test.first

[Model net.vyatta.test.second-a]
Modules=vyatta-test-second-a-v1
ModelSets=vyatta-v1`

const orderSecondBComp = `[Vyatta Component]
Name=net.vyatta.test.second-b
Description=Second Component (B)
ExecName=/opt/vyatta/sbin/second-b
ConfigFile=/etc/vyatta/second-b.conf
Before=net.vyatta.test.third
After=net.vyatta.test.first

[Model net.vyatta.test.second-b]
Modules=vyatta-test-second-b-v1
ModelSets=vyatta-v1`

const orderThirdComp = `[Vyatta Component]
Name=net.vyatta.test.third
Description=Third Component
ExecName=/opt/vyatta/sbin/third
ConfigFile=/etc/vyatta/third.conf
Before=net.vyatta.test.fourth

[Model net.vyatta.test.third-a]
Modules=vyatta-test-third-a-v1
ModelSets=open-v1

[Model net.vyatta.test.third-b]
Modules=vyatta-test-third-b-v1
ModelSets=vyatta-v1`

const orderFourthComp = `[Vyatta Component]
Name=net.vyatta.test.fourth
Description=Second Component (B)
ExecName=/opt/vyatta/sbin/fourth
ConfigFile=/etc/vyatta/fourth.conf
After=net.vyatta.test.first

[Model net.vyatta.test.fourth]
Modules=vyatta-test-fourth-v1
ModelSets=vyatta-v1`

func TestServiceOrdering(t *testing.T) {
	svcs, orderedSvcs := getTestServiceMap(t, "testdata/yang",
		orderSecondBComp,
		orderFourthComp,
		orderSecondAComp,
		orderFirstComp,
		orderThirdComp)

	checkOrderedService(t, "net.vyatta.test.first", 1, 1,
		orderedSvcs, svcs)
	checkOrderedService(t, "net.vyatta.test.second-a", 2, 3,
		orderedSvcs, svcs)
	checkOrderedService(t, "net.vyatta.test.second-b", 2, 3,
		orderedSvcs, svcs)
	checkOrderedService(t, "net.vyatta.test.third-b", 4, 4,
		orderedSvcs, svcs)
	checkOrderedService(t, "net.vyatta.test.fourth", 5, 5,
		orderedSvcs, svcs)
}

func checkOrderedService(
	t *testing.T,
	name string,
	earliest_1_indexed, latest_1_indexed int,
	orderedSvcs []string,
	serviceMap map[string]*service,
) {
	for pos, svc := range orderedSvcs {
		if svc == name {
			if earliest_1_indexed > (pos + 1) {
				t.Fatalf("Service %s too early in list", name)
			}
			if latest_1_indexed < (pos + 1) {
				t.Fatalf("Service %s too late in list", name)
			}
			if _, ok := serviceMap[name]; !ok {
				t.Fatalf("Service %s not found in service map", name)
			}
			return
		}
	}
	t.Fatalf("Service %s not found in ordered service list!", name)
}

const otherFirstSharingSameModuleComp = `[Vyatta Component]
Name=net.vyatta.test.other.first
Description=First Component
ExecName=/opt/vyatta/sbin/other-first-service
ConfigFile=/etc/vyatta/other-first.conf

[Model net.vyatta.test.other-first]
Modules=vyatta-test-first-v1
ModelSets=vyatta-v1`

func TestComponentsSharingModule(t *testing.T) {
	_, err := getModelSet(t, "testdata/yang",
		firstComp,
		otherFirstSharingSameModuleComp)
	if err != nil {
		expMsgs := assert.NewExpectedMessages(
			"net.vyatta.test.first",
			"net.vyatta.test.other-first",
			"cannot share 'urn:vyatta.com:test:vyatta-test-first:1'")
		expMsgs.ContainedIn(t, err.Error())
	} else {
		t.Fatalf("Module sharing not detected.\n")
	}
}

const otherFirstSharingSameModelComp = `[Vyatta Component]
Name=net.vyatta.test.other.first
Description=First Component
ExecName=/opt/vyatta/sbin/other-first-service
ConfigFile=/etc/vyatta/other-first.conf

[Model net.vyatta.test.first]
Modules=vyatta-test-first-v1
ModelSets=vyatta-v1`

func TestComponentsSharingModel(t *testing.T) {
	_, err := getModelSet(t, "testdata/yang",
		firstComp,
		otherFirstSharingSameModelComp)
	if err != nil {
		expMsgs := assert.NewExpectedMessages(
			"Model 'net.vyatta.test.first'",
			"defined twice for model set 'vyatta-v1'")
		expMsgs.ContainedIn(t, err.Error())
	} else {
		t.Fatalf("Shared model not detected.\n")
	}
}

const firstCompForUnassignedTest = `[Vyatta Component]
Name=net.vyatta.test.first
Description=First Component
ExecName=/opt/vyatta/sbin/first-service
ConfigFile=/etc/vyatta/first.conf

[Model net.vyatta.test.first]
Modules=vyatta-test-first-v1
ModelSets=vyatta-v1`

const secondCompForUnassignedTest = `[Vyatta Component]
Name=net.vyatta.test.second
Description=Second Component
ExecName=/opt/vyatta/sbin/second-service
ConfigFile=/etc/vyatta/second.conf

[Model net.vyatta.test.second]
Modules=vyatta-test-second-v1
ModelSets=vyatta-v1`

const defaultCompForUnassignedTest = `[Vyatta Component]
Name=net.vyatta.test.default
Description=Default Component
ExecName=/opt/vyatta/sbin/default-service
ConfigFile=/etc/vyatta/default.conf
DefaultComponent=true

[Model net.vyatta.test.default]
ModelSets=vyatta-v1`

const secondDefaultCompForUnassignedTest = `[Vyatta Component]
Name=net.vyatta.test.default2
Description=Default Component
ExecName=/opt/vyatta/sbin/default2-service
ConfigFile=/etc/vyatta/default2.conf
DefaultComponent=true

[Model net.vyatta.test.default2]
ModelSets=vyatta-v1`

const defaultCompWithModule = `[Vyatta Component]
Name=net.vyatta.test.default2
Description=Default Component
ExecName=/opt/vyatta/sbin/default2-service
ConfigFile=/etc/vyatta/default2.conf
DefaultComponent=true

[Model net.vyatta.test.default2]
Modules=vyatta-test-unassigned-a-v1
ModelSets=vyatta-v1`

func TestSingleDefaultComponentDetected(t *testing.T) {
	ms, err := getModelSet(t, "testdata/unassigned_yang",
		firstCompForUnassignedTest,
		secondCompForUnassignedTest,
		defaultCompForUnassignedTest)
	if err != nil {
		t.Fatalf("Unable to compile model set: %s\n", err.Error())
	}

	expDefSvcName := "net.vyatta.test.default"
	if ms.defaultService != expDefSvcName {
		t.Fatalf("Exp service name: %s\nAct service name: %s\n",
			expDefSvcName, ms.defaultService)
	}
}

func TestMultipleDefaultComponentsDetected(t *testing.T) {
	_, err := getModelSet(t, "testdata/unassigned_yang",
		firstCompForUnassignedTest,
		secondCompForUnassignedTest,
		defaultCompForUnassignedTest,
		secondDefaultCompForUnassignedTest)
	if err != nil {
		expMsgs := assert.NewExpectedMessages(
			"Can't have 2 default components",
			"'net.vyatta.test.default'",
			"'net.vyatta.test.default2'")
		expMsgs.ContainedIn(t, err.Error())
	} else {
		t.Fatalf("Duplicate default components not detected")
	}
}

func TestDefaultComponentWithModulesDetected(t *testing.T) {
	_, err := getModelSet(t, "testdata/unassigned_yang",
		firstCompForUnassignedTest,
		secondCompForUnassignedTest,
		defaultCompWithModule)
	if err != nil {
		expMsgs := assert.NewExpectedMessages(
			"Default component",
			"cannot have assigned namespaces")
		expMsgs.ContainedIn(t, err.Error())
	} else {
		t.Fatalf("Default components with assigned modules detected")
	}
}

func TestUnassignedNamespacesAssignedToDefaultComponent(t *testing.T) {
	serviceMap, _ := getTestServiceMap(t, "testdata/unassigned_yang",
		firstCompForUnassignedTest,
		secondCompForUnassignedTest,
		defaultCompForUnassignedTest)

	checkNumberOfServices(t, serviceMap, 3)

	checkServiceNamespaces(t, serviceMap,
		"net.vyatta.test.default",
		[]string{
			NsPfx + "vyatta-test-unassigned-a:1",
			NsPfx + "vyatta-test-unassigned-b:1",
			NsPfx + "vyatta-test-augment:1",
		},
		[]string{})
}

const checkSchemaSnippet = `
container checkCont {
	leaf checkLeaf {
		type string;
	}
}`

const requiredForCheckSchemaSnippet = `
container reqForCheckCont {
	leaf reqForCheckLeaf {
		type uint16;
	}
}`

func TestImportsRequiredForCheck(t *testing.T) {
	schemas := []*testutils.TestSchema{
		testutils.NewTestSchema("vyatta-test-check-v1", "check1").
			AddSchemaSnippet(checkSchemaSnippet),
		testutils.NewTestSchema("vyatta-required-for-check-v1", "required1").
			AddSchemaSnippet(requiredForCheckSchemaSnippet),
	}

	vciComp := conf.CreateTestDotComponentFile("test-check").
		AddModelWithCheckImport("net.vyatta.vci.test.test-check",
			[]string{"vyatta-test-check-v1"},
			[]string{conf.BaseModelSet},
			[]string{"vyatta-required-for-check-v1"})

	tmpYangDir := createYangDir(t, "checkTest", schemas)
	defer os.RemoveAll(tmpYangDir)

	serviceMap, _ := getTestServiceMap(t, tmpYangDir, vciComp.String())

	checkNumberOfServices(t, serviceMap, 1)

	checkServiceNamespaces(t, serviceMap,
		"net.vyatta.vci.test.test-check",
		[]string{NsPfx + "vyatta-test-check-v1"},
		[]string{NsPfx + "vyatta-required-for-check-v1"})
}
