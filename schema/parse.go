// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/danos/yang/parse"
)

func Parse(name, text string) (*parse.Tree, error) {
	return parse.Parse(name, text, configdCardinality)
}
