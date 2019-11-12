// Copyright (c) 2017,2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package data

import (
	"fmt"
	"testing"
)

func createBaseTree() *Node {
	t := New("root")
	for i := 0; i < 10; i++ {
		ch := New(fmt.Sprintf("Test%d", i))
		t.AddChild(ch)
		for i := 0; i < 5; i++ {
			ch.AddChild(New(fmt.Sprintf("TestCh%d", i)))
		}
	}
	return t
}

func TestChild(t *testing.T) {
	tree := createBaseTree()
	ch := tree.Child("Test0")
	if ch == nil {
		t.Fatal("did not find expected child")
	}
	ch = tree.Child("Test20")
	if ch != nil {
		t.Fatal("found unexpected child")
	}
}

func TestAddChild(t *testing.T) {
	tree := createBaseTree()
	tree.AddChild(New("TestChild"))
	if ch := tree.Child("TestChild"); ch == nil {
		t.Fatal("did not find expected child")
	}
}

func TestDeleteChild(t *testing.T) {
	tree := createBaseTree()
	tree.DeleteChild("Test0")
	if ch := tree.Child("Test0"); ch != nil {
		t.Fatal("found unexpected child")
	}
}

func TestChildren(t *testing.T) {
	children := make([]*Node, 0, 10)
	for i := 0; i < 10; i++ {
		children = append(children, New(fmt.Sprintf("Test%d", i)))
	}
	isElemOfChildren := func(n *Node) bool {
		for _, ch := range children {
			if ch.Name() == n.Name() {
				return true
			}
		}
		return false
	}
	tree := createBaseTree()
	tchildren := tree.Children()
	if len(tchildren) != len(children) {
		t.Fatal("unexpected number of children")
	}
	for _, ch := range tchildren {
		if !isElemOfChildren(ch) {
			t.Fatal("did not find expected child")
		}
	}
}

func TestClearChildren(t *testing.T) {
	tree := createBaseTree()
	tree.ClearChildren()
	if len(tree.Children()) != 0 {
		t.Fatal("unexpected number of children")
	}
}

func TestName(t *testing.T) {
	tree := createBaseTree()
	ch := tree.Child("Test0")
	if ch.Name() != "Test0" {
		t.Fatal("unexpected child name")
	}
}

func TestIndex(t *testing.T) {
	tree := createBaseTree()
	ch := tree.Child("Test0")
	if ch.Index() != 0 {
		t.Fatal("unexpected child index")
	}

	ch = tree.Child("Test9")
	if ch.Index() != 9 {
		t.Fatal(fmt.Sprintf("unexpected child index"))
	}
}

func TestSetIndex(t *testing.T) {
	tree := createBaseTree()
	ch := tree.Child("Test9")
	ch.SetIndex(10)
	if ch.Index() != 10 {
		t.Fatal(fmt.Sprintf("unexpected child index"))
	}
}

func TestComment(t *testing.T) {
	tree := createBaseTree()
	ch := tree.Child("Test0")
	ch.SetComment("My comment")
	if ch.Comment() != "My comment" {
		t.Fatal(fmt.Sprintf("unexpected comment"))
	}
}

func TestFlags(t *testing.T) {
	tree := createBaseTree()
	ch := tree.Child("Test0")
	ch.MarkDeleted(ClearChildFlags)
	if len(ch.Children()) != 0 {
		t.Fatal("unexpected number of children")
	}
	if !ch.Deleted() {
		t.Fatal("unexpected value for Deleted()")
	}
	ch.ClearDeleted()
	if ch.Deleted() {
		t.Fatal("unexpected value for Deleted()")
	}
	if !ch.Opaque() {
		t.Fatal("unexpected value for Opaque()")
	}
	ch.ClearOpaque()
	if ch.Opaque() {
		t.Fatal("unexpected value for Opaque()")
	}
	ch.MarkOpaque()
	if !ch.Opaque() {
		t.Fatal("unexpected value for Opaque()")
	}
}

func TestSetNoValidate(t *testing.T) {
	tree := createBaseTree()
	tree.SetNoValidate([]string{"Test0", "foo", "bar"})
	test0 := tree.Child("Test0")
	foo := test0.Child("foo")
	if foo == nil {
		t.Fatal("did not find expected child foo")
	}
	bar := foo.Child("bar")
	if bar == nil {
		t.Fatal("did not find expected child bar")
	}
}
