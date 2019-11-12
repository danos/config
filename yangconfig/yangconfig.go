// Copyright (c) 2019, AT&T Intellectual Property.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package yangconfig

import (
	"encoding/json"
	"github.com/danos/yang/compile"
	"io"
	"os"
	"path/filepath"
)

const (
	DefaultConfigFile = "/etc/vyatta/yang.conf"
)

type Feature struct {
	Location string `json:"location"`
	Enabled  bool   `json:"enabled"`
}

type Config struct {
	Yang     []string  `json:"yang"`
	Features []Feature `json:"features"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Load(r io.Reader) *Config {
	if r == nil {
		return c
	}
	reader := json.NewDecoder(r)
	cfg := &Config{}
	reader.Decode(&cfg)
	return c.merge(cfg)
}

func (c *Config) SystemConfig() *Config {
	return c.LoadFile(DefaultConfigFile)
}
func (c *Config) LoadFile(filename string) *Config {
	if filename == "" {
		return c
	}

	f, err := os.Open(filename)
	if err != nil {
		return c
	}
	defer f.Close()
	return c.Load(f)
}

func (c *Config) IncludeYangDirs(y ...string) *Config {
	c.Yang = appendUnique(c.Yang, y...)
	return c
}
func (c *Config) IncludeFeatures(f ...string) *Config {
	for _, feat := range f {
		c.Features = append(c.Features,
			Feature{Location: feat,
				Enabled: true})
	}
	return c
}

func (c *Config) IncludeDisabledFeatures(f ...string) *Config {
	for _, feat := range f {
		c.Features = append(c.Features,
			Feature{Location: feat,
				Enabled: false})
	}
	return c
}

func (c *Config) Save(w io.Writer) error {
	if w == nil {
		return nil
	}

	writer := json.NewEncoder(w)
	writer.SetIndent("", "  ")
	err := writer.Encode(c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) merge(cfg *Config) *Config {
	c.Yang = appendUnique(c.Yang, cfg.Yang...)
	for _, feat := range cfg.Features {
		c.Features = append(c.Features, feat)
	}
	return c
}

func appendUnique(d []string, n ...string) []string {
	inslice := func(v string, sl []string) bool {
		for _, s := range sl {
			if s == v {
				return true
			}
		}
		return false
	}

	for _, ns := range n {
		dir := filepath.Clean(ns)
		if !inslice(dir, d) {
			d = append(d, dir)
		}
	}
	return d
}

func (c *Config) YangLocator() compile.YangLocator {
	return compile.YangDirs(c.Yang...)
}

func (c *Config) FeaturesChecker() compile.FeaturesChecker {
	features := make([]compile.FeaturesChecker, 0)

	for _, f := range c.Features {
		feat := compile.FeaturesFromLocations(f.Enabled, f.Location)
		features = append(features, feat)
	}

	return compile.MultiFeatureCheckers(features...)
}
