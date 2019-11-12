// Copyright (c) 2017-2019, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/danos/yang/parse"
)

var configdcardinality = map[parse.NodeType]parse.Cardinality{
	parse.NodeConfigdCallRpc:      {'0', '1'},
	parse.NodeConfigdHelp:         {'0', '1'},
	parse.NodeConfigdValidate:     {'0', 'n'},
	parse.NodeConfigdNormalize:    {'0', '1'},
	parse.NodeConfigdSyntax:       {'0', 'n'},
	parse.NodeConfigdPriority:     {'0', '1'},
	parse.NodeConfigdAllowed:      {'0', '1'},
	parse.NodeConfigdBegin:        {'0', 'n'},
	parse.NodeConfigdEnd:          {'0', 'n'},
	parse.NodeConfigdCreate:       {'0', 'n'},
	parse.NodeConfigdDelete:       {'0', 'n'},
	parse.NodeConfigdUpdate:       {'0', 'n'},
	parse.NodeConfigdSubst:        {'0', 'n'},
	parse.NodeConfigdSecret:       {'0', '1'},
	parse.NodeConfigdErrMsg:       {'0', '1'},
	parse.NodeConfigdPHelp:        {'0', '1'},
	parse.NodeConfigdGetState:     {'0', 'n'},
	parse.NodeConfigdDeferActions: {'0', '1'},
	parse.NodeConfigdMust:         {'0', '1'},
	parse.NodeOpdHelp:             {'0', '1'},
	parse.NodeOpdAllowed:          {'0', '1'},
	parse.NodeOpdPatternHelp:      {'0', '1'},
}

var configdappliesto = map[parse.NodeType]map[parse.NodeType]struct{}{
	parse.NodeConfigdCallRpc: {
		parse.NodeRpc: struct{}{},
	},
	parse.NodeConfigdHelp: {
		parse.NodeGrouping:    struct{}{},
		parse.NodeContainer:   struct{}{},
		parse.NodeLeaf:        struct{}{},
		parse.NodeLeafList:    struct{}{},
		parse.NodeList:        struct{}{},
		parse.NodeChoice:      struct{}{},
		parse.NodeTypedef:     struct{}{},
		parse.NodeTyp:         struct{}{},
		parse.NodeBit:         struct{}{},
		parse.NodeEnum:        struct{}{},
		parse.NodeRefine:      struct{}{},
		parse.NodeOpdCommand:  struct{}{},
		parse.NodeOpdOption:   struct{}{},
		parse.NodeOpdArgument: struct{}{},
	},
	parse.NodeConfigdValidate: {
		parse.NodeGrouping:  struct{}{},
		parse.NodeContainer: struct{}{},
		parse.NodeLeaf:      struct{}{},
		parse.NodeLeafList:  struct{}{},
		parse.NodeList:      struct{}{},
		parse.NodeChoice:    struct{}{},
		parse.NodeRefine:    struct{}{},
		parse.NodeAugment:   struct{}{},
	},
	parse.NodeConfigdNormalize: {
		parse.NodeTyp: struct{}{},
	},
	parse.NodeConfigdSyntax: {
		parse.NodeTypedef: struct{}{},
		parse.NodeTyp:     struct{}{},
		parse.NodeBit:     struct{}{},
		parse.NodeEnum:    struct{}{},
		parse.NodeRefine:  struct{}{},
	},
	parse.NodeConfigdPriority: {
		parse.NodeGrouping:  struct{}{},
		parse.NodeContainer: struct{}{},
		parse.NodeLeaf:      struct{}{},
		parse.NodeLeafList:  struct{}{},
		parse.NodeList:      struct{}{},
		parse.NodeChoice:    struct{}{},
		parse.NodeRefine:    struct{}{},
	},
	parse.NodeConfigdAllowed: {
		parse.NodeGrouping:    struct{}{},
		parse.NodeContainer:   struct{}{},
		parse.NodeLeaf:        struct{}{},
		parse.NodeLeafList:    struct{}{},
		parse.NodeList:        struct{}{},
		parse.NodeChoice:      struct{}{},
		parse.NodeRefine:      struct{}{},
		parse.NodeOpdCommand:  struct{}{},
		parse.NodeOpdOption:   struct{}{},
		parse.NodeOpdArgument: struct{}{},
		// It is desirable to have configd:allowed apply to
		// types/typedefs too, but until it has an effect they
		// aren't listed here
	},
	parse.NodeConfigdBegin: {
		parse.NodeGrouping:  struct{}{},
		parse.NodeContainer: struct{}{},
		parse.NodeLeaf:      struct{}{},
		parse.NodeLeafList:  struct{}{},
		parse.NodeList:      struct{}{},
		parse.NodeChoice:    struct{}{},
		parse.NodeRefine:    struct{}{},
		parse.NodeAugment:   struct{}{},
	},
	parse.NodeConfigdEnd: {
		parse.NodeGrouping:  struct{}{},
		parse.NodeContainer: struct{}{},
		parse.NodeLeaf:      struct{}{},
		parse.NodeLeafList:  struct{}{},
		parse.NodeList:      struct{}{},
		parse.NodeChoice:    struct{}{},
		parse.NodeRefine:    struct{}{},
		parse.NodeAugment:   struct{}{},
	},
	parse.NodeConfigdCreate: {
		parse.NodeGrouping:  struct{}{},
		parse.NodeContainer: struct{}{},
		parse.NodeLeaf:      struct{}{},
		parse.NodeLeafList:  struct{}{},
		parse.NodeList:      struct{}{},
		parse.NodeChoice:    struct{}{},
		parse.NodeRefine:    struct{}{},
	},
	parse.NodeConfigdUpdate: {
		parse.NodeGrouping:  struct{}{},
		parse.NodeContainer: struct{}{},
		parse.NodeLeaf:      struct{}{},
		parse.NodeLeafList:  struct{}{},
		parse.NodeList:      struct{}{},
		parse.NodeChoice:    struct{}{},
		parse.NodeRefine:    struct{}{},
	},
	parse.NodeConfigdSubst: {
		parse.NodeGrouping:  struct{}{},
		parse.NodeContainer: struct{}{},
		parse.NodeLeaf:      struct{}{},
		parse.NodeLeafList:  struct{}{},
		parse.NodeList:      struct{}{},
		parse.NodeChoice:    struct{}{},
		parse.NodeRefine:    struct{}{},
	},
	parse.NodeConfigdDelete: {
		parse.NodeGrouping:  struct{}{},
		parse.NodeContainer: struct{}{},
		parse.NodeLeaf:      struct{}{},
		parse.NodeLeafList:  struct{}{},
		parse.NodeList:      struct{}{},
		parse.NodeChoice:    struct{}{},
		parse.NodeRefine:    struct{}{},
	},
	parse.NodeConfigdSecret: {
		parse.NodeTyp:      struct{}{},
		parse.NodeTypedef:  struct{}{},
		parse.NodeLeaf:     struct{}{},
		parse.NodeLeafList: struct{}{},
		parse.NodeBit:      struct{}{},
		parse.NodeRefine:   struct{}{},
	},
	parse.NodeConfigdPHelp: {
		parse.NodeLeaf:     struct{}{},
		parse.NodeLeafList: struct{}{},
		parse.NodeTypedef:  struct{}{},
		parse.NodeTyp:      struct{}{},
		parse.NodeBit:      struct{}{},
	},
	parse.NodeConfigdErrMsg: {
		parse.NodePattern: struct{}{},
		parse.NodeRange:   struct{}{},
		parse.NodeLength:  struct{}{},
	},
	parse.NodeConfigdGetState: {
		parse.NodeContainer: struct{}{},
		parse.NodeLeaf:      struct{}{},
		parse.NodeLeafList:  struct{}{},
		parse.NodeList:      struct{}{},
		parse.NodeAugment:   struct{}{},
	},
	parse.NodeConfigdDeferActions: {
		parse.NodeContainer: struct{}{},
		parse.NodeLeaf:      struct{}{},
		parse.NodeLeafList:  struct{}{},
		parse.NodeList:      struct{}{},
	},
	parse.NodeConfigdMust: {
		parse.NodeMust: struct{}{},
	},
	parse.NodeOpdAllowed: {
		parse.NodeGrouping:    struct{}{},
		parse.NodeRefine:      struct{}{},
		parse.NodeOpdCommand:  struct{}{},
		parse.NodeOpdOption:   struct{}{},
		parse.NodeOpdArgument: struct{}{},
	},
	parse.NodeOpdHelp: {
		parse.NodeGrouping:    struct{}{},
		parse.NodeTypedef:     struct{}{},
		parse.NodeTyp:         struct{}{},
		parse.NodeBit:         struct{}{},
		parse.NodeEnum:        struct{}{},
		parse.NodeRefine:      struct{}{},
		parse.NodeOpdCommand:  struct{}{},
		parse.NodeOpdOption:   struct{}{},
		parse.NodeOpdArgument: struct{}{},
	},
	parse.NodeOpdPatternHelp: {
		parse.NodeTypedef: struct{}{},
		parse.NodeTyp:     struct{}{},
		parse.NodeBit:     struct{}{},
	},
}

func configdCardinality(ntype parse.NodeType) map[parse.NodeType]parse.Cardinality {

	card := make(map[parse.NodeType]parse.Cardinality, len(configdcardinality))

	for k, v := range configdcardinality {
		if _, ok := configdappliesto[k][ntype]; ok {
			card[k] = v
		}
	}

	return card
}
