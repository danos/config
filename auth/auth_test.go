// Copyright (c) 2018-2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package auth

import (
	"fmt"
	"testing"

	"github.com/danos/utils/audit"
	"github.com/danos/utils/pathutil"
)

func TestMatchPath(t *testing.T) {
	type pathTest struct {
		rulepath  []string
		ruleperm  AuthAction
		reqpath   []string
		reqperm   AuthPerm
		expresult bool
	}
	testTbl := []pathTest{
		{
			rulepath:  []string{"interfaces", "dataplane", "dp0s9"},
			ruleperm:  AUTH_ALLOW,
			reqpath:   []string{"interfaces", "dataplane", "dp0s9"},
			reqperm:   P_DELETE,
			expresult: true,
		},
		{
			rulepath:  []string{"interfaces", "dataplane", "dp0s9"},
			ruleperm:  AUTH_ALLOW,
			reqpath:   []string{"interfaces"},
			reqperm:   P_DELETE,
			expresult: false,
		},
		{
			rulepath:  []string{"interfaces", "dataplane", "dp0s9"},
			ruleperm:  AUTH_ALLOW,
			reqpath:   []string{"interfaces"},
			reqperm:   P_CREATE,
			expresult: true,
		},
		{
			rulepath:  []string{"interfaces", "dataplane", "dp0s9"},
			ruleperm:  AUTH_DENY,
			reqpath:   []string{"interfaces"},
			reqperm:   P_CREATE,
			expresult: false,
		},
		{
			rulepath:  []string{"interfaces", "dataplane", "*"},
			ruleperm:  AUTH_ALLOW,
			reqpath:   []string{"interfaces", "dataplane", "dp0s9"},
			reqperm:   P_UPDATE,
			expresult: true,
		},
		{
			rulepath:  []string{"interfaces", "dataplane", "*"},
			ruleperm:  AUTH_DENY,
			reqpath:   []string{"interfaces", "dataplane", "dp0s9"},
			reqperm:   P_UPDATE,
			expresult: true,
		},
		{
			rulepath:  []string{"interfaces", "dataplane"},
			ruleperm:  AUTH_ALLOW,
			reqpath:   []string{""},
			reqperm:   P_UPDATE,
			expresult: true,
		},
		{
			rulepath:  []string{"interfaces", "dataplane"},
			ruleperm:  AUTH_DENY,
			reqpath:   []string{""},
			reqperm:   P_UPDATE,
			expresult: false,
		},
		{
			rulepath:  []string{"*"},
			ruleperm:  AUTH_DENY,
			reqpath:   []string{"interfaces", "dataplane", "dp0s9"},
			reqperm:   P_DELETE,
			expresult: true,
		},
		{
			rulepath:  []string{"interfaces", "dataplane", "dp0s9"},
			ruleperm:  AUTH_ALLOW,
			reqpath:   []string{"interfaces", "dataplane", "dp0s3"},
			reqperm:   P_CREATE,
			expresult: false,
		},
		{
			rulepath:  []string{},
			ruleperm:  AUTH_ALLOW,
			reqpath:   []string{"interfaces", "dataplane", "dp0s3"},
			reqperm:   P_CREATE,
			expresult: false,
		},
		{
			rulepath:  []string{"interfaces", "*", "dp0s9"},
			ruleperm:  AUTH_ALLOW,
			reqpath:   []string{"interfaces", "dataplane", "dp0s9"},
			reqperm:   P_CREATE,
			expresult: true,
		},
	}
	for _, test := range testTbl {
		match := matchPath(test.rulepath, test.reqpath, test.ruleperm, test.reqperm)
		if match != test.expresult {
			t.Fatalf("Unexpected matchPath result\n"+
				"    Rule: %s\n    RulePerm: %s\n"+
				"    Request: %s\n    ReqPerm: %s\n"+
				"    Result: %t\n    Expected: %t\n",
				test.rulepath, test.ruleperm.String(),
				test.reqpath, test.reqperm.String(),
				match, test.expresult)
		}
	}
}

func genLogReqPathMsg(uid uint32, path string, perm AuthPerm) string {
	return fmt.Sprintf("uid=%d req: {path=[%s], perm=%s}", uid, path, perm.String())
}

func TestLogReqPath(t *testing.T) {
	a := newAuthForTest()
	auditer := a.authGlobal.auditer.(*audit.TestAudit)
	adb, _ := a.load()

	// Generate the path being requested, and corresponding attributes
	path := []string{"set", "foo", "bar"}
	pathAttrs := pathutil.NewPathAttrs()
	pathAttrs.Attrs = append(pathAttrs.Attrs,
		pathutil.PathElementAttrs{Secret: false},
		pathutil.PathElementAttrs{Secret: false},
		pathutil.PathElementAttrs{Secret: false})

	a.LogReqPath(adb.Uid, path, &pathAttrs, P_READ, true)

	// Log another request, this time with a secret element
	pathAttrs.Attrs[2].Secret = true
	a.LogReqPath(adb.Uid, path, &pathAttrs, P_UPDATE, false)

	expUserLogs := audit.UserLogSlice{
		audit.UserLog{audit.LOG_TYPE_USER_CFG,
			genLogReqPathMsg(adb.Uid, "set foo bar", P_READ), 1},
		audit.UserLog{audit.LOG_TYPE_USER_CFG,
			genLogReqPathMsg(adb.Uid, "set foo **", P_UPDATE), 0},
	}
	audit.AssertUserLogSliceEqual(t, expUserLogs, auditer.GetUserLogs())
}

func TestLogReqPathRedactionFailure(t *testing.T) {
	a := newAuthForTest()
	auditer := a.authGlobal.auditer.(*audit.TestAudit)
	adb, _ := a.load()

	// Don't pass any path attributes
	path := []string{"delete", "bar", "baz"}
	a.LogReqPath(adb.Uid, path, nil, P_READ, true)

	// Log another request, this time with mismatched length of path and attrs
	pathAttrs := pathutil.NewPathAttrs()
	pathAttrs.Attrs = append(pathAttrs.Attrs, pathutil.PathElementAttrs{Secret: true})
	a.LogReqPath(adb.Uid, path, &pathAttrs, P_CREATE, false)

	expUserLogs := audit.UserLogSlice{
		audit.UserLog{audit.LOG_TYPE_USER_CFG,
			genLogReqPathMsg(adb.Uid, "<path redaction failed>", P_READ), 1},
		audit.UserLog{audit.LOG_TYPE_USER_CFG,
			genLogReqPathMsg(adb.Uid, "<path redaction failed>", P_CREATE), 0},
	}
	audit.AssertUserLogSliceEqual(t, expUserLogs, auditer.GetUserLogs())
}
