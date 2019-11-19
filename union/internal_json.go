// Copyright (c) 2018-2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2014-2015 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package union

import (
	"fmt"
)

type InternalJSONWriter struct {
	*JSONWriter
}

func (b *InternalJSONWriter) BeginList(n *List, empty bool, level int) {
	b.WriteByte('"')
	b.WriteString(n.Name())
	b.WriteString("\":{")
}
func (b *InternalJSONWriter) BeginListEntry(n *ListEntry, empty bool, level int, hideSecrets bool) {
	if redactListEntry(n, hideSecrets) {
		fmt.Fprintf(b, "\"********\":{")
	} else {
		fmt.Fprintf(b, "\"%s\":{", n.Name())
	}
}
func (b *InternalJSONWriter) EndList(n *List, empty bool, level int) {
	b.WriteByte('}')
}

func (n *node) MarshalInternalJSON(options ...UnionOption) []byte {
	var b = InternalJSONWriter{
		JSONWriter: new(JSONWriter),
	}
	b.WriteByte('{')
	n.Serialize(&b, nil, options...)
	b.WriteByte('}')
	return b.Bytes()
}
