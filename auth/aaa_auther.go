// Copyright (c) 2019, AT&T Intellectual Property Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package auth

import (
	"github.com/danos/aaa"
	"github.com/danos/utils/pathutil"
)

type AaaAuther struct {
	CommandAccounter
	CommandAuther
	DataAuther
	auth  *Auth
	proto *aaa.AAAProtocol
}

func NewAaaAuther(auth *Auth, proto *aaa.AAAProtocol) *AaaAuther {
	aaaAuther := &AaaAuther{
		auth:  auth,
		proto: proto,
	}
	return aaaAuther
}

func authEnvToMap(env *AuthEnv) map[string]string {
	return map[string]string{"tty": env.Tty}
}

func (a *AaaAuther) AccountCommand(uid uint32, groups []string, cmd []string, pathAttrs *pathutil.PathAttrs) {
	err := a.proto.Plugin.Account("conf-mode", uid, groups, cmd, pathAttrs, authEnvToMap(&a.auth.env))
	if err != nil {
		a.auth.authGlobal.Elog.Printf("Accounting error via AAA protocol %s: %s",
			a.proto.Cfg.Name, err)
	}
}

func (a *AaaAuther) AuthorizeCommand(uid uint32, groups []string, cmd []string, pathAttrs *pathutil.PathAttrs) bool {
	authed, err := a.proto.Plugin.Authorize("conf-mode", uid, groups, cmd, pathAttrs)
	if err != nil {
		a.auth.authGlobal.Elog.Printf("Authorization error via AAA protocol %s: %v",
			a.proto.Cfg.Name, err)
	}
	return authed
}

func (a *AaaAuther) AuthorizePath(uid uint32, groups []string, path []string, pathAttrs *pathutil.PathAttrs, perm AuthPerm) bool {
	// Data authorization not supported
	return true
}

func (a *AaaAuther) AuthorizeRead(uid uint32, groups []string, path []string, pathAttrs *pathutil.PathAttrs) bool {
	// Data authorization not supported
	return true
}

func (a *AaaAuther) AuthorizeCreate(uid uint32, groups []string, path []string, pathAttrs *pathutil.PathAttrs) bool {
	// Data authorization not supported
	return true
}

func (a *AaaAuther) AuthorizeUpdate(uid uint32, groups []string, path []string, pathAttrs *pathutil.PathAttrs) bool {
	// Data authorization not supported
	return true
}

func (a *AaaAuther) AuthorizeDelete(uid uint32, groups []string, path []string, pathAttrs *pathutil.PathAttrs) bool {
	// Data authorization not supported
	return true
}

func (a *AaaAuther) GetPerms(groups []string) map[string]string {
	// Data authorization not supported
	return map[string]string{}
}
