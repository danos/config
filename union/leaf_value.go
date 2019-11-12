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

type LeafValue struct {
	*node
	Schema schema.LeafValue
}

func (n *LeafValue) serialize(b Serializer, path []string, lvl int, opts *unionOptions) {
}

func NewLeafValue(overlay, underlay *data.Node, sch schema.LeafValue, parent Node, flags Flags) *LeafValue {
	out := new(LeafValue)
	out.node = newNode(overlay, underlay, sch, parent, out, flags)
	out.Schema = sch
	return out
}
