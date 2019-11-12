// Copyright (c) 2018-2019, AT&T Intellectual Property Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package acmd

import (
	"testing"

	"github.com/danos/encoding/rfc7951"
)

const sample = `{
   "vyatta-system-v1:system" : {
      "vyatta-system-acm-v1:acm" : {
         "vyatta-system-acm-configd-v1:create-default" : "allow",
         "vyatta-system-acm-configd-v1:notification-default" : "allow",
         "vyatta-system-acm-configd-v1:notification-ruleset" : {
            "rule" : [
               {
                  "module-name" : "*",
                  "path" : null,
                  "notification-name" : "*",
                  "operation" : "*",
                  "rpc-name" : null,
                  "action" : "deny",
                  "group" : [
                     "vyattaop"
                  ],
                  "tagnode" :399 
               },
               {
                  "module-name" : "*",
                  "path" : null,
                  "notification-name" : "vyatta-ifmgr-v1:interface-state",
                  "operation" : "*",
                  "rpc-name" : null,
                  "action" : "allow",
                  "group" : [
                     "vyattaop"
                  ],
                  "tagnode" : 5000
               },
               {
                  "module-name" : "vyatta-op-v1",
                  "path" : null,
                  "notification-name" : "*",
                  "group" : [
                     "vyattaop"
                  ],
                  "action" : "allow",
                  "tagnode" : 9999,
                  "operation" : "*",
                  "rpc-name" :null 
               }
            ]
         },
         "vyatta-system-acm-configd-v1:rpc-ruleset" : {
            "rule" : [
               {
                  "module-name" : "*",
                  "path" : null,
                  "notification-name" : null,
                  "operation" : "*",
                  "rpc-name" : "*",
                  "action" : "deny",
                  "group" : [
                     "vyattaop"
                  ],
                  "tagnode" :399 
               },
               {
                  "module-name" : "*",
                  "path" : null,
                  "notification-name" : null,
                  "operation" : "*",
                  "rpc-name" : "vyatta-op-v1:ping",
                  "action" : "allow",
                  "group" : [
                     "vyattaop"
                  ],
                  "tagnode" : 5000
               },
               {
                  "module-name" : "vyatta-op-v1",
                  "path" : null,
                  "notification-name" : null,
                  "group" : [
                     "vyattaop"
                  ],
                  "action" : "allow",
                  "tagnode" : 9999,
                  "operation" : "*",
                  "rpc-name" :"*" 
               }
            ]
         },
         "vyatta-system-acm-configd-v1:read-default" : "allow",
         "vyatta-system-acm-configd-v1:update-default" : "allow",
         "vyatta-system-acm-configd-v1:rpc-default" : "allow",
         "vyatta-system-acm-configd-v1:enable" : [
            null
         ],
         "vyatta-system-acm-configd-v1:delete-default" : "allow"
      }
   }
}`

const expect = `<!DOCTYPE busconfig PUBLIC
 "-//freedesktop//DTD D-BUS Bus Configuration 1.0//EN"
  "http://www.freedesktop.org/standards/dbus/1.0/busconfig.dtd">
 <busconfig>
    <policy context="default">
       <allow send_type="method_call"></allow>
       <allow receive_type="signal"></allow>
    </policy>
    <policy group="vyattaop">
       <allow receive_interface="yang.module.VyattaOpV1.Notification"></allow>
       <allow receive_interface="yang.module.VyattaIfmgrV1.Notification" receive_member="InterfaceState"></allow>
       <deny receive_type="signal" receive_interface="*"></deny>
       <allow send_interface="yang.module.VyattaOpV1.RPC"></allow>
       <allow send_interface="yang.module.VyattaOpV1.RPC" send_member="Ping"></allow>
       <deny send_type="method_call" send_interface="*"></deny>
    </policy>
 </busconfig>`

func TestTranslate(t *testing.T) {
	var aacfg AcmdConfig

	err := rfc7951.Unmarshal([]byte(sample), &aacfg)
	if err != nil {
		t.Fatalf("Unmarshal failed %s\n", err.Error())
	}

	if aacfg.System.Acm.AcmV1Config == nil {
		t.Fatalf("Sample data Unmarshal failed\n")
	}

	out, err := aacfg.System.Acm.AcmV1Config.translateToPolicy()
	if err != nil {
		t.Fatalf("Policy Rule translation failed %s\n", err.Error())
	}
	output := string(out)
	if output != expect {
		t.Fatalf("Expected:\n%s\nGot:\n%s\n", expect, output)
	}
}
