// Copyright (c) 2017,2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/danos/mgmterror"
	"github.com/danos/utils/pathutil"
	spawn "os/exec"
)

func callNormalizationScript(script, input string) (string, error) {

	if script == "" {
		return input, nil
	}

	args := strings.Split(script, " ")
	cmd := spawn.Command(args[0], args[1:]...)
	cmd.Stdin = bytes.NewBufferString(input)

	out, err := cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*spawn.ExitError); !ok {
			cerr := mgmterror.NewOperationFailedApplicationError()
			cerr.Message = err.Error()
			return "", cerr
		}
	}

	if !cmd.ProcessState.Success() {
		cerr := mgmterror.NewOperationFailedApplicationError()
		cerr.Message = string(out)
		return "", cerr
	}

	values := strings.Split(string(out), "\n")
	if values != nil && len(values) > 0 {
		return string(values[0]), nil
	}
	return input, nil
}

func normalizeValue(sn Node, value string) (string, error) {

	// NOTE: Consider ensuring current value is valid before normalizing
	switch sn.(type) {
	case LeafValue, ListEntry:

		var script string
		typ := sn.Type()
		ext := typ.(hasExtensions).ConfigdExt()

		// Override any with local one
		if ext.Normalize != "" {
			script = ext.Normalize

		} else {
			// Override default with type specific one from union
			switch u := typ.(type) {
			case Union:
				typ = u.MatchType(nil, []string{}, value)
				if typ == nil {
					break
				}
				ext := typ.(hasExtensions).ConfigdExt()
				if ext.Normalize != "" {
					script = ext.Normalize
				}
			}
		}

		if newVal, err := callNormalizationScript(script, value); err != nil {
			return "", err
		} else {
			return newVal, nil
		}
	}

	return value, nil
}

func NormalizePath(st Node, ps []string) ([]string, error) {
	var sn Node = st

	for i, v := range ps {
		sn = sn.SchemaChild(v)
		if sn == nil {
			cerr := mgmterror.NewUnknownElementApplicationError(v)
			cerr.Path = pathutil.Pathstr(ps[:i])
			return nil, cerr
		}
		if newV, err := normalizeValue(sn, v); err != nil {
			cerr := mgmterror.NewInvalidValueApplicationError()
			cerr.Path = pathutil.Pathstr(ps[:i])
			cerr.Message = fmt.Sprintf("Error normalizing value: %s",
				err.Error())
			return nil, cerr
		} else {
			ps[i] = newV
		}
	}
	return ps, nil
}
