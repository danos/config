// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"strings"

	"github.com/danos/yang/parse"
	yang "github.com/danos/yang/schema"
)

type Model interface {
	yang.Model
	Tree() Tree
	ServiceBus() string
}

type model struct {
	yang.Model
	tree       Tree
	serviceBus string
}

// Compile time check that the concrete type meets the interface
var _ Model = (*model)(nil)

func buildPrefixes(p parse.Node) map[string]string {
	prefixMap := make(map[string]string)

	for _, ch := range p.ChildrenByType(parse.NodeImport) {
		prefixMap[ch.ChildByType(parse.NodePrefix).Name()] = ch.Name()
	}
	return prefixMap
}

const yangd_module = `brocade-service-api-v1`
const serv_bus_ext = `service-bus`

func buildServiceBus(p parse.Node) string {
	prefixMap := buildPrefixes(p)
	for _, ch := range p.ChildrenByType(parse.NodeUnknown) {
		nameParts := strings.Split(ch.Statement(), ":")
		mod := prefixMap[nameParts[0]]
		ext := nameParts[1]
		if mod == yangd_module && ext == serv_bus_ext {
			return ch.Name()
		}
	}

	return ""
}

func (*CompilationExtensions) ExtendModel(
	p parse.Node, m yang.Model, t yang.Tree,
) (yang.Model, error) {

	serviceBus := buildServiceBus(p)
	return &model{m, t.(Tree), serviceBus}, nil
}

func (n *model) ServiceBus() string {
	return n.serviceBus
}

func (m *model) Tree() Tree {
	return m.tree
}
