// Copyright (c) 2021, AT&T Intellectual Property. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0
//

// The ComponentManager deals with communications between configd and the
// VCI components, and with the components' status (active / running / stopped
// etc).

package schema

import (
	"github.com/danos/yang/data/datanode"
	"github.com/danos/yang/data/encoding"
	yang "github.com/danos/yang/schema"
)

type OperationsManager interface {
	Dial() error
	SetConfigForModel(string, interface{}) error
	CheckConfigForModel(string, interface{}) error
	StoreConfigByModelInto(string, interface{}) error
	StoreStateByModelInto(string, interface{}) error
}

type ServiceManager interface {
	Close()
	IsActive(name string) (bool, error)
}

type componentMappings struct {
	components        map[string]*component
	nsMap             map[string]string
	orderedComponents []string
	defaultComponent  string
}

type component struct {
	name      string
	modMap    map[string]struct{}
	setFilter func(s yang.Node, d datanode.DataNode,
		children []datanode.DataNode) bool
	checkMap    map[string]struct{}
	checkFilter func(s yang.Node, d datanode.DataNode,
		children []datanode.DataNode) bool
}

func (c *component) FilterSetTree(n Node, dn datanode.DataNode) []byte {
	filteredCandidate := yang.FilterTree(n, dn, c.setFilter)
	return encoding.ToRFC7951(n, filteredCandidate)
}

func (c *component) FilterCheckTree(n Node, dn datanode.DataNode) []byte {
	filteredCandidate := yang.FilterTree(n, dn, c.checkFilter)
	return encoding.ToRFC7951(n, filteredCandidate)
}

func (c *component) HasConfiguration(n Node, dn datanode.DataNode) bool {
	return string(c.FilterSetTree(n, dn)) != "{}"
}

// ComponentManager encapsulates bus operations to/from components, and service
// queries against the components' service status.
type ComponentManager interface {
	OperationsManager
	ServiceManager
}

type compMgr struct {
	OperationsManager
	ServiceManager
}

var _ OperationsManager = (*compMgr)(nil)
var _ ServiceManager = (*compMgr)(nil)
var _ ComponentManager = (*compMgr)(nil)

func NewCompMgr(
	opsMgr OperationsManager,
	svcMgr ServiceManager,
) *compMgr {
	return &compMgr{
		OperationsManager: opsMgr,
		ServiceManager:    svcMgr,
	}
}
