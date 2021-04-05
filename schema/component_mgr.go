// Copyright (c) 2021, AT&T Intellectual Property. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0
//

// The ComponentManager deals with communications between configd and the
// VCI components, and with the components' status (active / running / stopped
// etc).

package schema

type OperationsManager interface {
	Dial() error
	SetConfigForModel(string, interface{}) error
	CheckConfigForModel(string, interface{}) error
	StoreConfigByModelInto(string, interface{}) error
	StoreStateByModelInto(string, interface{}) error
}

type ServiceManager interface {
	CloseSvcMgr()
	IsActive(name string) (bool, error)
}

// ComponentManager encapsulates bus operations to/from components, and service
// queries against the components' service status.
type ComponentManager interface {
	OperationsManager
	ServiceManager
}
