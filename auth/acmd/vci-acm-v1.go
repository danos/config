//Copyright (c) 2018-2019, AT&T Intellectual Property.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package acmd

import (
	"github.com/danos/config/auth"
)

type AcmV1Config struct {
	Enable              bool   `rfc7951:"vyatta-system-acm-configd-v1:enable,emptyleaf" json:"enabled"`
	LogReq              bool   `rfc7951:"vyatta-system-acm-configd-v1:log-requests,emptyleaf" json:"log-requests"`
	CreateDefault       Action `rfc7951:"vyatta-system-acm-configd-v1:create-default" json:"create-default"`
	ReadDefault         Action `rfc7951:"vyatta-system-acm-configd-v1:read-default" json:"read-default"`
	UpdateDefault       Action `rfc7951:"vyatta-system-acm-configd-v1:update-default" json:"update-default"`
	DeleteDefault       Action `rfc7951:"vyatta-system-acm-configd-v1:delete-default" json:"delete-default"`
	RpcDefault          Action `rfc7951:"vyatta-system-acm-configd-v1:rpc-default" json:"rpc-default"`
	NotificationDefault Action `rfc7951:"vyatta-system-acm-configd-v1:notification-default" json:"notification-default"`
	Ruleset             struct {
		Rule []*AcmV1Rule `rfc7951:"rule"`
	} `rfc7951:"vyatta-system-acm-configd-v1:ruleset" json:"-"`
	RpcRuleset struct {
		Rule []*AcmV1RpcRule `rfc7951:"rule"`
	} `rfc7951:"vyatta-system-acm-configd-v1:rpc-ruleset" json:"-"`
	NotificationRuleset struct {
		Rule []*AcmV1NotificationRule `rfc7951:"rule"`
	} `rfc7951:"vyatta-system-acm-configd-v1:notification-ruleset" json:"-"`
}

type AcmV1Rule struct {
	RuleNumber uint32   `rfc7951:"tagnode" json:"-"`
	Path       *string  `rfc7951:"path" json:"path"`
	Action     Action   `rfc7951:"action" json:"action"`
	Groups     []string `rfc7951:"group" json:"groups"`
	Log        bool     `rfc7951:"log,emptyleaf" json:"log"`
	Operation  AcmPerm  `rfc7951:"operation" json:"perm"`
}
type AcmV1RpcRule struct {
	RuleNumber uint32   `rfc7951:"rule-number" json:"-"`
	Rpc        *string  `rfc7951:"rpc-name" json:"rpc-name"`
	Module     string   `rfc7951:"module-name"`
	Action     Action   `rfc7951:"action" json:"action"`
	Groups     []string `rfc7951:"group" json:"groups"`
}
type AcmV1NotificationRule struct {
	RuleNumber   uint32   `rfc7951:"rule-number" json:"-"`
	Notification *string  `rfc7951:"notification-name" json:"notification-name"`
	Module       string   `rfc7951:"module-name"`
	Action       Action   `rfc7951:"action" json:"action"`
	Groups       []string `rfc7951:"group" json:"groups"`
}

// Translate the ruleset and rpc-ruleset rules into an Authdb
// suitable for saving to configruleset file
// Even though RPC rules are enforced by D-Bus, we still
// require them for non VCI clients.
func (a *AcmV1Config) translateToConfigRuleSet(crs *auth.Authdb) {
	crs.Enabled = a.Enable
	crs.LogReq = a.LogReq
	crs.CreateDefault = a.CreateDefault.translateToConfigRules()
	crs.ReadDefault = a.ReadDefault.translateToConfigRules()
	crs.UpdateDefault = a.UpdateDefault.translateToConfigRules()
	crs.DeleteDefault = a.DeleteDefault.translateToConfigRules()
	crs.RpcDefault = a.RpcDefault.translateToConfigRules()
	crs.ExecDefault = auth.AUTH_ALLOW

	// Translate the data node rules
	if len(a.Ruleset.Rule) > 0 {
		rules := make([]*auth.AuthRuleType, 0)
		for _, r := range a.Ruleset.Rule {
			var authtype auth.AuthType
			rule := &auth.AuthRule{}
			rule.Path = *r.Path
			authtype = auth.AUTH_T_DATA
			rule.Perm = r.Operation.translateToConfigRules()
			rule.Action = r.Action.translateToConfigRules()
			rule.Groups = append(rule.Groups, r.Groups...)
			ruletype := &auth.AuthRuleType{Type: authtype, Rule: rule}
			rules = append(rules, ruletype)
		}
		crs.Rules = rules
	}

	// Translate the RPC rules
	if len(a.RpcRuleset.Rule) > 0 {
		rules := make([]*auth.AuthRuleType, 0)
		for _, r := range a.RpcRuleset.Rule {
			var authtype auth.AuthType
			rule := &auth.AuthRule{}
			if r.Rpc != nil {
				rule.Rpc = *r.Rpc
			}
			authtype = auth.AUTH_T_RPC
			rule.Module = r.Module
			rule.Action = r.Action.translateToConfigRules()
			rule.Groups = append(rule.Groups, r.Groups...)
			ruletype := &auth.AuthRuleType{Type: authtype, Rule: rule}
			rules = append(rules, ruletype)
		}
		crs.RpcRules = rules
	}
}
