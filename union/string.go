// Copyright (c) 2018-2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2014-2015 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package union

import (
	"bytes"
)

type StringWriter struct {
	bytes.Buffer
}

func (b *StringWriter) endNode(empty bool, level int) {
	if empty {
		b.WriteByte('\n')
		return
	}
	b.writeLevelString(level)
	b.WriteString("}\n")
}
func (b *StringWriter) writeLevelString(level int) {
	for i := 0; i < level; i++ {
		b.WriteByte('\t')
	}
}
func (b *StringWriter) BeginContainer(n *Container, empty bool, level int) {
	b.writeLevelString(level)
	b.WriteString(n.Name())
	if empty {
		return
	}
	b.WriteString(" {\n")
}
func (b *StringWriter) EndContainer(n *Container, empty bool, level int) {
	b.endNode(empty, level)
}
func (b *StringWriter) BeginList(n *List, empty bool, level int) {}
func (b *StringWriter) EndList(n *List, empty bool, level int)   {}
func (b *StringWriter) BeginListEntry(n *ListEntry, empty bool, level int) {
	b.writeLevelString(level)
	b.WriteString(n.parent.Name())
	b.WriteByte(' ')
	b.WriteString(quote(n.Name()))
	if empty {
		return
	}
	b.WriteString(" {\n")
}
func (b *StringWriter) EndListEntry(n *ListEntry, empty bool, level int) {
	b.endNode(empty, level)
}

func (b *StringWriter) writeLeafValue(n Node, empty bool, level int, hideSecrets bool) {
	if empty {
		b.writeLevelString(level)
		b.WriteString(n.Name())
		b.WriteByte('\n')
		return
	}

	for _, v := range n.SortedChildren() {
		b.writeLevelString(level)
		b.WriteString(n.Name())
		b.WriteByte(' ')
		if hideSecrets && n.GetSchema().ConfigdExt().Secret {
			b.WriteString(quote("********"))
		} else {
			b.WriteString(escapeAndQuote(v.Name()))
		}
		b.WriteByte('\n')
	}
}

func (b *StringWriter) BeginLeaf(n *Leaf, empty bool, level int, hideSecrets bool) {}
func (b *StringWriter) WriteLeafValue(n *Leaf, empty bool, level int, hideSecrets bool) {
	b.writeLeafValue(n, empty, level, hideSecrets)
}
func (b *StringWriter) EndLeaf(n *Leaf, empty bool, level int) {}

func (b *StringWriter) BeginLeafList(n *LeafList, empty bool, level int, hideSecrets bool) {}
func (b *StringWriter) WriteLeafListValues(n *LeafList, empty bool, level int, hideSecrets bool) {
	b.writeLeafValue(n, empty, level, hideSecrets)
}
func (b *StringWriter) EndLeafList(n *LeafList, empty bool, level int) {}

func (b *StringWriter) PrintSep() {}
