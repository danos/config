// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package auth

import (
	"github.com/danos/utils/pathutil"
)

type TestCommandAccounter struct {
	CommandAccounter
	cmdAcctReqs TestAutherRequests
}

func (a *TestCommandAccounter) AccountCommand(
	uid uint32,
	groups []string,
	cmd []string,
	pathAttrs *pathutil.PathAttrs,
) {
	// For now we just log command accounting requests for later validation
	req := NewTestAutherCommandRequest(cmd, pathAttrs)
	a.cmdAcctReqs.Reqs = append(a.cmdAcctReqs.Reqs, req)
}

type TestCommandAuther struct {
	CommandAuther
	cmdReqs TestAutherRequests
}

func (a *TestCommandAuther) AuthorizeCommand(
	uid uint32,
	groups []string,
	cmd []string,
	pathAttrs *pathutil.PathAttrs,
) bool {
	// For now we just log command authorization requests for later validation
	req := NewTestAutherCommandRequest(cmd, pathAttrs)
	a.cmdReqs.Reqs = append(a.cmdReqs.Reqs, req)
	return true
}

type TestDataAuther struct {
	DataAuther
	rules []testRule
}

// <rulePerm> may consist of multiple operations, whereas <reqPerm> is the
// specific operation being requested.
func ruleOpMatches(rulePerm int, reqPerm AuthPerm) bool {
	val := rulePerm & int(reqPerm)
	if val > 0 {
		return true
	}
	return false
}

// To determine if <reqPath> is authorized, given <reqPerm>issions, we run
// through all the rules in turn, checking:
//
// - Does rule apply to the requested operation?  Skip if not
// - If rule covers all paths and requested path is '*', ALLOW
// - If rule's path is longer than requested path, cannot match.  Skip.
// - See if we match all of rule's path.  ALLOW if so.
//
// Default is to return DENY.
//
func (a *TestDataAuther) allowed(reqPath []string, reqPerm AuthPerm, pathAttrs *pathutil.PathAttrs) bool {
	for _, rule := range a.rules {
		if !ruleOpMatches(rule.perm, reqPerm) {
			continue
		}

		if len(reqPath) == 0 && len(rule.path) == 1 && rule.path[0] == "*" {
			return rule.action == true
		}

		// Can't match if rule's path is longer than one we are checking.
		if len(reqPath) < len(rule.path) {
			continue
		}

		matched := true
		for index, elem := range rule.path {
			if elem != "*" && elem != reqPath[index] {
				matched = false
				break
			}
		}
		if matched == true {
			return rule.action == true
		}
	}
	return Deny
}

func (a *TestDataAuther) AuthorizeRead(
	uid uint32,
	groups []string,
	path []string,
	pathAttrs *pathutil.PathAttrs,
) bool {
	return a.allowed(path, P_READ, pathAttrs)
}

func (a *TestDataAuther) AuthorizeCreate(
	uid uint32,
	groups []string,
	path []string,
	pathAttrs *pathutil.PathAttrs,
) bool {
	return a.allowed(path, P_CREATE, pathAttrs)
}

func (a *TestDataAuther) AuthorizeUpdate(
	uid uint32,
	groups []string,
	path []string,
	pathAttrs *pathutil.PathAttrs,
) bool {
	return a.allowed(path, P_UPDATE, pathAttrs)
}

func (a *TestDataAuther) AuthorizeDelete(
	uid uint32,
	groups []string,
	path []string,
	pathAttrs *pathutil.PathAttrs,
) bool {
	return a.allowed(path, P_DELETE, pathAttrs)
}

func (a *TestDataAuther) AuthorizePath(
	uid uint32,
	groups []string,
	path []string,
	pathAttrs *pathutil.PathAttrs,
	perm AuthPerm,
) bool {
	return a.allowed(path, perm, pathAttrs)
}

// Not yet implemented so just return nil for now.
func (a *TestDataAuther) GetPerms(groups []string) map[string]string {
	return nil
}
