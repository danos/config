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

type LeafList struct {
	*node
	Schema schema.LeafList
}

func (n *LeafList) serialize(b Serializer, path []string, lvl int, opts *unionOptions) {
	empty := n.Empty()
	hideSecrets := opts.shouldHideSecrets(path)
	if !opts.includeDefaults && n.def() {
		return
	}
	b.BeginLeafList(n, empty, lvl, hideSecrets)
	b.WriteLeafListValues(n, empty, lvl, hideSecrets)
	b.EndLeafList(n, empty, lvl)
}

func NewLeafList(overlay, underlay *data.Node, sch schema.LeafList, parent Node, flags Flags) *LeafList {
	out := new(LeafList)
	out.node = newNode(overlay, underlay, sch, parent, out, flags)
	out.Schema = sch
	return out
}
