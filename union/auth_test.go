// Copyright (c) 2019, AT&T Intellectual Property Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package union

import (
	"os/user"
	"strconv"

	"github.com/danos/config/auth"
)

type testAuther struct {
	Auther
	a           auth.Auther
	showSecrets bool
}

func newTestAuther(a auth.Auther, showSecrets bool) *testAuther {
	return &testAuther{a: a, showSecrets: showSecrets}
}

func (a *testAuther) authorize(perm auth.AuthPerm, path []string) bool {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		panic(err)
	}

	// Not implemented: user's groups and path attributes
	// Neither are currently required
	return a.a.AuthorizePath(uint32(uid), []string{}, path, nil, perm)
}

func (a *testAuther) AuthRead(path []string) bool {
	return a.authorize(auth.P_READ, path)
}

func (a *testAuther) AuthCreate(path []string) bool {
	return a.authorize(auth.P_CREATE, path)
}

func (a *testAuther) AuthUpdate(path []string) bool {
	return a.authorize(auth.P_UPDATE, path)
}

func (a *testAuther) AuthDelete(path []string) bool {
	return a.authorize(auth.P_DELETE, path)
}

func (a *testAuther) AuthReadSecrets(path []string) bool {
	return a.showSecrets
}
