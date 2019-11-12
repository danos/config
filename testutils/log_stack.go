// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// Utilities used by configd unit tests to log the stack.

package testutils

import (
	"runtime"
	"testing"
)

func LogStack(t *testing.T) {
	LogStackInternal(t, false)
}

func LogStackFatal(t *testing.T) {
	LogStackInternal(t, true)
}

func LogStackInternal(t *testing.T, fatal bool) {
	stack := make([]byte, 4096)
	runtime.Stack(stack, false)
	if fatal {
		t.Fatalf("%s", stack)
	} else {
		t.Logf("%s", stack)
	}
}
