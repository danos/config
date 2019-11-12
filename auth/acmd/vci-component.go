//Copyright (c) 2018-2019, AT&T Intellectual Property.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package acmd

import (
	"encoding/json"
	"io/ioutil"
	"sync/atomic"

	"github.com/danos/config/auth"
	"github.com/danos/encoding/rfc7951"
)

const (
	acmd_config_file = "/opt/vyatta/etc/acmd.conf"
	policy_file      = "/usr/share/dbus-1/vci-local.conf"
)

type AcmdConfig struct {
	System struct {
		Acm struct {
			*AcmV1Config
		} `rfc7951:"vyatta-system-acm-v1:acm,omitempty"`
	} `rfc7951:"vyatta-system-v1:system"`
}

type Config struct {
	state *State
	data  atomic.Value
}

func NewConfig(state *State) *Config {
	config := &Config{
		state: state,
	}
	config.data.Store(&AcmdConfig{})
	config.loadComponentConfig()
	return config
}

func (c *Config) Get() *AcmdConfig {
	return c.data.Load().(*AcmdConfig)
}

func (c *Config) Check(cfg *AcmdConfig) error {
	return nil
}

func (c *Config) Set(cfg *AcmdConfig) error {
	err := c.Check(cfg)
	if err != nil {
		return err
	}
	c.data.Store(cfg)
	err = c.saveComponentConfig(cfg)
	if err != nil {
		return err
	}

	if c.state != nil {
		c.state.set(cfg)
	}

	err = c.saveToConfigRuleset(cfg)
	if err != nil {
		return err
	}

	return applyPolicyRules(cfg)
}

func applyPolicyRules(cfg *AcmdConfig) error {
	pol, err := cfg.System.Acm.AcmV1Config.translateToPolicy()
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(policy_file, pol, 0644); err != nil {
		return err
	}
	return reloadPolicyRules()
}

func (c *Config) saveToConfigRuleset(cfg *AcmdConfig) error {
	adb := &auth.Authdb{}
	cfg.System.Acm.AcmV1Config.translateToConfigRuleSet(adb)
	d, err := json.Marshal(adb)

	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(auth.Authrulefile, d, 0644); err != nil {
		return err
	}

	return nil
}

func (c *Config) saveComponentConfig(cfg *AcmdConfig) error {
	d, err := rfc7951.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(acmd_config_file, d, 0644); err != nil {
		return err
	}
	return nil
}

func (c *Config) loadComponentConfig() error {
	b, err := ioutil.ReadFile(acmd_config_file)
	if err != nil {
		return err
	}

	var cfg AcmdConfig
	err = rfc7951.Unmarshal(b, &cfg)
	if err != nil {
		return err
	}

	c.data.Store(&cfg)

	return nil
}

type State struct {
	settings atomic.Value
}

func NewState() *State {
	state := &State{}
	state.settings.Store(&AcmdConfig{})
	return state
}

func (s *State) Get() *AcmdConfig {
	return s.settings.Load().(*AcmdConfig)
}

func (s *State) set(v *AcmdConfig) {

	s.settings.Store(v)
}
