// Copyright (c) 2017-2021, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"time"

	"github.com/danos/config/data"
	yang "github.com/danos/yang/schema"
)

type ConfigMultiplexerFn func([][]byte, ModelSet) (*data.Node, error)

// Needs to match configd: (*commitctx) LogCommitTime()
type commitTimeLogFn func(string, time.Time)

type ModelSet interface {
	yang.ModelSet
	ExtendedNode
	PathDescendant([]string) *TmplCompat
	OpdPathDescendant([]string) *TmplCompat

	// TODO - should these get moved too? Yes, along with componentExtensions
	// and perhaps create compMgr in yangd and pass back to caller in main.go?
	GetModelNameForNamespace(string) (string, bool)
	GetDefaultComponentModuleMap() map[string]struct{}
}

type modelSet struct {
	yang.ModelSet
	*extensions
	*state
	compMappings *componentMappings
}

// Compile time check that the concrete type meets the interface
var _ ModelSet = (*modelSet)(nil)

type namespaceToComponent func(string) *component

// For now there is an implicit assumption that we are only dealing with the
// single 'vyatta-v1' model set.  As and when we support multiple model sets
// we should probably pass the required model set name in to this function,
// probably provided initially by the call to start yangd that provides the
// YANG directory to be parsed, as we will have a separate YANG directory
// per modelset.
const VyattaV1ModelSet = "vyatta-v1"

func (c *CompilationExtensions) ExtendModelSet(
	m yang.ModelSet,
) (yang.ModelSet, error) {

	compMappings, err := createComponentMappings(
		m, VyattaV1ModelSet, c.ComponentConfig)
	if err != nil {
		return nil, err
	}

	ext := newExtend(nil)
	return &modelSet{
			m, ext, newState(m, ext),
			compMappings},
		nil
}

func (m *modelSet) GetModelNameForNamespace(ns string) (string, bool) {
	for svcName, svc := range m.compMappings.components {
		if _, ok := svc.modMap[ns]; ok {
			return svcName, true
		}
	}
	return "", false
}

func (m *modelSet) GetDefaultComponentModuleMap() map[string]struct{} {
	return m.compMappings.components[m.compMappings.defaultComponent].modMap
}
