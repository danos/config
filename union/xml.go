// Copyright (c) 2018-2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2014-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package union

import (
	"bytes"
	"encoding/xml"
	"fmt"

	"github.com/danos/config/data"
	"github.com/danos/config/schema"
	"github.com/danos/yang/data/encoding"
)

type XMLWriter struct {
	*xml.Encoder
}

func (enc *XMLWriter) BeginContainer(n *Container, empty bool, level int) {
	name := xml.Name{Space: n.Schema.Namespace(), Local: n.Name()}
	enc.EncodeToken(xml.StartElement{Name: name})
}
func (enc *XMLWriter) EndContainer(n *Container, empty bool, level int) {
	name := xml.Name{Space: n.Schema.Namespace(), Local: n.Name()}
	enc.EncodeToken(xml.EndElement{Name: name})
}
func (enc *XMLWriter) BeginList(n *List, empty bool, level int) {}
func (enc *XMLWriter) EndList(n *List, empty bool, level int)   {}
func (enc *XMLWriter) BeginListEntry(n *ListEntry, empty bool, level int, hideSecrets bool) {
	sch := n.Schema
	pname := xml.Name{Space: sch.Namespace(), Local: n.parent.Name()}
	enc.EncodeToken(xml.StartElement{Name: pname})

	//create list 'key' nodes
	//TODO: fix this for multi part keys
	kname := xml.Name{Space: sch.Namespace(), Local: sch.Keys()[0]}
	enc.EncodeToken(xml.StartElement{Name: kname})
	if redactListEntry(n, hideSecrets) {
		enc.EncodeToken(xml.CharData("********"))
	} else {
		enc.EncodeToken(xml.CharData([]byte(n.Data().Name())))
	}
	enc.EncodeToken(xml.EndElement{Name: kname})
}
func (enc *XMLWriter) EndListEntry(n *ListEntry, empty bool, level int) {
	name := xml.Name{Space: n.GetSchema().Namespace(), Local: n.parent.Name()}
	enc.EncodeToken(xml.EndElement{Name: name})
}

func (enc *XMLWriter) BeginLeaf(n *Leaf, empty bool, level int, hideSecrets bool) {}

func (enc *XMLWriter) WriteLeafValue(n *Leaf, empty bool, level int, hideSecrets bool) {
	enc.writeLeafValue(n, empty, level, hideSecrets)
}
func (enc *XMLWriter) writeLeafValue(n Node, empty bool, level int, hideSecrets bool) {
	name := xml.Name{Space: n.GetSchema().Namespace(), Local: n.Name()}
	if empty {
		enc.EncodeToken(xml.StartElement{Name: name})
		enc.EncodeToken(xml.EndElement{Name: name})
		return
	}
	hide := hideSecrets && n.GetSchema().ConfigdExt().Secret
	vals := n.SortedChildren()
	for _, v := range vals {
		//Errors only occur if we run out of buffer
		//just ignore them, there is no error path
		//here.
		enc.EncodeToken(xml.StartElement{Name: name})
		if hide {
			enc.EncodeToken(xml.CharData("********"))
		} else {
			enc.EncodeToken(xml.CharData([]byte(v.Name())))
		}
		enc.EncodeToken(xml.EndElement{Name: name})
	}
}
func (enc *XMLWriter) EndLeaf(n *Leaf, empty bool, level int) {}

func (enc *XMLWriter) BeginLeafList(n *LeafList, empty bool, level int, hideSecrets bool) {}

func (enc *XMLWriter) WriteLeafListValues(n *LeafList, empty bool, level int, hideSecrets bool) {
	enc.writeLeafValue(n, empty, level, hideSecrets)
}

func (enc *XMLWriter) EndLeafList(n *LeafList, empty bool, level int) {}

func (enc *XMLWriter) PrintSep() {}

func (n *node) addEnclosingStartElements(enc *XMLWriter) {
	var tokens []xml.Token
	for p := n.Parent(); p != nil && p.Name() != "root"; p = p.Parent() {
		switch v := p.(type) {
		case *Container:
			xName := xml.Name{Space: v.Schema.Namespace(),
				Local: v.Name()}
			tokens = append(tokens, xml.StartElement{Name: xName})

		case *ListEntry:
			sch := v.Schema

			kname := xml.Name{Space: sch.Namespace(), Local: sch.Keys()[0]}
			tokens = append(tokens, xml.EndElement{Name: kname})
			tokens = append(tokens, xml.CharData([]byte(v.Data().Name())))
			tokens = append(tokens, xml.StartElement{Name: kname})

			pname := xml.Name{Space: sch.Namespace(), Local: v.parent.Name()}
			tokens = append(tokens, xml.StartElement{Name: pname})
		}
	}
	for i := len(tokens) - 1; i >= 0; i-- {
		enc.EncodeToken(tokens[i])
	}
}

func (n *node) addEnclosingEndElements(enc *XMLWriter) {
	for p := n.Parent(); p != nil && p.Name() != "root"; p = p.Parent() {
		switch v := p.(type) {
		case *Container:
			enc.EncodeToken(xml.EndElement{
				Name: xml.Name{Space: v.Schema.Namespace(),
					Local: v.Name()}})
		case *ListEntry:
			name := xml.Name{Space: v.GetSchema().Namespace(),
				Local: v.parent.Name()}
			enc.EncodeToken(xml.EndElement{Name: name})
		}
	}
}

// ToXML - return <n> serialized in XML format.
//
// For 'xml', we just return the node and entries under it, wrapped
// in the tag for the rootName.
func (n *node) ToXML(rootName string, options ...UnionOption) []byte {
	var b bytes.Buffer
	enc := &XMLWriter{xml.NewEncoder(&b)}
	enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: rootName}})
	n.Serialize(enc, nil, options...)
	enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: rootName}})
	enc.Flush()
	return b.Bytes()
}

// ToNETCONF - return <n> serialized in NETCONF format.
//
// For 'netconf', we need to fill in the enclosing tags up to root,
// eg if <n> represents '/interfaces/dataplane/dp0s1/address', we need to return
// tags for interfaces, dataplane *and* tagnode (dp0s1)
func (n *node) ToNETCONF(rootName string, options ...UnionOption) []byte {
	var b bytes.Buffer
	enc := &XMLWriter{xml.NewEncoder(&b)}

	enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: rootName}})
	n.addEnclosingStartElements(enc)
	n.Serialize(enc, nil, options...)
	n.addEnclosingEndElements(enc)
	enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: rootName}})

	enc.Flush()
	return b.Bytes()
}

func UnmarshalXML(schemaRoot schema.Node, xml_input []byte) (root Node, err error) {

	root = NewNode(data.New("root"), data.New("root"), schemaRoot, nil, 0)
	if root == nil {
		return nil, fmt.Errorf("Invalid schema provided")
	}

	datatree, err := encoding.UnmarshalXML(schemaRoot, xml_input)
	if err != nil {
		return nil, err
	}

	return root, yangDataIntoTree(root, datatree)
}
