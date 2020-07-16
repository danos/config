// Copyright (c) 2019-2020, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"fmt"

	"github.com/danos/yang/parse"
	yang "github.com/danos/yang/schema"
)

type Node interface {
	yang.Node
	ExtendedNode
}

type ExtendedNode interface {
	hasExtensions
	hasState
	SchemaChild(name string) Node
	HelpMap() map[string]string
}

func Descendant(node Node, path []string) Node {

	if node == nil || len(path) == 0 {
		return node
	}

	return Descendant(schemaChild(node, path[0]), path[1:])
}

func schemaChild(n Node, name string) Node {
	ch := n.Child(name)
	if ch == nil {
		return nil
	}
	return ch.(Node)
}

func (n *modelSet) SchemaChild(name string) Node {
	return schemaChild(n, name)
}

func (n *tree) SchemaChild(name string) Node {
	return schemaChild(n, name)
}

func (n *container) SchemaChild(name string) Node {
	return schemaChild(n, name)
}

func (n *list) SchemaChild(name string) Node {
	return schemaChild(n, name)
}

func (n *listEntry) SchemaChild(name string) Node {
	return schemaChild(n, name)
}

func (n *leaf) SchemaChild(name string) Node {
	return schemaChild(n, name)
}

func (n *leafList) SchemaChild(name string) Node {
	return schemaChild(n, name)
}

func (n *leafValue) SchemaChild(name string) Node {
	return schemaChild(n, name)
}

func (n *opdCommand) SchemaChild(name string) Node {
	return schemaChild(n, name)
}

func (n *opdOption) SchemaChild(name string) Node {
	return schemaChild(n, name)
}

func (n *opdOptionValue) SchemaChild(name string) Node {
	return schemaChild(n, name)
}

func (n *opdArgument) SchemaChild(name string) Node {
	return schemaChild(n, name)
}

type Rpc interface {
	yang.Rpc
	Script() string
}

type rpc struct {
	yang.Rpc
	script string
}

// Compile time check that the concrete type meets the interface
var _ Rpc = (*rpc)(nil)

func (*CompilationExtensions) ExtendRpc(
	p parse.Node, y yang.Rpc,
) (yang.Rpc, error) {

	return &rpc{y, parseExtensions(p).CallRpc}, nil
}

func (r *rpc) Script() string { return r.script }

type Notification interface {
	yang.Notification
}

type notification struct {
	yang.Notification
}

var _ Notification = (*notification)(nil)

func (*CompilationExtensions) ExtendNotification(
	p parse.Node, y yang.Notification,
) (yang.Notification, error) {
	return &notification{y}, nil
}

type Tree interface {
	yang.Tree
	ExtendedNode
}

type tree struct {
	yang.Tree
	*extensions
	*state
}

// Compile time check that the concrete type meets the interface
var _ Tree = (*tree)(nil)

func (*CompilationExtensions) ExtendTree(
	p parse.Node, y yang.Tree,
) (yang.Tree, error) {

	if p == nil {
		ext := newExtend(nil)
		return &tree{y, ext, newState(y, ext)}, nil
	}
	ext := newExtend(parseExtensions(p))
	return &tree{y, ext, newState(y, ext)}, nil
}

func NewTree(children []Node) (Tree, error) {
	ych := make([]yang.Node, len(children), len(children))
	for i, ch := range children {
		ych[i] = ch
	}
	ytree, err := yang.NewTree(ych)
	if err != nil {
		return nil, err
	}
	ext := newExtend(nil)
	return &tree{ytree, ext, newState(ytree, ext)}, nil
}

type Container interface {
	yang.Container
	ExtendedNode
}

type container struct {
	yang.Container
	*extensions
	*state
}

// Compile time check that the concrete type meets the interface
var _ Container = (*container)(nil)

func (*CompilationExtensions) ExtendContainer(
	p parse.Node, y yang.Container,
) (yang.Container, error) {

	ext := newExtend(parseExtensions(p))
	return &container{y, ext, newState(y, ext)}, nil
}

type List interface {
	yang.List
	ExtendedNode
}

type list struct {
	yang.List
	*extensions
	*state

	entry *listEntry
}

// Compile time check that the concrete type meets the interface
var _ List = (*list)(nil)

func (*CompilationExtensions) ExtendList(
	p parse.Node, y yang.List,
) (yang.List, error) {

	pext := parseExtensions(p)
	if p.OrdBy() == "user" {
		if extensionsContainAny(pext, "create", "update", "delete") {
			return nil, fmt.Errorf("only begin and end extensions allowed on ordered-by user list")
		}

		hasCommitExt := func(n parse.Node) bool {
			pext := parseExtensions(n)
			return extensionsContainAny(pext, "create", "update", "delete", "begin", "end")
		}
		var verifyNoCommitActions func(n parse.Node) error
		verifyNoCommitActions = func(n parse.Node) error {
			for _, ch := range n.Children() {
				if hasCommitExt(ch) {
					return fmt.Errorf(
						"action extension not allowed in " +
							"ordered-by user list descendant")
				}
				err := verifyNoCommitActions(ch)
				if err != nil {
					return err
				}
			}
			return nil
		}
		err := verifyNoCommitActions(p)
		if err != nil {
			return nil, err
		}
	}

	ext := newExtend(parseExtensions(p))
	state := newState(y, ext)
	l := &list{
		List:       y,
		extensions: ext,
		state:      state,
	}
	yentry := y.Child("").(yang.ListEntry)
	entry := &listEntry{
		ListEntry:  yentry,
		list:       l,
		extensions: ext,
		state:      newState(yentry, ext),
	}
	l.entry = entry
	return l, nil
}

type ListEntry interface {
	yang.ListEntry
	ExtendedNode
}

type listEntry struct {
	yang.ListEntry
	list List
	*extensions
	*state
}

// Compile time check that the concrete type meets the interface
var _ ListEntry = (*listEntry)(nil)

func (n *list) Child(name string) yang.Node {
	return n.entry
}

func isElemOf(list []string, elem string) bool {
	for _, v := range list {
		if v == elem {
			return true
		}
	}
	return false
}

type Leaf interface {
	yang.Leaf
	ExtendedNode
}

type leaf struct {
	yang.Leaf
	*extensions
	*state
}

// Compile time check that the concrete type meets the interface
var _ Leaf = (*leaf)(nil)

func (*CompilationExtensions) ExtendLeaf(
	p parse.Node, y yang.Leaf,
) (yang.Leaf, error) {

	ext := newExtend(parseExtensions(p))
	return &leaf{y, ext, newState(y, ext)}, nil
}

func (n *leaf) Child(name string) yang.Node {
	y := n.Leaf.Child(name)
	if y == nil {
		return nil
	}
	return &leafValue{y.(yang.LeafValue), n.extensions, n.state}
}

func (n *leaf) DefaultChild(name string) yang.Node {
	y := n.Leaf.DefaultChild(name)
	if y == nil {
		return nil
	}
	return &leafValue{y.(yang.LeafValue), n.extensions, n.state}
}

type Choice interface {
	yang.Choice
	ExtendedNode
}

type LeafList interface {
	yang.LeafList
	ExtendedNode
}

type leafList struct {
	yang.LeafList
	*extensions
	*state
}

// Compile time check that the concrete type meets the interface
var _ LeafList = (*leafList)(nil)

func (*CompilationExtensions) ExtendLeafList(
	p parse.Node, y yang.LeafList,
) (yang.LeafList, error) {

	ext := newExtend(parseExtensions(p))
	return &leafList{y, ext, newState(y, ext)}, nil
}

func (n *leafList) Child(name string) yang.Node {
	y := n.LeafList.Child(name)
	if y == nil {
		return nil
	}
	return &leafValue{y.(yang.LeafValue), n.extensions, n.state}
}

func (n *leafList) DefaultChild(name string) yang.Node {
	y := n.LeafList.DefaultChild(name)
	if y == nil {
		return nil
	}
	return &leafValue{y.(yang.LeafValue), n.extensions, n.state}
}

type LeafValue interface {
	yang.LeafValue
	ExtendedNode
}

type leafValue struct {
	yang.LeafValue
	*extensions
	*state
}

// Compile time check that the concrete type meets the interface
var _ LeafValue = (*leafValue)(nil)

type OpdArgument interface {
	yang.OpdArgument
	ExtendedNode
}

type opdArgument struct {
	yang.OpdArgument
	*extensions
	*state
}

// Compile time check that the concrete type meets the interface
var _ OpdArgument = (*opdArgument)(nil)

func (*CompilationExtensions) ExtendOpdArgument(
	p parse.Node, y yang.OpdArgument,
) (yang.OpdArgument, error) {

	ext := newExtend(parseExtensions(p))
	return &opdArgument{y, ext, newState(y, ext)}, nil
}

type OpdCommand interface {
	yang.OpdCommand
	ExtendedNode
}

type opdCommand struct {
	yang.OpdCommand
	*extensions
	*state
}

// Compile time check that the concrete type meets the interface
var _ OpdCommand = (*opdCommand)(nil)

func (*CompilationExtensions) ExtendOpdCommand(
	p parse.Node, y yang.OpdCommand,
) (yang.OpdCommand, error) {

	ext := newExtend(parseExtensions(p))
	return &opdCommand{y, ext, newState(y, ext)}, nil
}

type OpdOption interface {
	yang.OpdOption
	ExtendedNode
}

type opdOption struct {
	yang.OpdOption
	*extensions
	*state
}

// Compile time check that the concrete type meets the interface
var _ OpdOption = (*opdOption)(nil)

func (n *opdOption) Child(name string) yang.Node {
	y := n.OpdOption.Child(name)
	if y == nil {
		return nil
	}
	if _, ok := n.Type().(Empty); ok || n.Type() == nil {
		return y
	}
	ext := newExtend(n.extensions.ext)
	return &opdOptionValue{
		y.(yang.OpdOptionValue),
		n,
		ext,
		newState(y, ext),
	}
}

func (*CompilationExtensions) ExtendOpdOption(
	p parse.Node, y yang.OpdOption,
) (yang.OpdOption, error) {

	ext := newExtend(parseExtensions(p))
	return &opdOption{y, ext, newState(y, ext)}, nil
}

func (n *opdOption) DefaultChild(name string) yang.Node {
	y := n.OpdOption.DefaultChild(name)
	if y == nil {
		return nil
	}
	return &opdOptionValue{y.(yang.OpdOptionValue), n, n.extensions, n.state}
}

type OpdOptionValue interface {
	yang.OpdOptionValue
	ExtendedNode
}

type opdOptionValue struct {
	yang.OpdOptionValue
	opdOption OpdOption
	*extensions
	*state
}

// Compile time check that the concrete type meets the interface
var _ OpdOptionValue = (*opdOptionValue)(nil)
