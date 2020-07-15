// Copyright (c) 2017-2020, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2016-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/danos/yang/parse"
)

type ConfigdExt struct {
	Help           string
	Validate       []string
	Syntax         []string
	Priority       uint
	Normalize      string
	Allowed        string
	Begin          []string
	End            []string
	Create         []string
	Update         []string
	Delete         []string
	Subst          []string
	Secret         bool
	DeferActions   string
	Must           string
	PatternHelp    []string
	CallRpc        string
	GetState       []string
	OpdHelp        string
	OpdAllowed     string
	OpdPatternHelp []string
}

var emptyExt = &ConfigdExt{}

func elementOf(e string, list []string) bool {
	for _, le := range list {
		if e == le {
			return true
		}
	}
	return false
}

func mergeStringList(base, new []string) []string {
	switch {
	case len(base) == 0:
		if len(new) == 0 {
			return nil
		}
		out := make([]string, len(new))
		copy(out, new)
		return out
	case len(new) == 0:
		out := make([]string, len(base))
		copy(out, base)
		return out
	default:
		sz := len(base)
		for _, n := range new {
			if elementOf(n, base) {
				continue
			}
			sz++
		}
		out := make([]string, len(base), sz)
		copy(out, base)
		for _, n := range new {
			if elementOf(n, base) {
				continue
			}
			out = append(out, n)
		}
		return out
	}
}

func (ext *ConfigdExt) getHelp() string {
	if ext == nil {
		return ""
	}
	if ext.Help != "" {
		return ext.Help
	}
	if ext.OpdHelp != "" {
		return ext.OpdHelp
	}
	return ""
}

// Get the help text for a node, if absent, return
// <No help text available>
func (ext *ConfigdExt) GetHelp() string {
	const nohelp = "<No help text available>"
	if h := ext.getHelp(); h != "" {
		return h
	}
	return nohelp
}

// Get the help text as is, suitable for a type node
func (ext *ConfigdExt) GetTypeHelp() string {
	return ext.getHelp()
}

func (ext *ConfigdExt) Override(in *ConfigdExt) *ConfigdExt {
	if in == nil || in == emptyExt {
		return ext
	}

	if ext == emptyExt {
		ext = &ConfigdExt{}
	}
	if in.Help != "" {
		ext.Help = in.Help
	}
	if len(in.Syntax) > 0 {
		ext.Syntax = in.Syntax
	}
	if len(in.PatternHelp) > 0 {
		ext.PatternHelp = in.PatternHelp
	}
	if len(in.OpdPatternHelp) > 0 {
		ext.OpdPatternHelp = in.OpdPatternHelp
	}
	if len(in.Create) > 0 {
		ext.Create = in.Create
	}
	if len(in.Update) > 0 {
		ext.Update = in.Update
	}
	if len(in.Delete) > 0 {
		ext.Delete = in.Delete
	}
	if len(in.Subst) > 0 {
		ext.Subst = in.Subst
	}
	if len(in.Validate) > 0 {
		ext.Validate = in.Validate
	}
	if len(in.GetState) > 0 {
		ext.GetState = in.GetState
	}
	if in.Normalize != "" {
		ext.Normalize = in.Normalize
	}
	if in.Allowed != "" {
		ext.Allowed = in.Allowed
	}
	if len(in.Begin) > 0 {
		ext.Begin = in.Begin
	}
	if len(in.End) > 0 {
		ext.End = in.End
	}
	if in.Priority != 0 {
		ext.Priority = in.Priority
	}
	if in.Secret == true {
		ext.Secret = true
	}
	if in.CallRpc != "" {
		ext.CallRpc = in.CallRpc
	}
	if in.DeferActions != "" {
		ext.DeferActions = in.DeferActions
	}
	if in.Must != "" {
		ext.Must = in.Must
	}
	if in.OpdHelp != "" {
		ext.OpdHelp = in.OpdHelp
	}
	if in.OpdAllowed != "" {
		ext.OpdAllowed = in.OpdAllowed
	}
	return ext
}

//Types require merging of base data into new type
func (ext *ConfigdExt) Merge(in *ConfigdExt) *ConfigdExt {
	if in == nil || in == emptyExt {
		return ext
	}
	if ext == emptyExt {
		ext = &ConfigdExt{}
	}
	if in.Help != "" {
		ext.Help = in.Help
	}
	ext.Syntax = mergeStringList(ext.Syntax, in.Syntax)
	ext.PatternHelp = mergeStringList(ext.PatternHelp, in.PatternHelp)
	ext.Create = mergeStringList(ext.Create, in.Create)
	ext.Update = mergeStringList(ext.Update, in.Update)
	ext.Delete = mergeStringList(ext.Delete, in.Delete)
	ext.Subst = mergeStringList(ext.Subst, in.Subst)
	ext.Validate = mergeStringList(ext.Validate, in.Validate)
	ext.GetState = mergeStringList(ext.GetState, in.GetState)
	if in.Normalize != "" {
		ext.Normalize = in.Normalize
	}
	if in.Allowed != "" {
		ext.Allowed = in.Allowed
	}
	ext.Begin = mergeStringList(ext.Begin, in.Begin)
	ext.End = mergeStringList(ext.End, in.End)
	if in.Priority != 0 {
		ext.Priority = in.Priority
	}
	if in.Secret == true {
		ext.Secret = true
	}
	if in.CallRpc != "" {
		ext.CallRpc = in.CallRpc
	}
	if in.DeferActions != "" {
		ext.DeferActions = in.DeferActions
	}
	if in.Must != "" {
		ext.Must = in.Must
	}
	if in.OpdHelp != "" {
		ext.OpdHelp = in.OpdHelp
	}
	if in.OpdAllowed != "" {
		ext.OpdAllowed = in.OpdAllowed
	}
	ext.OpdPatternHelp = mergeStringList(ext.OpdPatternHelp, in.OpdPatternHelp)
	return ext
}

func copyStringList(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func (ext *ConfigdExt) Copy() *ConfigdExt {
	if ext == emptyExt {
		return ext
	}
	return &ConfigdExt{
		Help:           ext.Help,
		Validate:       copyStringList(ext.Validate),
		Syntax:         copyStringList(ext.Syntax),
		Priority:       ext.Priority,
		Normalize:      ext.Normalize,
		Allowed:        ext.Allowed,
		Begin:          copyStringList(ext.Begin),
		End:            copyStringList(ext.End),
		Create:         copyStringList(ext.Create),
		Delete:         copyStringList(ext.Delete),
		Update:         copyStringList(ext.Update),
		Subst:          copyStringList(ext.Subst),
		Secret:         ext.Secret,
		DeferActions:   ext.DeferActions,
		Must:           ext.Must,
		PatternHelp:    copyStringList(ext.PatternHelp),
		CallRpc:        ext.CallRpc,
		GetState:       copyStringList(ext.GetState),
		OpdAllowed:     ext.OpdAllowed,
		OpdHelp:        ext.OpdHelp,
		OpdPatternHelp: copyStringList(ext.OpdPatternHelp),
	}
}

type ValidateCtx struct {
	Noexec                bool
	CurPath               []string
	Path                  string
	Sid                   string
	St                    ModelSet
	IncompletePathIsValid bool
}

func (v ValidateCtx) AllowIncompletePaths() bool {
	return v.IncompletePathIsValid
}

func (v ValidateCtx) ErrorHelpText() []string {
	p := make([]string, 0)

	if v.St != nil {
		sn := v.St.PathDescendant(v.CurPath[:len(v.CurPath)-1])
		if sn != nil {
			hm := sn.Node.HelpMap()
			for k, _ := range hm {
				p = append(p, k)
			}
		}
	}
	sort.Strings(p)
	return p
}

// We need the extras to be separate since we can't include the yang.Node interface twice.
type hasExtensions interface {
	ConfigdExt() *ConfigdExt
	checkSyntax(ValidateCtx) error
}

type extensions struct {
	ext *ConfigdExt
}

func extUintByTypeOne(n parse.Node, typ parse.NodeType) uint {
	if ch := n.ChildByType(typ); ch != nil {
		return ch.ArgUint()
	}
	return 0
}

func extBoolByTypeOne(n parse.Node, typ parse.NodeType) bool {
	if ch := n.ChildByType(typ); ch != nil {
		return ch.ArgBool()
	}
	return false
}

func extByTypeOne(n parse.Node, typ parse.NodeType) string {
	if ch := n.ChildByType(typ); ch != nil {
		return ch.ArgString()
	}
	return ""
}

func extByTypeMany(n parse.Node, typ parse.NodeType) []string {
	children := n.ChildrenByType(typ)
	if len(children) == 0 {
		return nil
	}
	exts := make([]string, len(children))
	for i, v := range children {
		exts[i] = v.ArgString()
	}
	return exts
}

func numExtensions(p parse.Node) int {
	var count int
	for _, ch := range p.Children() {
		ty := ch.Type()
		if ty.IsConfigdNode() || ty.IsOpdExtension() {
			count++
		}
	}
	return count
}

func parseExtensions(p parse.Node) *ConfigdExt {
	if numExtensions(p) == 0 {
		return emptyExt
	}
	return &ConfigdExt{
		Help:           extByTypeOne(p, parse.NodeConfigdHelp),
		Validate:       extByTypeMany(p, parse.NodeConfigdValidate),
		Syntax:         extByTypeMany(p, parse.NodeConfigdSyntax),
		Priority:       extUintByTypeOne(p, parse.NodeConfigdPriority),
		Normalize:      extByTypeOne(p, parse.NodeConfigdNormalize),
		Allowed:        extByTypeOne(p, parse.NodeConfigdAllowed),
		Begin:          extByTypeMany(p, parse.NodeConfigdBegin),
		End:            extByTypeMany(p, parse.NodeConfigdEnd),
		Create:         extByTypeMany(p, parse.NodeConfigdCreate),
		Delete:         extByTypeMany(p, parse.NodeConfigdDelete),
		Update:         extByTypeMany(p, parse.NodeConfigdUpdate),
		Subst:          extByTypeMany(p, parse.NodeConfigdSubst),
		Secret:         extBoolByTypeOne(p, parse.NodeConfigdSecret),
		PatternHelp:    extByTypeMany(p, parse.NodeConfigdPHelp),
		CallRpc:        extByTypeOne(p, parse.NodeConfigdCallRpc),
		GetState:       extByTypeMany(p, parse.NodeConfigdGetState),
		DeferActions:   extByTypeOne(p, parse.NodeConfigdDeferActions),
		Must:           extByTypeOne(p, parse.NodeConfigdMust),
		OpdHelp:        extByTypeOne(p, parse.NodeOpdHelp),
		OpdAllowed:     extByTypeOne(p, parse.NodeOpdAllowed),
		OpdPatternHelp: extByTypeMany(p, parse.NodeOpdPatternHelp),
	}
}

func extensionsContainAny(ext *ConfigdExt, match ...string) bool {

	for _, v := range match {
		switch v {
		case "create":
			if len(ext.Create) > 0 {
				return true
			}
		case "update":
			if len(ext.Update) > 0 {
				return true
			}
		case "delete":
			if len(ext.Delete) > 0 {
				return true
			}
		case "begin":
			if len(ext.Begin) > 0 {
				return true
			}
		case "end":
			if len(ext.End) > 0 {
				return true
			}
		default:
			panic(fmt.Errorf("Unrecognised extension"))
		}
	}
	return false
}

var emptyExtend = &extensions{emptyExt}

func newExtend(ext *ConfigdExt) *extensions {
	if ext == nil || ext == emptyExt {
		return emptyExtend
	}
	return &extensions{ext}
}

func (e *extensions) ConfigdExt() *ConfigdExt {
	return e.ext
}

func pathstr(path []string) string {
	var str string
	for _, v := range path {
		str += "/" + strings.Replace(url.QueryEscape(v), "+", "%20", -1)
	}
	return str
}

func (e *extensions) checkSyntax(ctx ValidateCtx) error {
	if ctx.Noexec {
		return nil
	}

	path := pathstr(ctx.CurPath)
	for _, syn := range e.ext.Syntax {
		_, err := execCmd(ctx.Sid, path, syn)
		if err != nil {
			return err
		}
	}
	return nil
}
