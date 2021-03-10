// Copyright (c) 2019, 2021, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/danos/yang/parse"
	yang "github.com/danos/yang/schema"
)

type Model interface {
	yang.Model
	Tree() Tree
}

type model struct {
	yang.Model
	tree Tree
}

// Compile time check that the concrete type meets the interface
var _ Model = (*model)(nil)

func (*CompilationExtensions) ExtendModel(
	p parse.Node, m yang.Model, t yang.Tree,
) (yang.Model, error) {

	return &model{m, t.(Tree)}, nil
}

func (m *model) Tree() Tree {
	return m.tree
}
