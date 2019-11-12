// Copyright (c) 2017-2019, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package union

import (
	"github.com/danos/config/data"
	"github.com/danos/config/schema"
	"github.com/danos/yang/data/datanode"
)

type internalBaseNode interface {
	setHook()
	emptyNonDefault() bool
	deleted() bool
	opaque() bool
	def() bool
	copyUp() Node
	addChild(*data.Node) Node
	hasUnderlay() bool
	hasOverlay() bool
	markDeleted(clearChildFlagsWhenDeletingParent bool)
	deleteIfEmpty(checkAuth bool)
	deleteEmptyParent(checkAuth bool)
	setIndex(uint64)
	index() uint64
	getnode() *node
}

//baseNode describes the interface for generic unioning functions
type baseNode interface {
	internalBaseNode
	Data() *data.Node
	Name() string
	Children() map[string]Node
	NumChildren() int
	SortedChildren() []Node
	Child(name string) Node
	Parent() Node
	Added() bool
	Changed() bool
	Updated() bool
	Empty() bool
	GetSchema() schema.Node
	GetStateJson([]string) ([][]byte, error)
	GetStateJsonWithWarnings(
		[]string,
		schema.StateLogger,
	) ([][]byte, []error)
	Default() bool
	Module() string
}

//internalSchemaNode describes the internal API for each node type
type internalSchemaNode interface {
	get(path, curPath []string) ([]string, error)
	set(path, curPath []string) error
	show(path, curPath []string, opts *unionOptions) (string, error)
	deleteEverythingUnder(path, curPath []string) error
	deleteCheckAuth(path, curPath []string) error
	deleteWalkerCheckAuth(auth Auther, root Node, path []string) error
	exists(path, curPath []string) error
	notExists(path, curPath []string) error
	validateDeletePath(path, curPath []string) error
	validateNotExistsSet(path, curPath []string) error
	descendant(path, curPath []string) (Node, error)
	//getHelp(path, curPath []string, fromSchema bool) map[string]string
	isDefault(path, curPath []string) (bool, error)
	serialize(b Serializer, path []string, lvl int, opts *unionOptions)
	getHelp(auth Auther, fromSchema bool, path, curPath []string) (map[string]string, error)
}

//schemaNode describes the external API for each node type but hides the implementation details
type schemaNode interface {
	internalSchemaNode
	Get(auth Auther, path []string) ([]string, error)
	Set(auth Auther, path []string) error
	Show(path []string, options ...UnionOption) (string, error)
	Delete(auth Auther, path []string, checkAuth bool) error
	Exists(auth Auther, path []string) error
	Descendant(auth Auther, path []string) (Node, error)
	//GetHelp(path []string, fromSchema bool) map[string]string
	IsDefault(auth Auther, path []string) (bool, error)
	Merge() *data.Node
	MergeWithoutDefaults() *data.Node
	Serialize(b Serializer, path []string, options ...UnionOption)
	ToJSON(options ...UnionOption) []byte
	ToRFC7951(options ...UnionOption) []byte
	MarshalInternalJSON(options ...UnionOption) []byte
	ToNETCONF(rootName string, options ...UnionOption) []byte
	ToXML(rootName string, options ...UnionOption) []byte
	Marshal(rootName, encoding string, options ...UnionOption) (string, error)
	GetHelp(auth Auther, fromSchema bool, path []string) (map[string]string, error)
}

//Node describes the full external API
type Node interface {
	baseNode
	schemaNode
	datanode.DataNode
}
