// Copyright (c) 2017, 2019-2020, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2014-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/danos/vci/conf"
	"github.com/danos/yang/compile"
	"github.com/danos/yang/parse"
	"github.com/danos/yang/xpath"
	"github.com/danos/yang/xpath/xutils"
	"github.com/danos/yangd"
)

// Schema Template with '%s' at end for insertion of schema for each test.
const schemaTemplate = `
module test-configd-compile {
	namespace "urn:vyatta.com:test:configd-compile";
	prefix test;
	organization "Brocade Communications Systems, Inc.";
	revision 2014-12-29 {
		description "Test schema for configd";
	}
	%s
}
`

// Create ModelSet structure from multiple buffers, each buffer
// represents a single yang module.
func GetConfigSchema(bufs ...[]byte) (ModelSet, error) {
	return GetSchemaWithDispatcher(nil, nil, compile.IsConfig, bufs...)
}

func GetConfigSchemaWithWarnings(bufs ...[]byte,
) (ModelSet, []xutils.Warning, error) {
	return GetSchemaWithDispatcherAndWarnings(
		nil, nil, compile.IsConfig, bufs...)
}

func GetConfigSchemaWithWarningsAndCustomFunctions(
	userFnChecker xpath.UserCustomFunctionCheckerFn,
	bufs ...[]byte,
) (ModelSet, []xutils.Warning, error) {
	return GetSchemaWithDispatcherAndWarningsAndCustomFunctions(
		nil, nil, compile.IsConfig, userFnChecker, bufs...)
}

func GetConfigSchemaWithDispatcher(
	disp yangd.Dispatcher,
	comps []*conf.ServiceConfig,
	bufs ...[]byte,
) (ModelSet, error) {

	return GetSchemaWithDispatcher(disp, comps, compile.IsConfig, bufs...)
}

func GetSchema(filter compile.SchemaFilter, bufs ...[]byte) (ModelSet, error) {
	return GetSchemaWithDispatcher(nil, nil, filter, bufs...)
}

func GetSchemaWithDispatcher(
	disp yangd.Dispatcher,
	comps []*conf.ServiceConfig,
	filter compile.SchemaFilter,
	bufs ...[]byte,
) (ModelSet, error) {

	const name = "schema"
	modules := make(map[string]*parse.Tree)
	for index, b := range bufs {
		t, err := Parse(name+strconv.Itoa(index), string(b))
		if err != nil {
			return nil, err
		}
		mod := t.Root.Argument().String()
		modules[mod] = t
	}
	return CompileModules(modules, "", false, filter,
		&CompilationExtensions{disp, comps})
}

func GetSchemaWithDispatcherAndWarnings(
	disp yangd.Dispatcher,
	comps []*conf.ServiceConfig,
	filter compile.SchemaFilter,
	bufs ...[]byte,
) (ModelSet, []xutils.Warning, error) {

	const name = "schema"
	modules := make(map[string]*parse.Tree)
	for index, b := range bufs {
		t, err := Parse(name+strconv.Itoa(index), string(b))
		if err != nil {
			return nil, nil, err
		}
		mod := t.Root.Argument().String()
		modules[mod] = t
	}
	return CompileModulesWithWarnings(modules, "", false, filter,
		&CompilationExtensions{disp, comps})
}

func GetSchemaWithDispatcherAndWarningsAndCustomFunctions(
	disp yangd.Dispatcher,
	comps []*conf.ServiceConfig,
	filter compile.SchemaFilter,
	userFnChecker xpath.UserCustomFunctionCheckerFn,
	bufs ...[]byte,
) (ModelSet, []xutils.Warning, error) {

	const name = "schema"
	modules := make(map[string]*parse.Tree)
	for index, b := range bufs {
		t, err := Parse(name+strconv.Itoa(index), string(b))
		if err != nil {
			return nil, nil, err
		}
		mod := t.Root.Argument().String()
		modules[mod] = t
	}
	return CompileModulesWithWarningsAndCustomFunctions(
		modules, "", false, filter, &CompilationExtensions{disp, comps},
		userFnChecker)
}

func expectValidationError(
	t *testing.T,
	schema_text *bytes.Buffer,
	nodeName, value string,
	expectList ...string,
) {
	checkValidationError(t, true, schema_text, nodeName, value, expectList...)
}

func dontExpectValidationError(
	t *testing.T,
	schema_text *bytes.Buffer,
	nodeName, value string,
	expectList ...string,
) {
	checkValidationError(t, false, schema_text, nodeName, value, expectList...)
}

func checkValidationError(
	t *testing.T,
	expectToFind bool,
	schema_text *bytes.Buffer,
	nodeName, value string,
	expectList ...string,
) {
	ms, err := GetSchema(compile.Include(compile.IsConfig, compile.IsOpd),
		schema_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	ctx := ValidateCtx{Sid: "", CurPath: []string{nodeName, value}, St: ms}
	node := ms.Child(nodeName)
	if node == nil {
		t.Fatalf("Unable to find node with name '%s'\n", nodeName)
	}

	err = node.Validate(ctx, []string{nodeName}, []string{value})
	if err == nil {
		t.Fatalf("Unexpected success:\n  Expect: %s\n\n", expectList)
	}

	expStr := "Expect"
	if !expectToFind {
		expStr = "Don't expect"
	}

	for _, expect := range expectList {
		if !strings.Contains(err.Error(), expect) == expectToFind {
			t.Fatalf(
				"Unexpected error string:\nActual:\n\n%s\n\n%s:\n\n%s\n\n",
				err.Error(), expStr, expect)
		}
	}
}

func expectValidationSuccess(
	t *testing.T,
	schema_text *bytes.Buffer,
	nodeName, value string,
) {
	ms, err := GetConfigSchema(schema_text.Bytes())
	if err != nil {
		t.Errorf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	ctx := ValidateCtx{Sid: "", CurPath: []string{nodeName, value}, St: ms}
	node := ms.Child(nodeName)

	err = node.Validate(ctx, []string{nodeName}, []string{value})
	if err != nil {
		t.Errorf("Unexpected failure: %s\n\n", err.Error())
	}
}

// Test Service Manager
//
// Used to replace the default service manager with one where we can control the
// results, eg for the call to IsActive().
type testSvcManager struct{}

func enableTestSvcManager() svcMgrFnType {
	oldFn := svcMgrFn
	svcMgrFn = func() SvcManager {
		return createTestSvcManager()
	}
	return oldFn
}

func disableTestSvcManager(oldFn svcMgrFnType) { svcMgrFn = oldFn }

func createTestSvcManager() *testSvcManager {
	return &testSvcManager{}
}

func (tsm *testSvcManager) Close()                            { return }
func (tsm *testSvcManager) Start(name string) error           { return nil }
func (tsm *testSvcManager) Reload(name string) error          { return nil }
func (tsm *testSvcManager) ReloadOrRestart(name string) error { return nil }
func (tsm *testSvcManager) Restart(name string) error         { return nil }
func (tsm *testSvcManager) ReloadServices() error             { return nil }
func (tsm *testSvcManager) Stop(name string) error            { return nil }
func (tsm *testSvcManager) Enable(name string) error          { return nil }
func (tsm *testSvcManager) Disable(name string) error         { return nil }

// For now, assume any component is active.
func (tsm *testSvcManager) IsActive(name string) (bool, error) {
	return true, nil
}

// Test Dispatcher and Service objects to allow injection of customised
// functionality.

type testDispatcher struct{}

type testService struct{}

func (d *testDispatcher) NewService(name string) (yangd.Service, error) {
	return &testService{}, nil
}

func (s *testService) GetRunning(path string) ([]byte, error) {
	return nil, nil
}

func (s *testService) GetState(path string) ([]byte, error) {
	return nil, nil
}

func (s *testService) ValidateCandidate(candidate []byte) error {
	return nil
}

func (s *testService) SetRunning(candidate []byte) error {
	return nil
}

func getComponentConfigs(t *testing.T, dotCompFiles ...string,
) (configs []*conf.ServiceConfig) {

	for _, file := range dotCompFiles {
		cfg, err := conf.ParseConfiguration([]byte(file))
		if err != nil {
			t.Fatalf("Unexpected component config parse failure:\n  %s\n\n",
				err.Error())
		}
		configs = append(configs, cfg)
	}

	return configs
}

func getTestServiceMap(t *testing.T, yangDir string, dotCompFiles ...string,
) (map[string]*service, []string) {

	compExt := &CompilationExtensions{
		Dispatcher: &testDispatcher{},
		ComponentConfig: getComponentConfigs(
			t, dotCompFiles...),
	}

	ms, err := CompileDir(
		&compile.Config{
			YangDir:      yangDir,
			CapsLocation: "",
			Filter:       compile.IsConfig},
		compExt,
	)
	if err != nil {
		t.Fatalf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}
	return ms.(*modelSet).services, ms.(*modelSet).orderedServices
}

func getModelSet(t *testing.T, yangDir string, dotCompFiles ...string,
) (*modelSet, error) {

	compExt := &CompilationExtensions{
		Dispatcher: &testDispatcher{},
		ComponentConfig: getComponentConfigs(
			t, dotCompFiles...),
	}

	ms, err := CompileDir(
		&compile.Config{
			YangDir:      yangDir,
			CapsLocation: "",
			Filter:       compile.IsConfig},
		compExt,
	)
	if ms == nil {
		return nil, err
	}
	return ms.(*modelSet), err
}

func checkNumberOfServices(
	t *testing.T,
	serviceMap map[string]*service,
	numSvcs int) {

	if len(serviceMap) != numSvcs {
		t.Fatalf("Unexpected number of services found: exp %d, got %d\n",
			numSvcs, len(serviceMap))
	}
}

func checkServiceNamespaces(
	t *testing.T,
	serviceMap map[string]*service,
	modelName string,
	expSetAndCheckNamespaces []string,
	expCheckOnlyNamespaces []string,
) {

	service, ok := serviceMap[modelName]
	if !ok {
		// Only an error if there are any namespaces to check.  Otherwise
		// this is a model for a different model set.
		if len(expSetAndCheckNamespaces) != 0 {
			t.Fatalf("Unable to find service '%s'\n", modelName)
		}
		return
	}

	// First check 'owned' namespaces that will be sent to component's Set()
	// function on commit
	checkNamespacesInMap(t, service.modMap, modelName, expSetAndCheckNamespaces)
}

// Ensure exact match for namespaces in modMap
func checkNamespacesInMap(
	t *testing.T,
	modMap map[string]struct{},
	modelName string,
	expNamespaces []string,
) {
	var ns string
	if len(expNamespaces) != len(modMap) {
		t.Fatalf("%s: Expected %d namespaces, but found %d\n",
			modelName, len(expNamespaces), len(modMap))
	}
	for _, ns = range expNamespaces {
		if _, ok := modMap[ns]; !ok {
			t.Fatalf("Unable to find '%s' namespace in:\n%v",
				ns, modMap)
			return
		}
	}
}

func dumpServiceMap(serviceMap map[string]*service) {
	for svcName, svc := range serviceMap {
		fmt.Printf("S: %s\n", svcName)
		for ns, _ := range svc.modMap {
			fmt.Printf("\tNS: %s\n", ns)
		}
	}
}
