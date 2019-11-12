// Copyright (c) 2018-2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2014-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package commit

import (
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/danos/config/diff"
	"github.com/danos/config/schema"
	"github.com/danos/mgmterror"
	"github.com/danos/utils/exec"
	"github.com/danos/utils/pathutil"
)

type CfgNode struct {
	*diff.Node
	Path        []string
	Parent      *CfgNode
	CfgChildren []*CfgNode
	SubChanged  bool
	ctx         Context
	ignore      bool
	deferred    bool
}

func (c *CfgNode) IsList() bool {
	_, ok := c.Node.Schema().(schema.List)
	return ok
}

func (c *CfgNode) IsListEntry() bool {
	_, ok := c.Node.Schema().(schema.ListEntry)
	return ok
}

func (c *CfgNode) IsLeaf() bool {
	_, ok := c.Node.Schema().(schema.Leaf)
	return ok
}

func (c *CfgNode) IsLeafList() bool {
	_, ok := c.Node.Schema().(schema.LeafList)
	return ok
}

func (c *CfgNode) IsLeafValue() bool {
	_, ok := c.Node.Schema().(schema.LeafValue)
	return ok
}

func (c *CfgNode) GetValues() []string {
	children := c.Node.Children()
	out := make([]string, 0, len(children))
	for _, ch := range children {
		if ch.Deleted() {
			continue
		}
		out = append(out, ch.Name())
	}
	return out
}

func (c *CfgNode) GetAddedValues() []string {
	var out []string
	for _, ch := range c.Node.Children() {
		if ch.Added() {
			out = append(out, ch.Name())
		}
	}
	return out
}

func (c *CfgNode) GetDeletedValues() []string {
	var out []string
	for _, ch := range c.Node.Children() {
		if ch.Deleted() {
			out = append(out, ch.Name())
		}
	}
	return out
}

func (c *CfgNode) BeginEnd() bool {
	e := c.Schema().ConfigdExt()
	return len(e.Begin) > 0 || len(e.End) > 0
}

func (c *CfgNode) PreOrder() []*CfgNode {
	var preord func(c *CfgNode, nodes []*CfgNode) []*CfgNode
	preord = func(c *CfgNode, nodes []*CfgNode) []*CfgNode {
		if c == nil {
			return nodes
		}
		nodes = append(nodes, c)
		for _, ch := range c.CfgChildren {
			nodes = preord(ch, nodes)
		}
		return nodes
	}
	nodes := make([]*CfgNode, 0)
	return preord(c, nodes)
}

func (c *CfgNode) PreOrderFilterBE() []*CfgNode {
	var preord func(n *CfgNode, nodes []*CfgNode) []*CfgNode
	preord = func(n *CfgNode, nodes []*CfgNode) []*CfgNode {
		if n == nil {
			return nodes
		}
		nodes = append(nodes, n)
		if n == c || !n.BeginEnd() {
			for _, ch := range n.CfgChildren {
				nodes = preord(ch, nodes)
			}
		}
		return nodes
	}
	nodes := make([]*CfgNode, 0)
	return preord(c, nodes)
}

func (c *CfgNode) PostOrder() []*CfgNode {
	var postord func(c *CfgNode, nodes []*CfgNode) []*CfgNode
	postord = func(c *CfgNode, nodes []*CfgNode) []*CfgNode {
		if c == nil {
			return nodes
		}
		for _, ch := range c.CfgChildren {
			nodes = postord(ch, nodes)
		}
		nodes = append(nodes, c)
		return nodes
	}
	nodes := make([]*CfgNode, 0)
	return postord(c, nodes)
}

func (c *CfgNode) PostOrderFilterBE() []*CfgNode {
	var postord func(n *CfgNode, nodes []*CfgNode) []*CfgNode
	postord = func(n *CfgNode, nodes []*CfgNode) []*CfgNode {
		if n == nil {
			return nodes
		}
		if n == c || !n.BeginEnd() {
			for _, ch := range n.CfgChildren {
				nodes = postord(ch, nodes)
			}
		}
		nodes = append(nodes, n)
		return nodes
	}
	nodes := make([]*CfgNode, 0)
	return postord(c, nodes)
}

func (c *CfgNode) Changed() bool {
	return c.Added() || c.Deleted() || c.Updated()
}

func (c *CfgNode) commitaction() string {
	updated := func() bool {
		// Preserve old behaviour of the COMMIT_ACTION that
		// many scripts rely on. The original Updated()
		// function returned false for anything other than
		// Leaf and LeafList.
		switch c.Schema().(type) {
		case schema.Leaf, schema.LeafList:
			return c.Updated()
		}
		return false
	}
	switch {
	case c.Deleted():
		return "DELETE"
	case c.Added() || updated():
		return "SET"
	default:
		return "ACTIVE"
	}
}

func (c *CfgNode) ExecActs(acts []string, action string) ([]*exec.Output, []error, bool) {
	outs := make([]*exec.Output, 0)
	errs := make([]error, 0)
	var t time.Time
	for _, act := range acts {
		if act == "" {
			continue
		}

		isDeferred := c.deferred && action != "commit" && action != "register-defer"
		runDeferred := c.ctx.RunDeferred()
		if runDeferred != isDeferred {
			if c.ctx.Debug() {
				fmt.Println("ignoring", action, c.Path, ":", act)
			}
			continue // Skip it
		}

		if c.ctx.Debug() {
			fmt.Println(action, c.Path, ":", act, "(start)")
			t = time.Now()
		}
		//register-defer is not known to the old config code
		//just pretend it is a begin action
		if action == "register-defer" {
			action = "begin"
		}
		out, err := exec.Exec(exec.Env(c.ctx.Sid(), c.Path, action, c.commitaction()), c.Path, act)
		if err != nil {
			if c.ctx.Debug() {
				fmt.Println("FAILED:", action, c.Path, ":", act)
			}
			errs = append(errs, err)
			if out == nil {
				// Error last as if it contains a newline, reminaing output
				// is seemingly lost.
				c.ctx.LogError(
					fmt.Sprintf(
						"FAILED: %s %s %s: Error with no output. err: %s",
						action, c.Path, act, err))
			}
		}
		if out != nil {
			outs = append(outs, out)
		}
		if c.ctx.Debug() {
			fmt.Println(action, c.Path, ":", act, "- took", time.Since(t),
				"(end)")
		}
	}
	return outs, errs, len(errs) == 0
}

func (c *CfgNode) ExecCreate() ([]*exec.Output, []error, bool) {
	action := "create"
	acts := c.Schema().ConfigdExt().Create
	if len(acts) == 0 {
		acts = c.Schema().ConfigdExt().Update
		action = "update"
	}
	return c.ExecActs(acts, action)
}

func (c *CfgNode) ExecValidate() ([]*exec.Output, []error, bool) {
	acts := c.Schema().ConfigdExt().Validate
	return c.ExecActs(acts, "commit")
}

func (c *CfgNode) ExecUpdate() ([]*exec.Output, []error, bool) {
	acts := c.Schema().ConfigdExt().Update
	return c.ExecActs(acts, "update")
}

func (c *CfgNode) ExecDelete() ([]*exec.Output, []error, bool) {
	acts := c.Schema().ConfigdExt().Delete
	return c.ExecActs(acts, "delete")
}

func (c *CfgNode) ExecBegin() ([]*exec.Output, []error, bool) {
	acts := c.Schema().ConfigdExt().Begin
	return c.ExecActs(acts, "begin")
}

func (c *CfgNode) ExecEnd() ([]*exec.Output, []error, bool) {
	acts := c.Schema().ConfigdExt().End
	return c.ExecActs(acts, "end")
}

func (c *CfgNode) ExecRegisterDefer() ([]*exec.Output, []error, bool) {
	var acts []string
	register := c.Schema().ConfigdExt().DeferActions
	if register != "" {
		acts = []string{register}
	}
	return c.ExecActs(acts, "register-defer")
}

func (c *CfgNode) UpdateLeaf() ([]*exec.Output, []error, bool) {
	var ok bool
	outs, errs := make([]*exec.Output, 0), make([]error, 0)
	fns := []exec.ExecFunc{c.ExecRegisterDefer, c.ExecBegin, c.ExecCreate, c.ExecEnd}
	if _, ok := c.Schema().Type().(schema.Empty); ok && c.Added() {
		for _, fn := range fns {
			outs, errs, ok = exec.AppendOutput(fn, outs, errs)
			if !ok {
				return outs, errs, ok
			}
		}
		return outs, errs, true
	}
	for _, v := range c.GetAddedValues() {
		c.Path = append(c.Path, v)
		for _, fn := range fns {
			outs, errs, ok = exec.AppendOutput(fn, outs, errs)
			if !ok {
				c.Path = c.Path[:len(c.Path)-1]
				return outs, errs, ok
			}

		}
		c.Path = c.Path[:len(c.Path)-1]
	}
	return outs, errs, true
}

func (c *CfgNode) UpdateLeafList() ([]*exec.Output, []error, bool) {
	var ok bool
	outs, errs := make([]*exec.Output, 0), make([]error, 0)
	delfns := []exec.ExecFunc{c.ExecBegin, c.ExecDelete, c.ExecEnd}
	for _, v := range c.GetDeletedValues() {
		c.Path = append(c.Path, v)
		for _, fn := range delfns {
			outs, errs, ok = exec.AppendOutput(fn, outs, errs)
			if !ok {
				c.Path = c.Path[:len(c.Path)-1]
				return outs, errs, ok
			}

		}
		c.Path = c.Path[:len(c.Path)-1]
	}
	addfns := []exec.ExecFunc{c.ExecRegisterDefer, c.ExecBegin, c.ExecCreate, c.ExecEnd}
	for _, v := range c.GetAddedValues() {
		c.Path = append(c.Path, v)
		for _, fn := range addfns {
			outs, errs, ok = exec.AppendOutput(fn, outs, errs)
			if !ok {
				c.Path = c.Path[:len(c.Path)-1]
				return outs, errs, ok
			}

		}
		c.Path = c.Path[:len(c.Path)-1]
	}
	return outs, errs, true
}

func (c *CfgNode) updateChildren() ([]*exec.Output, []error, bool) {
	var ok bool
	outs, errs := make([]*exec.Output, 0), make([]error, 0)
	for _, n := range c.PreOrderFilterBE() {
		switch {
		case n.IsLeaf():
			outs, errs, ok = exec.AppendOutput(n.UpdateLeaf, outs, errs)
			if !ok {
				return outs, errs, ok
			}
		case n.IsLeafList():
			outs, errs, ok = exec.AppendOutput(n.UpdateLeaf, outs, errs)
			if !ok {
				return outs, errs, ok
			}
		case n.Deleted():
			continue
		case n != c && n.BeginEnd():
			outs, errs, ok = exec.AppendOutput(n.Update, outs, errs)
			if !ok {
				return outs, errs, ok
			}
		case !n.Changed():
			continue
		case n.IsList() || n.IsLeafValue():
			continue
		default:
			if n.Added() {
				outs, errs, ok = exec.AppendOutput(n.ExecCreate, outs, errs)
				if !ok {
					return outs, errs, ok
				}
			} else {
				outs, errs, ok = exec.AppendOutput(n.ExecUpdate, outs, errs)
				if !ok {
					return outs, errs, ok
				}
			}
		}
	}
	return outs, errs, true
}

func (c *CfgNode) UpdateList() ([]*exec.Output, []error, bool) {
	var ok bool
	outs, errs := make([]*exec.Output, 0), make([]error, 0)
	for _, n := range c.CfgChildren {
		outs, errs, ok = exec.AppendOutput(n.Update, outs, errs)
		if !ok {
			return outs, errs, ok
		}
	}
	return outs, errs, true
}

func (c *CfgNode) Update() ([]*exec.Output, []error, bool) {
	var ok bool
	outs, errs := make([]*exec.Output, 0), make([]error, 0)
	if !c.SubChanged {
		return outs, errs, true
	}
	switch c.Schema().(type) {
	case schema.Leaf:
		return c.UpdateLeaf()
	case schema.LeafList:
		return c.UpdateLeafList()
	case schema.List:
		return c.UpdateList()
	}

	fns := []exec.ExecFunc{c.ExecRegisterDefer, c.ExecBegin, c.deleteChildren, c.updateChildren, c.ExecEnd}
	for _, fn := range fns {
		outs, errs, ok = exec.AppendOutput(fn, outs, errs)
		if !ok {
			return outs, errs, ok
		}
	}

	return outs, errs, true
}

func (c *CfgNode) DeleteLeaf() ([]*exec.Output, []error, bool) {
	var ok bool
	outs, errs := make([]*exec.Output, 0), make([]error, 0)
	fns := []exec.ExecFunc{c.ExecRegisterDefer, c.ExecBegin, c.ExecDelete, c.ExecEnd}
	if _, ok := c.Schema().Type().(schema.Empty); ok && c.Deleted() {
		for _, fn := range fns {
			outs, errs, ok = exec.AppendOutput(fn, outs, errs)
			if !ok {
				return outs, errs, ok
			}
		}
		return outs, errs, true
	}
	for _, v := range c.GetDeletedValues() {
		for _, fn := range fns {
			c.Path = append(c.Path, v)
			outs, errs, ok = exec.AppendOutput(fn, outs, errs)
			if !ok {
				c.Path = c.Path[:len(c.Path)-1]
				return outs, errs, ok
			}
			c.Path = c.Path[:len(c.Path)-1]
		}
	}
	return outs, errs, true
}

func (c *CfgNode) deleteChildren() ([]*exec.Output, []error, bool) {
	var ok bool
	outs, errs := make([]*exec.Output, 0), make([]error, 0)
	for _, n := range c.PostOrderFilterBE() {
		switch {
		case n.IsLeaf():
			fallthrough
		case n.IsLeafList():
			outs, errs, ok = exec.AppendOutput(n.DeleteLeaf, outs, errs)
			if !ok {
				return outs, errs, ok
			}
		case !n.Deleted():
			continue
		case n != c && n.BeginEnd():
			outs, errs, ok = exec.AppendOutput(n.Delete, outs, errs)
			if !ok {
				return outs, errs, ok
			}
		case n.IsList() || n.IsLeafValue():
			continue
		default:
			outs, errs, ok = exec.AppendOutput(n.ExecDelete, outs, errs)
			if !ok {
				return outs, errs, ok
			}
		}
	}
	return outs, errs, true
}

func (c *CfgNode) DeleteList() ([]*exec.Output, []error, bool) {
	var ok bool
	outs, errs := make([]*exec.Output, 0), make([]error, 0)
	for _, n := range c.CfgChildren {
		outs, errs, ok = exec.AppendOutput(n.Delete, outs, errs)
		if !ok {
			return outs, errs, ok
		}
	}
	return outs, errs, true
}

func (c *CfgNode) Delete() ([]*exec.Output, []error, bool) {
	var ok bool
	outs, errs := make([]*exec.Output, 0), make([]error, 0)
	if !c.SubChanged {
		return outs, errs, true
	}
	switch c.Schema().(type) {
	case schema.Leaf, schema.LeafList:
		return c.DeleteLeaf()
	case schema.List:
		return c.DeleteList()
	}

	fns := []exec.ExecFunc{c.ExecRegisterDefer, c.ExecBegin, c.deleteChildren, c.ExecEnd}
	for _, fn := range fns {
		outs, errs, ok = exec.AppendOutput(fn, outs, errs)
		if !ok {
			return outs, errs, ok
		}
	}

	return outs, errs, true
}

func (c *CfgNode) ValidateExec() ([]*exec.Output, []error, bool) {
	var nWorker = runtime.NumCPU()
	validateWork := make(chan validateJob, nWorker*10)
	responses := make(chan validateResponse, nWorker)
	var wg sync.WaitGroup
	var ok bool
	outs, errs := make([]*exec.Output, 0), make([]error, 0)

	if c.ctx.Debug() {
		fmt.Printf("Starting %d validation workers\n", nWorker)
	}
	for i := 0; i < nWorker; i++ {
		go func(id int) {
			for {
				req, ok := <-validateWork
				if !ok {
					return
				}
				if req.node.ctx.Debug() {
					fmt.Printf("Worker %d running validate for node %s\n",
						id, req.node.Path)
				}
				outs, errs, success := req.job()
				req.resp <- validateResponse{
					outs:    outs,
					errs:    errs,
					success: success,
				}
				req.wg.Done()
			}
		}(i)
	}

	//start jobs
	vn := validateNode{
		wg:   &wg,
		node: c,
		req:  validateWork,
		resp: responses,
	}

	wg.Add(1)
	go func() {
		vn.validateParallelExec()
		wg.Done()
	}()
	go func() {
		wg.Wait()
		close(responses)
	}()

	var done bool
	for !done {
		select {
		case resp, cont := <-responses:
			outs = append(outs, resp.outs...)
			errs = append(errs, resp.errs...)
			ok = ok && resp.success
			done = !cont
		}
	}
	close(validateWork)
	sort.Sort(outsByPath(outs))
	sort.Sort(errsByPath(errs))
	return outs, errs, ok

}

func (c *CfgNode) Validate() ([]*exec.Output, []error, bool) {
	startTime := time.Now()
	outs, errs, _ := c.ValidateExec()
	c.ctx.LogCommitTime("Validation scripts", startTime)

	validateSchemaFn := func() ([]*exec.Output, []error, bool) {
		// We ignore c.ctx.Debug() for schema validation as this is very
		// noisy, to the extent that the log is overwhelmed and we lose
		// other valuable information.  If users need XPATH machine debug,
		// they should use XYANG off-box / send us the config on which to
		// run XYANG.
		return schema.ValidateSchemaWithLog(c.Schema(), c.Node, false,
			c.ctx.LogCommitTime)
	}
	outs, errs, _ = exec.AppendOutput(validateSchemaFn, outs, errs)

	return outs, errs, len(errs) == 0
}

func (c *CfgNode) DeleteChild(ch *CfgNode) {
	var idx int
	idx = -1
	for i, child := range c.CfgChildren {
		if child.Data().Name() == ch.Data().Name() {
			idx = i
			break
		}
	}
	if idx >= 0 {
		c.CfgChildren = append(c.CfgChildren[:idx], c.CfgChildren[idx+1:]...)
	}
}

func (c *CfgNode) SetSubtreeChanged() {
	if c == nil || c.SubChanged {
		return
	}
	c.SubChanged = true
	c.Parent.SetSubtreeChanged()
}

// Ignore interfaces are for detecting descendants added to the commit
// tree. These descendants are added to restore running action scripts
// on list key nodes which existing users depend on.
//
// Yes, this is a cludge.
func (c *CfgNode) IsIgnore() bool {
	return c.ignore
}

func (c *CfgNode) SetIgnore(i bool) {
	c.ignore = i
}

type validateJob struct {
	wg   *sync.WaitGroup
	node *CfgNode
	job  func() ([]*exec.Output, []error, bool)
	resp chan validateResponse
}

type validateResponse struct {
	outs    []*exec.Output
	errs    []error
	success bool
}

type validateNode struct {
	wg   *sync.WaitGroup
	node *CfgNode
	req  chan validateJob
	resp chan validateResponse
}

func (n *validateNode) enqueueJob() {
	acts := n.node.Schema().ConfigdExt().Validate
	if len(acts) == 0 {
		return
	}

	if n.node.ctx.Debug() {
		fmt.Println("enqueuing validation job for", n.node.Path)
	}

	n.wg.Add(1)
	n.req <- validateJob{
		job: func() ([]*exec.Output, []error, bool) {
			return n.node.ExecActs(acts, "commit")
		},
		resp: n.resp,
		wg:   n.wg,
		node: n.node,
	}
}

func (n *validateNode) validateParallelLeafExec() {
	if _, ok := n.node.Schema().Type().(schema.Empty); ok {
		n.enqueueJob()
		return
	}

	for _, v := range n.node.GetValues() {
		cpy := *n.node
		cpy.Path = pathutil.CopyAppend(cpy.Path, v)
		validateCpy := validateNode{
			wg:   n.wg,
			node: &cpy,
			req:  n.req,
			resp: n.resp,
		}
		validateCpy.enqueueJob()
	}
}

func (n *validateNode) enqueueChildJobs(children []*CfgNode) {
	for _, child := range children {
		validateChild := validateNode{
			wg:   n.wg,
			node: child,
			req:  n.req,
			resp: n.resp,
		}
		validateChild.validateParallelExec()
	}
}

func (n *validateNode) validateParallelExec() {
	switch n.node.Schema().(type) {
	case schema.Leaf, schema.LeafList:
		n.validateParallelLeafExec()
	case schema.List:
		n.enqueueChildJobs(n.node.CfgChildren)
	default:
		n.enqueueJob()
		n.enqueueChildJobs(n.node.CfgChildren)
	}
}

type outsByPath []*exec.Output

func (s outsByPath) Len() int      { return len(s) }
func (s outsByPath) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s outsByPath) Less(i, j int) bool {
	return pathutil.Pathstr(s[i].Path) < pathutil.Pathstr(s[j].Path)
}

type errsByPath []error

func (s errsByPath) Len() int      { return len(s) }
func (s errsByPath) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s errsByPath) Less(i, j int) bool {
	// If we have Path comparable errors then compare them
	// otherwise return an arbitrary order
	ei, iok := s[i].(*mgmterror.ExecError)
	ej, jok := s[j].(*mgmterror.ExecError)
	if iok && jok {
		return ei.Path < ej.Path
	}
	return !iok
}
