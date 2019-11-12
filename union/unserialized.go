// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2014-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package union

import (
	"fmt"
	"reflect"

	"github.com/danos/config/data"
	"github.com/danos/config/schema"
	"github.com/danos/mgmterror"
	"github.com/danos/utils/pathutil"
	"github.com/danos/yang/data/datanode"
	"github.com/danos/yang/data/encoding"
	yang "github.com/danos/yang/schema"
)

type unserialized interface {
	name() string
	values() ([]string, error)
	unserializedChildren(yang.Node) ([]unserialized, error)
}

type callbackCtx struct {
	path []string
}

type callbackFn func(*callbackCtx) error

func callWalkerCallback(full_path []string, userFn callbackFn) error {
	return userFn(&callbackCtx{path: full_path})
}

func newUnknownSchemaError(schema_type, name string, path []string) error {
	err := mgmterror.NewOperationFailedApplicationError()
	err.Path = pathutil.Pathstr(append(path, name))
	err.Message = fmt.Sprintf("Schema type (%s) not handled when getting path",
		schema_type)
	return err
}

func newEmptyLeafWithValue(name string) error {
	err := mgmterror.NewInvalidValueApplicationError()
	err.Message = fmt.Sprintf("Value found for empty leaf: %s", name)
	return err
}

func convertToDataNode(node unserialized, sn yang.Node) (datanode.DataNode, error) {

	var err error
	children := []datanode.DataNode{}
	vals := []string{}
	name := node.name()

	switch sn.(type) {

	// We don't expect to be called for LeafValues
	default:
		panic(fmt.Errorf("Attempt to convert unexpected schema type"))

	// Leaf and LeafList nodes have values
	case schema.Leaf, schema.LeafList:
		vals, err = node.values()
		if err != nil {
			return nil, err
		}

		// This is the only place where we differentiate between and empty leaf
		// and a value of ""
		if _, ok := sn.(schema.Leaf); ok {
			if _, isEmpty := sn.Type().(schema.Empty); isEmpty {
				if len(vals) > 0 && (len(vals) != 1 || vals[0] != "") {
					return nil, newEmptyLeafWithValue(node.name())
				}
			}
		}

	// Container, List and ListEntry nodes have children
	case schema.Container, schema.List, schema.ListEntry:
		ukids, err := node.unserializedChildren(sn)
		if err != nil {
			return nil, err
		}
		children = make([]datanode.DataNode, len(ukids), len(ukids))
		for i, ch := range ukids {
			csn := sn.Child(ch.name())
			if csn == nil {
				return nil, yang.NewSchemaMismatchError(ch.name(), []string{sn.Name()})
			}
			children[i], err = convertToDataNode(ch, csn)
			if err != nil {
				return nil, err
			}
		}
	}

	return datanode.CreateDataNode(name, children, vals), nil
}

func validateDataNode(n datanode.DataNode, sn yang.Node) error {

	if _, errs, ok := yang.ValidateSchema(sn, n, false /* dbg */); !ok {
		var errList mgmterror.MgmtErrorList
		errList.MgmtErrorListAppend(errs...)
		return errList
	}

	return nil
}

// If we are adding state data into the tree, we do not want to create
// presence nodes.  These will be created, if needed, by the previously
// added config data.  State data will then be added to these presence
// containers if they exist, otherwise it won't be added. This also avoids
// trying to recreate existing presence containers.
func yangDataIntoTree(ut Node, node datanode.DataNode) error {
	return yangDataIntoTreeInternal(ut, node, false)
}

func yangStateDataIntoTree(ut Node, node datanode.DataNode) error {
	return yangDataIntoTreeInternal(ut, node, true)
}

func yangDataIntoTreeInternal(
	ut Node,
	node datanode.DataNode,
	stateData bool,
) error {
	// We use a closure for the callback so we don't have to pass down context
	// (the ut in this case) through the full walk.
	callback := func(ctx *callbackCtx) error {
		return ut.Set(nil, ctx.path)
	}

	// The top level node has already been processed,
	// just process the children of this node into the ut.
	return processChildren(node, ut.GetSchema(), []string{}, callback,
		stateData)
}

// Process each child of a node, skipping the key node for ListEntry nodes,
// because the value of the key is encoded as the value of the ListEntry itself.
func processChildren(
	node datanode.DataNode,
	sn yang.Node,
	path []string,
	userFn callbackFn,
	stateData bool) error {

	child_list, err := childrenSkipListKey(node, sn, path)
	if err != nil {
		return err
	}

	for _, c := range child_list {
		cn := sn.Child(c.YangDataName())
		if err := processNode(c, cn, path, userFn, stateData); err != nil {
			return err
		}
	}
	return nil
}

func childrenSkipListKey(
	node datanode.DataNode,
	sn yang.Node,
	path []string,
) ([]datanode.DataNode, error) {

	list := node.YangDataChildren()

	switch schema_node := sn.(type) {
	case schema.Tree, schema.Container, schema.List:
		return list, nil

	case schema.ListEntry:
		key_name := schema_node.Keys()[0]

		if len(list) == 0 {
			// Not even the key?
			return nil, yang.NewMissingKeyError(path)
		}
		filtered := make([]datanode.DataNode, 0, len(list)-1)
		for _, v := range list {
			if v.YangDataName() != key_name {
				filtered = append(filtered, v)
			}
		}
		return filtered, nil

	default:
		return nil, newUnknownSchemaError(reflect.TypeOf(sn).String(),
			node.YangDataName(), path)
	}
}

// Process a node based on it's schema type
func processNode(
	node datanode.DataNode,
	sn yang.Node,
	parent_path []string,
	userFn callbackFn,
	stateData bool) error {

	switch schema_node := sn.(type) {

	case schema.LeafList, schema.Leaf:
		path := append(parent_path, node.YangDataName())
		return processValues(node, sn, path, userFn)

	case schema.ListEntry:
		// List Entries are a very special case. They have the name of
		// their key_value and we skip creating the child node for the key
		key_value, err := getKeyValue(node, schema_node)
		if err != nil {
			return err
		}
		path := append(parent_path, key_value)

		child_list, err := childrenSkipListKey(node, sn, path)
		if err != nil {
			return err
		}

		if len(child_list) == 0 {
			// If the list entry exists with just the key, callback here
			return callWalkerCallback(path, userFn)
		}
		return processChildren(node, sn, path, userFn, stateData)

	case schema.Container:
		path := append(parent_path, node.YangDataName())
		if schema_node.Presence() && !stateData {
			// For presence containers, we need to create them,
			// but only if configuration data. This ensures that
			// there are no merge conflicts of the config and
			// state trees, and that any state will only be present
			// in the merged tree if parent presence nodes exist in
			// the configuration.
			if err := callWalkerCallback(path, userFn); err != nil {
				return err
			}
		}
		return processChildren(node, sn, path, userFn, stateData)

	case schema.List:
		path := append(parent_path, node.YangDataName())
		return processChildren(node, sn, path, userFn, stateData)

	default:
		return newUnknownSchemaError(reflect.TypeOf(schema_node).String(),
			node.YangDataName(), parent_path)
	}
}

func getKeyValue(node datanode.DataNode, ln schema.ListEntry) (string, error) {

	// NOTE: We only deal with lists with exactly one key
	key_name := ln.Keys()[0]

	for _, ch := range node.YangDataChildrenNoSorting() {
		if ch.YangDataName() == key_name {
			values := ch.YangDataValuesNoSorting()
			if (len(values) != 1) || (values[0] == "") {
				return "", fmt.Errorf("Unexpected value(s) for list key: %s\\%s\n",
					node.YangDataName(), key_name)
			}
			return values[0], nil
		}
	}

	return "", yang.NewMissingKeyError([]string{ln.Name(), key_name})
}

func processValues(
	node datanode.DataNode, sn yang.Node, path []string, userFn callbackFn) error {

	values := node.YangDataValues()

	val_count := len(values)

	switch schema_node := sn.(type) {
	case schema.Leaf:
		if _, isEmpty := schema_node.Type().(schema.Empty); isEmpty {
			if val_count > 0 && (val_count != 1 || values[0] != "") {
				return newEmptyLeafWithValue(node.YangDataName())
			}
			return callWalkerCallback(path, userFn)

		} else if val_count > 1 {
			return fmt.Errorf("More than one entry for non-list: %s\n", node.YangDataName())
		}
	}

	for _, value := range values {
		if err := callWalkerCallback(append(path, value), userFn); err != nil {
			return err
		}
	}
	return nil
}

func dumpTree(n unserialized, tabbing string, sn yang.Node) {

	values, _ := n.values()
	if sn == nil {
		fmt.Printf("DATA:     %s%s/%s ....?  ERROR - SCHEMA NOT FOUND\n",
			tabbing, n.name(), values)
		return
	}

	list, err := n.unserializedChildren(sn)
	if err != nil {
		fmt.Printf("DATA:     %s%s/%s ....> ERROR GETTING CHILDREN %s\n",
			tabbing, n.name(), values, err.Error())
		return
	}

	fmt.Printf("DATA:     %s%s/%s (%d)\n",
		tabbing, n.name(), values, len(list))

	for _, c := range list {
		dumpTree(c, tabbing+"    ", sn.Child(c.name()))
	}
}

func dumpDataTree(dn datanode.DataNode, tabbing string, sn yang.Node) {

	values := dn.YangDataValues()
	if sn == nil {
		fmt.Printf("DATA:     %s%s/%s ....?  ERROR - SCHEMA NOT FOUND\n",
			tabbing, dn.YangDataName(), values)
		return
	}

	list := dn.YangDataChildren()
	fmt.Printf("DATA:     %s%s/%s (%d)\n",
		tabbing, dn.YangDataName(), values, len(list))

	for _, cdn := range list {
		dumpDataTree(cdn, tabbing+"    ", sn.Child(cdn.YangDataName()))
	}
}

func dumpSchemaTree(n yang.Node, tabbing string) {
	fmt.Printf("SCHEMA:   %s%s/\n", tabbing, n.Name())
	for _, c := range n.Children() {
		dumpSchemaTree(c, tabbing+"  ")
	}
}

// We can completely avoid the rat's nest of ever increasing APIs for
// unmarshalling different encoding types with different validation
// requirements by using the generic unmarshaller interface provided
// by the 'encoder'.
type Unmarshaller struct {
	um encoding.Unmarshaller
}

func NewUnmarshaller(enc encoding.EncType) *Unmarshaller {
	return &Unmarshaller{
		um: encoding.NewUnmarshaller(enc)}
}

func (um *Unmarshaller) SetValidation(
	valType yang.ValidationType,
) *Unmarshaller {
	um.um.SetValidation(valType)
	return um
}

func (um *Unmarshaller) Unmarshal(
	ms schema.ModelSet,
	input []byte,
) (root Node, err error) {

	root = NewNode(data.New("root"), data.New("root"), ms, nil, 0)
	if root == nil {
		err := mgmterror.NewOperationFailedApplicationError()
		err.Message = "Invalid schema provided"
		return nil, err
	}

	datatree, err := um.um.Unmarshal(root.GetSchema(), input)
	if err != nil {
		return nil, err
	}

	return root, yangDataIntoTree(root, datatree)
}
