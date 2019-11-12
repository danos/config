// Copyright (c) 2017,2019, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"strings"

	yang "github.com/danos/yang/schema"
)

func GetHelp(n yang.Node) string {
	h := n.(hasExtensions).ConfigdExt().GetHelp()
	switch n.Status() {
	case yang.Deprecated:
		h += " [Deprecated]"
	case yang.Obsolete:
		return ""
	}

	// TODO: Add [Required] for mandatory nodes

	return h
}

const runhelp = "Execute the current command"
const runkey = "<Enter>"

func getTypePatternHelp(t yang.Type) map[string]string {
	switch v := t.(type) {
	case Boolean:
		m := make(map[string]string)
		h := v.ConfigdExt().GetTypeHelp()
		m["true"] = h
		m["false"] = h
		return m
	case Enumeration:
		return v.getHelpMap()
	case Decimal64:
		h := v.ConfigdExt().GetTypeHelp()
		strs := make([]string, 0, len(v.Rbs()))
		for _, rb := range v.Rbs() {
			strs = append(strs, rb.String())
		}
		rstr := strings.Join(strs, " | ")
		rstr = "<" + rstr + ">"
		return map[string]string{rstr: h}
	case Integer:
		h := v.ConfigdExt().GetTypeHelp()
		strs := make([]string, 0, len(v.Rbs()))
		for _, rb := range v.Rbs() {
			strs = append(strs, rb.String())
		}
		rstr := strings.Join(strs, " | ")
		rstr = "<" + rstr + ">"
		return map[string]string{rstr: h}
	case Uinteger:
		h := v.ConfigdExt().GetTypeHelp()
		strs := make([]string, 0, len(v.Rbs()))
		for _, rb := range v.Rbs() {
			strs = append(strs, rb.String())
		}
		rstr := strings.Join(strs, " | ")
		rstr = "<" + rstr + ">"
		return map[string]string{rstr: h}
	case String:
		m := make(map[string]string)
		h := v.ConfigdExt().GetTypeHelp()
		phs := v.ConfigdExt().PatternHelp
		phs = append(phs, v.ConfigdExt().OpdPatternHelp...)
		for _, ph := range phs {
			if ph == "" {
				ph = "<text>"
			}
			m[ph] = h
		}
		if len(m) == 0 && len(v.Pats()) > 0 && len(v.Pats()[0]) > 0 {
			m["<pattern>"] = h
		} else if len(phs) == 0 {
			m["<text>"] = h
		}
		return m

	case Union:
		m := make(map[string]string)
		h := v.ConfigdExt().GetTypeHelp()
		for _, t := range v.Typs() {
			for k, h2 := range getTypePatternHelp(t) {
				if h2 == "" {
					h2 = h
				}
				m[k] = h2
			}
		}
		return m

	case Leafref:
		m := make(map[string]string)
		h := v.ConfigdExt().GetTypeHelp()
		phs := v.ConfigdExt().PatternHelp
		phs = append(phs, v.ConfigdExt().OpdPatternHelp...)
		for _, ph := range phs {
			if ph == "" {
				ph = "<text>"
			}
			m[ph] = h
		}
		return m

	default:
		return nil
	}
}

func getPatternHelp(n yang.Node) map[string]string {
	switch v := n.(type) {
	case Leaf, LeafList, OpdOption:
		return getTypePatternHelp(n.Type())
	case ListEntry:
		return getPatternHelp(v.Child(v.Keys()[0]))
	case List:
		return getPatternHelp(v.Child("Dummy"))
	case OpdArgument:
		tp := getTypePatternHelp(n.Type())
		for a, b := range tp {
			if b == "" {
				tp[a] = GetHelp(n)
			}
		}
		return tp
	default:
		return nil
	}
}

func getPatternHelpMap(n yang.Node) map[string]string {
	if n == nil {
		return make(map[string]string)
	}
	h := GetHelp(n)
	m := getPatternHelp(n)
	for k, v := range m {
		if v == "" {
			m[k] = h
		}
	}
	if m == nil {
		m = make(map[string]string)
	}
	if _, ok := n.Type().(Empty); !ok && len(m) == 0 {
		m["<text>"] = h
	}
	if n.HasPresence() {
		m[runkey] = runhelp
	}
	return m
}

func getHelpMap(n yang.Node) map[string]string {
	m := make(map[string]string)
	if n == nil {
		return m
	}
	children := n.Children()
	args := n.Arguments()
	if len(children) < 1 && n.Parent() != nil {
		children = n.Parent().Children()
		args = n.Parent().Arguments()
	}
	for _, ch := range children {
		if args != nil && len(args) > 0 && args[0] == ch.Name() {
			for k, v := range getPatternHelp(ch) {
				m[k] = v
			}
		} else {
			h := GetHelp(ch)
			if h != "" {
				m[ch.Name()] = h
			}
		}
	}
	if n.HasPresence() {
		m[runkey] = runhelp
	}
	return m
}

// We have to override because a tree has presence, but we don't want
// the runkey and runhelp in the CLI
func (n *modelSet) HelpMap() map[string]string {
	m := make(map[string]string)
	if n == nil {
		return m
	}
	children := n.Children()
	for _, ch := range children {
		h := GetHelp(ch)
		if h != "" {
			m[ch.Name()] = h
		}
	}
	return m
}

func (n *tree) HelpMap() map[string]string {
	return getHelpMap(n)
}

func (n *container) HelpMap() map[string]string {
	return getHelpMap(n)
}

func (n *list) HelpMap() map[string]string {
	return getPatternHelpMap(n)
}

func (n *listEntry) HelpMap() map[string]string {
	m := make(map[string]string)
	if n == nil {
		return m
	}
	children := n.Children()
	for _, ch := range children {
		if isElemOf(n.Keys(), ch.Name()) {
			continue
		}
		h := GetHelp(ch)
		if h != "" {
			m[ch.Name()] = h
		}
	}
	if n.HasPresence() {
		m[runkey] = runhelp
	}
	return m
}

func (n *leaf) HelpMap() map[string]string {
	return getPatternHelpMap(n)
}

func (n *leafList) HelpMap() map[string]string {
	return getPatternHelpMap(n)
}

func (y *leafValue) HelpMap() map[string]string {
	return map[string]string{
		runkey: runhelp,
	}
}

func (n *opdArgument) HelpMap() map[string]string {
	return getHelpMap(n)
}

func (n *opdCommand) HelpMap() map[string]string {
	return getHelpMap(n)
}

func (n *opdOption) HelpMap() map[string]string {
	if _, ok := n.Type().(Empty); !ok {
		return getPatternHelpMap(n)
	}
	return getHelpMap(n)
}

func (n *opdOptionValue) HelpMap() map[string]string {
	return getHelpMap(n)
}
