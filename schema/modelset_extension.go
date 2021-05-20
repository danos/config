// Copyright (c) 2017-2021, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	yang "github.com/danos/yang/schema"
)

type ModelSet interface {
	yang.ModelSet
	ExtendedNode
	PathDescendant([]string) *TmplCompat
	OpdPathDescendant([]string) *TmplCompat
}

type modelSet struct {
	yang.ModelSet
	*extensions
	*state
}

// Compile time check that the concrete type meets the interface
var _ ModelSet = (*modelSet)(nil)

func (c *CompilationExtensions) ExtendModelSet(
	m yang.ModelSet,
) (yang.ModelSet, error) {

	ext := newExtend(nil)
	return &modelSet{
			m, ext, newState(m, ext)},
		nil
}
