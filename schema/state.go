// Copyright (c) 2017-2019, 2021, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2016-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"fmt"
	"os"
	"strings"
	"time"

	spawn "os/exec"

	"github.com/danos/mgmterror"
	"github.com/danos/utils/pathutil"
	yang "github.com/danos/yang/schema"
)

type StateLogger interface {
	Printf(format string, a ...interface{})
	Println(a ...interface{})
}

type getStateFn func(path []string, logger StateLogger) ([]byte, error)

type hasState interface {
	GetStateJson(path []string) ([][]byte, error)
	GetStateJsonWithWarnings(
		path []string,
		logger StateLogger,
	) ([][]byte, []error)
	StateChildren() []yang.Node
	addStateFn(getStateFn)
	hasState() bool
}

type state struct {
	getStateFns   []getStateFn
	stateChildren []yang.Node
}

func (s *state) addStateFn(fn getStateFn) {
	s.getStateFns = append(s.getStateFns, fn)
}

func newState(node yang.Node, ext *extensions) *state {
	var stateFn []getStateFn
	var stateCh []yang.Node

	for _, v := range ext.ConfigdExt().GetState {
		stateScript := v
		newFn := func(path []string, logger StateLogger) ([]byte, error) {
			if logger != nil {
				logger.Printf("%s: %v %s\n",
					stateLogMsgPrefix, path, stateScript)
			}
			return getJsonFromStateScript(stateScript, path)
		}
		stateFn = append(stateFn, newFn)
	}

	for _, c := range node.Children() {
		if c.Config() && c.HasPresence() {
			continue
		}
		if c.(ExtendedNode).hasState() {
			stateCh = append(stateCh, c)
		}
	}

	return &state{stateFn, stateCh}
}

func (s *state) StateChildren() []yang.Node {
	return s.stateChildren
}

func (s *state) hasState() bool {
	return len(s.StateChildren()) > 0 || len(s.getStateFns) > 0
}

func resolvePath(name, path string) string {
	if strings.Contains(name, "/") {
		// Contains a path separator, return unaltered
		return name
	}

	pathDirs := strings.Split(path, ":")
	for _, dir := range pathDirs {
		// Build possible absolute path and check if it exists
		abs := dir + "/" + name
		d, _ := os.Stat(abs)
		if d != nil {
			// Found a match
			return abs
		}
	}

	// Not found, return unaltered
	return name
}

func getJsonFromStateScript(getState string, path []string) ([]byte, error) {

	getStateArgs := strings.Split(getState, " ")

	scriptPaths := "/bin:/usr/bin:/sbin:/usr/sbin:/opt/vyatta/bin:/opt/vyatta/sbin"
	name := resolvePath(getStateArgs[0], scriptPaths)
	c := spawn.Command(name, getStateArgs[1:]...)
	c.Env = append(os.Environ(),
		"PATH="+scriptPaths,
		"CONFIGD_PATH="+pathutil.Pathstr(path))

	// TODO: Seperate stdout from stderr
	output, err := c.CombinedOutput()
	if err != nil {
		if _, ok := err.(*spawn.ExitError); !ok {
			cerr := mgmterror.NewOperationFailedApplicationError()
			cerr.Path = pathutil.Pathstr(getStateArgs)
			cerr.Message = fmt.Sprintf("Failure to spawn GetState process: %s", err.Error())
			return nil, cerr
		}
	}
	if !c.ProcessState.Success() {
		cerr := mgmterror.NewOperationFailedApplicationError()
		cerr.Path = pathutil.Pathstr(getStateArgs)
		cerr.Message = fmt.Sprintf("GetState failure: %s", string(output))
		return nil, cerr
	}

	return output, nil
}

// GetStateJson - retrieve state information from component for given path.
//
// NB: It's possible the component function may return an empty JSON string.
//     In such cases, we ignore it here as it's pointless passing it up the
//     calling tree.
func (s *state) GetStateJson(path []string) ([][]byte, error) {
	var all_json_state [][]byte

	for _, fn := range s.getStateFns {
		json_state, err := fn(path, nil)
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(string(json_state), "{}") {
			continue
		}
		all_json_state = append(all_json_state, json_state)
	}
	return all_json_state, nil
}

// GetStateJsonWithWarnings - retrieve state info, inc warnings, for given path
//
// NB: Rather than fail on an error, we run all state functions and return
//     as much JSON as we can, along with relevant warnings.  No error here
//     is considered fatal.
func (s *state) GetStateJsonWithWarnings(
	path []string,
	logger StateLogger,
) ([][]byte, []error) {

	var allJsonState [][]byte
	var warnings []error

	start := time.Now()
	count := 0

	for _, fn := range s.getStateFns {
		count++
		jsonState, err := fn(path, logger)
		if err != nil {
			cerr := mgmterror.NewOperationFailedApplicationError()
			cerr.Path = pathutil.Pathstr(path)
			cerr.Message = fmt.Sprintf("Failed to run state fn. Error: %s", err)
			warnings = append(warnings, cerr)
			continue
		}
		if strings.HasPrefix(string(jsonState), "{}") {
			continue
		}
		allJsonState = append(allJsonState, jsonState)
	}
	if count > 0 {
		logStateTime(logger, fmt.Sprintf("%v %d script(s)", path, count),
			start)
	}
	return allJsonState, warnings
}

const (
	stateLogMsgPrefix = "STATE"
	msgPadToLength    = 20
	// 20 + 3 extra for luck
	msgPadding = "                      "
)

func msgPad(msg string) string {
	msgLen := len(msg)
	padLen := 0
	if msgLen < msgPadToLength {
		padLen = msgPadToLength - msgLen
	}
	return msg + ": " + msgPadding[:padLen]
}

func logStateTime(logger StateLogger, msg string, startTime time.Time) {
	if logger == nil {
		return
	}
	logger.Printf("%s: %s%s", stateLogMsgPrefix, msgPad(msg),
		time.Since(startTime).Round(time.Millisecond))
}

func logStateEvent(logger StateLogger, msg string) {
	if logger == nil {
		return
	}
	logger.Printf("%s: %s", stateLogMsgPrefix, msg)
}
