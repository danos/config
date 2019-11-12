// Copyright (c) 2017-2019, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2015 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package union

import (
	"errors"
	"strings"

	"github.com/danos/utils/pathutil"
)

type Serializer interface {
	BeginContainer(n *Container, empty bool, level int)
	EndContainer(n *Container, empty bool, level int)

	BeginList(n *List, empty bool, level int)
	EndList(n *List, empty bool, level int)
	BeginListEntry(n *ListEntry, empty bool, level int)
	EndListEntry(n *ListEntry, empty bool, level int)

	BeginLeaf(n *Leaf, empty bool, level int, hideSecrets bool)
	WriteLeafValue(n *Leaf, empty bool, level int, hideSecrets bool)
	EndLeaf(n *Leaf, empty bool, level int)

	BeginLeafList(n *LeafList, empty bool, level int, hideSecrets bool)
	WriteLeafListValues(n *LeafList, empty bool, level int, hideSecrets bool)
	EndLeafList(n *LeafList, empty bool, level int)

	PrintSep()
}

func quote(in string) string {
	if strings.ContainsAny(in, "*}{;\011\012\013\014\015 ") {
		return "\"" + in + "\""
	}
	return in
}

// escapeAndQuote Escapes double quotes, and encloses strings with quotes
// Similar to quote() but in addition we first escape any unescaped double
// quotes and then we add enclosing double quotes for all the instances in
// quote() and additionally for strings containing any double quotes.  This
// is needed for when we serialise user data (eg leaf values) or show / load
// will fail.
func escapeAndQuote(in string) string {
	in = escapeUnescapedDoubleQuotes(in)
	if strings.ContainsAny(in, "*}{;\011\012\013\014\015 \"") {
		in = "\"" + in + "\""
	}

	return in
}

// escapeUnescapedDoubleQuotes Fixup unescaped double quotes
func escapeUnescapedDoubleQuotes(input string) string {
	// Replace " with \" BUT NOT if already escaped.
	//    " -> \"
	//   \" -> \" (unchanged)
	//  \\" -> \\\"
	out := ""
	prev_elem_was_backslash := false
	for _, elem := range input {
		if need_to_escape_quote(elem, prev_elem_was_backslash) {
			out = out + "\\"
		}
		out = out + string(elem)
		if elem == '\\' {
			prev_elem_was_backslash = !prev_elem_was_backslash
		} else {
			prev_elem_was_backslash = false
		}
	}
	return out
}

func need_to_escape_quote(elem rune, prev_elem_was_backslash bool) bool {
	if elem != '"' {
		return false
	}
	return !prev_elem_was_backslash
}

func isElemOf(list []string, elem string) bool {
	for _, v := range list {
		if v == elem {
			return true
		}
	}
	return false
}

func (n *node) serializeIsEmpty(defaults bool) bool {
	if defaults {
		return n.Empty()
	}
	return n.emptyNonDefault()
}

func (n *node) serializeChildren(b Serializer, cpath []string, lvl int, opts *unionOptions) {
	n.serializeChildrenSkip(b, cpath, lvl, opts, nil)
}

func (n *node) serializeChildrenSkip(b Serializer, cpath []string, lvl int, opts *unionOptions, skipList []string) {
	children := make([]Node, 0)
	for _, ch := range n.SortedChildren() {
		if isElemOf(skipList, ch.Name()) {
			continue
		}
		if !opts.includeDefaults && ch.def() {
			continue
		}
		children = append(children, ch)
	}

	first := true

	for _, ch := range children {
		npath := pathutil.CopyAppend(cpath, ch.Name())
		if !authorize(opts.auth, npath, "read") {
			continue
		}
		if !first {
			b.PrintSep()
		} else {
			first = false
		}
		ch.serialize(b, npath, lvl, opts)
	}
}

func (n *node) Marshal(rootName, encoding string, options ...UnionOption) (string, error) {
	var outb []byte
	switch encoding {
	case "json":
		outb = n.ToJSON(options...)
	case "rfc7951":
		outb = n.ToRFC7951(options...)
	case "internal":
		outb = n.MarshalInternalJSON(options...)
	case "netconf":
		outb = n.ToNETCONF(rootName, options...)
	case "xml":
		outb = n.ToXML(rootName, options...)
	default:
		return "", errors.New("Invalid encoding requested")
	}

	return string(outb), nil
}
