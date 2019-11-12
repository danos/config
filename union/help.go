// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package union

import (
	"github.com/danos/config/schema"
	"github.com/danos/utils/pathutil"
)

// Return set of help options, including:
//
// - schema help text if <fromSchema> is true
//
// - any existing configured values unless they are marked secret and we
//   don't have permission to view them.
//
func getHelpMap(auth Auther, n Node, path []string, fromSchema bool) map[string]string {
	var m map[string]string
	sch := n.GetSchema()
	if fromSchema {
		m = sch.(schema.ExtendedNode).HelpMap()
	} else {
		m = make(map[string]string)
	}

	if n.GetSchema().ConfigdExt().Secret {
		if !authorize(auth, pathutil.CopyAppend(path, n.Name()), "secrets") {
			return m
		}
	}

	children := n.Children()
	for _, ch := range children {
		if !authorize(auth, pathutil.CopyAppend(path, ch.Name()), "read") {
			continue
		}
		if ch.def() {
			continue
		}
		h := schema.GetHelp(ch.GetSchema())
		if _, ok := m[ch.Name()]; !ok {
			m[ch.Name()] = h
		}
	}
	return m
}

func (n *node) getHelp(auth Auther, fromSchema bool, path, curPath []string) (map[string]string, error) {
	m := make(map[string]string)
	err := n.walkPath(
		func(ch Node, hd string, tl []string) error {
			var err error
			m, err = ch.getHelp(auth, fromSchema, tl, append(curPath, hd))
			return err
		},
		func(ch Node, hd string, tl []string) error {
			if fromSchema {
				sch := schema.Descendant(n.GetSchema(), path)
				if sch == nil {
					return nil
				}
				m = sch.(schema.ExtendedNode).HelpMap()
			}
			return nil

		},
		func(last Node) error {
			m = getHelpMap(auth, last, curPath, fromSchema)
			return nil
		},
		path,
	)
	return m, err
}

func (n *node) GetHelp(auth Auther, fromSchema bool, path []string) (map[string]string, error) {
	if !fromSchema && !authorize(auth, path, "read") {
		return nil, autherr
	}
	return n.getHelp(auth, fromSchema, path, make([]string, 0, len(path)))
}
