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

type ListEntry struct {
	*node
	Schema schema.ListEntry
}

func (n *ListEntry) serialize(b Serializer, path []string, lvl int, opts *unionOptions) {
	empty := n.serializeIsEmpty(opts.includeDefaults)
	b.BeginListEntry(n, empty, lvl)
	if empty {
		b.EndListEntry(n, empty, lvl)
		return
	}
	n.serializeChildrenSkip(b, path, lvl+1, opts, n.Schema.Keys())
	b.EndListEntry(n, empty, lvl)
}

func NewListEntry(overlay, underlay *data.Node, sch schema.ListEntry, parent Node, flags Flags) *ListEntry {
	out := new(ListEntry)
	out.node = newNode(overlay, underlay, sch, parent, out, flags)
	out.Schema = sch
	return out
}
