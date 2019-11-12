// Copyright (c) 2018-2019, AT&T Intellectual Property. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/danos/utils/pathutil"
)

func shouldRedact(sn Node) bool {
	switch sn.(type) {
	case LeafValue, ListEntry:
		return sn.ConfigdExt().Secret
	}

	return false
}

func AttrsForPath(st Node, path []string) *pathutil.PathAttrs {
	var sn Node = st
	attrs := pathutil.NewPathAttrs()

	for _, v := range path {
		sn = sn.SchemaChild(v)
		if sn == nil {
			// Don't return any attributes for an invalid path
			return nil
		}
		attr := pathutil.NewPathElementAttrs()
		attr.Secret = shouldRedact(sn)
		attrs.Attrs = append(attrs.Attrs, attr)
	}
	return &attrs
}
