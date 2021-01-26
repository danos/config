// Copyright (c) 2019-2021, AT&T Intellectual Property.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package platform

import (
	"fmt"
	"github.com/go-ini/ini"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultBaseDir      = "/lib/vyatta-platform"
	PlatformDefinitions = "platforms"
	PlatformDefSuffix   = ".platform"
	Yang                = "yang"
	DefaultActiveDir    = "/run/vyatta-platform"
	ActiveYang          = "yang"
	Features            = "features"
	DisabledFeatures    = "features-disabled"
)

type Platforms struct {
	platformDir string
	activeDir   string
	Platforms   map[string]*Definition
}

func NewPlatform() *Platforms {
	return &Platforms{platformDir: DefaultBaseDir,
		activeDir: DefaultActiveDir,
		Platforms: make(map[string]*Definition)}
}

func (p *Platforms) PlatformBaseDir(dir string) *Platforms {
	p.platformDir = dir

	return p
}

func (p *Platforms) ActiveDir(dir string) *Platforms {

	p.activeDir = dir
	return p
}

func (p *Platforms) LoadDefinitions() *Platforms {

	files, _ := filepath.Glob(filepath.Join(p.platformDir, PlatformDefinitions, "*"+PlatformDefSuffix))

	const platformSectionPrefix = `Platform `
	const baseSectionPrefix = `Base `

	for _, f := range files {
		iniFiles, err := ini.Load(f)
		if err != nil {
			// Ignore troublesome files
			fmt.Printf("Error processing platform file %s: %s\n", f, err)
			continue
		}
		bases := make(map[string]*Definition)
		for _, section := range iniFiles.Sections() {
			name := section.Name()
			switch {
			case strings.HasPrefix(name, baseSectionPrefix):
				base := strings.Trim(name[len(baseSectionPrefix):], " ")
				processSection(base, section, bases)
			case strings.HasPrefix(name, platformSectionPrefix):
				platform := strings.Trim(name[len(platformSectionPrefix):], " ")
				processSection(platform, section, p.Platforms)
				config, ok := p.Platforms[platform]
				if !ok {
					config = newDefinition()
				}
				b := section.Key("Base").Strings(",")
				for _, bk := range b {
					ref, ok := bases[bk]
					if !ok {
						continue
					}
					config.Yang = appendUnique(config.Yang, ref.Yang, validateFile)
					config.Features = appendUnique(config.Features, ref.Features, validateFeature)
					config.DisabledFeatures = appendUnique(config.DisabledFeatures, ref.DisabledFeatures, validateFeature)
				}
				p.Platforms[platform] = config
			}
		}
	}
	return p
}

func (p *Platforms) CreatePlatform(platform string) (*Platforms, error) {
	if _, ok := p.Platforms[platform]; ok {
		if err := checkActions(p.createFeaturesDir(platform),
			p.createDisabledFeaturesDir(platform),
			p.createActiveYang(platform)); err != nil {
			return p, err
		}
	}

	return p, nil
}

func (p *Platforms) createActiveYang(platform string) actionChecker {
	actv := filepath.Join(p.activeDir, ActiveYang)
	src := filepath.Join(p.platformDir, Yang)

	return func() error {
		if err := os.MkdirAll(actv, os.ModePerm); err != nil {
			return err
		}

		for _, y := range p.Platforms[platform].Yang {
			err := copyFile(filepath.Join(src, y),
				filepath.Join(actv, y))
			if err != nil {
				return err
			}

		}
		return nil
	}
}

func (p *Platforms) createFeaturesDir(platform string) actionChecker {
	featsDir := filepath.Join(p.activeDir, Features)
	return createFeatures(featsDir, p.Platforms[platform].Features)
}

func (p *Platforms) createDisabledFeaturesDir(platform string) actionChecker {
	disabledFeatsDir := filepath.Join(p.activeDir, DisabledFeatures)
	return createFeatures(disabledFeatsDir, p.Platforms[platform].DisabledFeatures)

}

func createFeatures(dir string, features []string) actionChecker {
	return func() error {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
		for _, f := range features {
			parts := strings.Split(f, ":")
			if len(parts) != 2 {
				continue
			}
			modDir := filepath.Join(dir, parts[0])
			if err := os.MkdirAll(modDir, os.ModePerm); err != nil {
				return err
			}
			if _, err := os.Create(filepath.Join(modDir, parts[1])); err != nil {
				return err
			}
		}

		return nil
	}
}

type Definition struct {
	Yang             []string
	Features         []string
	DisabledFeatures []string
}

func newDefinition() *Definition {
	return &Definition{
		Yang:             make([]string, 0),
		Features:         make([]string, 0),
		DisabledFeatures: make([]string, 0)}
}

// Process the Yang, Features and DisbaledFeatures keys of a section
func processSection(platformName string, section *ini.Section, cfg map[string]*Definition) {
	config, ok := cfg[platformName]
	if !ok {
		config = newDefinition()
	}
	yang := section.Key("Yang").Strings(",")
	config.Yang = appendUnique(config.Yang, yang, validateFile)
	feats := section.Key("Features").Strings(",")
	config.Features = appendUnique(config.Features, feats, validateFeature)
	disabledfeats := section.Key("DisabledFeatures").Strings(",")
	config.DisabledFeatures = appendUnique(config.DisabledFeatures, disabledfeats, validateFeature)

	cfg[platformName] = config

}

type validator func(s string) bool

func validateFeature(s string) bool {

	rslt := true
	if strings.Count(s, ":") != 1 {
		rslt = false
	}

	if strings.ContainsAny(s, "\n ") {
		rslt = false
	}

	if !rslt {
		fmt.Printf("Ignoring invalid Yang feature: '%s'\n", s)
	}
	return rslt
}

func validateFile(s string) bool {

	if strings.ContainsAny(s, "\n ") {
		fmt.Printf("Ignoring invalid Yang file: '%s'\n", s)
		return false
	}
	return true
}

func appendUnique(orig, new []string, validator validator) []string {
	exists := func(sl []string, s string) bool {
		for _, aa := range sl {
			if aa == s {
				return true
			}
		}
		return false
	}
	for _, a := range new {
		if !validator(a) {
			continue
		}
		if !exists(orig, a) {
			orig = append(orig, a)
		}
	}
	return orig
}

func copyFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)

	if err != nil {
		return err
	}
	err = d.Sync()
	if err != nil {
		return err
	}
	return nil
}

type actionChecker func() error

func emptyDir(dir string) actionChecker {
	return func() error {
		return os.RemoveAll(dir)
	}
}

func checkActions(actions ...actionChecker) error {
	for _, a := range actions {
		if err := a(); err != nil {
			return err
		}
	}
	return nil
}
