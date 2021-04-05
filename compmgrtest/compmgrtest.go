// Copyright (c) 2021, AT&T Intellectual Property.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package compmgrtest

import (
	"fmt"
	"testing"

	"github.com/danos/encoding/rfc7951"
	"github.com/danos/mgmterror"
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

func NewLogEntry(fn string, params ...string) TestLogEntry {
	return TestLogEntry{fn: fn, params: params}
}

// TestCompMgr allows replacement of ComponentManager functionality at a high
// level where we don't want an actual (d)bus or components created. Instead
// we want to verify that the correct tuple of {config, model} has been passed
// to the bus (for validate and commit) or provide the ability to return the
// expected config (or an error) for 'get' operations.
type TestCompMgr struct {
	//OperationsManager
	t *testing.T

	validatedConfig map[string]string
	committedConfig map[string]string
	currentState    map[string]string

	testLog []TestLogEntry
}

// Compile time check that the concrete type meets the interface
//var _ ComponentManager = (*TestCompMgr)(nil)

func NewTestCompMgr(t *testing.T) *TestCompMgr {
	compMgr := &TestCompMgr{t: t}
	compMgr.testLog = make([]TestLogEntry, 0)

	compMgr.validatedConfig = make(map[string]string, 0)
	compMgr.committedConfig = make(map[string]string, 0)
	compMgr.currentState = make(map[string]string, 0)

	return compMgr
}

func (tcm *TestCompMgr) Dial() error { return nil }

// Config / state management.

func (tcm *TestCompMgr) ValidatedConfig(model string) string {
	cfg, ok := tcm.validatedConfig[model]
	if !ok {
		tcm.t.Fatalf("No validated config for %s", model)
	}
	return cfg
}

func (tcm *TestCompMgr) CommittedConfig(model string) string {
	cfg, ok := tcm.committedConfig[model]
	if !ok {
		tcm.t.Fatalf("No committed config for %s", model)
	}
	return cfg
}

func (tcm *TestCompMgr) CurrentState(model string) string {
	cfg, ok := tcm.currentState[model]
	if !ok {
		tcm.t.Fatalf("No current state for %s", model)
	}
	return cfg
}

func (tcm *TestCompMgr) SetCurrentState(model, stateJson string) {
	tcm.currentState[model] = stateJson
}

// Log management.

func (tcm *TestCompMgr) addLogEntry(fn string, params ...string) {
	fmt.Printf("Add %s\n", fn)
	tcm.testLog = append(tcm.testLog, NewLogEntry(fn, params...))
}

func (tcm *TestCompMgr) ClearLogEntries() {
	fmt.Printf("Clear log\n")
	tcm.testLog = nil
}

func (tcm *TestCompMgr) filteredLogEntries(filter string) []TestLogEntry {
	retLog := make([]TestLogEntry, 0)

	for _, entry := range tcm.testLog {
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
	actualLog := tcm.testLog
	fmt.Printf("Entries: %d\n", len(tcm.testLog))
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
	for _, entry := range tcm.testLog {
		t.Logf("%s:\n", entry.fn)
		for _, param := range entry.params {
			t.Logf("\t%s\n", param)
		}
	}
	t.Logf("---  END TEST LOG  ---\n")
}

func (tcm *TestCompMgr) marshal(object interface{}) (string, error) {
	if s, ok := object.(string); ok {
		return s, nil
	}
	buf, err := rfc7951.Marshal(object)
	if err != nil {
		return "", mgmterror.NewMalformedMessageError()
	}
	return string(buf), nil
}

func (tcm *TestCompMgr) unmarshal(encodedData string, object interface{}) error {
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

func (tcm *TestCompMgr) SetConfigForModel(
	modelName string,
	object interface{},
) error {
	var err error

	cfg, err := tcm.marshal(object)

	tcm.committedConfig[modelName] = string(cfg)

	fmt.Printf("\tadd log entry\n")
	tcm.addLogEntry(SetRunning, modelName, cfg)

	return err
}

func (tcm *TestCompMgr) CheckConfigForModel(
	modelName string,
	object interface{},
) error {

	cfg, err := tcm.marshal(object)

	tcm.validatedConfig[modelName] = string(cfg)

	tcm.addLogEntry(Validate, modelName, cfg)

	return err
}

func (tcm *TestCompMgr) StoreConfigByModelInto(
	modelName string,
	object interface{},
) error {
	err := tcm.unmarshal(tcm.committedConfig[modelName], object)

	tcm.addLogEntry(GetRunning, modelName, fmt.Sprintf("%v", object))

	return err
}

func (tcm *TestCompMgr) StoreStateByModelInto(
	modelName string,
	object interface{},
) error {
	err := tcm.unmarshal(tcm.currentState[modelName], object)

	tcm.addLogEntry(GetState, modelName, fmt.Sprintf("%v", object))

	return err
}

func (tcm *TestCompMgr) CloseSvcMgr() { return }

// For now, assume any component is active.
func (tcm *TestCompMgr) IsActive(name string) (bool, error) {
	return true, nil
}
