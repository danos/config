// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"time"

	"github.com/danos/mgmterror"
	"github.com/danos/utils/exec"
	"github.com/danos/yang/data/datanode"
	yang "github.com/danos/yang/schema"
)

func init() {
	exec.NewExecError = func(path []string, err string) error {
		return mgmterror.NewExecError(path, err)
	}
}

func ValidateSchemaWithLog(
	sn Node,
	dn datanode.DataNode,
	debug bool,
	logFn commitTimeLogFn,
) (
	[]*exec.Output,
	[]error,
	bool,
) {
	yangValStart := time.Now()
	outs, errs, ok := yang.ValidateSchema(sn, dn, debug)
	if !ok {
		return outs, errs, ok
	}
	if logFn != nil {
		logFn("YANG validation", yangValStart)
	}

	if ms, ok := sn.(ModelSet); ok {
		service_errors := ms.ServiceValidation(dn, logFn)
		if len(service_errors) > 0 {
			ok = false
			errs = append(errs, service_errors...)
		}
	}

	return outs, errs, ok
}

func ValidateSchema(sn Node, dn datanode.DataNode, debug bool) (
	[]*exec.Output, []error, bool) {

	return ValidateSchemaWithLog(sn, dn, debug, nil)
}
