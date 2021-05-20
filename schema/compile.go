// Copyright (c) 2017-2021, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/danos/yang/compile"
	"github.com/danos/yang/parse"
	"github.com/danos/yang/xpath"
	"github.com/danos/yang/xpath/xutils"
)

// Previously contained information passed through YANG compiler. Left in for
// now in case of future use given full removal requires changes across 3
// repositories.
type CompilationExtensions struct {
}

func (e *CompilationExtensions) NodeCardinality(
	ntype parse.NodeType,
) map[parse.NodeType]parse.Cardinality {

	return configdCardinality(ntype)
}

func CompileDir(cfg *compile.Config, ext *CompilationExtensions) (ModelSet, error) {
	if ext == nil {
		ext = &CompilationExtensions{}
	}
	ms, err := compile.CompileDir(ext, cfg)
	if err != nil {
		return nil, err
	}
	return ms.(ModelSet), nil
}

func CompileDirWithWarnings(
	cfg *compile.Config,
	ext *CompilationExtensions,
) (ModelSet, []xutils.Warning, error,
) {
	if ext == nil {
		ext = &CompilationExtensions{}
	}
	ms, warns, err := compile.CompileDirWithWarnings(ext, cfg)
	if err != nil {
		return nil, warns, err
	}
	return ms.(ModelSet), warns, nil
}

// If you need to be able to look up prefix / module mappings after the
// schema has been compiled, then this function provides the module and
// submodule data to do this.
func CompileDirKeepMods(
	cfg *compile.Config,
	ext *CompilationExtensions,
) (ModelSet, error, map[string]*parse.Module, map[string]*parse.Module) {

	if ext == nil {
		ext = &CompilationExtensions{}
	}
	ms, err, mods, submods := compile.CompileDirKeepMods(ext, cfg)
	if err != nil {
		return nil, err, nil, nil
	}
	return ms.(ModelSet), nil, mods, submods
}

func ParseModules(list ...string) (map[string]*parse.Tree, error) {

	return compile.ParseModules(configdCardinality, list...)
}

func ParseModuleDir(dir string) (map[string]*parse.Tree, error) {
	return compile.ParseModuleDir(dir, configdCardinality)
}

func CompileModules(
	mods map[string]*parse.Tree,
	capabilities string,
	skipUnknown bool,
	filter compile.SchemaFilter,
	ext *CompilationExtensions,
) (ModelSet, error) {
	ms, _, err := CompileModulesWithWarnings(
		mods, capabilities, skipUnknown, filter, ext)
	return ms, err
}

func CompileModulesWithWarnings(
	mods map[string]*parse.Tree,
	capabilities string,
	skipUnknown bool,
	filter compile.SchemaFilter,
	ext *CompilationExtensions,
) (ModelSet, []xutils.Warning, error) {

	if ext == nil {
		ext = &CompilationExtensions{}
	}

	ms, warns, err := compile.CompileModulesWithWarnings(
		ext,
		mods,
		capabilities,
		skipUnknown,
		compile.SchemaFilter(filter),
	)
	if err != nil {
		return nil, warns, err
	}
	return ms.(ModelSet), warns, nil
}

func CompileModulesWithWarningsAndCustomFunctions(
	mods map[string]*parse.Tree,
	capabilities string,
	skipUnknown bool,
	filter compile.SchemaFilter,
	ext *CompilationExtensions,
	userFnChecker xpath.UserCustomFunctionCheckerFn,
) (ModelSet, []xutils.Warning, error) {

	if ext == nil {
		ext = &CompilationExtensions{}
	}

	ms, warns, err := compile.CompileModulesWithWarningsAndCustomFunctions(
		ext,
		mods,
		capabilities,
		skipUnknown,
		compile.SchemaFilter(filter),
		userFnChecker,
	)
	if err != nil {
		return nil, warns, err
	}
	return ms.(ModelSet), warns, nil
}

func CompileParseTrees(
	mods map[string]*parse.Tree,
	features compile.FeaturesChecker,
	skipUnknown bool,
	filter compile.SchemaFilter,
	ext *CompilationExtensions,
) (ModelSet, error) {

	if ext == nil {
		ext = &CompilationExtensions{}
	}

	ms, err := compile.CompileParseTrees(
		ext,
		mods,
		features,
		skipUnknown,
		compile.SchemaFilter(filter),
	)
	if err != nil {
		return nil, err
	}
	return ms.(ModelSet), nil
}

func ParseYang(locations compile.YangLocator) (map[string]*parse.Tree, error) {
	return compile.ParseYang(configdCardinality, locations)
}
