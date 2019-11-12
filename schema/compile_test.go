// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func assertErrorContains(t *testing.T, err error, expected string) {
	if err == nil {
		t.Errorf(
			"Unexpected success when parsing schema and expecting:\n  %s",
			expected)
		return
	}
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("Unexpected error output:\n    expect: %s\n    actual=%s",
			expected, err.Error())
	}
}

func expandExt(ext string) string {
	if ext != "" {
		ext = "configd:" + ext + " 'dummy';"
	}
	return ext
}

func compileOrdByUserListWithExtensions(listExt, leafExt string) (ModelSet, error) {

	listExt = expandExt(listExt)
	leafExt = expandExt(leafExt)

	list_text := fmt.Sprintf(
		`list testList {
             ordered-by user;
             key listKey;
             leaf listKey {
                 type string;
                 %s
             }
             %s
         }`, leafExt, listExt)
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate, list_text))

	return GetConfigSchema(schema_text.Bytes())
}

const onlyBeginOrEndAllowed = "only begin and end extensions allowed on ordered-by user list"

func TestRejectCreateOnOrdByUserList(t *testing.T) {
	_, err := compileOrdByUserListWithExtensions("create", "")
	assertErrorContains(t, err, onlyBeginOrEndAllowed)
}

func TestRejectUpdateOnOrdByUserList(t *testing.T) {
	_, err := compileOrdByUserListWithExtensions("update", "")
	assertErrorContains(t, err, onlyBeginOrEndAllowed)
}

func TestRejectDeleteOnOrdByUserList(t *testing.T) {
	_, err := compileOrdByUserListWithExtensions("delete", "")
	assertErrorContains(t, err, onlyBeginOrEndAllowed)
}

const noAction = "action extension not allowed in ordered-by user"

func TestRejectCreateOnOrdByUserListLeaf(t *testing.T) {
	_, err := compileOrdByUserListWithExtensions("", "create")
	assertErrorContains(t, err, noAction)
}

func TestRejectUpdateOnOrdByUserListLeaf(t *testing.T) {
	_, err := compileOrdByUserListWithExtensions("", "update")
	assertErrorContains(t, err, noAction)
}
func TestRejectDeleteOnOrdByUserListLeaf(t *testing.T) {
	_, err := compileOrdByUserListWithExtensions("", "delete")
	assertErrorContains(t, err, noAction)
}

func TestRejectBeginOnOrdByUserListLeaf(t *testing.T) {
	_, err := compileOrdByUserListWithExtensions("", "begin")
	assertErrorContains(t, err, noAction)
}

func TestRejectEndOnOrdByUserListLeaf(t *testing.T) {
	_, err := compileOrdByUserListWithExtensions("", "end")
	assertErrorContains(t, err, noAction)
}
