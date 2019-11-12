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

type Leaf struct {
	*node
	Schema schema.Leaf
}

func (n *Leaf) serialize(b Serializer, path []string, lvl int, opts *unionOptions) {
	empty := n.Empty()
	hideSecrets := opts.shouldHideSecrets(path)
	if !opts.includeDefaults && n.def() {
		return
	}
	b.BeginLeaf(n, empty, lvl, hideSecrets)
	b.WriteLeafValue(n, empty, lvl, hideSecrets)
	b.EndLeaf(n, empty, lvl)
}

func (n *Leaf) setHook() {
	n.Data().ClearChildren()
	n.Data().MarkOpaque()
}

func NewLeaf(overlay, underlay *data.Node, sch schema.Leaf, parent Node, flags Flags) *Leaf {
	out := new(Leaf)
	out.node = newNode(overlay, underlay, sch, parent, out, flags)
	out.Schema = sch
	return out
}
