// Copyright (c) 2017-2019 by AT&T Intellectual Property
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This file contains tests relating to the submodules, specifically to
// their assignment to modelsets / services.

package schema

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/danos/vci/conf"
	"github.com/danos/yang/testutils"
)

// Tests for submodules:
//
// 'component' here refers to a 'new' VCI component that explicitly owns
// modules, so not provisiond.
//
// 1) Module + submodule in same component (both VCI and provd)
// 2) Module + submodule in different component
// 3) Module in component, submodule in provisiond
// 4) Module in provisiond, submodule in separate component
// 5) Module + submodule in 1 component, second submodule in other comp
//    augments first submodule
//
// Need to verify filter function set up correctly
// Ideally check ownership with augment to sibling submodule

// Tests are more readable if the schemas are included in the same file as
// the tests, rather than in separate YANG files.  As the underlying
// infrastructure expects separate YANG files, this function creates a
// temporary directory with yang files in it.
func createYangDir(
	t *testing.T,
	dirname string,
	schemas []*testutils.TestSchema,
) string {
	dir, err := ioutil.TempDir("", dirname)
	if err != nil {
		t.Fatalf("Unable to create Yang Dir (%s): %s\n",
			dirname, err.Error())
		return ""
	}

	for _, schema := range schemas {
		if err := ioutil.WriteFile(
			filepath.Join(dir, schema.Name.Namespace+".yang"),
			[]byte(testutils.ConstructSchema(*schema)), 0666); err != nil {
			os.RemoveAll(dir)
			t.Fatalf("Unable to create %s: %s\n",
				schema.Name.Namespace+".yang", err.Error())
			return ""
		}
	}

	return dir
}

const mod1SchemaSnippet = `
	container mod1Cont {
	leaf mod1Leaf {
		type string;
	}
}`

const submod1SchemaSnippet = `
	container submod1Cont {
	leaf submod1Leaf {
		type string;
	}
}`

func TestModuleAndSubmoduleInSameVCIComponent(t *testing.T) {

	schemas := []*testutils.TestSchema{
		testutils.NewTestSchema("vyatta-test-mod1-v1", "mod1").
			AddInclude("vyatta-test-submod1-v1").
			AddSchemaSnippet(mod1SchemaSnippet),
		testutils.NewTestSchema("vyatta-test-submod1-v1", "submod1").
			AddBelongsTo("vyatta-test-mod1-v1", "mod1").
			AddSchemaSnippet(submod1SchemaSnippet),
	}

	vciCompWithModuleAndSubmodule := conf.
		CreateTestDotComponentFile("mod-and-submod").
		AddModel("net.vyatta.vci.test.mod-and-submod",
			[]string{"vyatta-test-mod1-v1", "vyatta-test-submod1-v1"},
			[]string{conf.BaseModelSet})

	tmpYangDir := createYangDir(t,
		"modAndSubmodSameComp", schemas)
	defer os.RemoveAll(tmpYangDir)

	serviceMap, _ := getTestServiceMap(t, tmpYangDir,
		vciCompWithModuleAndSubmodule.String())

	checkNumberOfServices(t, serviceMap, 1)

	checkService(t, serviceMap,
		"net.vyatta.vci.test.mod-and-submod",
		[]string{
			NsPfx + "vyatta-test-mod1-v1",
			"vyatta-test-submod1-v1@" + NsPfx + "vyatta-test-mod1-v1",
		})
}

func TestModuleAndSubmoduleInDefaultComp(t *testing.T) {

	schemas := []*testutils.TestSchema{
		testutils.NewTestSchema("vyatta-test-mod1-v1", "mod1").
			AddInclude("vyatta-test-submod1-v1").
			AddSchemaSnippet(mod1SchemaSnippet),
		testutils.NewTestSchema("vyatta-test-submod1-v1", "submod1").
			AddBelongsTo("vyatta-test-mod1-v1", "mod1").
			AddSchemaSnippet(submod1SchemaSnippet),
	}

	defaultComp := conf.
		CreateTestDotComponentFile("default").
		SetDefault().
		AddModel("net.vyatta.vci.test.default",
			[]string{},
			[]string{conf.BaseModelSet})

	tmpYangDir := createYangDir(t,
		"modAndSubmodInDefaultComp", schemas)
	defer os.RemoveAll(tmpYangDir)

	serviceMap, _ := getTestServiceMap(t, tmpYangDir,
		defaultComp.String())

	checkNumberOfServices(t, serviceMap, 1)

	checkService(t, serviceMap,
		"net.vyatta.vci.test.default",
		[]string{
			NsPfx + "vyatta-test-mod1-v1",
			"vyatta-test-submod1-v1@" + NsPfx + "vyatta-test-mod1-v1",
		})
}

func TestSubmoduleInDifferentComponent(t *testing.T) {

	schemas := []*testutils.TestSchema{
		testutils.NewTestSchema("vyatta-test-mod1-v1", "mod1").
			AddInclude("vyatta-test-submod1-v1").
			AddSchemaSnippet(mod1SchemaSnippet),
		testutils.NewTestSchema("vyatta-test-submod1-v1", "submod1").
			AddBelongsTo("vyatta-test-mod1-v1", "mod1").
			AddSchemaSnippet(submod1SchemaSnippet),
	}

	moduleComp := conf.
		CreateTestDotComponentFile("module").
		AddModel("net.vyatta.vci.test.mod",
			[]string{"vyatta-test-mod1-v1"},
			[]string{conf.BaseModelSet})
	submoduleComp := conf.
		CreateTestDotComponentFile("submodule").
		AddModel("net.vyatta.vci.test.submod",
			[]string{"vyatta-test-submod1-v1"},
			[]string{conf.BaseModelSet})

	tmpYangDir := createYangDir(t,
		"modAndSubmodDiffComp", schemas)
	defer os.RemoveAll(tmpYangDir)

	serviceMap, _ := getTestServiceMap(t, tmpYangDir,
		moduleComp.String(), submoduleComp.String())

	checkNumberOfServices(t, serviceMap, 2)

	checkService(t, serviceMap,
		"net.vyatta.vci.test.mod",
		[]string{NsPfx + "vyatta-test-mod1-v1"})
	checkService(t, serviceMap,
		"net.vyatta.vci.test.submod",
		[]string{"vyatta-test-submod1-v1@" + NsPfx + "vyatta-test-mod1-v1"})
}

func TestModuleInVCICompSubmoduleInDefaultComp(t *testing.T) {

	schemas := []*testutils.TestSchema{
		testutils.NewTestSchema("vyatta-test-mod1-v1", "mod1").
			AddInclude("vyatta-test-submod1-v1").
			AddSchemaSnippet(mod1SchemaSnippet),
		testutils.NewTestSchema("vyatta-test-submod1-v1", "submod1").
			AddBelongsTo("vyatta-test-mod1-v1", "mod1").
			AddSchemaSnippet(submod1SchemaSnippet),
	}

	moduleComp := conf.
		CreateTestDotComponentFile("module").
		AddModel("net.vyatta.vci.test.mod",
			[]string{"vyatta-test-mod1-v1"},
			[]string{conf.BaseModelSet})
	defaultComp := conf.
		CreateTestDotComponentFile("default").
		SetDefault().
		AddModel("net.vyatta.vci.test.default",
			[]string{},
			[]string{conf.BaseModelSet})

	tmpYangDir := createYangDir(t,
		"modInCompSubmodInDflt", schemas)
	defer os.RemoveAll(tmpYangDir)

	serviceMap, _ := getTestServiceMap(t, tmpYangDir,
		moduleComp.String(), defaultComp.String())

	checkNumberOfServices(t, serviceMap, 2)

	checkService(t, serviceMap,
		"net.vyatta.vci.test.mod",
		[]string{NsPfx + "vyatta-test-mod1-v1"})
	checkService(t, serviceMap,
		"net.vyatta.vci.test.default",
		[]string{"vyatta-test-submod1-v1@" + NsPfx + "vyatta-test-mod1-v1"})
}

func TestModuleInDefaultCompSubmoduleInVCIComp(t *testing.T) {

	schemas := []*testutils.TestSchema{
		testutils.NewTestSchema("vyatta-test-mod1-v1", "mod1").
			AddInclude("vyatta-test-submod1-v1").
			AddSchemaSnippet(mod1SchemaSnippet),
		testutils.NewTestSchema("vyatta-test-submod1-v1", "submod1").
			AddBelongsTo("vyatta-test-mod1-v1", "mod1").
			AddSchemaSnippet(submod1SchemaSnippet),
	}

	moduleComp := conf.
		CreateTestDotComponentFile("submodule").
		AddModel("net.vyatta.vci.test.submod",
			[]string{"vyatta-test-submod1-v1"},
			[]string{conf.BaseModelSet})
	defaultComp := conf.
		CreateTestDotComponentFile("default").
		SetDefault().
		AddModel("net.vyatta.vci.test.default",
			[]string{},
			[]string{conf.BaseModelSet})

	tmpYangDir := createYangDir(t,
		"modInDfltSubmodInComp", schemas)
	defer os.RemoveAll(tmpYangDir)

	serviceMap, _ := getTestServiceMap(t, tmpYangDir,
		moduleComp.String(), defaultComp.String())

	checkNumberOfServices(t, serviceMap, 2)

	checkService(t, serviceMap,
		"net.vyatta.vci.test.default",
		[]string{NsPfx + "vyatta-test-mod1-v1"})
	checkService(t, serviceMap,
		"net.vyatta.vci.test.submod",
		[]string{"vyatta-test-submod1-v1@" + NsPfx + "vyatta-test-mod1-v1"})
}

// Testing of augment / refine doesn't belong here as it needs configuration
// to check correct 'destination' component.
