// Copyright (c) 2018-2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package union

import (
	"testing"

	"github.com/danos/config/data"
	"github.com/danos/config/schema"
	"github.com/danos/yang/compile"
	"github.com/danos/yang/parse"
	yang "github.com/danos/yang/schema"
)

const baseSchema = `
module test-union {
    namespace "urn:vyatta.com:test:union";
    prefix utest;
    organization "Brocade Communications Systems, Inc.";
    contact
        "Brocade Communications Systems, Inc.
         Postal: 130 Holger Way
                 San Jose, CA 95134
         E-mail: support@Brocade.com
         Web: www.brocade.com";
    revision 2015-03-11 {
        description "Test schema for unions";
    }
	grouping leaves {
		leaf nondefault {
			type string;
		}
		leaf default {
			type string;
			default "foo";
		}
		leaf default2 {
			type string;
			default "bar";
		}
	}
	grouping group {
		list testlist {
			key key;
			leaf key {
				type string;
			}
			uses leaves;
		}
		uses leaves;
	}
	grouping internal {
		container internalnonpresence {
			uses group;
		}
		container internalpresence {
			presence "Has presence";
			uses group;
		}
	}
	
	container nonpresence {
		uses group;
		uses internal;
	}
	container presence {
		presence "Has presence";
		uses group;
		uses internal;
	}
	uses group;
}
`

func getSchema(buf []byte) (schema.ModelSet, error) {
	const name = "schema"
	t, err := schema.Parse(name, string(buf))
	if err != nil {
		return nil, err
	}

	var mods = map[string]*parse.Tree{name: t}
	st, err := schema.CompileModules(mods, "", false, compile.IsConfig, nil)
	if err != nil {
		return nil, err
	}
	return st, nil
}

func getFullSchema(buf []byte) (schema.ModelSet, error) {
	const name = "schema"
	t, err := schema.Parse(name, string(buf))
	if err != nil {
		return nil, err
	}

	var mods = map[string]*parse.Tree{name: t}
	st, err := schema.CompileModules(mods, "", false,
		compile.Include(compile.IsConfig, compile.IncludeState(true)), nil)
	if err != nil {
		return nil, err
	}
	return st, nil
}

func getInitialTree(t *testing.T) Node {
	st, err := getSchema([]byte(baseSchema))
	if err != nil {
		t.Fatal(err)
	}

	return NewNode(data.New("root"), data.New("root"), st, nil, 0)
}

func generateTree(t *testing.T) Node {
	st, err := getSchema([]byte(baseSchema))
	if err != nil {
		t.Fatal(err)
	}

	root := data.New("root")

	return NewNode(data.New("root"), root, st, nil, 0)
}

func serializeTree(root Node) string {
	var b StringWriter
	root.Serialize(&b, nil, IncludeDefaults)
	return b.String()
}

func badSerialization(t *testing.T, got, expected string) {
	t.Fatalf("Bad serialization:\n%s\nExpected:\n%s\n", got, expected)
}

func testSerialize(t *testing.T, tree Node, expected string) {
	out := serializeTree(tree)
	//fmt.Println(out)
	if out != expected {
		badSerialization(t, out, expected)
	}
}

//Base tests just verify the behavior of defaults at various node types in the system.
func TestSerialize(t *testing.T) {
	const initialExpectedTree = `default foo
default2 bar
nonpresence {
	default foo
	default2 bar
	internalnonpresence {
		default foo
		default2 bar
	}
}
`
	root := getInitialTree(t)
	testSerialize(t, root, initialExpectedTree)
}

func TestTopPresence(t *testing.T) {
	const topLevelPresenceTree = `default foo
default2 bar
nonpresence {
	default foo
	default2 bar
	internalnonpresence {
		default foo
		default2 bar
	}
}
presence {
	default foo
	default2 bar
	internalnonpresence {
		default foo
		default2 bar
	}
}
`
	//set presence
	root := getInitialTree(t)
	root.Data().AddChild(data.New("presence"))
	testSerialize(t, root, topLevelPresenceTree)
}

func TestInternalPresenceTree1(t *testing.T) {
	const internalPresenceTree1 = `default foo
default2 bar
nonpresence {
	default foo
	default2 bar
	internalnonpresence {
		default foo
		default2 bar
	}
	internalpresence {
		default foo
		default2 bar
	}
}
presence {
	default foo
	default2 bar
	internalnonpresence {
		default foo
		default2 bar
	}
}
`
	//set presence
	//set nonpresence internalpresence
	root := getInitialTree(t)
	root.Data().AddChild(data.New("presence"))
	root.Data().AddChild(data.New("nonpresence"))
	root.Child("nonpresence").Data().AddChild(data.New("internalpresence"))
	testSerialize(t, root, internalPresenceTree1)

}

func TestInternalPresenceTree2(t *testing.T) {
	const internalPresenceTree2 = `default foo
default2 bar
nonpresence {
	default foo
	default2 bar
	internalnonpresence {
		default foo
		default2 bar
	}
	internalpresence {
		default foo
		default2 bar
	}
}
presence {
	default foo
	default2 bar
	internalnonpresence {
		default foo
		default2 bar
	}
	internalpresence {
		default foo
		default2 bar
	}
}
`
	//set presence
	//set nonpresence internalpresence
	//set presence internalpresence
	root := getInitialTree(t)
	root.Data().AddChild(data.New("presence"))
	root.Data().AddChild(data.New("nonpresence"))
	root.Child("nonpresence").Data().AddChild(data.New("internalpresence"))
	root.Child("presence").Data().AddChild(data.New("internalpresence"))
	testSerialize(t, root, internalPresenceTree2)
}

func TestListPresenceTree(t *testing.T) {
	const listPresenceExpectedTree = `default foo
default2 bar
nonpresence {
	default foo
	default2 bar
	internalnonpresence {
		default foo
		default2 bar
	}
}
presence {
	default foo
	default2 bar
	internalnonpresence {
		default foo
		default2 bar
	}
	testlist one {
		default foo
		default2 bar
	}
	testlist two {
		default foo
		default2 bar
	}
}
`
	//set presence testlist one
	//set presence testlist two
	root := getInitialTree(t)
	root.Data().AddChild(data.New("presence"))
	root.Child("presence").Data().AddChild(data.New("testlist"))
	root.Child("presence").Child("testlist").Data().AddChild(data.New("one"))
	root.Child("presence").Child("testlist").Data().AddChild(data.New("two"))
	testSerialize(t, root, listPresenceExpectedTree)
}

func TestListNonPresenceTree(t *testing.T) {
	const listExpectedTree = `default foo
default2 bar
nonpresence {
	default foo
	default2 bar
	internalnonpresence {
		default foo
		default2 bar
	}
	testlist one {
		default foo
		default2 bar
	}
	testlist two {
		default foo
		default2 bar
	}
}
`
	//set nonpresence testlist one
	//set nonpresence testlist two
	root := getInitialTree(t)
	root.Data().AddChild(data.New("nonpresence"))
	root.Child("nonpresence").Data().AddChild(data.New("testlist"))
	root.Child("nonpresence").Child("testlist").Data().AddChild(data.New("one"))
	root.Child("nonpresence").Child("testlist").Data().AddChild(data.New("two"))
	testSerialize(t, root, listExpectedTree)
}

func TestName(t *testing.T) {
}

func TestChildren(t *testing.T) {
}

func TestNumChildren(t *testing.T) {
}

func TestSortedChildren(t *testing.T) {
}

func TestChild(t *testing.T) {
}

func TestEmpty(t *testing.T) {
}

func compareLists(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if b[i] != v {
			return false
		}
	}
	return true
}

type TestCase func()

func genTestGet(t *testing.T, node Node, path []string, expected []string) TestCase {
	return func() {
		out, err := node.Get(nil, path)
		if err != nil {
			t.Fatal(err)
		}
		if !compareLists(out, expected) {
			t.Fatal("got", out, "expected", expected, "from get", path)
		}
	}
}

func checkValidPathErr(t *testing.T, path []string, err error, valid bool) {
	switch {
	case valid && err != nil:
		t.Fatal(err)
	case !valid && err == nil:
		t.Fatal(yang.NewNodeExistsError(path))
	}
}

func genTestSet(t *testing.T, node Node, path []string, valid bool) TestCase {
	return func() {
		err := node.Set(nil, path)
		checkValidPathErr(t, path, err, valid)
	}
}

func genTestDelete(t *testing.T, node Node, path []string, valid bool) TestCase {
	return func() {
		err := node.Delete(nil, path, DontCheckAuth)
		checkValidPathErr(t, path, err, valid)
	}
}

func genTestExists(t *testing.T, node Node, path []string, valid bool) TestCase {
	return func() {
		err := node.Exists(nil, path)
		checkValidPathErr(t, path, err, valid)
	}
}

func genTestOnlyOverlay(t *testing.T, node Node, path []string, shouldPass bool) TestCase {
	return func() {
		n, err := node.Descendant(nil, path)
		if err != nil && shouldPass {
			t.Fatal(err)
		} else if !shouldPass {
			return
		}
		if !n.hasOverlay() {
			t.Fatal("missing overlay")
		}
		if n.hasUnderlay() {
			t.Fatal("unexpected underlay")
		}
	}
}

func genTestOverlayDeleted(t *testing.T, node Node, path []string) TestCase {
	return func() {
		n, err := node.Descendant(nil, path)
		if err != nil {
			t.Fatal(err)
		}
		if !n.deleted() {
			t.Fatal("node should be deleted")
		}
	}
}

func genTestOverlayDeletedandOpaque(t *testing.T, node Node, path []string) TestCase {
	return func() {
		n, err := node.Descendant(nil, path)
		if err != nil {
			t.Fatal(err)
		}
		if !n.hasOverlay() {
			t.Fatal("missing overlay")
		}
		if n.hasUnderlay() {
			t.Fatal("unexpected underlay")
		}
	}
}

func genTestOverlayOpaque(t *testing.T, node Node, path []string) TestCase {
	return func() {
		n, err := node.Descendant(nil, path)
		if err != nil {
			t.Fatal(err)
		}
		if !n.hasOverlay() {
			t.Fatal("missing overlay")
		}
		if n.hasUnderlay() {
			t.Fatal("unexpected underlay")
		}
	}
}

func testSequence(tcs ...TestCase) {
	for _, tc := range tcs {
		tc()
	}
}

func TestUnion(t *testing.T) {
	//This is a sanity check of the general workings of the tree.
	root := getInitialTree(t)
	//set presence nondefault someval
	testSequence(
		genTestSet(t, root, []string{"presence", "nondefault", "someval"}, true),
		genTestGet(t, root, []string{"presence", "nondefault"}, []string{"someval"}),
		genTestGet(t, root, []string{"presence"},
			[]string{"default", "default2", "internalnonpresence", "nondefault"}),
		genTestExists(t, root, []string{"presence", "nondefault", "someval"}, true),
		genTestExists(t, root, []string{"presence", "nondefault", "someval2"}, false),
		genTestDelete(t, root, []string{"presence", "nondefault", "someval"}, true),
		genTestExists(t, root, []string{"presence", "nondefault", "someval"}, false),
		genTestDelete(t, root, []string{"presence"}, true),
	)
}

//The following test cases are numbered because what they check would be too long for
//a function name, node states
//  1) created in candidate
//  2) deleted from running
//  3) 1) and then deleted
//  4) 2) and then 1)
//Each of these 4 cases will be tested on all node types (except root)
// Container, List, ListEntry, Leaf, LeafList, LeafValue
// Container and Leaf defaults will also be tested

func test1(t *testing.T, path []string, validSet bool) TestCase {
	return func() {
		root := getInitialTree(t)
		testSequence(
			genTestSet(t, root, path, validSet),
			genTestOnlyOverlay(t, root, path, validSet),
		)
	}
}

func TestContainer(t *testing.T) {
	testSequence(
		//1
		test1(t, []string{"presence"}, true),
		//2
		//3
		//4
	)
}
func TestDefaultContainer(t *testing.T) {
	testSequence(
		//1
		test1(t, []string{"nonpresence"}, false),
		//2
		//3
		//4
	)
}

func TestList(t *testing.T) {
	testSequence(
		//1
		test1(t, []string{"presence", "testlist"}, false),
		//2
		//3
		//4
	)
}

func TestListEntry(t *testing.T) {
	testSequence(
		//1
		test1(t, []string{"presence", "testlist", "one"}, true),
		//2
		//3
		//4
	)
}

func TestLeaf(t *testing.T) {
	testSequence(
		//1
		test1(t, []string{"nondefault", "foobar"}, true),
	//2
	//3
	//4
	)
}

func TestDefaultLeaf(t *testing.T) {
	testSequence(
	//1
	//2
	//3
	//4
	)
}

func TestLeafList(t *testing.T) {
	testSequence(
	//1
	//2
	//3
	//4
	)
}

func TestLeafValue(t *testing.T) {
	testSequence(
	//1
	//2
	//3
	//4
	)
}
