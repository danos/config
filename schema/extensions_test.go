// Copyright (c) 2018-2019, AT&T Intellectual Property Inc. All rights reserved.
//
// Copyright (c) 2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"bytes"
	"fmt"
	"testing"

	yang "github.com/danos/yang/schema"
)

type checkFn func(t *testing.T, actual Node)

type NodeChecker struct {
	Name   string
	checks []checkFn
}

func (l NodeChecker) GetName() string {
	return l.Name
}

func (expected NodeChecker) check(t *testing.T, actual Node) {
	for _, checker := range expected.checks {
		checker(t, actual)
	}
}

func (n NodeChecker) String() string {
	return n.Name
}

func CheckName(expected_name string) checkFn {
	return func(t *testing.T, actual Node) {
		if expected_name != actual.Name() {
			t.Errorf("Node name does not match\n  expect=%s\n  actual=%s",
				expected_name, actual.Name())
		}
	}
}

func findChildByName(nl []yang.Node, name string) Node {
	for _, v := range nl {
		if v.Name() == name {
			return v.(Node)
		}
	}
	return nil
}

func checkAllNodes(t *testing.T, node_name string, expected []NodeChecker, actual []yang.Node) {
	if len(expected) != len(actual) {
		t.Errorf("Node %s child count does not match\n  expect=%d - %s\n  actual=%d - %s",
			node_name, len(expected), expected, len(actual), actual)
	}
	for _, exp := range expected {
		actualLeaf := findChildByName(actual, exp.GetName())
		if actualLeaf == nil {
			t.Errorf("Expected leaf not found: %s\n", exp.GetName())
			continue
		}
		exp.check(t, actualLeaf)
	}
}

func CheckChildren(node_name string, expected []NodeChecker) checkFn {
	return func(t *testing.T, actual Node) {
		checkAllNodes(t, node_name, expected, actual.Children())
	}
}

func stringListsMatch(expect, actual []string) bool {
	if len(expect) != len(actual) {
		return false
	}

	for i, _ := range expect {
		if actual[i] != expect[i] {
			return false
		}
	}

	return true
}

func CheckConfigdEnd(expect ...string) checkFn {
	return func(t *testing.T, node Node) {
		actual := node.ConfigdExt().End
		if !stringListsMatch(expect, actual) {
			t.Errorf("Create list mismatch for %s\n    expect: %s\n    actual: %s\n",
				node.Name(), expect, actual)
		}
	}
}

func CheckConfigdCreate(expect ...string) checkFn {
	return func(t *testing.T, node Node) {
		actual := node.ConfigdExt().Create
		if !stringListsMatch(expect, actual) {
			t.Errorf("Create list mismatch for %s\n    expect: %s\n    actual: %s\n",
				node.Name(), expect, actual)
		}
	}
}

func CheckConfigdAllowed(expect string) checkFn {
	return func(t *testing.T, node Node) {
		actual := node.ConfigdExt().Allowed
		if actual != expect {
			t.Errorf(
				"Allowed mismatch for %s\n    expect: %s\n    actual: %s\n",
				node.Name(), expect, actual)
		}
	}
}

func CheckConfigdValidate(expect ...string) checkFn {
	return func(t *testing.T, node Node) {
		actual := node.ConfigdExt().Validate
		if !stringListsMatch(expect, actual) {
			t.Errorf(
				"Validate mismatch for %s\n    expect: %s\n    actual: %s\n",
				node.Name(), expect, actual)
		}
	}
}

func NewContainerChecker(
	name string,
	children []NodeChecker,
	checks ...checkFn,
) NodeChecker {
	checkType := func(t *testing.T, node Node) {
		if _, ok := node.(Container); !ok {
			t.Errorf("Node type is not Container")
		}
	}
	checkList := append([]checkFn{
		CheckName(name),
		checkType,
		CheckChildren(name, children)},
		checks...)
	return NodeChecker{name, checkList}
}

func NewLeafChecker(name string, checks ...checkFn) NodeChecker {
	checkType := func(t *testing.T, node Node) {
		if _, ok := node.(Leaf); !ok {
			t.Errorf("Node type is not Leaf")
		}
	}
	checkList := append([]checkFn{
		checkType,
		CheckName(name)},
		checks...)
	return NodeChecker{name, checkList}
}

func getSchemaNode(
	t *testing.T,
	schema_text *bytes.Buffer,
	nodeName string,
) Node {

	ms, err := GetConfigSchema(schema_text.Bytes())
	if err != nil {
		t.Errorf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	return ms.SchemaChild(nodeName)
}

func TestReplaceRefine(t *testing.T) {

	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`grouping g1 {
			leaf one {
				type string;
					configd:create "stuff";
					configd:allowed "original";
			}
		}
		container c1 {
			uses g1 {
				refine one {
					configd:create "more";
					configd:allowed "new";
				}
			}
		}`))

	expected := NewContainerChecker(
		"c1",
		[]NodeChecker{
			NewLeafChecker("one",
				CheckConfigdCreate("more"),
				CheckConfigdAllowed("new")),
		})

	actual := getSchemaNode(t, schema_text, "c1")

	expected.check(t, actual)
}

func TestConfigdEndSpecialRefine(t *testing.T) {

	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`grouping g1 {
			leaf one {
				type string;
					configd:end "original1";
					configd:end "original2";
			}
		}
		container c1 {
			uses g1 {
				refine one {
					configd:end "new1";
					configd:end "new2";
				}
			}
		}`))

	expected := NewContainerChecker(
		"c1",
		[]NodeChecker{
			NewLeafChecker("one",
				CheckConfigdEnd("new1", "new2")),
		})

	actual := getSchemaNode(t, schema_text, "c1")

	expected.check(t, actual)
}

func TestConfigdEndSpecialNotRefine(t *testing.T) {

	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`grouping g1 {
			leaf one {
				type string;
					configd:end "original1";
					configd:end "original2";
			}
		}
		container c1 {
			uses g1 {
				refine one {
					mandatory true;
				}
			}
		}`))

	expected := NewContainerChecker(
		"c1",
		[]NodeChecker{
			NewLeafChecker("one",
				CheckConfigdEnd("original1", "original2")),
		})

	actual := getSchemaNode(t, schema_text, "c1")

	expected.check(t, actual)
}

func TestConfigdEndAugment(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(schemaTemplate,
		`container c1 {
			description "Test augment of configd:end";
		}
		augment /c1 {
			configd:end "augmented";
		}`))

	expected := NewContainerChecker("c1", []NodeChecker{}, CheckConfigdEnd("augmented"))
	actual := getSchemaNode(t, schema_text, "c1")
	expected.check(t, actual)
}

func TestConfigdValidateAugment(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(schemaTemplate,
		`container c1 {
			description "Test augment of configd:validate";
			configd:validate "original";
		}
		augment /c1 {
			configd:validate "augmented";
		}`))

	expected := NewContainerChecker("c1",
		[]NodeChecker{},
		CheckConfigdValidate("original", "augmented"))

	actual := getSchemaNode(t, schema_text, "c1")
	expected.check(t, actual)
}
