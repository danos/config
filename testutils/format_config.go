// Copyright (c) 2018-2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// Utilities used by configd unit tests to format configuration

package testutils

import (
	"fmt"
	"strings"
)

// Set of helper functions to produce correctly formatted config and that
// isolates test code that relies on correctly formatted config from subsequent
// format changes so that we are testing content not format.
func Prefix(entry, pfx string) string {
	tmp := strings.Replace(entry, "\n", "\n"+pfx, -1)
	return pfx + tmp[:len(tmp)-len(pfx)]
}

func Tab(entry string) string {
	return Prefix(entry, "\t")
}

func Add(entry string) string {
	return Prefix(entry, "+")
}

func Rem(entry string) string {
	return Prefix(entry, "-")
}

// Initially the +/- for changed lines get added right in front of the
// element being changed.  This function pulls them to the front of the line
// and inserts a leading space on unchanged lines.  Completely blank lines
// (other than leading tabs) do NOT get a leading space.
func FormatAsDiff(entry string) (diffs string) {
	lines := strings.Split(entry, "\n")
	for _, line := range lines {
		trimmed := strings.Trim(line, "\t")
		if len(trimmed) > 0 {
			if trimmed[0] == '+' || trimmed[0] == '-' {
				// Iteratively move + or - ahead of tabs
				for line[0] == '\t' {
					line = strings.Replace(line, "\t+", "+\t", 1)
					line = strings.Replace(line, "\t-", "-\t", 1)
				}
			} else {
				line = " " + line
			}
		}
		diffs += line + "\n"
	}
	return diffs
}

func FormatAsDiffNoTrailingLine(entry string) (diffs string) {
	return strings.TrimSuffix(FormatAsDiff(entry), "\n")
}

func FormatCtxDiffHunk(ctxPath, hunk string) string {
	ctx := "[edit"
	if len(ctxPath) > 0 {
		ctx += " " + ctxPath
	}
	ctx += "]\n"
	return ctx + FormatAsDiffNoTrailingLine(hunk)
}

// ListEntries and Containers are handled exactly the same way.
func contOrListEntry(name string, entries []string) (retStr string) {
	retStr = name
	if len(entries) == 0 {
		return retStr + "\n"
	}

	retStr += " {\n"
	for _, entry := range entries {
		retStr += Tab(entry)
	}
	retStr += "}\n"
	return retStr
}

func Root(rootEntries ...string) (rootStr string) {
	return strings.Join(rootEntries, "")
}

func Cont(name string, contEntries ...string) (contStr string) {
	return contOrListEntry(name, contEntries)
}

func ListEntry(name string, leaves ...string) (listEntryStr string) {
	return contOrListEntry(name, leaves)
}

func listOrLeafList(name string, entries []string) (retStr string) {
	for _, entry := range entries {
		// Deal with +/- prefix
		if entry[0] == '+' || entry[0] == '-' {
			retStr += fmt.Sprintf("%c%s %s", entry[0], name, entry[1:])
		} else {
			retStr += name + " " + entry
		}
	}
	return retStr
}

func List(name string, listEntries ...string) (listStr string) {
	return listOrLeafList(name, listEntries)
}

func LeafList(name string, leafListEntries ...string) (leafListStr string) {
	return listOrLeafList(name, leafListEntries)
}

func LeafListEntry(name string) string { return name + "\n" }

func Leaf(name, value string) string { return name + " " + value + "\n" }

func EmptyLeaf(name string) string { return name + "\n" }
