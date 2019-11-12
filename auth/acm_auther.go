// Copyright (c) 2018-2019, AT&T Intellectual Property Inc.
// All rights reserved.
//
// Copyright (c) 2014-2015, 2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package auth

import (
	"fmt"
	"strconv"

	"github.com/danos/utils/pathutil"
)

type AcmAuther struct {
	CommandAuther
	DataAuther
	auth *Auth
}

func NewAcmAuther(auth *Auth) *AcmAuther {
	acmAuther := &AcmAuther{
		auth: auth,
	}
	return acmAuther
}

func (a *AcmAuther) AuthorizeCommand(uid uint32, groups []string, cmd []string, pathAttrs *pathutil.PathAttrs) bool {
	// Command authorization not supported
	return true
}

func (a *AcmAuther) AuthorizePath(uid uint32, groups []string, path []string, pathAttrs *pathutil.PathAttrs, perm AuthPerm) bool {
	var result bool
	adb, _ := a.auth.load()
	if !adb.Enabled {
		result = true
		return result
	}
	if uid == adb.Uid {
		return true
	}
	defer func() { a.auth.LogReqPath(uid, path, pathAttrs, perm, result) }()

	for _, rulet := range adb.Rules {
		if rulet.Type != AUTH_T_DATA {
			continue
		}
		rule := rulet.Rule
		rpath := pathutil.Makepath(rule.Path)
		if matchPath(rpath, path, rule.Action, perm) &&
			matchGroup(groups, rule.Groups) &&
			(rule.Perm&perm == perm) {
			switch {
			case rule.Action&AUTH_DENY == AUTH_DENY:
				result = false
				a.auth.Log(uid, rule, result)
				return result
			case rule.Action&AUTH_ALLOW == AUTH_ALLOW:
				result = true
				a.auth.Log(uid, rule, result)
				return result
			}
		}
	}
	result = adb.authorizeDefault(perm)
	return result
}

func (a *AcmAuther) AuthorizeRead(uid uint32, groups []string, path []string, pathAttrs *pathutil.PathAttrs) bool {
	return a.AuthorizePath(uid, groups, path, pathAttrs, P_READ)
}

func (a *AcmAuther) AuthorizeCreate(uid uint32, groups []string, path []string, pathAttrs *pathutil.PathAttrs) bool {
	return a.AuthorizePath(uid, groups, path, pathAttrs, P_CREATE)
}

func (a *AcmAuther) AuthorizeUpdate(uid uint32, groups []string, path []string, pathAttrs *pathutil.PathAttrs) bool {
	return a.AuthorizePath(uid, groups, path, pathAttrs, P_UPDATE)
}

func (a *AcmAuther) AuthorizeDelete(uid uint32, groups []string, path []string, pathAttrs *pathutil.PathAttrs) bool {
	return a.AuthorizePath(uid, groups, path, pathAttrs, P_DELETE)
}

func (a *AcmAuther) GetPerms(groups []string) map[string]string {
	m := make(map[string]string)

	adb, _ := a.auth.load()
	if !adb.Enabled {
		m["DEFAULT"] = strconv.Itoa(int(P_CREATE | P_READ | P_UPDATE | P_DELETE | P_EXECUTE))
		return m
	}

	defaultperm := 0
	if adb.CreateDefault == AUTH_ALLOW {
		defaultperm = defaultperm | int(P_CREATE)
	}
	if adb.ReadDefault == AUTH_ALLOW {
		defaultperm = defaultperm | int(P_READ)
	}
	if adb.UpdateDefault == AUTH_ALLOW {
		defaultperm = defaultperm | int(P_UPDATE)
	}
	if adb.DeleteDefault == AUTH_ALLOW {
		defaultperm = defaultperm | int(P_DELETE)
	}
	if adb.ExecDefault == AUTH_ALLOW {
		defaultperm = defaultperm | int(P_EXECUTE)
	}

	for i, rulet := range adb.Rules {
		if rulet.Type&AUTH_T_DATA != AUTH_T_DATA {
			continue
		}

		rule := rulet.Rule

		if !matchGroup(groups, rule.Groups) {
			continue
		}

		perm := defaultperm
		if rule.Action&AUTH_DENY == AUTH_DENY {
			perm = perm & ^int(rule.Perm)
		} else {
			perm = perm | int(rule.Perm)
		}

		key := fmt.Sprintf("%d: %s", i, rule.Path)
		m[key] = strconv.Itoa(perm)
	}
	m["DEFAULT"] = strconv.Itoa(defaultperm)
	return m
}
