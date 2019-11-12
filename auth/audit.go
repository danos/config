// Copyright (c) 2019, AT&T Intellectual Property Inc. All rights reserved.
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

func (a *AuditAccounter) AccountCommand(
	uid uint32, groups []string, cmd []string, pathAttrs *pathutil.PathAttrs,
) {
	cmd, _ = pathutil.RedactPath(cmd, pathAttrs)
	a.auth.authGlobal.auditer.LogUserCmd(
		fmt.Sprintf("run: %s, for user: %d", strings.Join(cmd, " "), uid), 1)
}
