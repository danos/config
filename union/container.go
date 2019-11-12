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

type Container struct {
	*node
	Schema schema.Container
}

func (n *Container) serialize(b Serializer, cpath []string, lvl int, opts *unionOptions) {
	empty := n.serializeIsEmpty(opts.includeDefaults)
	b.BeginContainer(n, empty, lvl)
	if empty {
		b.EndContainer(n, empty, lvl)
		return
	}
	n.serializeChildren(b, cpath, lvl+1, opts)
	b.EndContainer(n, empty, lvl)
}

func NewContainer(overlay, underlay *data.Node, sch schema.Container, parent Node, flags Flags) *Container {
	out := new(Container)
	out.node = newNode(overlay, underlay, sch, parent, out, flags)
	out.Schema = sch
	return out
}
