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

type Choice struct {
	*node
	Schema schema.Choice
}

func (n *Choice) serialize(b Serializer, path []string, lvl int, opts *unionOptions) {
}

func NewChoice(overlay, underlay *data.Node, sch schema.Choice, parent Node, flags Flags) *Choice {
	out := new(Choice)
	out.node = newNode(overlay, underlay, sch, parent, out, flags)
	out.Schema = sch
	return out
}
