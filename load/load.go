// Copyright (c) 2017-2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0
package load

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/danos/config/data"
	"github.com/danos/config/parse"
	"github.com/danos/config/schema"
	"github.com/danos/config/union"
	"github.com/danos/utils/pathutil"
)

func Load(fname string, st schema.ModelSet) (*data.Node, error, []error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err, nil
	}
	defer f.Close()
	return LoadFile(fname, f, st)
}

func LoadNoValidate(fname string) (*data.Node, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadFileNoValidate(fname, f)
}

func LoadFile(
	fname string,
	f io.Reader,
	st schema.ModelSet,
) (*data.Node, error, []error) {

	text, err := ioutil.ReadAll(f)
	if err != nil && err != io.EOF {
		return nil, err, nil
	}
	return LoadString(fname, string(text), st)
}

func LoadFileNoValidate(fname string, f io.Reader) (*data.Node, error) {
	text, err := ioutil.ReadAll(f)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return LoadStringNoValidate(fname, string(text))
}

// LoadString - create tree from given config, noting any rejected paths
//
// Rejected paths are not a fatal error so we return them separately to
// other, fatal, errors.
func LoadString(
	name, text string,
	st schema.ModelSet,
) (*data.Node, error, []error) {
	return loadStringInternal(name, text, st, true)
}

// LoadStringNoNormalize - load from string, ignoring any normalization.
//
// Useful for XYANG tool and any other environment where normalization
// scripts may not be present.
func LoadStringNoNormalize(
	name, text string,
	st schema.ModelSet,
) (*data.Node, error, []error) {
	return loadStringInternal(name, text, st, false)
}

func loadStringInternal(
	name, text string,
	st schema.ModelSet,
	normalize bool,
) (*data.Node, error, []error) {
	paths, err := getPaths(name, text)
	if err != nil {
		return nil, err, []error{}
	}

	//Build a data.Node representation of these paths using
	//the UnionTree abstraction
	can, run := data.New("root"), data.New("root")
	ut := union.NewNode(can, run, st, nil, 0)
	var invalidPaths []error
	for _, path := range paths {
		normalizedPath := path
		if normalize {
			normalizedPath, err = schema.NormalizePath(st, path)
			if err != nil {
				invalidPaths = append(invalidPaths, err)
				continue
			}
		}
		if err = ut.Set(nil, normalizedPath); err != nil {
			invalidPaths = append(invalidPaths, err)
		}
	}
	return can, nil, invalidPaths
}

// LoadStringNoValidate - create tree from given config, without performing any
// validation at all
//
// Note that this doesn't even check that the config conforms to the schema
func LoadStringNoValidate(name, text string) (*data.Node, error) {
	paths, err := getPaths(name, text)
	if err != nil {
		return nil, err
	}

	tree := data.New("root")
	for _, path := range paths {
		tree.SetNoValidate(path)
	}
	return tree, nil
}

func getPaths(name, text string) ([][]string, error) {
	t, err := parse.Parse(name, text)
	if err != nil {
		return nil, err
	}
	//Generate paths from parsed nodes
	paths := make([][]string, 0)
	var buildpaths func(n *parse.Node, path []string)
	buildpaths = func(n *parse.Node, path []string) {
		if n.HasArg {
			path = append(path, n.Id, n.Arg)
		} else {
			path = append(path, n.Id)
		}
		if len(n.Children) == 0 {
			paths = append(paths, pathutil.Copypath(path))
		}
		for _, ch := range n.Children {
			buildpaths(ch, path)
		}
	}

	for _, ch := range t.Root.Children {
		buildpaths(ch, []string{})
	}
	return paths, nil
}
