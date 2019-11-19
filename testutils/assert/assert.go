// Copyright (c) 2019, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2016-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// Useful test functions for validating (mostly) string outputs match
// what is expected.

package assert

import (
	"bytes"
	"github.com/danos/mgmterror"
	"github.com/danos/utils/exec"
	"io"
	"os"
	"strings"
	"testing"
)

func init() {
	exec.NewExecError = func(path []string, err string) error {
		return mgmterror.NewExecError(path, err)
	}
}

type ExpectedError struct {
	expected string
}

func NewExpectedError(expect string) *ExpectedError {
	return &ExpectedError{expected: expect}
}

func (e *ExpectedError) Matches(t *testing.T, actual error) {
	if actual == nil {
		t.Fatalf("Unexpected success")
	}

	CheckStringDivergence(t, e.expected, actual.Error())
}

type ExpectedMessages struct {
	expected []string
}

func NewExpectedMessages(expect ...string) *ExpectedMessages {
	return &ExpectedMessages{expected: expect}
}

func (e *ExpectedMessages) ContainedIn(t *testing.T, actual string) {
	if len(actual) == 0 {
		t.Fatalf("No output in which to search for expected message(s).")
		return
	}

	for _, exp := range e.expected {
		if !strings.Contains(actual, exp) {
			t.Fatalf("Actual output doesn't contain expected output:\n"+
				"Exp:\n%s\nAct:\n%v\n", exp, actual)
		}
	}
}

func (e *ExpectedMessages) NotContainedIn(t *testing.T, actual string) {
	if len(actual) == 0 {
		t.Fatalf("No output in which to search for expected message(s).")
		return
	}

	for _, exp := range e.expected {
		if strings.Contains(actual, exp) {
			t.Fatalf("Actual output contain unexpected output:\n"+
				"NotExp:\n%s\nAct:\n%v\n", exp, actual)
		}
	}
}

// Check each expected message appears in at least one of the actual strings.
func (e *ExpectedMessages) ContainedInAny(t *testing.T, actual []string) {
	if len(actual) == 0 {
		t.Fatalf("No output in which to search for expected message(s).")
		return
	}

outerLoop:
	for _, exp := range e.expected {
		for _, act := range actual {
			if strings.Contains(act, exp) {
				continue outerLoop
			}
		}

		t.Fatalf("Actual output doesn't contain expected output:\n"+
			"Exp:\n%s\nAct:\n%v\n", exp, actual)
	}
}

// Very useful when debugging outputs that don't match up.
func CheckStringDivergence(t *testing.T, expOut, actOut string) {
	if expOut == actOut {
		return
	}

	var expOutCopy = expOut
	var act bytes.Buffer
	var charsToDump = 10
	var expCharsToDump = 10
	var actCharsLeft, expCharsLeft int
	for index, char := range actOut {
		if len(expOutCopy) > 0 {
			if char == rune(expOutCopy[0]) {
				act.WriteByte(byte(char))
			} else {
				act.WriteString("###") // Mark point of divergence.
				expCharsLeft = len(expOutCopy)
				actCharsLeft = len(actOut) - index
				if expCharsLeft < charsToDump {
					expCharsToDump = expCharsLeft
				}
				if actCharsLeft < charsToDump {
					charsToDump = actCharsLeft
				}
				act.WriteString(actOut[index : index+charsToDump])
				break
			}
		} else {
			t.Logf("Expected output terminates early.\n")
			t.Fatalf("Exp:\n%s\nGot extra:\n%s\n",
				expOut[:index], act.String()[index:])
		}
		expOutCopy = expOutCopy[1:]
	}

	// When expOut is longer than actOut, need to update the expCharsToDump
	if len(expOutCopy) < charsToDump {
		expCharsToDump = len(expOutCopy)
	}

	// Useful to print whole output first for reference (useful when debugging
	// when you don't want to have to construct the expected output up front).
	t.Logf("Actual output:\n%s\n--- ENDS ---\n", actOut)

	// After that we then print up to the point of divergence so it's easy to
	// work out what went wrong ...
	t.Fatalf("Unexpected output.\nGot:\n%s\nExp at ###:\n'%s ...'\n",
		act.String(), expOutCopy[:expCharsToDump])
}

type actionFn func() ([]*exec.Output, []error, bool)

// For some tests we need to capture stdout for validation with expected
// output.  Code here is based on various similar code snippets found by
// googling StackOverflow and the like.
func RunTestAndCaptureStdout(
	fn actionFn,
) (out []*exec.Output, retErr []error, result bool, debug string) {

	// Save 'stdout' so we can restore later.
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, false, ""
	}
	os.Stdout = w
	outC := make(chan string)

	// Set up go routine to collect stdout.
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		r.Close()
		if err != nil {
			return
		}
		outC <- buf.String()
	}()

	result = true

	// Clean up in a deferred call so we can recover.
	defer func() {
		// Close pipe, restore stdout, get output.
		w.Close()
		os.Stdout = stdout
		debug = <-outC

		err := recover()
		if err != nil {
			panic(err)
		}
	}()

	// Run our test
	out, retErr, result = fn()

	return out, retErr, result, debug
}
