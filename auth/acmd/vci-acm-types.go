//Copyright (c) 2018-2019, AT&T Intellectual Property.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package acmd

import (
	"encoding/json"
	"fmt"
	"github.com/danos/config/auth"
	"github.com/danos/encoding/rfc7951"
)

type Action string

const (
	action_deny   = "deny"
	action_permit = "permit"
	action_allow  = "allow"
	action_log    = "log"
)

func NewAction(value string) (*Action, error) {
	a := new(Action)
	err := a.set(value)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (a *Action) Validate() error {
	if a == nil {
		return nil
	}
	value := a.String()
	values := map[string]struct{}{
		action_deny:  struct{}{},
		action_allow: struct{}{},
	}
	if _, ok := values[value]; ok {
		return nil
	}

	return fmt.Errorf("Invalid action: %s", value)
}

func (a *Action) MarshalJSON() ([]byte, error) {
	values := map[string]auth.AuthAction{
		action_deny:  auth.AUTH_DENY,
		action_allow: auth.AUTH_ALLOW,
	}
	r, ok := values[a.String()]
	if !ok {
		r = auth.AUTH_DENY
	}
	return json.Marshal(r)
}

func (a *Action) UnmarshalJSON(value []byte) error {
	var v string
	if err := rfc7951.Unmarshal(value, &v); err != nil {
		return err
	}
	return a.set(v)
}

func (p *Action) translateToConfigRules() auth.AuthAction {
	values := map[string]auth.AuthAction{
		action_deny:  auth.AUTH_DENY,
		action_allow: auth.AUTH_ALLOW,
	}
	if r, ok := values[p.String()]; ok {
		return r
	}
	return auth.AUTH_DENY

}

func (a *Action) set(value string) error {
	*a = Action(value)
	return a.Validate()
}

func (a *Action) String() string {
	if a != nil {
		return string(*a)
	}
	return "<nil>"
}

type AcmPerm string

const (
	perm_create = "create"
	perm_read   = "read"
	perm_update = "update"
	perm_delete = "delete"
	perm_all    = "*"
)

func NewAcmPerm(value string) (*AcmPerm, error) {
	p := new(AcmPerm)
	err := p.set(value)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *AcmPerm) MarshalJSON() ([]byte, error) {
	values := map[string]auth.AuthPerm{
		perm_create: auth.P_CREATE,
		perm_read:   auth.P_READ,
		perm_update: auth.P_UPDATE,
		perm_delete: auth.P_DELETE,
		perm_all:    auth.P_CREATE | auth.P_READ | auth.P_UPDATE | auth.P_DELETE | auth.P_EXECUTE,
	}
	r, ok := values[p.String()]
	if !ok {
		r = 0
	}
	return json.Marshal(r)

}

func (p *AcmPerm) UnmarshalJSON(value []byte) error {
	var v string
	if err := rfc7951.Unmarshal(value, &v); err != nil {
		return err
	}
	return p.set(v)
}

func (p *AcmPerm) translateToConfigRules() auth.AuthPerm {
	values := map[string]auth.AuthPerm{
		perm_create: auth.P_CREATE,
		perm_read:   auth.P_READ,
		perm_update: auth.P_UPDATE,
		perm_delete: auth.P_DELETE,
		perm_all:    auth.P_CREATE | auth.P_READ | auth.P_UPDATE | auth.P_DELETE | auth.P_EXECUTE,
	}
	if r, ok := values[p.String()]; ok {
		return r
	}
	return auth.AUTH_DENY

}

func (p *AcmPerm) Validate() error {
	if p == nil {
		return nil
	}
	value := p.String()
	values := map[string]struct{}{
		perm_create: struct{}{},
		perm_read:   struct{}{},
		perm_update: struct{}{},
		perm_delete: struct{}{},
		perm_all:    struct{}{},
	}
	if _, ok := values[value]; ok {
		return nil
	}

	return fmt.Errorf("Invalid perm: %s", value)
}

func (p *AcmPerm) set(value string) error {
	*p = AcmPerm(value)
	return p.Validate()
}

func (p *AcmPerm) String() string {
	if p != nil {
		return string(*p)
	}
	return "<nil>"
}
