// Copyright (c) 2018-2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0
package diff

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/danos/config/data"
	"github.com/danos/config/schema"
	"github.com/danos/utils/natsort"
	"github.com/danos/utils/pathutil"
	"github.com/danos/yang/data/datanode"
)

type status int

const (
	unchanged status = iota
	added
	deleted
)

type Node struct {
	new    *data.Node
	old    *data.Node
	schema schema.Node
	parent *Node
}

type ByUser []*Node

func (b ByUser) Len() int           { return len(b) }
func (b ByUser) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByUser) Less(i, j int) bool { return b[i].Index() < b[j].Index() }

type BySystem []*Node

func (b BySystem) Len() int           { return len(b) }
func (b BySystem) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b BySystem) Less(i, j int) bool { return natsort.Less(b[i].Name(), b[j].Name()) }

func (n *Node) buildChild(name string) *Node {
	newch := n.new.Child(name)
	oldch := n.old.Child(name)
	sch := n.schema.SchemaChild(name)
	if newch != nil && newch.Deleted() {
		newch = nil
	}
	return NewNode(newch, oldch, sch, n)
}

func (n *Node) Data() *data.Node {
	if n.new != nil {
		return n.new
	}
	return n.old
}

func (n *Node) Index() uint64 {
	return n.Data().Index()
}

func (n *Node) IsDefault() bool {
	return n.Data().Default()
}

func (n *Node) Schema() schema.Node {
	return n.schema
}

func (n *Node) Parent() *Node {
	return n.parent
}

func (n *Node) Name() string {
	return n.Data().Name()
}

func (n *Node) Child(name string) *Node {
	return n.buildChild(name)
}

func (n *Node) getStatus() status {
	switch {
	case n.Deleted():
		return deleted
	case n.Added():
		return added
	default:
		return unchanged
	}
}

func (n *Node) Added() bool {
	return n.new != nil && n.old == nil ||
		(n.new != nil && n.old != nil && !n.new.Default() && n.old.Default())
}

func (n *Node) Deleted() bool {
	return n.new == nil && n.old != nil
}

func (n *Node) Updated() bool {
	//Updated means we have a child that has changed
	//it turns out that this accounts for all 3 cases
	//previously handled.

	//technically this is deleted, but defaults are special
	//and need to be considered updated
	if n.new != nil && n.old != nil && n.new.Default() && !n.old.Default() {
		return true
	}
	for _, ch := range n.UnsortedChildren() {
		if ch.Changed() {
			return true
		}
	}
	return false
}

func (n *Node) Changed() bool {
	return n.Added() || n.Deleted() || n.Updated()
}

func (n *Node) getSortedChildren(parent *data.Node) []*data.Node {
	children := parent.Children()
	return n.sortDataChildren(children)
}

func (n *Node) sortDataChildren(children []*data.Node) []*data.Node {
	switch n.schema.OrdBy() {
	case "user":
		sort.Sort(data.ByUser(children))
	default:
		sort.Sort(data.BySystem(children))
	}
	for i, ch := range children {
		ch.SetIndex(uint64(i))
	}
	return children
}

func (n *Node) sortChildren(children []*Node) []*Node {
	switch n.schema.OrdBy() {
	case "user":
		sort.Sort(ByUser(children))
	default:
		sort.Sort(BySystem(children))
	}
	return children
}

func (n *Node) traverseChildren(
	children []*data.Node,
	fn func(*Node),
	skipFn func(*Node) bool,
	buildFn func(string) *Node,
) {
	for _, ch := range children {
		ch := buildFn(ch.Name())
		if ch == nil {
			continue
		}
		if skipFn(ch) {
			continue
		}
		fn(ch)
	}
}

func (n *Node) traverseDiffChildren(
	fn func(*Node),
	skip func(*Node) bool,
) {
	seen := make(map[string]struct{})
	travFn := func(ch *Node) {
		seen[ch.Name()] = struct{}{}
		fn(ch)
	}
	skipFn := func(ch *Node) bool {
		_, ok := seen[ch.Name()]
		if ok {
			return ok
		}
		return skip != nil && skip(ch)
	}
	children := n.old.Children()
	children = append(children, n.new.Children()...)
	n.traverseChildren(
		children,
		travFn,
		skipFn,
		n.buildChild)
}

// ordered-by-user lists and leaf-lists require rather more complex logic
// when traversing to ensure we get entries in the right order, and the
// correct number of times.
//
// We first traverse the list of old nodes, taking into account the fact that
// if the index has changed for a node between old and new node, then it's as
// if the node doesn't exist in the new list and the node will appear as
// deleted in the old list.
//
// We then traverse the list of new nodes.  As with old nodes, we play the
// same game with the index, but there's an added twist.  If the node is
// only marked as updated, rather than Added or Deleted, we ignore it as it
// will already be in the list of old nodes as Updated.
func (n *Node) traverseDiffChildrenUser(
	fn func(*Node),
	skip func(*Node) bool,
) {
	//user ordered children need to be treated
	//specially otherwise the differences aren't
	//reflected in the output. Deleteing an entry
	//in the middle of the list re-indexes the list,
	//so all nodes after it appear to be recreated.
	//TODO: it would be nice if we could only show
	//the one deletion during serialization.

	skipFn := func(ch *Node) bool {
		return skip != nil && skip(ch)
	}

	var dch *Node
	for _, ch := range n.getSortedChildren(n.old) {
		sch := n.schema.SchemaChild(ch.Name())
		if sch == nil {
			continue
		}
		new := n.new.Child(ch.Name())
		if new != nil && new.Index() != ch.Index() {
			new = nil
		}
		dch = NewNode(new, ch, sch, n)
		if skipFn(dch) {
			continue
		}
		fn(dch)
	}
	for _, ch := range n.getSortedChildren(n.new) {
		sch := n.schema.SchemaChild(ch.Name())
		if sch == nil {
			continue
		}
		old := n.old.Child(ch.Name())
		if old != nil && ch.Index() != old.Index() {
			old = nil
		}
		dch = NewNode(ch, old, sch, n)
		if skipFn(dch) || !(dch.Added() || dch.Deleted()) {
			continue
		}
		fn(dch)
	}
}

func (n *Node) Children() []*Node {
	return n.sortChildren(n.children())
}

func (n *Node) UnsortedChildren() []*Node {
	return n.children()
}

func (n *Node) children() []*Node {

	out := make([]*Node, 0)
	travFn := func(n *Node) {
		out = append(out, n)
	}
	if n.schema.OrdBy() == "user" {
		n.traverseDiffChildrenUser(
			travFn,
			nil,
		)
	} else {
		n.traverseDiffChildren(
			travFn,
			nil,
		)
	}
	return out
}

func (n *Node) YangDataName() string {
	return n.Name()
}

func (n *Node) YangDataChildren() []datanode.DataNode {
	children := func(n *Node) []*Node {
		return n.Children()
	}
	return n.yangDataChildren(children)
}

func (n *Node) YangDataChildrenNoSorting() []datanode.DataNode {
	children := func(n *Node) []*Node {
		return n.UnsortedChildren()
	}
	return n.yangDataChildren(children)
}

func (n *Node) yangDataChildren(
	getChildrenFn func(n *Node) []*Node,
) []datanode.DataNode {

	out := make([]datanode.DataNode, 0)

	if sch, ok := n.schema.(schema.ListEntry); ok {
		name := sch.Keys()[0]
		new_node := datanode.CreateDataNode(name, nil, []string{n.Name()})
		out = append(out, new_node)
	}

	for _, child := range getChildrenFn(n) {
		if child.Deleted() {
			continue
		}
		out = append(out, child)
	}
	return out
}

func (n *Node) YangDataValues() []string {
	children := func(n *Node) []*Node {
		return n.Children()
	}
	return n.yangDataValues(children)
}

func (n *Node) YangDataValuesNoSorting() []string {
	children := func(n *Node) []*Node {
		return n.UnsortedChildren()
	}
	return n.yangDataValues(children)
}

func (n *Node) yangDataValues(getChildrenFn func(n *Node) []*Node) []string {
	out := make([]string, 0)
	for _, child := range getChildrenFn(n) {
		if child.Deleted() {
			continue
		}
		out = append(out, child.Name())
	}
	return out

}

func (n *Node) Descendant(path []string) *Node {
	if len(path) == 0 {
		return n
	}
	hd, tl := path[0], path[1:]
	ch := n.Child(hd)
	if ch == nil {
		return ch
	}
	return ch.Descendant(tl)
}

func (n *Node) EmptyNonDefault() bool {
	for _, ch := range n.UnsortedChildren() {
		if ch.IsDefault() {
			continue
		}
		return false
	}
	return true
}

func (n *Node) Empty() bool {
	return len(n.UnsortedChildren()) == 0
}

func (n *Node) serializeChildren(
	w io.Writer,
	path []string,
	ctxdiff bool,
	lvl int,
) {
	var diffChildren []*Node
	for _, ch := range n.Children() {
		if ctxdiff {
			if ch.getStatus() != unchanged {
				diffChildren = append(diffChildren, ch)
				continue
			}
		}
		ch.serialize(w, path, ctxdiff, lvl)
	}
	if ctxdiff && len(diffChildren) > 0 {
		writeCtxdiff(w, path)
		for _, ch := range diffChildren {
			ch.serialize(w, path, false, lvl)
		}
	}
}

func (n *Node) serializeContainer(
	w io.Writer,
	path []string,
	ctxdiff bool,
	lvl int,
) {
	if n.serializeCtxDiff(w, path, ctxdiff, true, lvl) {
		return
	}
	printStatus(w, n)
	printLevel(w, lvl)
	fmt.Fprint(w, n.Name())
	if n.Empty() {
		fmt.Fprint(w, "\n")
		return
	}
	fmt.Fprintln(w, " {")
	n.serializeChildren(w, path, ctxdiff, lvl+1)
	printStatus(w, n)
	printLevel(w, lvl)
	fmt.Fprintln(w, "}")
}

func (n *Node) serializeList(
	w io.Writer,
	path []string,
	ctxdiff bool,
	lvl int,
) {
	if n.serializeCtxDiff(w, path, ctxdiff, false, lvl) {
		return
	}
	n.serializeChildren(w, path, ctxdiff, lvl)
}

func (n *Node) serializeListEntry(
	w io.Writer,
	path []string,
	ctxdiff bool,
	lvl int,
) {
	if n.serializeCtxDiff(w,
		pathutil.CopyAppend(path, n.parent.Name()), ctxdiff, true, lvl) {
		return
	}
	printStatus(w, n)
	printLevel(w, lvl)
	fmt.Fprintf(w, "%s %s", n.parent.Name(), quote(n.Name()))
	if n.Empty() {
		fmt.Fprint(w, "\n")
		return
	}
	fmt.Fprintln(w, " {")
	n.serializeChildren(w, path, ctxdiff, lvl+1)
	printStatus(w, n)
	printLevel(w, lvl)
	fmt.Fprintln(w, "}")
}

func (n *Node) serializeLeaf(
	w io.Writer,
	path []string,
	ctxdiff bool,
	lvl int,
) {
	if n.serializeCtxDiff(w, path, ctxdiff, false, lvl) {
		return
	}
	if _, isEmpty := n.schema.Type().(schema.Empty); isEmpty {
		printStatus(w, n)
		printLevel(w, lvl)
		fmt.Fprintf(w, "%s\n", n.Name())
		return
	}
	n.serializeChildren(w, path, ctxdiff, lvl)
}

func (n *Node) serializeLeafList(
	w io.Writer,
	path []string,
	ctxdiff bool,
	lvl int,
) {
	if n.serializeCtxDiff(w, path, ctxdiff, false, lvl) {
		return
	}
	n.serializeChildren(w, path, ctxdiff, lvl)
}

func (n *Node) serializeLeafValue(
	w io.Writer,
	path []string,
	ctxdiff bool,
	lvl int,
) {
	if n.serializeCtxDiff(w, path, ctxdiff, false, lvl) {
		return
	}
	printStatus(w, n)
	printLevel(w, lvl)
	fmt.Fprintf(w, "%s %s\n", n.parent.Name(), quote(n.Name()))
}

func (n *Node) serializeCtxDiff(
	w io.Writer,
	path []string,
	ctxdiff, append bool,
	lvl int,
) bool {
	if ctxdiff && n.getStatus() == unchanged {
		if append {
			path = pathutil.CopyAppend(path, n.Name())
		}
		n.serializeChildren(w, path, ctxdiff, lvl)
		return true
	}
	return false
}

func (n *Node) serialize(w io.Writer, path []string, ctxdiff bool, lvl int) {
	switch n.schema.(type) {
	case schema.Container:
		n.serializeContainer(w, path, ctxdiff, lvl)
	case schema.List:
		n.serializeList(w, path, ctxdiff, lvl)
	case schema.ListEntry:
		n.serializeListEntry(w, path, ctxdiff, lvl)
	case schema.Leaf:
		n.serializeLeaf(w, path, ctxdiff, lvl)
	case schema.LeafList:
		n.serializeLeafList(w, path, ctxdiff, lvl)
	case schema.LeafValue:
		n.serializeLeafValue(w, path, ctxdiff, lvl)
	case schema.Tree:
		n.serializeChildren(w, path, ctxdiff, lvl)
	}
}

// If there's nothing to serialize (node is nil) this isn't an error, and
// we just return the empty string.
func (n *Node) Serialize(ctxdiff bool) string {
	if n == nil {
		return ""
	}
	var buf bytes.Buffer
	n.serialize(&buf, nil, ctxdiff, 0)
	return buf.String()
}

func quote(in string) string {
	if strings.ContainsAny(in, "*}{;\011\012\013\014\015 ") {
		return "\"" + in + "\""
	}
	return in
}

func writeCtxdiff(w io.Writer, path []string) {
	fmt.Fprint(w, "[edit")
	for _, elem := range path {
		fmt.Fprintf(w, " %s", elem)
	}
	fmt.Fprintln(w, "]")
}

func printLevel(w io.Writer, lvl int) {
	for i := 0; i < lvl; i++ {
		fmt.Fprint(w, "\t")
	}
}

func printStatus(w io.Writer, n *Node) {
	switch n.getStatus() {
	case added:
		fmt.Fprint(w, "+")
	case deleted:
		fmt.Fprint(w, "-")
	default:
		fmt.Fprint(w, " ")
	}
}

func NewNode(new, old *data.Node, sch schema.Node, parent *Node) *Node {
	switch {
	case sch == nil:
		return nil
	case new == nil && old == nil:
		return nil
	}
	return &Node{
		new:    new,
		old:    old,
		schema: sch,
		parent: parent,
	}
}

func CreateChangedNSMap(
	new, old *data.Node,
	sch schema.Node,
	parent *Node,
) *map[string]bool {
	switch {
	case sch == nil:
		return nil
	case new == nil && old == nil:
		return nil
	}
	diffs := &Node{
		new:    new,
		old:    old,
		schema: sch,
		parent: parent,
	}
	return diffs.walk()
}

func (n *Node) QualifiedNamespace() string {
	if n.schema.Submodule() == "" {
		return n.schema.Namespace()
	}
	return n.schema.Submodule() + "@" + n.schema.Namespace()
}

func (n *Node) walkChildren(
	nsMap *map[string]bool,
	path []string,
) {
	for _, ch := range n.Children() {
		ch.walkInternal(nsMap, path)
	}
}

// Update all node types on addition / deletion.
// includeUpdates allows leaf nodes to be flagged when they are updated -
// for other nodes only addition or deletion matters.
func updateMap(nsMap *map[string]bool, n *Node, includeUpdates bool) {
	if n.Added() || n.Deleted() || (includeUpdates && n.Updated()) {
		(*nsMap)[n.QualifiedNamespace()] = true
	}
}

func (n *Node) walkContainer(
	nsMap *map[string]bool,
	path []string,
) {
	if n.schema.HasPresence() {
		updateMap(nsMap, n, false)
	}
	if n.Empty() {
		return
	}
	n.walkChildren(nsMap, path)
}

func (n *Node) walkList(
	nsMap *map[string]bool,
	path []string,
) {
	n.walkChildren(nsMap, path)
}

func (n *Node) walkListEntry(
	nsMap *map[string]bool,
	path []string,
) {
	updateMap(nsMap, n, false)
	if n.Empty() {
		return
	}
	n.walkChildren(nsMap, path)
}

func (n *Node) walkLeaf(
	nsMap *map[string]bool,
	path []string,
) {
	if _, isEmpty := n.schema.Type().(schema.Empty); isEmpty {
		updateMap(nsMap, n, true)
		return
	}
	n.walkChildren(nsMap, path)
}

func (n *Node) walkLeafList(
	nsMap *map[string]bool,
	path []string,
) {
	n.walkChildren(nsMap, path)
}

func (n *Node) walkLeafValue(
	nsMap *map[string]bool,
	path []string,
) {
	updateMap(nsMap, n.parent, true)
}

func (n *Node) walkInternal(nsMap *map[string]bool, path []string) {
	switch n.schema.(type) {
	case schema.Container:
		n.walkContainer(nsMap, path)
	case schema.List:
		n.walkList(nsMap, path)
	case schema.ListEntry:
		n.walkListEntry(nsMap, path)
	case schema.Leaf:
		n.walkLeaf(nsMap, path)
	case schema.LeafList:
		n.walkLeafList(nsMap, path)
	case schema.LeafValue:
		n.walkLeafValue(nsMap, path)
	case schema.Tree:
		n.walkChildren(nsMap, path)
	}
}

// If there's nothing to walk (node is nil) this isn't an error, and
// we just return the empty string.
func (n *Node) walk() *map[string]bool {
	if n == nil {
		return nil
	}
	nsMap := make(map[string]bool, 0)
	n.walkInternal(&nsMap, nil)
	return &nsMap
}
