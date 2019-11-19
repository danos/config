// Copyright (c) 2019, AT&T Intellectual Property.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0
//

package union

func redactListEntry(n *ListEntry, hideSecrets bool) bool {
	sch := n.Schema
	keynode := sch.SchemaChild(sch.Keys()[0])
	if hideSecrets && keynode.ConfigdExt().Secret {
		return true
	}
	return false
}
