// Copyright (c) 2018-2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package union

import (
	"github.com/danos/config/data"
	"github.com/danos/config/schema"
)

type List struct {
	*node
	Schema schema.List
}

func (n *List) serialize(b Serializer, path []string, lvl int, opts *unionOptions) {
	empty := n.Empty()
	b.BeginList(n, empty, lvl)
	n.serializeChildren(b, path, lvl, opts)
	b.EndList(n, empty, lvl)
}

func NewList(overlay, underlay *data.Node, sch schema.List, parent Node, flags Flags) *List {
	out := new(List)
	out.node = newNode(overlay, underlay, sch, parent, out, flags)
	out.Schema = sch
	return out
}
