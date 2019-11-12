// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

type TmplCompat struct {
	Node Node
	Val  bool
}

func pathDescendant(spec Node, path []string) *TmplCompat {
	if len(path) == 0 {
		switch n := spec.(type) {
		default:
			return &TmplCompat{Node: n}
		case *listEntry:
			return &TmplCompat{Node: n.list, Val: true}
		case *leafValue:
			return &TmplCompat{Node: n, Val: true}
		}
	}

	c := spec.Child(path[0])
	if c == nil {
		return nil
	}
	return pathDescendant(c.(Node), path[1:])
}

func (ms *modelSet) PathDescendant(path []string) *TmplCompat {
	return pathDescendant(ms, path)
}

func opdPathDescendant(spec Node, path []string) *TmplCompat {
	if len(path) == 0 {
		switch n := spec.(type) {
		default:
			return nil
		case *modelSet:
			return &TmplCompat{Node: n}
		case *tree:
			return &TmplCompat{Node: n}
		case *opdOption:
			_, ok := n.Type().(Empty)
			return &TmplCompat{Node: n, Val: ok}
		case *opdOptionValue:
			return &TmplCompat{Node: n.opdOption, Val: true}
		case *opdArgument:
			return &TmplCompat{Node: n, Val: true}
		case *opdCommand:
			return &TmplCompat{Node: n}
		}
	}

	c := spec.Child(path[0])
	if c == nil {
		return nil
	}
	return opdPathDescendant(c.(Node), path[1:])
}

func (ms *modelSet) OpdPathDescendant(path []string) *TmplCompat {
	return opdPathDescendant(ms, path)
}
