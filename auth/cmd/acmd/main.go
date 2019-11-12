//Copyright (c) 2018-2019, AT&T Intellectual Property.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"github.com/danos/config/auth/acmd"
	"github.com/danos/vci"
)

func main() {
	state := acmd.NewState()
	config := acmd.NewConfig(state)
	comp := vci.NewComponent("net.vyatta.vci.acmd")
	comp.Model("net.vyatta.vci.acmd.v1").
		Config(config).
		State(state)
	comp.Run()
	comp.Wait()
}
