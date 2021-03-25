// Copyright (c) 2018-2020, AT&T Intellectual Copyright. All rights reserved.
//
// Copyright (c) 2014-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package commit

import (
	"container/heap"
	"fmt"
	"time"

	"github.com/danos/config/data"
	"github.com/danos/config/diff"
	"github.com/danos/config/schema"
	"github.com/danos/mgmterror"
	"github.com/danos/utils/exec"
	"github.com/danos/utils/pathutil"
)

func init() {
	exec.NewExecError = func(path []string, err string) error {
		return mgmterror.NewExecError(path, err)
	}
}

type EffectiveDatabase interface {
	Set([]string) error
	Delete([]string) error
}

type Context interface {
	Log(...interface{})
	LogError(...interface{})
	LogCommitMsg(string)
	LogCommitTime(string, time.Time)
	LogAudit(string)
	Debug() bool
	MustDebugThreshold() int
	Sid() string
	Uid() uint32
	Running() *data.Node
	Candidate() *data.Node
	Schema() schema.Node
	RunDeferred() bool
	Effective() EffectiveDatabase
}

type ComponentContext interface {
	CompMgr() schema.ComponentManager
}

func getComponentManager(ctx Context) schema.ComponentManager {
	cmpCtx, hasCmpCtx := ctx.(ComponentContext)
	if !hasCmpCtx {
		return nil
	}
	return cmpCtx.CompMgr()
}

func runPrioTrees(
	ctx Context,
	setq *MinHeap,
	delq *MaxHeap,
) ([]*exec.Output, []error, int, int) {
	var successes, failures int
	var outs []*exec.Output
	var errs []error
	ctx.Log("execute deletes")
	start := time.Now()
	for !delq.Empty() {
		n := heap.Pop(delq).(*PrioNode)
		ctx.Log(n.Priority)
		couts, cerrs, ok := n.RunDeleteActions(ctx)
		if !ok {
			failures++
		} else {
			successes++
		}
		outs = append(outs, couts...)
		errs = append(errs, cerrs...)
	}
	ctx.LogCommitTime("Delete Actions", start)
	ctx.Log("execute sets")
	start = time.Now()
	for !setq.Empty() {
		n := heap.Pop(setq).(*PrioNode)
		ctx.Log(n.Priority)
		couts, cerrs, ok := n.RunUpdateActions(ctx)
		if !ok {
			failures++
		} else {
			successes++
		}
		outs = append(outs, couts...)
		errs = append(errs, cerrs...)
	}
	ctx.LogCommitTime("Set Actions", start)
	return outs, errs, successes, failures
}

func Changed(ctx Context) bool {
	t := buildCommitTree(ctx, nil,
		diff.NewNode(ctx.Candidate(), ctx.Running(), ctx.Schema(), nil),
		true, false)
	return t != nil && len(t.UnsortedChildren()) != 0
}

func Validate(ctx Context) ([]*exec.Output, []error, bool) {
	start := time.Now()
	t := buildCommitTree(ctx, nil,
		diff.NewNode(ctx.Candidate(), ctx.Running(), ctx.Schema(), nil),
		false, true)
	op, errs, ok := t.Validate(getComponentManager(ctx))
	ctx.LogCommitTime("Validate OVERALL", start)
	return op, errs, ok
}

func Commit(ctx Context) ([]*exec.Output, []error, int, int) {
	var outs []*exec.Output
	var errs []error

	cfgtree := buildCommitTree(ctx, nil,
		diff.NewNode(ctx.Candidate(), ctx.Running(), ctx.Schema(), nil),
		true, false)
	if cfgtree == nil {
		e := mgmterror.NewOperationFailedProtocolError()
		e.Message = "No changes to commit"
		errs := append(errs, e)
		return nil, errs, 0, 1
	}

	proot := &PrioNode{
		Parent:   nil,
		Cfg:      cfgtree,
		Priority: 0,
	}
	proot = buildPrioTree(cfgtree, proot)
	setq, delq := buildQueues(proot)
	out, err, successes, failures := runPrioTrees(ctx, setq, delq)
	outs = append(outs, out...)
	errs = append(errs, err...)

	return outs, errs, successes, failures
}

func buildPrioTree(root *CfgNode, parent *PrioNode) *PrioNode {
	attachprionode := func(root *CfgNode, parent *PrioNode, prio uint) *PrioNode {
		proot := &PrioNode{
			Priority: prio,
			Cfg:      root,
			Parent:   parent,
			Children: make(PrioNodes, 0),
		}
		parent.Children = append(parent.Children, proot)
		return proot
	}

	proot := parent
	var priority uint

	priority = root.Schema().ConfigdExt().Priority
	cnodes := make([]*CfgNode, len(root.CfgChildren))
	copy(cnodes, root.CfgChildren)
	if priority != 0 {
		//do what cstore does skip list and only add the keys
		if !root.IsList() && !root.IsLeafValue() {
			pprio := parent.Priority
			if priority <= pprio {
				path := root.Path
				fmt.Printf("Warning: priority inversion %s(%d) <= %s(%d)\n"+
					"         changing %s to (%d)\n",
					path, priority, parent.Cfg.Path, pprio, path, pprio+1)
				priority = pprio + 1
			}
			proot = attachprionode(root, parent, priority)
			root.Parent.DeleteChild(root)
		}
	}
	for _, ch := range cnodes {
		buildPrioTree(ch, proot)
	}
	return proot
}

func buildQueues(proot *PrioNode) (setq *MinHeap, delq *MaxHeap) {
	setq = &MinHeap{make(PrioNodes, 0)}
	heap.Init(setq)
	delq = &MaxHeap{make(PrioNodes, 0)}
	heap.Init(delq)

	var buildq func(n *PrioNode)
	buildq = func(n *PrioNode) {
		if n == nil {
			return
		}
		if n.Cfg != nil {
			if n.Cfg.Deleted() {
				heap.Push(delq, n)
			} else {
				heap.Push(setq, n)
			}
		}

		// For creation, need to add children in reverse order
		// if ListEntry or LeafList due to FILO queues.
		//
		// For delete, need to add children in order.
		setch := []*PrioNode{}
		for _, ch := range n.Children {
			if ch.Cfg.Deleted() {
				buildq(ch)
			} else {
				setch = append(setch, ch)
			}
		}
		for ch := len(setch) - 1; ch >= 0; ch-- {
			buildq(setch[ch])
		}
	}
	buildq(proot)

	return setq, delq
}

func buildCommitTree(
	ctx Context,
	parent *CfgNode,
	diffNode *diff.Node,
	skipUnchanged,
	skipDeleted bool,
) *CfgNode {
	node := &CfgNode{
		Parent:      parent,
		Node:        diffNode,
		CfgChildren: make([]*CfgNode, 0),
		ctx:         ctx,
	}

	if node.Schema() == nil {
		panic(fmt.Errorf("Missing Schema"))
	}
	if parent != nil {
		if parent.IsList() {
			node.deferred = parent.deferred
		} else {
			node.deferred = parent.deferred ||
				node.Schema().ConfigdExt().DeferActions != ""
		}
	} else {
		node.deferred = node.Schema().ConfigdExt().DeferActions != ""
	}

	if skipDeleted && node.Deleted() {
		return nil
	}
	if parent == nil {
		node.Path = []string{}
	} else {
		node.Path = pathutil.CopyAppend(parent.Path, node.Name())
	}

	var child *CfgNode
	// If a ListEntry (aka List), add a child for each key
	if l, ok := diffNode.Schema().(schema.ListEntry); ok {
		for _, k := range l.Keys() {
			d := data.New(k)
			d.AddChild(data.New(diffNode.Name()))
			var n *diff.Node
			if diffNode.Added() {
				n = diff.NewNode(d, nil, diffNode.Schema().SchemaChild(k), diffNode)
			} else if diffNode.Deleted() {
				n = diff.NewNode(nil, d, diffNode.Schema().SchemaChild(k), diffNode)
			} else {
				n = diff.NewNode(d, d, diffNode.Schema().SchemaChild(k), diffNode)
			}
			child = buildCommitTree(ctx, node, n, skipUnchanged, skipDeleted)
			if child == nil {
				continue
			}
			child.SetIgnore(true)
			node.CfgChildren = append(node.CfgChildren, child)
		}
	}
	for _, n := range node.Children() {
		child = buildCommitTree(ctx, node, n, skipUnchanged, skipDeleted)
		if child == nil {
			continue
		}
		node.CfgChildren = append(node.CfgChildren, child)
	}

	if skipUnchanged && !node.Changed() && len(node.CfgChildren) == 0 {
		return nil
	}
	return node
}
