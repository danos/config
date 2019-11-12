// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0
//
// Tests on diff node functionality for ordered-by-user lists and
// leaf-lists.  Note that while in some cases you might think that the
// change shown would be a no-op, it may need to show up as a remove followed
// by an add to maintain existing behaviour.

package diff_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/danos/config/data"
	"github.com/danos/config/diff"
	"github.com/danos/config/load"
	"github.com/danos/config/schema"
	. "github.com/danos/config/testutils"
	"github.com/danos/config/testutils/assert"
	"github.com/danos/utils/pathutil"
)

const schemaTemplate = `
module test-configd-diff {
	namespace "urn:vyatta.com:test:configd-diff";
	prefix test;
	organization "Brocade Communications Systems, Inc.";
	contact
		"Brocade Communications Systems, Inc.
		 Postal: 130 Holger Way
		         San Jose, CA 95134
		 E-mail: support@Brocade.com
		 Web: www.brocade.com";
	revision 2014-12-29 {
		description "Test schema for configd";
	}
	%s
}
`

func compare(t1, t2 *data.Node, st schema.Tree, spath string) string {
	dtree := diff.NewNode(t1, t2, st, nil)
	dtree = dtree.Descendant(pathutil.Makepath(spath))
	return fmt.Sprintf("%s\n", dtree.Serialize(false))
}

func getDiff(t *testing.T, oldCfg, newCfg, schema string) string {
	sch := bytes.NewBufferString(fmt.Sprintf(schemaTemplate, schema))
	st, err := GetConfigSchema(sch.Bytes())
	if err != nil {
		t.Fatalf("Unable to get schema tree: %s", err.Error())
		return ""
	}

	old, err, _ := load.LoadString("oldCfg", oldCfg, st)
	if err != nil {
		t.Fatalf("Unable to load oldCfg: %s", err.Error())
		return ""
	}

	new, err, _ := load.LoadString("newCfg", newCfg, st)
	if err != nil {
		t.Fatalf("Unable to load newCfg: %s", err.Error())
		return ""
	}

	return compare(new, old, st, "")
}

const ordByUserListSchema = `
container testCont {
	list aList {
		ordered-by user;
		key name;
		leaf name {
			type string;
		}
		leaf aLeaf {
			type string;
		}
		leaf anotherLeaf {
			type string;
		}
	}
}`

// Different config variants
var A1_Cfg = Cont("testCont",
	List("aList",
		ListEntry("A",
			Leaf("aLeaf", "One"))))

var B1_Cfg = Cont("testCont",
	List("aList",
		ListEntry("B",
			Leaf("aLeaf", "One"))))

var A1B1_Cfg = Cont("testCont",
	List("aList",
		ListEntry("A",
			Leaf("aLeaf", "One")),
		ListEntry("B",
			Leaf("aLeaf", "One"))))

var B1A1_Cfg = Cont("testCont",
	List("aList",
		ListEntry("B",
			Leaf("aLeaf", "One")),
		ListEntry("A",
			Leaf("aLeaf", "One"))))

var A1B2_Cfg = Cont("testCont",
	List("aList",
		ListEntry("A",
			Leaf("aLeaf", "One")),
		ListEntry("B",
			Leaf("aLeaf", "Two"))))

var A2B1_Cfg = Cont("testCont",
	List("aList",
		ListEntry("A",
			Leaf("aLeaf", "Two")),
		ListEntry("B",
			Leaf("aLeaf", "One"))))

// Ordered List tests

func TestOrdByUserListChangeLastEntry(t *testing.T) {
	expect := FormatAsDiff(
		Cont("testCont",
			List("aList",
				ListEntry("A",
					Leaf("aLeaf", "One")),
				ListEntry("B",
					Rem(Leaf("aLeaf", "One")),
					Add(Leaf("aLeaf", "Two"))))))

	actual := getDiff(t, A1B1_Cfg, A1B2_Cfg, ordByUserListSchema)

	assert.CheckStringDivergence(t, expect, actual)
}

func TestOrdByUserListChangeFirstEntry(t *testing.T) {
	expect := FormatAsDiff(
		Cont("testCont",
			List("aList",
				ListEntry("A",
					Rem(Leaf("aLeaf", "One")),
					Add(Leaf("aLeaf", "Two"))),
				ListEntry("B",
					Leaf("aLeaf", "One")))))

	actual := getDiff(t, A1B1_Cfg, A2B1_Cfg, ordByUserListSchema)

	assert.CheckStringDivergence(t, expect, actual)
}

func TestOrdByUserListDeleteFirstEntry(t *testing.T) {
	expect := FormatAsDiff(
		Cont("testCont",
			List("aList",
				Rem(ListEntry("A",
					Leaf("aLeaf", "One"))),
				Add(ListEntry("B",
					Leaf("aLeaf", "One"))),
				Rem(ListEntry("B",
					Leaf("aLeaf", "One"))))))

	actual := getDiff(t, A1B1_Cfg, B1_Cfg, ordByUserListSchema)

	assert.CheckStringDivergence(t, expect, actual)
}

func TestOrdByUserListDeleteLastEntry(t *testing.T) {
	expect := FormatAsDiff(
		Cont("testCont",
			List("aList",
				ListEntry("A",
					Leaf("aLeaf", "One")),
				Rem(ListEntry("B",
					Leaf("aLeaf", "One"))))))

	actual := getDiff(t, A1B1_Cfg, A1_Cfg, ordByUserListSchema)

	assert.CheckStringDivergence(t, expect, actual)
}

func TestOrdByUserListSwapEntry(t *testing.T) {
	expect := FormatAsDiff(
		Cont("testCont",
			List("aList",
				Rem(ListEntry("A",
					Leaf("aLeaf", "One"))),
				Add(ListEntry("B",
					Leaf("aLeaf", "One"))),
				Rem(ListEntry("B",
					Leaf("aLeaf", "One"))),
				Add(ListEntry("A",
					Leaf("aLeaf", "One"))))))

	actual := getDiff(t, A1B1_Cfg, B1A1_Cfg, ordByUserListSchema)

	assert.CheckStringDivergence(t, expect, actual)
}

// Ordered-by-user leaf-list tests

const ordByUserLeafListSchema = `
container testCont {
	leaf-list aLeafList {
		ordered-by user;
		type string;
    }
}`

var LL_1_Cfg = Cont("testCont",
	LeafList("aLeafList",
		LeafListEntry("One")))

var LL_2_Cfg = Cont("testCont",
	LeafList("aLeafList",
		LeafListEntry("Two")))

var LL_12_Cfg = Cont("testCont",
	LeafList("aLeafList",
		LeafListEntry("One"),
		LeafListEntry("Two")))

var LL_13_Cfg = Cont("testCont",
	LeafList("aLeafList",
		LeafListEntry("One"),
		LeafListEntry("Three")))

var LL_21_Cfg = Cont("testCont",
	LeafList("aLeafList",
		LeafListEntry("Two"),
		LeafListEntry("One")))

func TestOrdByUserLeafListChangeLastEntry(t *testing.T) {
	expect := FormatAsDiff(
		Cont("testCont",
			LeafList("aLeafList",
				LeafListEntry("One"),
				Rem(LeafListEntry("Two")),
				Add(LeafListEntry("Three")))))

	actual := getDiff(t, LL_12_Cfg, LL_13_Cfg, ordByUserLeafListSchema)

	assert.CheckStringDivergence(t, expect, actual)
}

func TestOrdByUserLeafListDeleteFirstEntry(t *testing.T) {
	expect := FormatAsDiff(
		Cont("testCont",
			LeafList("aLeafList",
				Rem(LeafListEntry("One")),
				Add(LeafListEntry("Two")),
				Rem(LeafListEntry("Two")))))

	actual := getDiff(t, LL_12_Cfg, LL_2_Cfg, ordByUserLeafListSchema)

	assert.CheckStringDivergence(t, expect, actual)
}

func TestOrdByUserLeafListDeleteLastEntry(t *testing.T) {
	expect := FormatAsDiff(
		Cont("testCont",
			LeafList("aLeafList",
				LeafListEntry("One"),
				Rem(LeafListEntry("Two")))))

	actual := getDiff(t, LL_12_Cfg, LL_1_Cfg, ordByUserLeafListSchema)

	assert.CheckStringDivergence(t, expect, actual)
}

func TestOrdByUserLeafListSwapEntry(t *testing.T) {
	expect := FormatAsDiff(
		Cont("testCont",
			LeafList("aLeafList",
				Rem(LeafListEntry("One")),
				Add(LeafListEntry("Two")),
				Rem(LeafListEntry("Two")),
				Add(LeafListEntry("One")))))

	actual := getDiff(t, LL_12_Cfg, LL_21_Cfg, ordByUserLeafListSchema)

	assert.CheckStringDivergence(t, expect, actual)
}
