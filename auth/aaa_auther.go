// Copyright (c) 2019-2020, AT&T Intellectual Property Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package auth

import (
	"errors"

	"github.com/danos/aaa"
	"github.com/danos/utils/guard"
	"github.com/danos/utils/pathutil"
)

func authEnvToMap(env *AuthEnv) map[string]string {
	return map[string]string{"tty": env.Tty}
}

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

type aaaTask struct {
	a *AaaAuther
	t aaa.AAATask
}

func (a aaaTask) AccountStart() error {
	err := guard.CatchPanicErrorOnly(func() error {
		return a.t.AccountStart()
	})
	if err != nil {
		a.a.auth.authGlobal.Elog.Printf("Start accounting error via AAA protocol %s: %s",
			a.a.proto.Cfg.Name, err)
	}
	return err
}

func (a aaaTask) AccountStop(taskErr *error) error {
	err := guard.CatchPanicErrorOnly(func() error {
		return a.t.AccountStop(taskErr)
	})
	if err != nil {
		a.a.auth.authGlobal.Elog.Printf("Stop accounting error via AAA protocol %s: %s",
			a.a.proto.Cfg.Name, err)
	}
	return err
}

func (a *AaaAuther) newTask(
	uid uint32, groups []string, cmd []string, pathAttrs *pathutil.PathAttrs,
) (*aaaTask, error) {
	t, err := guard.CatchPanic(func() (interface{}, error) {
		return a.proto.Plugin.NewTask(
			"conf-mode", uid, groups, cmd, pathAttrs, authEnvToMap(&a.auth.env))
	})
	if t == nil && err == nil {
		err = errors.New("No task object")
	}
	if err != nil {
		a.auth.authGlobal.Elog.Printf("Accounting error via AAA protocol %s: %s",
			a.proto.Cfg.Name, err)
		return nil, err
	}
	return &aaaTask{a, t.(aaa.AAATask)}, err
}

func (a *AaaAuther) NewTaskAccounter(
	uid uint32, groups []string, cmd []string, pathAttrs *pathutil.PathAttrs,
) TaskAccounter {
	t, _ := a.newTask(uid, groups, cmd, pathAttrs)
	return t
}

func (a *AaaAuther) AccountCommand(uid uint32, groups []string, cmd []string, pathAttrs *pathutil.PathAttrs) {
	if t := a.NewTaskAccounter(uid, groups, cmd, pathAttrs); t != nil {
		t.AccountStop(nil)
	}
}

func (a *AaaAuther) AuthorizeCommand(uid uint32, groups []string, cmd []string, pathAttrs *pathutil.PathAttrs) bool {
	authed, err := guard.CatchPanicBoolError(func() (bool, error) {
		return a.proto.Plugin.Authorize("conf-mode", uid, groups, cmd, pathAttrs)
	})
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
