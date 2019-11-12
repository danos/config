// Copyright (c) 2017-2019, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2014-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package union

import (
	"bytes"
	"encoding/json"

	"github.com/danos/config/data"
	"github.com/danos/config/schema"
	"github.com/danos/mgmterror"
	"github.com/danos/yang/data/encoding"
)

type JSONWriter struct {
	bytes.Buffer
	rfc7951    bool
	moduleName []string
}

func (b *JSONWriter) pushName(n Node) string {
	if b.moduleName == nil {
		b.moduleName = make([]string, 0)
	}
	name := n.Module()
	newname := name
	if len(b.moduleName) > 0 {
		if name == b.moduleName[len(b.moduleName)-1] {
			newname = ""
		}
	}
	b.moduleName = append(b.moduleName, name)

	return newname
}

func (b *JSONWriter) popName() {
	if b.rfc7951 {
		if len(b.moduleName) > 0 {
			b.moduleName = b.moduleName[:len(b.moduleName)-1]
		}
	}
}

func (b *JSONWriter) currentModuleName() string {
	switch len(b.moduleName) {
	case 0:
		return ""
	case 1:
		return b.moduleName[0]
	default:
		if b.moduleName[len(b.moduleName)-1] !=
			b.moduleName[len(b.moduleName)-2] {
			return b.moduleName[len(b.moduleName)-1]
		}
		return ""
	}
}
func (b *JSONWriter) handleModuleName(n Node) {
	if b.rfc7951 {
		b.pushName(n)
		if nm := b.currentModuleName(); nm != "" {
			b.WriteString(nm)
			b.WriteString(":")
		}
	}
}

func (b *JSONWriter) writeValue(n Node) {
	switch t := n.GetSchema().Type().(type) {
	case schema.Integer:
		if b.rfc7951 && t.BitWidth() > 32 {
			// 64-bit integers encoded as JSON string
			buf, _ := json.Marshal(n.Name())
			b.Write(buf)
		} else {
			// Write the raw value out as a native JSON type
			b.WriteString(n.Name())
		}
	case schema.Uinteger:
		if b.rfc7951 && t.BitWidth() > 32 {
			// 64-bit integers encoded as JSON string
			buf, _ := json.Marshal(n.Name())
			b.Write(buf)
		} else {
			// Write the raw value out as a native JSON type
			b.WriteString(n.Name())
		}
	case schema.Boolean:
		// Write the raw value out as a native JSON type
		b.WriteString(n.Name())

	default:
		// Treat as a string, with appropriate escaping and quotes.
		// Note that Decimal64 is our variable precision floating point and
		// must be encoded as a string, as per draft-ietf-netmod-yang-json-05

		// json.Marshal won't err on a string
		buf, _ := json.Marshal(n.Name())
		b.Write(buf)
	}
}

func (b *JSONWriter) BeginContainer(n *Container, empty bool, level int) {
	b.WriteByte('"')
	b.handleModuleName(n)
	b.WriteString(n.Name())
	b.WriteString("\":{")
}

func (b *JSONWriter) EndContainer(n *Container, empty bool, level int) {
	b.WriteByte('}')
	b.popName()
}

func (b *JSONWriter) BeginList(n *List, empty bool, level int) {
	b.WriteByte('"')
	b.handleModuleName(n)
	b.WriteString(n.Name())
	b.WriteString("\":[")
}

func (b *JSONWriter) BeginListEntry(n *ListEntry, empty bool, level int) {
	sch := n.Schema
	b.WriteString("{\"")
	b.WriteString(sch.Keys()[0])
	b.WriteString("\":")
	b.writeValue(n)
	if !empty {
		b.WriteByte(',')
	}
}

func (b *JSONWriter) EndListEntry(n *ListEntry, empty bool, level int) {
	b.WriteByte('}')
}

func (b *JSONWriter) EndList(n *List, empty bool, level int) {
	b.WriteByte(']')
	b.popName()
}

func (b *JSONWriter) BeginLeaf(n *Leaf, empty bool, level int, hideSecrets bool) {
	b.WriteByte('"')
	b.handleModuleName(n)
	b.WriteString(n.Name())
	b.WriteString("\":")
}

func (b *JSONWriter) writeNullLeafValue(n *Leaf) {
	if b.rfc7951 {
		if _, ok := n.GetSchema().Type().(schema.Empty); ok {
			b.WriteString("[null]")
			return
		}
	}
	b.WriteString("null")

}
func (b *JSONWriter) WriteLeafValue(
	n *Leaf,
	empty bool,
	level int,
	hideSecrets bool,
) {
	vals := n.SortedChildren()
	switch len(vals) {
	case 0:
		b.writeNullLeafValue(n)
	default:
		if hideSecrets && n.GetSchema().ConfigdExt().Secret {
			b.WriteString(quote("********"))
		} else {
			b.writeValue(vals[0])
		}
	}
}

func (b *JSONWriter) EndLeaf(n *Leaf, empty bool, level int) {
	b.popName()
}

func (b *JSONWriter) BeginLeafList(
	n *LeafList,
	empty bool,
	level int,
	hideSecrets bool,
) {
	b.WriteByte('"')
	b.handleModuleName(n)
	b.WriteString(n.Name())
	b.WriteString("\":")
}

func (b *JSONWriter) WriteLeafListValues(
	n *LeafList,
	empty bool,
	level int,
	hideSecrets bool,
) {
	vals := n.SortedChildren()
	if len(vals) == 0 {
		b.WriteString("null")
		return
	}
	hide := hideSecrets && n.GetSchema().ConfigdExt().Secret
	b.WriteByte('[')
	for i, v := range vals {
		if hide {
			b.WriteString(quote("********"))
		} else {
			b.writeValue(v)
		}
		if i != len(vals)-1 {
			b.WriteByte(',')
		}
	}
	b.WriteByte(']')
}

func (b *JSONWriter) EndLeafList(n *LeafList, empty bool, level int) {
	b.popName()
}

func (b *JSONWriter) PrintSep() {
	b.WriteByte(',')
}

func (n *node) encodeJSON(b *JSONWriter, options ...UnionOption) []byte {
	switch n.specialized.(type) {
	case *ListEntry:
		//ListEntries already put the '{}' around the
		//object, doing it twice is not legal JSON syntax.
		n.Serialize(b, nil, options...)
	default:
		b.WriteByte('{')
		n.Serialize(b, nil, options...)
		b.WriteByte('}')
	}

	return b.Bytes()
}

func (n *node) ToJSON(options ...UnionOption) []byte {
	return n.encodeJSON(&JSONWriter{}, options...)
}

func (n *node) ToRFC7951(options ...UnionOption) []byte {
	return n.encodeJSON(&JSONWriter{rfc7951: true}, options...)
}

// Take the JSON message and create a UnionTree using the given schema and
// message content.
func UnmarshalJSON(schemaRoot schema.Node, jsonInput []byte) (Node, error) {
	return unmarshalJSONInternal(schemaRoot, jsonInput, true /* validate */)
}

func UnmarshalJSONWithoutValidation(
	schemaRoot schema.Node,
	jsonInput []byte,
) (Node, error) {
	return unmarshalJSONInternal(schemaRoot, jsonInput,
		false /* don't validate */)
}

func unmarshalJSONInternal(
	schemaRoot schema.Node,
	jsonInput []byte,
	validate bool,
) (Node, error) {

	root := NewNode(data.New("root"), data.New("root"), schemaRoot, nil, 0)
	if root == nil {
		err := mgmterror.NewOperationFailedApplicationError()
		err.Message = "Invalid schema provided"
		return nil, err
	}

	if validate {
		if err := unmarshalJSONIntoNode(root, jsonInput); err != nil {
			return nil, err
		}
	} else {
		err := UnmarshalJSONIntoNodeWithoutValidation(
			root, encoding.Config /* dummy value */, jsonInput)
		if err != nil {
			return nil, err
		}
	}
	return root, nil
}

func unmarshalJSONIntoNode(ut Node, jsonInput []byte) (err error) {

	datatree, err := encoding.UnmarshalJSON(ut.GetSchema(), jsonInput)
	if err != nil {
		return err
	}

	return yangDataIntoTree(ut, datatree)
}

func UnmarshalJSONIntoNodeWithoutValidation(
	ut Node,
	cfgOrState encoding.ConfigOrState,
	jsonInput []byte,
) (err error) {

	datatree, err := encoding.UnmarshalJSONWithoutValidation(
		ut.GetSchema(), cfgOrState /* IGNORED */, jsonInput)
	if err != nil {
		return err
	}

	if cfgOrState == encoding.State {
		return yangStateDataIntoTree(ut, datatree)
	}
	return yangDataIntoTree(ut, datatree)
}

// UnmarshalJSONConfigsWithoutValidation - merge JSON configs into data tree
// Function is of type schema.ConfigMultiplexerFn
//
// Only used for test (from configd) - so needs to be exported, and can't
// be exported from _test.go file as that doesn't work outside this package
// (only works for exporting from pkg foo to foo_test AFAICT).
func UnmarshalJSONConfigsWithoutValidation(
	configs [][]byte,
	ms schema.ModelSet,
) (*data.Node, error) {

	root := NewNode(data.New("root"), data.New("root"), ms, nil, 0)
	for _, config := range configs {
		err := UnmarshalJSONIntoNodeWithoutValidation(
			root, encoding.Config, config)
		if err != nil {
			return nil, err
		}
	}
	return root.Data(), nil
}

func UnmarshalRFC7951(schemaRoot schema.Node, jsonInput []byte) (Node, error) {
	root := NewNode(data.New("root"), data.New("root"), schemaRoot, nil, 0)
	if root == nil {
		err := mgmterror.NewOperationFailedApplicationError()
		err.Message = "Invalid schema provided"
		return nil, err
	}

	if err := unmarshalRFC7951IntoNode(root, jsonInput); err != nil {
		return nil, err
	}
	return root, nil
}

func unmarshalRFC7951IntoNode(ut Node, jsonInput []byte) (err error) {
	datatree, err := encoding.UnmarshalRFC7951(ut.GetSchema(), jsonInput)
	if err != nil {
		return err
	}

	return yangDataIntoTree(ut, datatree)
}
