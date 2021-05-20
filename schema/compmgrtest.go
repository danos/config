// Copyright (c) 2021, AT&T Intellectual Property.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"fmt"
	"testing"

	"github.com/danos/encoding/rfc7951"
	"github.com/danos/mgmterror"
	yang "github.com/danos/yang/schema"
)

// TestLog provides the ability for tests to verify which components had
// which operations performed on them (validate, set config, get config, get
// state), and in which order.
const (
	NoFilter   = ""
	SetRunning = "SetRunning"
	GetRunning = "GetRunning"
	GetState   = "GetState"
	Validate   = "Validate"
)

type TestLogEntry struct {
	fn     string
	params []string
}

func NewTestLogEntry(fn string, params ...string) TestLogEntry {
	return TestLogEntry{fn: fn, params: params}
}

// TestCompMgr allows replacement of ComponentManager functionality at a high
// level where we don't want an actual (d)bus or components created. Instead
// we want to verify that the correct tuple of {config, model} has been passed
// to the bus (for validate and commit) or provide the ability to return the
// expected config (or an error) for 'get' operations.
type testCompParams struct {
	t *testing.T

	validatedConfig map[string]string
	committedConfig map[string]string
	currentState    map[string]string

	testLog []TestLogEntry
}

type TestCompMgr struct {
	*compMgr

	tcmParams *testCompParams
}

// Compile time check that the concrete type meets the interface
var _ ComponentManager = (*TestCompMgr)(nil)

func NewTestCompMgr(
	t *testing.T,
	ms yang.ModelSet,
	mappings *ComponentMappings,
) *TestCompMgr {

	var tcmParams testCompParams

	tcmParams.t = t
	tcmParams.testLog = make([]TestLogEntry, 0)

	tcmParams.validatedConfig = make(map[string]string, 0)
	tcmParams.committedConfig = make(map[string]string, 0)
	tcmParams.currentState = make(map[string]string, 0)

	tcm := &TestCompMgr{
		compMgr: NewCompMgr(
			newTestOpsMgr(&tcmParams),
			newTestSvcMgr(&tcmParams),
			ms,
			mappings,
		),
		tcmParams: &tcmParams,
	}

	return tcm
}

// Config / state management.

func (tcm *TestCompMgr) ValidatedConfig(model string) string {
	cfg, ok := tcm.tcmParams.validatedConfig[model]
	if !ok {
		tcm.tcmParams.t.Fatalf("No validated config for %s", model)
	}
	return cfg
}

func (tcm *TestCompMgr) CommittedConfig(model string) string {
	cfg, ok := tcm.tcmParams.committedConfig[model]
	if !ok {
		tcm.tcmParams.t.Fatalf("No committed config for %s", model)
	}
	return cfg
}

func (tcm *TestCompMgr) CurrentState(model string) string {
	cfg, ok := tcm.tcmParams.currentState[model]
	if !ok {
		tcm.tcmParams.t.Fatalf("No current state for %s", model)
	}
	return cfg
}

func (tcm *TestCompMgr) SetCurrentState(model, stateJson string) {
	tcm.tcmParams.currentState[model] = stateJson
}

// Log management.

func (tom *testOpsMgr) addLogEntry(fn string, params ...string) {
	fmt.Printf("Add %s\n", fn)
	tom.tcmParams.testLog = append(tom.tcmParams.testLog,
		NewTestLogEntry(fn, params...))
}

func (tcm *TestCompMgr) ClearLogEntries() {
	fmt.Printf("Clear log\n")
	tcm.tcmParams.testLog = nil
}

func (tcm *TestCompMgr) filteredLogEntries(filter string) []TestLogEntry {
	retLog := make([]TestLogEntry, 0)

	for _, entry := range tcm.tcmParams.testLog {
		if entry.fn == filter {
			retLog = append(retLog, entry)
		}
	}

	return retLog
}

func (tcm *TestCompMgr) CheckLogEntries(
	t *testing.T,
	name string,
	entries []TestLogEntry,
	filter string,
) {
	actualLog := tcm.tcmParams.testLog
	fmt.Printf("Entries: %d\n", len(tcm.tcmParams.testLog))
	if filter != NoFilter {
		actualLog = tcm.filteredLogEntries(filter)
	}
	if len(entries) != len(actualLog) {
		t.Logf("\nTEST: %s\n", name)
		t.Logf("\nExp: %d entries\nGot: %d\n",
			len(entries), len(actualLog))
		tcm.dumpLog(t)
		t.Fatalf("---\n")
		return
	}

	for ix, entry := range entries {
		if entry.fn != actualLog[ix].fn {
			t.Logf("\nTEST: %s\n", name)
			tcm.dumpLog(t)
			t.Fatalf("\nExp fn: %s\nGot fn: %s\n", entry.fn, actualLog[ix].fn)
			return
		}
		for iy, param := range entry.params {
			if param != actualLog[ix].params[iy] {
				t.Logf("\nTEST: %s\n", name)
				tcm.dumpLog(t)
				t.Fatalf("\nExp param: %s\nGot param: %s\n",
					param, actualLog[ix].params[iy])
				return
			}
		}
	}
}

func (tcm *TestCompMgr) dumpLog(t *testing.T) {
	t.Logf("--- START TEST LOG ---\n")
	for _, entry := range tcm.tcmParams.testLog {
		t.Logf("%s:\n", entry.fn)
		for _, param := range entry.params {
			t.Logf("\t%s\n", param)
		}
	}
	t.Logf("---  END TEST LOG  ---\n")
}

// TestOpsMgr
type testOpsMgr struct {
	tcmParams *testCompParams
}

func newTestOpsMgr(tcmParams *testCompParams) *testOpsMgr {
	return &testOpsMgr{tcmParams: tcmParams}
}

func (tom *testOpsMgr) marshal(object interface{}) (string, error) {
	if s, ok := object.(string); ok {
		return s, nil
	}
	buf, err := rfc7951.Marshal(object)
	if err != nil {
		return "", mgmterror.NewMalformedMessageError()
	}
	return string(buf), nil
}

func (tom *testOpsMgr) unmarshal(encodedData string, object interface{}) error {
	if s, ok := object.(*string); ok {
		*s = encodedData
		return nil
	}
	err := rfc7951.Unmarshal([]byte(encodedData), object)
	if err != nil {
		return mgmterror.NewMalformedMessageError()
	}
	return nil
}

func (tom *testOpsMgr) Dial() error { return nil }

func (tom *testOpsMgr) SetConfigForModel(
	modelName string,
	object interface{},
) error {
	var err error

	cfg, err := tom.marshal(object)

	tom.tcmParams.committedConfig[modelName] = string(cfg)

	fmt.Printf("\tadd log entry\n")
	tom.addLogEntry(SetRunning, modelName, cfg)

	return err
}

func (tom *testOpsMgr) CheckConfigForModel(
	modelName string,
	object interface{},
) error {

	cfg, err := tom.marshal(object)

	tom.tcmParams.validatedConfig[modelName] = string(cfg)

	tom.addLogEntry(Validate, modelName, cfg)

	return err
}

func (tom *testOpsMgr) StoreConfigByModelInto(
	modelName string,
	object interface{},
) error {
	err := tom.unmarshal(tom.tcmParams.committedConfig[modelName], object)

	tom.addLogEntry(GetRunning, modelName, fmt.Sprintf("%v", object))

	return err
}

func (tom *testOpsMgr) StoreStateByModelInto(
	modelName string,
	object interface{},
) error {
	err := tom.unmarshal(tom.tcmParams.currentState[modelName], object)

	tom.addLogEntry(GetState, modelName, fmt.Sprintf("%v", object))

	return err
}

// TestSvcMgr
type testSvcMgr struct {
	tcmParams *testCompParams
}

func newTestSvcMgr(tcmParams *testCompParams) *testSvcMgr {
	return &testSvcMgr{tcmParams: tcmParams}
}

func (tsm *testSvcMgr) Close() { return }

// For now, assume any component is active.
func (tsm *testSvcMgr) IsActive(name string) (bool, error) {
	return true, nil
}
