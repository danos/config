// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2014 by Brocade Communications Systems, Inc.
// All rights reserved.

// Portions Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// SPDX-License-Identifier: MPL-2.0 and BSD-3-Clause

package parse

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
)

type Node struct {
	Pos
	Id       string
	Arg      string
	Children []*Node
	HasArg   bool
}

func printilvl(buf *bytes.Buffer, level int) {
	for i := 0; i < level; i++ {
		buf.WriteByte('\t')
	}
}

func (n *Node) serialize(buf *bytes.Buffer, level int) {
	printilvl(buf, level)
	if n.HasArg {
		buf.WriteString(n.Id + " " + n.Arg)
	} else {
		buf.WriteString(n.Id)
	}
	if len(n.Children) == 0 {
		buf.WriteByte('\n')
		return
	}
	buf.WriteString(" {\n")
	for _, ch := range n.Children {
		ch.serialize(buf, level+1)
	}
	printilvl(buf, level)
	buf.WriteByte('}')
	buf.WriteByte('\n')
}

func (n *Node) String() string {
	var buf bytes.Buffer
	n.serialize(&buf, 0)
	return buf.String()
}

// Tree is the representation of a single parsed template.
type Tree struct {
	Root      *Node // top-level root of the tree.
	ParseName string
	text      string // text parsed to create the template (or its parent)
	lex       *lexer
	token     [3]item // three-token lookahead for parser.
	peekCount int
}

func Parse(name, text string) (*Tree, error) {
	t := New(name)
	t.text = text
	_, err := t.Parse(text)
	return t, err
}

func (t *Tree) String() string {
	return t.text
}

// next returns the next token.
func (t *Tree) next() item {
	if t.peekCount > 0 {
		t.peekCount--
	} else {
		t.token[0] = t.lex.nextItem()
	}
	return t.token[t.peekCount]
}

// backup backs the input stream up one token.
func (t *Tree) backup() {
	t.peekCount++
}

// backup2 backs the input stream up two tokens.
// The zeroth token is already there.
func (t *Tree) backup2(t1 item) {
	t.token[1] = t1
	t.peekCount = 2
}

// backup3 backs the input stream up three tokens
// The zeroth token is already there.
func (t *Tree) backup3(t2, t1 item) { // Reverse order: we're pushing back.
	t.token[1] = t1
	t.token[2] = t2
	t.peekCount = 3
}

// peek returns but does not consume the next token.
func (t *Tree) peek() item {
	if t.peekCount > 0 {
		return t.token[t.peekCount-1]
	}
	t.peekCount = 1
	t.token[0] = t.lex.nextItem()
	return t.token[0]
}

// nextNonSpace returns the next non-space token.
func (t *Tree) nextNonSpace() (token item) {
	for {
		token = t.next()
		if token.typ != itemSep {
			break
		}
	}
	return token
}

// peekNonSpace returns but does not consume the next non-space token.
func (t *Tree) peekNonSpace() (token item) {
	for {
		token = t.next()
		if token.typ != itemSep {
			break
		}
	}
	t.backup()
	return token
}

// Parsing.
// New allocates a new parse tree with the given name.
func New(name string) *Tree {
	return &Tree{ParseName: name}
}

// ErrorContext returns a textual representation of the location of the node in the input text.
func (t *Tree) ErrorContext(n *Node) (location, context string) {
	pos := int(n.Position())
	context = n.String()
	return t.ErrorContextPosition(pos, context)
}
func (t *Tree) ErrorContextPosition(pos int, ctx string) (location, context string) {
	text := t.text[:pos]
	byteNum := strings.LastIndex(text, "\n")
	if byteNum == -1 {
		byteNum = pos // On first line.
	} else {
		byteNum++ // After the newline.
		byteNum = pos - byteNum
	}
	lineNum := 1 + strings.Count(text, "\n")
	context = ctx
	if len(context) > 20 {
		context = fmt.Sprintf("%.20s...", context)
	}
	if ctx == "" {
		return fmt.Sprintf("%s:%d:%d", t.ParseName, lineNum, byteNum), context
	}
	return fmt.Sprintf("%s:%d:%d: %s", t.ParseName, lineNum, byteNum, ctx), context
}

// errorf formats the error and terminates processing.
func (t *Tree) errorf(format string, args ...interface{}) {
	t.Root = nil
	pos := int(t.lex.lastPos)
	text := t.lex.input[:t.lex.lastPos]
	byteNum := strings.LastIndex(text, "\n")
	if byteNum == -1 {
		byteNum = pos // On first line.
	} else {
		byteNum++ // After the newline.
		byteNum = pos - byteNum
	}
	format = fmt.Sprintf("yang: %s:%d:%d: %s", t.ParseName, t.lex.lineNumber(), byteNum, format)
	panic(fmt.Errorf(format, args...))
}

// error terminates processing.
func (t *Tree) error(err error) {
	t.errorf("%s", err)
}

// expect consumes the next token and guarantees it has the required type.
func (t *Tree) expect(expected itemType, context string) item {
	token := t.nextNonSpace()
	if token.typ != expected {
		t.unexpected(token, context)
	}
	return token
}

// expectOneOf consumes the next token and guarantees it has one of the required types.
func (t *Tree) expectOneOf(expected1, expected2 itemType, context string) item {
	token := t.nextNonSpace()
	if token.typ != expected1 && token.typ != expected2 {
		t.unexpected(token, context)
	}
	return token
}

// unexpected complains about the token and terminates processing.
func (t *Tree) unexpected(token item, context string) {
	t.errorf("unexpected %s in %s", token, context)
}

// recover is the handler that turns panics into returns from the top level of Parse.
func (t *Tree) recover(errp *error) {
	e := recover()
	if e != nil {
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
		if t != nil {
			t.stopParse()
		}
		*errp = e.(error)
	}
	return
}

// startParse initializes the parser, using the lexer.
func (t *Tree) startParse(lex *lexer) {
	t.Root = nil
	t.lex = lex
}

// stopParse terminates parsing.
func (t *Tree) stopParse() {
	t.lex = nil
}

func (t *Tree) Parse(text string) (tree *Tree, err error) {
	defer t.recover(&err)
	t.startParse(lex(t.ParseName, text))
	t.text = text
	t.parse()
	t.stopParse()
	return t, nil
}

func (t *Tree) NewNode(id item, arg string, hasarg bool, children []*Node) *Node {
	return &Node{Id: id.val, Arg: arg, HasArg: hasarg, Children: children}
}

/*
<file>       ::= <stmtStar>
<stmtStar>   ::= ""
               | <stmtStar> <stmt>
<stmt>       ::= <id> <arg> <stmtBody>
               | <id> <stmtBody>
<id>         ::= [-[:alnum:]_]+
<arg>        ::= ".*"
               | [^{[:space:]]+
<stmtBody>   ::= EOS
               | '{' <stmtStar> '}
*/

//file:
//	stmtStar
func (t *Tree) parse() {
	t.Root = &Node{Id: "root", HasArg: false, Children: t.stmtStar("root")}
	t.expect(itemEOF, "file")
	return
}

//stmt:
//	identifier argument stmtBody
//|	identifier stmtBody
func (t *Tree) stmt(ctx string) *Node {
	var arg string
	var hasarg bool
	id := t.expect(itemString, ctx)
	i := t.peekNonSpace()
	switch i.typ {
	case itemLeftBrace:
		break
	case itemEOS:
		break
	default:
		arg = t.argument("argument of " + id.val)
		hasarg = true
	}

	body := t.stmtBody("body of " + id.val + " " + arg)
	n := t.NewNode(id, arg, hasarg, body)
	return n
}

//argument:
//	string
func (t *Tree) argument(ctx string) string {
	var i item
	var s string

	i = t.peekNonSpace()
	switch i.typ {
	case itemLeftBrace:
		fallthrough
	case itemString:
		i = t.nextNonSpace()
		s = i.val
		return s
	default:
		t.unexpected(i, ctx)
	}

	return s
}

//stmtBody:
//	EOS
//| '{' stmtStar '}'
func (t *Tree) stmtBody(ctx string) []*Node {
	var out []*Node
	delim := t.expectOneOf(itemEOS, itemLeftBrace, ctx)
	switch delim.typ {
	case itemLeftBrace:
		out = t.stmtStar(ctx)
		t.expect(itemRightBrace, ctx)
	default:
	}

	return out
}

//stmtStar
//	stmtStar stmt
func (t *Tree) stmtStar(ctx string) []*Node {
	//0 stmts
	if i := t.peekNonSpace(); i.typ == itemRightBrace || i.typ == itemEOF {
		return nil
	}

	//1 or more stmts
	out := make([]*Node, 0)
	for n := t.stmt(ctx); n != nil; n = t.stmt(ctx) {
		out = append(out, n)
		if i := t.peekNonSpace(); i.typ == itemRightBrace || i.typ == itemEOF {
			break
		}
	}
	if len(out) == 0 {
		return nil
	} else {
		return out
	}
}
