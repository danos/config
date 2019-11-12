// Copyright (c) 2017-2019, AT&T Intellectual Propery. All rights reserved
//
// Copyright (c) 2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// ModelSetSpec - utility type to handle model set creation from schemas in
// varying formats, with or without associated VCI component configuration.

package testutils

import (
	"strconv"
	"testing"

	"github.com/danos/config/schema"
	"github.com/danos/yang/compile"
	"github.com/danos/yang/parse"
)

type ModelSetSpec struct {
	// Provided at setup
	t               *testing.T
	skipUnknown     bool
	schemas         [][]byte
	schemaDir       string
	capabilities    string
	extensions      *schema.CompilationExtensions
	featuresChecker compile.FeaturesChecker

	// Derived
	modules    map[string]*parse.Tree
	ms, msFull schema.ModelSet
	features   compile.FeaturesChecker
}

func NewModelSetSpec(t *testing.T) *ModelSetSpec {
	return &ModelSetSpec{
		t: t}
}

func (mss *ModelSetSpec) SetSchemas(schemas ...[]byte) *ModelSetSpec {
	mss.schemas = schemas
	return mss
}

func (mss *ModelSetSpec) SetSchemaDir(schemaDir string) *ModelSetSpec {
	mss.schemaDir = schemaDir
	return mss
}

func (mss *ModelSetSpec) SetSkipUnknown() *ModelSetSpec {
	mss.skipUnknown = true
	return mss
}

func (mss *ModelSetSpec) SetExtensions(
	ext *schema.CompilationExtensions,
) *ModelSetSpec {
	mss.extensions = ext
	return mss
}

func (mss *ModelSetSpec) SetCapabilities(
	capabilities string,
) *ModelSetSpec {
	mss.capabilities = capabilities
	return mss
}

func (mss *ModelSetSpec) SetFeaturesChecker(
	checker compile.FeaturesChecker,
) *ModelSetSpec {
	mss.featuresChecker = checker
	return mss
}

func (mss *ModelSetSpec) generateFeaturesChecker() error {
	if mss.capabilities == "" {
		mss.features = mss.featuresChecker
		return nil
	}

	fc, err := CreateFeaturesChecker(mss.capabilities)

	if err != nil {
		return err
	}
	mss.features = compile.MultiFeatureCheckers(fc, mss.featuresChecker)

	return nil
}

func (mss *ModelSetSpec) generateModelSetsFromSchemaDir() (
	schema.ModelSet, schema.ModelSet, error) {

	var err error
	mss.ms, err = schema.CompileDir(
		&compile.Config{
			YangLocations: compile.YangDirs(mss.schemaDir),
			Features:      mss.features},
		nil)

	if err != nil {
		return nil, nil, err
	}

	mss.msFull, err = schema.CompileDir(
		&compile.Config{
			YangLocations: compile.YangDirs(mss.schemaDir),
			Features:      mss.features,
			Filter:        compile.IsConfigOrState()},
		mss.extensions)
	if err != nil {
		return nil, nil, err
	}

	return mss.ms, mss.msFull, nil
}

func (mss *ModelSetSpec) parseModules() error {
	const name = "schema"
	modules := make(map[string]*parse.Tree)
	for index, b := range mss.schemas {
		tree, err := schema.Parse(name+strconv.Itoa(index), string(b))
		if err != nil {
			return err
		}
		mod := tree.Root.Argument().String()
		modules[mod] = tree
	}
	mss.modules = modules
	return nil
}

func (mss *ModelSetSpec) GenerateModelSets() (
	schema.ModelSet, schema.ModelSet, error) {

	var err error

	if err = mss.generateFeaturesChecker(); err != nil {
		return nil, nil, err
	}

	if mss.schemaDir != "" {
		return mss.generateModelSetsFromSchemaDir()
	}

	if err = mss.parseModules(); err != nil {
		return nil, nil, err
	}
	mss.ms, err = schema.CompileParseTrees(mss.modules, mss.features, mss.skipUnknown,
		compile.Include(compile.IsConfig, compile.IncludeState(false)),
		mss.extensions)
	if err != nil {
		return nil, nil, err
	}

	// MUST reparse modules as CompileModules modifies the data.
	// TODO - rework this so parsing is done *inside* CompileModules to
	//        completely remove the temptation to make this mistake.
	if err := mss.parseModules(); err != nil {
		return nil, nil, err
	}
	mss.msFull, err = schema.CompileParseTrees(mss.modules, mss.features, mss.skipUnknown,
		compile.Include(compile.IsConfig, compile.IncludeState(true)),
		mss.extensions)
	if err != nil {
		return nil, nil, err
	}

	return mss.ms, mss.msFull, nil
}

func (mss *ModelSetSpec) GetCfgOnlyModelSet() schema.ModelSet {
	return mss.ms
}

func (mss *ModelSetSpec) GetFullModelSet() schema.ModelSet {
	return mss.msFull
}
