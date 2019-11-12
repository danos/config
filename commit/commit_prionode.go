// Copyright (c) 2019, AT&T Intellectual Copyright. All rights reserved.
//
// Copyright (c) 2014-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package commit

import (
	"fmt"

	"github.com/danos/config/schema"
	"github.com/danos/utils/exec"
	"github.com/danos/utils/pathutil"
)

func tryRedactPath(cfg *CfgNode, path []string) []string {
	for cfg.Parent != nil {
		cfg = cfg.Parent
	}
	pathAttrs := schema.AttrsForPath(cfg.Schema(), path)
	rpath, _ := pathutil.RedactPath(path, pathAttrs)
	return rpath
}

type PrioNode struct {
	Priority uint
	Cfg      *CfgNode
	Parent   *PrioNode
	Children PrioNodes
}

type commitAction int

const (
	created commitAction = iota
	deleted
	updated
)

func (c commitAction) String() string {
	switch c {
	case created:
		return "created"
	case deleted:
		return "deleted"
	case updated:
		return "updated"
	default:
		return "unknown"
	}
}

// Determine the action that is happening for a node
// on the delete list. Each node has a few state bits,
// each True (T) or False (F) for IsDefault(), Deleted(), etc.
// Nodes on the delete list can have the states in the following
// Table. The table maps these to the operation/action that results
// in the states.
//
// IsDefault  Deleted  Added  Updated  Changed
//
//     F         F       T       T        T    [1] Replace default with
//                                             explicit value
//
//     F         T       F       T        T    [2] Delete a value, no default
//                                             on node to replace it
//
//
//     T         F       F       T        T    [3] Delete a value, replace it
//                                             with the nodes default value
//
//     T         F       T       T        T    [4] Inital creation of a node with
//                                             instantiation of the default value
//
// These states have been collapsed down into the switch statement, each case
// is evaluated in order, first match wins.
//
func deleteAction(n *CfgNode) commitAction {
	switch {
	case n.IsDefault() && n.Added():
		// [4] Initial instantiation of a node with default value
		return created

	case n.IsDefault():
		// [3] Delete value, replacing with default
		return updated

	default:
		// States [1], [2], and everything else
		// Node is being explicitly deleted, or default is being
		// removed, to be replaced by an explicit value
		return deleted
	}
}

func createAction(n *CfgNode) commitAction {
	switch {
	case n.Added():
		return created

	default:
		return updated
	}
}

type getAction = func(n *CfgNode) commitAction

type change struct {
	path   []string
	action commitAction
}

func changeEntry(n *CfgNode, act getAction) *change {
	return changeEntryWithPath(n, act, n.Path)
}

func changeEntryWithPath(n *CfgNode, act getAction, path []string) *change {
	action := updated

	if act != nil {
		action = act(n)
	}

	return &change{path: path,
		action: action}
}

func (pnode *PrioNode) getDlist() []*change {
	dlist := make([]*change, 0)
	nodes := pnode.Cfg.PostOrder()
	for _, n := range nodes {
		if (n.Parent != nil && n.Parent.Deleted()) || n.IsIgnore() {
			continue
		}
		switch n.Schema().(type) {
		case schema.Tree, schema.Container, schema.List, schema.ListEntry:
			if n.Deleted() {
				dlist = append(dlist, changeEntry(n, deleteAction))
			}
		case schema.Leaf:
			if n.Deleted() || (n.IsDefault() && n.Updated()) {
				dlist = append(dlist, changeEntry(n, deleteAction))
				continue
			}
			for _, v := range n.GetDeletedValues() {
				path := pathutil.CopyAppend(n.Path, v)
				dlist = append(dlist,
					changeEntryWithPath(n, deleteAction, path))
			}
		case schema.LeafList:
			//TODO: this is to preserve order by user order...
			//Lists need something similar, but for now this
			//keeps the proper order.
			if n.Deleted() || len(n.GetDeletedValues()) > 0 {
				dlist = append(dlist, changeEntry(n, deleteAction))
			}
		}
	}
	return dlist
}

func (pnode *PrioNode) RunDeleteActions(ctx Context) ([]*exec.Output, []error, bool) {
	pnode.SetChanged()
	dlist := pnode.getDlist()
	effective := ctx.Effective()

	ctx.Log("run delete actions for priority root", pnode.Cfg.Path)
	outs, errs, ok := pnode.Cfg.Delete()

	for _, dl := range dlist {
		effective.Delete(dl.path)
		ctx.LogAudit(fmt.Sprintf(
			"configuration path %s %s by user %d",
			tryRedactPath(pnode.Cfg, dl.path), dl.action, pnode.Cfg.ctx.Uid()))
	}
	//log
	return outs, errs, ok

}

func (pnode *PrioNode) getClist() []*change {
	nodes := pnode.Cfg.PreOrder()
	clist := make([]*change, 0)
	for _, n := range nodes {
		if n.IsDefault() || n.IsIgnore() {
			continue
		}
		switch n.Schema().(type) {
		case schema.Tree, schema.Container, schema.List, schema.ListEntry:
			if !n.EmptyNonDefault() {
				continue
			}
			if n.Changed() && !n.Deleted() {
				clist = append(clist, changeEntry(n, createAction))
			}
		case schema.Leaf:
			if _, ok := n.Schema().Type().(schema.Empty); ok && !n.Deleted() {
				clist = append(clist, changeEntry(n, createAction))
				continue
			}
			for _, v := range n.Node.Children() {
				if v.IsDefault() || !v.Added() {
					continue
				}
				p := pathutil.CopyAppend(n.Path, v.Name())
				clist = append(clist,
					changeEntryWithPath(n, createAction, p))
			}
		case schema.LeafList:
			//recreate entire leaflist preserves user order
			//TODO: lists need something similar?
			for _, v := range n.GetValues() {
				p := pathutil.CopyAppend(n.Path, v)
				clist = append(clist, changeEntryWithPath(n, createAction, p))
			}
		}
	}
	return clist
}

func (pnode *PrioNode) SetChanged() {
	nodes := pnode.Cfg.PreOrder()
	for _, n := range nodes {
		n.SetSubtreeChanged()
	}
}

func (pnode *PrioNode) RunUpdateActions(ctx Context) ([]*exec.Output, []error, bool) {
	effective := ctx.Effective()
	ctx.Log("run update actions for priority root", pnode.Cfg.Path)

	pnode.SetChanged()
	dlist := pnode.getDlist()
	clist := pnode.getClist()

	//recursively run update actions
	outs, errs, ok := pnode.Cfg.Update()

	//copy commit list to effective tree
	for _, dl := range dlist {
		effective.Delete(dl.path)
		ctx.LogAudit(fmt.Sprintf("configuration path %s %s by user %d",
			tryRedactPath(pnode.Cfg, dl.path), dl.action, pnode.Cfg.ctx.Uid()))
	}
	for _, cl := range clist {
		effective.Set(cl.path)
		ctx.LogAudit(fmt.Sprintf("configuration path %s %s by user %d",
			tryRedactPath(pnode.Cfg, cl.path), cl.action, pnode.Cfg.ctx.Uid()))
	}
	//log
	return outs, errs, ok
}
