// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package union

import (
	"errors"
)

var autherr = errors.New("access denied")

type Auther interface {
	AuthRead([]string) bool
	AuthCreate([]string) bool
	AuthUpdate([]string) bool
	AuthDelete([]string) bool
	AuthReadSecrets([]string) bool
}

func authorize(auth Auther, path []string, action string) bool {
	//If auther is nil then we want to allow everything
	if auth == nil {
		return true
	}
	switch action {
	case "read":
		return auth.AuthRead(path)
	case "create":
		return auth.AuthCreate(path)
	case "update":
		return auth.AuthUpdate(path)
	case "delete":
		return auth.AuthDelete(path)
	case "secrets":
		return auth.AuthReadSecrets(path)
	default:
		return false
	}
}
