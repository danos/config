// Copyright (c) 2019-2020, AT&T Intellectual Property Inc. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package auth

import (
	"fmt"
	"strings"

	"github.com/danos/utils/pathutil"
)

type AuditAccounter struct {
	CommandAccounter
	auth *Auth
}

func NewAuditAccounter(a *Auth) *AuditAccounter {
	return &AuditAccounter{auth: a}
}

type taskAuditer struct {
	auth        *Auth
	redactedCmd string
	uid         uint32
}

func (a taskAuditer) AccountStart() error {
	// No-op for audit accounting
	return nil
}

func (a taskAuditer) AccountStop(_ *error) error {
	a.auth.authGlobal.auditer.LogUserCmd(
		fmt.Sprintf("run: %s, for user: %d", a.redactedCmd, a.uid), 1)
	return nil
}

func (a *AuditAccounter) NewTaskAccounter(
	uid uint32, groups []string, cmd []string, pathAttrs *pathutil.PathAttrs,
) TaskAccounter {
	cmd, _ = pathutil.RedactPath(cmd, pathAttrs)
	return taskAuditer{
		auth:        a.auth,
		redactedCmd: strings.Join(cmd, " "),
		uid:         uid,
	}
}

func (a *AuditAccounter) AccountCommand(
	uid uint32, groups []string, cmd []string, pathAttrs *pathutil.PathAttrs,
) {
	a.NewTaskAccounter(uid, groups, cmd, pathAttrs).AccountStop(nil)
}
