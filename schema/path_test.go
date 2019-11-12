// Copyright (c) 2018-2019, AT&T Intellectual Property. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/danos/utils/pathutil"
	"testing"
)

func checkPathAttrs(t *testing.T, attrs *pathutil.PathAttrs, exp_attrs *pathutil.PathAttrs) {
	if attrs == nil {
		if exp_attrs != nil {
			t.Fatalf("Expected PathAttrs")
		}
		return
	} else if exp_attrs == nil {
		t.Fatalf("Unexpectedly got PathAttrs")
	}

	if len(attrs.Attrs) != len(exp_attrs.Attrs) {
		t.Fatalf("Incorrect PathAttrs len: expected %v, got %v",
			len(exp_attrs.Attrs), len(attrs.Attrs))
	}

	for i, v := range attrs.Attrs {
		if v != exp_attrs.Attrs[i] {
			t.Fatalf("PathElementAttr mismatch at index %v\nexpected: %v\ngot: %v",
				i, exp_attrs.Attrs[i], v)
		}
	}
}

const simple_schema_snippet = `
	leaf test-leaf {
		configd:secret "true";
		type string;
	}`

const complex_schema_snippet = `
	container test-container {
		list test-list {
			key test-key;
			leaf test-key {
				type string;
			}
			container sub-container {
				leaf-list test-leaf-list {
					configd:secret "true";
					type string;
				}
			}
			leaf test-leaf {
				configd:secret "true";
				type string;
			}
		}
	}`

func TestAttrsForPathSuccessSimple(t *testing.T) {
	st := buildSchema(t, simple_schema_snippet)
	attrs := AttrsForPath(st, []string{"test-leaf", "bar"})

	exp_attrs := pathutil.NewPathAttrs()
	exp_attrs.Attrs = append(exp_attrs.Attrs, pathutil.PathElementAttrs{Secret: false})
	exp_attrs.Attrs = append(exp_attrs.Attrs, pathutil.PathElementAttrs{Secret: true})

	checkPathAttrs(t, attrs, &exp_attrs)
}

func TestAttrsForPathFailSimple(t *testing.T) {
	st := buildSchema(t, simple_schema_snippet)
	attrs := AttrsForPath(st, []string{"foo", "bar", "baz"})
	checkPathAttrs(t, attrs, nil)
}

func TestAttrsForPathSuccessComplex(t *testing.T) {
	st := buildSchema(t, complex_schema_snippet)
	attrs := AttrsForPath(st, []string{"test-container"})

	exp_attrs := pathutil.NewPathAttrs()
	exp_attrs.Attrs = append(exp_attrs.Attrs, pathutil.PathElementAttrs{Secret: false})
	checkPathAttrs(t, attrs, &exp_attrs)

	attrs = AttrsForPath(st, []string{"test-container", "test-list", "a", "sub-container", "test-leaf-list", "b"})
	exp_attrs.Attrs = nil
	for _, v := range []bool{false, false, false, false, false, true} {
		exp_attrs.Attrs = append(exp_attrs.Attrs, pathutil.PathElementAttrs{Secret: v})
	}
	checkPathAttrs(t, attrs, &exp_attrs)

	attrs = AttrsForPath(st, []string{"test-container", "test-list", "a", "test-leaf", "c"})
	exp_attrs.Attrs = nil
	for _, v := range []bool{false, false, false, false, true} {
		exp_attrs.Attrs = append(exp_attrs.Attrs, pathutil.PathElementAttrs{Secret: v})
	}
	checkPathAttrs(t, attrs, &exp_attrs)
}

func TestAttrsForPathFailComplex(t *testing.T) {
	st := buildSchema(t, complex_schema_snippet)
	attrs := AttrsForPath(st, []string{"test-container", "test-list", "b", "baz"})
	checkPathAttrs(t, attrs, nil)

	attrs = AttrsForPath(st, []string{"test-container", "test-list", "c", "sub-container", "foo", "bar"})
	checkPathAttrs(t, attrs, nil)
}
