// Copyright (c) 2018-2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2014-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

/* Package data is the base storage for the union data structure.
 * It provides the basic nary tree structure. */
package data

import (
	"sync/atomic"

	"github.com/danos/utils/natsort"
)

const (
	flagDeleted uint32 = 1 << iota
	flagOpaque
	flagDefault
)

const (
	ClearChildFlags     = true
	DontClearChildFlags = false
)

type ByUser []*Node

func (b ByUser) Len() int           { return len(b) }
func (b ByUser) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByUser) Less(i, j int) bool { return b[i].Index() < b[j].Index() }

type BySystem []*Node

func (b BySystem) Len() int           { return len(b) }
func (b BySystem) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b BySystem) Less(i, j int) bool { return natsort.Less(b[i].Name(), b[j].Name()) }

type AtomicNode struct {
	atomic.Value
}

func (t *AtomicNode) Load() *Node {
	return t.Value.Load().(*Node)
}
func (t *AtomicNode) Store(n *Node) {
	t.Value.Store(n)
}

func NewAtomicNode(n *Node) *AtomicNode {
	a := &AtomicNode{}
	if n == nil {
		a.Store(New("root"))
	} else {
		a.Store(n)
	}
	return a
}

type Node struct {
	name    string
	comment string
	flags   uint32
	//32bit hole
	children    map[string]*Node
	idx         uint64
	nxtChildIdx uint64
}

func New(name string) *Node {
	return &Node{
		name:     name,
		children: make(map[string]*Node),
	}
}

func (n *Node) Copy() *Node {
	return &Node{
		name:     n.name,
		comment:  n.comment,
		children: make(map[string]*Node),
	}
}

func (n *Node) Child(name string) *Node {
	if n == nil {
		return nil
	}
	return n.children[name]
}

func (n *Node) AddChild(child *Node) {
	if child == nil {
		return
	}
	child.SetIndex(n.nxtChildIdx)
	/* 64bit counter this is 34 million years at
	 * current rpc rates to overflow, so I don't care about overflow
	 * This is lifetime of the session only. */
	n.nxtChildIdx++
	n.children[child.Name()] = child
}

func (n *Node) DeleteChild(name string) {
	delete(n.children, name)
}

func (n *Node) ClearChildren() {
	n.children = make(map[string]*Node)
}

func (n *Node) ChildNames() []string {
	if n == nil {
		return nil
	}
	children := make([]string, 0, len(n.children))
	for _, v := range n.children {
		children = append(children, v.Name())
	}
	return children
}

func (n *Node) Children() []*Node {
	if n == nil {
		return nil
	}
	/* Return a list of nodes in iteration order of the map;
	 * this means there is no guaranteed ordering. The upper
	 * layer is expected to do sorting based on schema information. */
	children := make([]*Node, 0, len(n.children))
	for _, v := range n.children {
		children = append(children, v)
	}
	return children
}

func (n *Node) NumChildren() int {
	if n == nil {
		return 0
	}
	return len(n.children)
}

func (n *Node) ChildMap() map[string]*Node {
	if n == nil {
		return make(map[string]*Node)
	}
	return n.children
}

func (n *Node) Name() string {
	return n.name
}

func (n *Node) Index() uint64 {
	return n.idx
}

func (n *Node) SetIndex(idx uint64) {
	n.idx = idx
}

func (n *Node) Comment() string {
	return n.comment
}

func (n *Node) SetComment(comment string) {
	n.comment = comment
}

func (n *Node) Deleted() bool {
	return n.flags&flagDeleted == flagDeleted
}

func (n *Node) Opaque() bool {
	return n.flags&flagOpaque == flagOpaque
}

func (n *Node) Default() bool {
	return n.flags&flagDefault == flagDefault
}

// Depending on the type of delete, we may or may not clear the children's
// flags.  When checking authorization on each node, we don't clear, eg
// when we are doing the likes of a load operation and
// deleting everything in the tree which we are authorized to delete.  If we
// cleared the flags on the children, then when we add the new config on top
// of previous deletions, they will reappear if their parent is recreated.
func (n *Node) MarkDeleted(clearChildFlagsWhenDeletingParent bool) {
	n.flags = n.flags | flagDeleted | flagOpaque
	if clearChildFlagsWhenDeletingParent {
		n.ClearChildren()
	}
}

func (n *Node) MarkOpaque() {
	n.flags = n.flags | flagOpaque
}

func (n *Node) MarkDefault() {
	n.flags = n.flags | flagDefault
}

func (n *Node) ClearDeleted() {
	n.flags = n.flags &^ flagDeleted
}

func (n *Node) ClearOpaque() {
	n.flags = n.flags &^ flagOpaque
}

func (n *Node) ClearDefault() {
	n.flags = n.flags &^ flagDefault
}

func (n *Node) SetNoValidate(path []string) {
	n.setNoValidateInternal(path, make([]string, 0, len(path)))
}

func (n *Node) setNoValidateInternal(path, curPath []string) error {
	return n.walkPath(
		func(ch *Node, hd string, tl []string) error {
			return ch.setNoValidateInternal(tl, append(curPath, hd))
		},
		func(ch *Node, hd string, tl []string) error {
			ch = New(hd)
			n.AddChild(ch)
			return ch.setNoValidateInternal(tl, append(curPath, hd))
		},
		func(_ *Node) error {
			return nil
		},
		path,
	)
}

type walker func(*Node, string, []string) error
type laster func(*Node) error

//walkPath is a generic algorithm for walking the configuration path to
//preform actions. The actions are abstracted in the chExists, chNotExists
//and last functions. chExists is called when a child node is found. chNotExists
//is called in the other case. last is called at the end of the path and
//signifies the end of recursion down the path
func (n *Node) walkPath(
	chExists walker,
	chNotExists walker,
	last laster,
	path []string,
) error {
	if len(path) == 0 {
		return last(n)
	}
	hd, tl := path[0], path[1:]
	ch := n.Child(hd)
	if ch == nil || (ch.Deleted() && !ch.Default()) {
		return chNotExists(ch, hd, tl)
	}
	return chExists(ch, hd, tl)
}
