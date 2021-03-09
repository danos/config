// Copyright (c) 2019, AT&T Intellectual Property Inc. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package auth

import (
	"fmt"
	"testing"

	"github.com/danos/utils/audit"
	"github.com/danos/utils/pathutil"
)

func genAuditCmdAccounterMsg(cmd string, uid uint32) string {
	return fmt.Sprintf("run: %s, for user: %d", cmd, uid)
}

func TestAccountCommandAudit(t *testing.T) {
	a := newAuthForTest()
	auditer := a.authGlobal.auditer.(*audit.TestAudit)

	// Generate the command, and corresponding attributes
	cmd := []string{"delete", "foo", "bar"}
	pathAttrs := pathutil.NewPathAttrs()
	pathAttrs.Attrs = append(pathAttrs.Attrs,
		pathutil.PathElementAttrs{Secret: false},
		pathutil.PathElementAttrs{Secret: false},
		pathutil.PathElementAttrs{Secret: false})

	a.AccountCommand(1000, []string{}, cmd, &pathAttrs)

	// Log another command, this time with a secret element
	pathAttrs.Attrs[2].Secret = true
	a.AccountCommand(1000, []string{}, cmd, &pathAttrs)

	expUserLogs := audit.UserLogSlice{
		audit.UserLog{
			Type: audit.LOG_TYPE_USER_CMD,
			Msg: genAuditCmdAccounterMsg("delete foo bar", 1000),
			Result: 1,
		},
		audit.UserLog{
			Type: audit.LOG_TYPE_USER_CMD,
			Msg: genAuditCmdAccounterMsg("delete foo **", 1000),
			Result: 1,
		},
	}
	audit.AssertUserLogSliceEqual(t, expUserLogs, auditer.GetUserLogs())
}
