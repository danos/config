// Copyright (c) 2019-2020, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/danos/mgmterror"
	"github.com/danos/yang/parse"
	yang "github.com/danos/yang/schema"
)

type hasRangeErrMsg interface {
	getRangeErrMsg() string
}

func parseRangeErrMsg(p parse.Node, base yang.Type) string {

	if rng := p.ChildByType(parse.NodeRange); rng != nil {
		// A local range will override the error message
		return rng.Cmsg()

	} else if base != nil {
		// If parent has a configd:error-message then we use it
		if b, ok := base.(hasRangeErrMsg); ok {
			return b.getRangeErrMsg()
		}
	}

	return ""
}

func parseLengthErrMsg(p parse.Node, base yang.Type) string {

	if len := p.ChildByType(parse.NodeLength); len != nil {
		// A local range will override the error message
		return len.Cmsg()

	} else if base != nil {
		// If parent has a configd:error-message then we use it
		if b, ok := base.(*ystring); ok {
			return b.lengthErrMsg
		}
	}

	return ""
}

func parsePatternErrMsgs(p parse.Node, base yang.Type) map[string]string {

	errMsgs := make(map[string]string)

	// Inherit the parent pattern errors
	if b, ok := base.(*ystring); ok {
		for k, v := range b.patternErrMsgs {
			errMsgs[k] = v
		}
	}

	// Add any local pattern errors
	pats := p.ChildrenByType(parse.NodePattern)
	for _, p := range pats {
		msg := p.Cmsg()
		if msg != "" {
			errMsgs[p.Argument().String()] = msg
		}
	}

	return errMsgs
}

type Type interface {
	yang.Type
	hasExtensions
}

type Binary interface {
	yang.Binary
	hasExtensions
}

type binary struct {
	yang.Binary
	*extensions
}

type Boolean interface {
	yang.Boolean
	hasExtensions
}

type boolean struct {
	yang.Boolean
	*extensions
}

type Decimal64 interface {
	yang.Decimal64
	hasExtensions
	hasRangeErrMsg
}

type decimal64 struct {
	yang.Decimal64
	*extensions
	rangeErrMsg string
}

// Compile time check that the concrete type meets the interface
var _ Decimal64 = (*decimal64)(nil)

func (d *decimal64) getRangeErrMsg() string {
	return d.rangeErrMsg
}

func newDecimal64(
	p parse.Node, base yang.Type, y yang.Decimal64, ext *extensions,
) (yang.Type, error) {

	rangeErrMsg := parseRangeErrMsg(p, base)
	return &decimal64{y, ext, rangeErrMsg}, nil
}

func (d *decimal64) Validate(ictx yang.ValidateCtx, path []string, s string) error {

	err := d.Decimal64.Validate(ictx, path, s)

	if ictx == nil {
		return err
	}
	ctx := ictx.(ValidateCtx)
	ctx.CurPath = path

	if err != nil {
		if d.rangeErrMsg != "" {
			return newRangeError(ctx, append(path, s), d.rangeErrMsg)
		}
		return err
	}
	return d.checkSyntax(ctx)
}

type Empty interface {
	yang.Empty
	hasExtensions
}

type empty struct {
	yang.Empty
	*extensions
}

type Enumeration interface {
	yang.Enumeration
	hasExtensions
	getHelpMap() map[string]string
}

type enumeration struct {
	yang.Enumeration
	*extensions
	helpMap map[string]string
}

// Compile time check that the concrete type meets the interface
var _ Enumeration = (*enumeration)(nil)

func (e *enumeration) getHelpMap() map[string]string {
	helpMap := make(map[string]string)
	localHelp := e.ConfigdExt().GetHelp()

	for val, help := range e.helpMap {
		if help == "" {
			help = localHelp
		}
		helpMap[val] = help
	}

	return helpMap
}

func parseEnumHelpMap(p parse.Node, base yang.Type) map[string]string {
	helpMap := make(map[string]string)

	if base != nil {
		return base.(*enumeration).helpMap

	} else {
		// Enums are only allowed on the base enumeration type
		enums := p.ChildrenByType(parse.NodeEnum)
		for _, en := range enums {
			if en.Status() == "obsolete" {
				continue
			}
			val := en.ArgString()

			// We want the empty string here as it gets overridden
			// in the outermost type
			help := parseExtensions(en).Help
			if help == "" {
				help = parseExtensions(en).OpdHelp
			}
			helpMap[val] = help
		}
	}
	return helpMap
}

func newEnumeration(
	p parse.Node, base yang.Type, y yang.Enumeration, ext *extensions,
) (yang.Type, error) {

	helpMap := parseEnumHelpMap(p, base)
	return &enumeration{y, ext, helpMap}, nil
}

type Identityref interface {
	yang.Identityref
	hasExtensions
	getHelpMap(parentPrefix string) map[string]string
}

type identityref struct {
	yang.Identityref
	*extensions
	helpMap map[string]string
}

// Compile time check that the concrete type meets the interface
var _ Identityref = (*identityref)(nil)

func (i *identityref) getHelpMap(parentPrefix string) map[string]string {
	pfx := parentPrefix + ":"
	helpMap := make(map[string]string)
	localHelp := i.ConfigdExt().GetHelp()

	for val, help := range i.helpMap {
		if help == "" {
			help = localHelp
		}
		val = strings.TrimPrefix(val, pfx)
		helpMap[val] = help
	}

	return helpMap
}

func getIdentities(p parse.Node, helpMap map[string]string) {
	for _, i := range p.ChildrenByType(parse.NodeIdentity) {
		if i.Status() != "obsolete" {
			nm := i.Root().Name() + ":" + i.Name()
			help := parseExtensions(i).Help
			if help == "" {
				help = parseExtensions(i).OpdHelp
			}
			helpMap[nm] = help
		}
		getIdentities(i, helpMap)
	}
}

func parseIdentityHelpMap(p parse.Node, base yang.Type, idef yang.Identityref) map[string]string {
	helpMap := make(map[string]string)

	if base != nil {
		return base.(*identityref).helpMap
	}
	for _, ident := range p.ChildrenByType(parse.NodeIdentity) {
		getIdentities(ident, helpMap)
	}
	return helpMap
}

func newIdentityref(
	p parse.Node, base yang.Type, y yang.Identityref, ext *extensions,
) (yang.Type, error) {

	helpMap := parseIdentityHelpMap(p, base, y)
	return &identityref{y, ext, helpMap}, nil
}

type Integer interface {
	yang.Integer
	hasExtensions
	hasRangeErrMsg
}

type integer struct {
	yang.Integer
	*extensions
	rangeErrMsg string
}

// Compile time check that the concrete type meets the interface
var _ Integer = (*integer)(nil)

func (i *integer) getRangeErrMsg() string {
	return i.rangeErrMsg
}

func newInteger(
	p parse.Node, base yang.Type, y yang.Integer, ext *extensions,
) (yang.Type, error) {

	rangeErrMsg := parseRangeErrMsg(p, base)
	return &integer{y, ext, rangeErrMsg}, nil
}

func (i *integer) Validate(ictx yang.ValidateCtx, path []string, s string) error {

	err := i.Integer.Validate(ictx, path, s)

	if ictx == nil {
		return err
	}
	ctx := ictx.(ValidateCtx)
	ctx.CurPath = path

	if err != nil {
		if i.rangeErrMsg != "" {
			return newRangeError(ctx, append(path, s), i.rangeErrMsg)
		}
		return err
	}

	return i.checkSyntax(ctx)
}

type Uinteger interface {
	yang.Uinteger
	hasExtensions
	hasRangeErrMsg
}

type uinteger struct {
	yang.Uinteger
	*extensions
	rangeErrMsg string
}

// Compile time check that the concrete type meets the interface
var _ Uinteger = (*uinteger)(nil)

func (u *uinteger) getRangeErrMsg() string {
	return u.rangeErrMsg
}

func newUinteger(
	p parse.Node, base yang.Type, y yang.Uinteger, ext *extensions,
) (yang.Type, error) {

	rangeErrMsg := parseRangeErrMsg(p, base)
	return &uinteger{y, ext, rangeErrMsg}, nil
}

func (u *uinteger) Validate(ictx yang.ValidateCtx, path []string, s string) error {

	err := u.Uinteger.Validate(ictx, path, s)

	if ictx == nil {
		return err
	}
	ctx := ictx.(ValidateCtx)
	ctx.CurPath = path

	if err != nil {
		if u.rangeErrMsg != "" {
			return newRangeError(ctx, append(path, s), u.rangeErrMsg)
		}
		return err
	}

	return u.checkSyntax(ctx)
}

type String interface {
	yang.String
	hasExtensions
}

type ystring struct {
	yang.String
	*extensions
	patternErrMsgs map[string]string
	lengthErrMsg   string
}

// Compile time check that the concrete type meets the interface
var _ String = (*ystring)(nil)

func newString(
	p parse.Node, base yang.Type, y yang.String, ext *extensions,
) (yang.Type, error) {

	patternErrMsgs := parsePatternErrMsgs(p, base)
	lengthErrMsg := parseLengthErrMsg(p, base)
	return &ystring{y, ext, patternErrMsgs, lengthErrMsg}, nil
}

func (s *ystring) Validate(ictx yang.ValidateCtx, path []string, value string) error {

	err := s.String.Validate(ictx, path, value)

	if ictx == nil {
		return err
	}
	ctx := ictx.(ValidateCtx)
	ctx.CurPath = path
	if err == nil {
		return s.checkSyntax(ctx)
	}

	switch merr := err.(type) {
	case *mgmterror.InvalidValueApplicationError:
		pattern := merr.Info.FindMgmtErrorTag(mgmterror.VyattaNamespace,
			"pattern")
		if pattern != "" {
			// Override default pattern message if
			// error-message or pattern-help is present.
			message := merr.Info.FindMgmtErrorTag(mgmterror.VyattaNamespace,
				"message")
			if errMsg, ok := s.patternErrMsgs[pattern]; ok && errMsg != "" {
				merr.Message = errorMessage(ctx, errMsg)
			} else if message == "" && len(s.ext.PatternHelp) > 0 {
				merr.Message = errorMessage(ctx,
					fmt.Sprintf("Must match %s", s.ext.PatternHelp[0]))
			}
			break
		}

		length := merr.Info.FindMgmtErrorTag(mgmterror.VyattaNamespace,
			"length")
		if length != "" {
			if s.lengthErrMsg != "" {
				merr.Message = errorMessage(ctx, s.lengthErrMsg)
			}
			break
		}
	}
	return err
}

type Union interface {
	yang.Union
	hasExtensions
}

type union struct {
	yang.Union
	*extensions
}

type InstanceId interface {
	yang.InstanceId
	hasExtensions
}

type instanceId struct {
	yang.InstanceId
	*extensions
}

type Leafref interface {
	yang.Leafref
	hasExtensions
}

type leafref struct {
	yang.Leafref
	*extensions
}

type Bits interface {
	yang.Bits
	hasExtensions
}

type bits struct {
	yang.Bits
	*extensions
}

func mergeExtensions(a, b *ConfigdExt) *ConfigdExt {
	if a == nil {
		return b
	}

	ext := a.Copy()
	ext = ext.Merge(b)
	return (ext)
}

func (*CompilationExtensions) ExtendMust(
	p parse.Node, m parse.Node,
) (string, error) {

	return parseExtensions(m).Must, nil
}

func (*CompilationExtensions) ExtendType(
	p parse.Node, base yang.Type, t yang.Type,
) (yang.Type, error) {

	pext := parseExtensions(p)
	if base != nil {
		pext = mergeExtensions(base.(Type).ConfigdExt(), pext)
	}

	if len(pext.Syntax) > 0 {
		switch t.(type) {
		case yang.Boolean, yang.Empty, yang.Enumeration,
			yang.Bits, yang.Leafref, yang.InstanceId, yang.Identityref:
			return nil, fmt.Errorf("configd:syntax restriction is not valid for this type")
		case yang.Union:
			return nil, fmt.Errorf("cannot restrict configd:syntax of a union type - " +
				"restrictions must be applied to members instead")
		}
	}

	ext := newExtend(pext)
	switch y := t.(type) {
	case yang.Binary:
		return &binary{y, ext}, nil
	case yang.Boolean:
		return &boolean{y, ext}, nil
	case yang.Decimal64:
		return newDecimal64(p, base, y, ext)
	case yang.Empty:
		return &empty{y, ext}, nil
	case yang.Enumeration:
		return newEnumeration(p, base, y, ext)
	case yang.Identityref:
		return newIdentityref(p, base, y, ext)
	case yang.Integer:
		return newInteger(p, base, y, ext)
	case yang.Uinteger:
		return newUinteger(p, base, y, ext)
	case yang.String:
		return newString(p, base, y, ext)
	case yang.Union:
		return &union{y, ext}, nil
	case yang.InstanceId:
		return &instanceId{y, ext}, nil
	case yang.Leafref:
		return &leafref{y, ext}, nil
	case yang.Bits:
		return &bits{y, ext}, nil
	default:
		panic(fmt.Errorf("Unexpected type for extension: %s", reflect.TypeOf(t)))
	}
}
