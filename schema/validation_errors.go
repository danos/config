// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"os"
	"strings"

	"github.com/danos/mgmterror"
	"github.com/danos/utils/pathutil"
	spawn "os/exec"
)

func configdEnv(sid string, path []string, action string, cact string) []string {
	var env = []string{
		"vyatta_htmldir=/opt/vyatta/share/html",
		"vyatta_datadir=/opt/vyatta/share",
		"vyatta_op_templates=/opt/vyatta/share/vyatta-op/templates",
		"vyatta_sysconfdir=/opt/vyatta/etc",
		"vyatta_sharedstatedir=/opt/vyatta/com",
		"vyatta_sbindir=/opt/vyatta/sbin",
		"vyatta_cfg_templates=/opt/vyatta/share/vyatta-cfg/templates",
		"vyatta_bindir=/opt/vyatta/bin",
		"vyatta_libdir=/opt/vyatta/lib",
		"vyatta_localstatedir=/opt/vyatta/var",
		"vyatta_libexecdir=/opt/vyatta/libexec",
		"vyatta_prefix=/opt/vyatta",
		"vyatta_datarootdir=/opt/vyatta/share",
		"vyatta_configdir=/opt/vyatta/config",
		"vyatta_infodir=/opt/vyatta/share/info",
		"vyatta_localedir=/opt/vyatta/share/locale",
		"PATH=/usr/local/bin:/usr/bin:/bin:/usr/local/sbin:/usr/sbin:/sbin:/opt/vyatta/bin:/opt/vyatta/bin/sudo-users:/opt/vyatta/sbin",
		"PERL5LIB=/opt/vyatta/share/perl5",
	}

	if sid != "" {
		env = append(env, "VYATTA_CONFIG_SID="+sid)
	}
	if cact != "" {
		env = append(env, "COMMIT_ACTION="+cact)
	}
	env = append(env, "CONFIGD_PATH="+pathutil.Pathstr(path))
	env = append(env, "CONFIGD_EXT="+action)
	return env
}

func execCmd(sid, path, c string) (string, error) {

	var env []string
	env = append(env, os.Environ()...)
	env = append(env, configdEnv(sid, pathutil.Makepath(path), "syntax", "")...)

	var interpreter string
	if sid == "" {
		interpreter = "/bin/bash"
	} else {
		interpreter = "/opt/vyatta/bin/cliexec"
	}

	cmd := spawn.Command(interpreter, "-c", c)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*spawn.ExitError); !ok {
			return "", err
		}
	}
	if !cmd.ProcessState.Success() {
		cerr := mgmterror.NewOperationFailedApplicationError()
		cerr.Path = path
		cerr.Message = string(out)
		return "", cerr
	}
	return string(out), nil
}

func errorMessage(ctx ValidateCtx, msg string) string {
	// It's extremely costly on low-end platforms to call echo on every pattern
	// in the interface type union until you get a match, on each and every
	// interface, only to always throw it away (we ignore union errors as the
	// next type might match ...).  So, unless the configd:error-message really
	// needs expansion, don't do it!
	if !strings.Contains(msg, "@") {
		return msg
	}
	// If the following Replacer is modified, you may need to give
	// some attention to the one in:
	// configd/cmd/cfgcli/completefns.go: getCompReply()
	//
	// Escape backslashes and double quotes in the error message
	// to ensure they display correctly post bash processing
	escapedMessage := strings.NewReplacer(
		`\`, `\\`, `"`, `\"`, `<`, `\<`, `>`, `\>`).Replace(msg)
	out, _ := execCmd(ctx.Sid, ctx.Path, "echo -ne "+escapedMessage)
	return out
}

func newRangeError(ctx ValidateCtx, path []string, msg string) error {
	merr := mgmterror.NewInvalidValueApplicationError()
	merr.Message = errorMessage(ctx, msg)
	merr.Path = pathutil.Pathstr(path)
	return merr
}
