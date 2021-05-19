// Copyright (c) 2021, AT&T Intellectual Property. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0
//

// The ComponentManager deals with communications between configd and the
// VCI components, and with the components' status (active / running / stopped
// etc).

package schema

import (
	"fmt"
	"time"

	"github.com/danos/vci/conf"
	"github.com/danos/yang/data/datanode"
	"github.com/danos/yang/data/encoding"
	yang "github.com/danos/yang/schema"
)

type OperationsManager interface {
	Dial() error
	SetConfigForModel(string, interface{}) error
	CheckConfigForModel(string, interface{}) error
	StoreConfigByModelInto(string, interface{}) error
	StoreStateByModelInto(string, interface{}) error
}

type ServiceManager interface {
	Close()
	IsActive(name string) (bool, error)
}

type componentMappings struct {
	modelSetName      string
	components        map[string]*component
	nsMap             map[string]string
	orderedComponents []string
	defaultComponent  string
}

func createComponentMappings(
	m yang.ModelSet,
	modelSetName string,
	compConfig []*conf.ServiceConfig,
) (*componentMappings, error) {

	modelToNamespaceMap, globalNSMap, defaultComponent, err :=
		getModelToNamespaceMapsForModelSet(
			m, modelSetName, compConfig)
	if err != nil {
		return nil, err
	}

	componentMap := getComponentMap(modelToNamespaceMap)

	orderedComponents, err := getOrderedComponentsList(
		modelSetName, defaultComponent, compConfig)
	if err != nil {
		return nil, err
	}

	if len(componentMap) != len(orderedComponents) {
		return nil, fmt.Errorf(
			"Mismatch between number of ordered (%d) "+
				"and unordered (%d) components.",
			len(orderedComponents), len(componentMap))
	}

	return &componentMappings{
			modelSetName:      modelSetName,
			components:        componentMap,
			nsMap:             globalNSMap,
			orderedComponents: orderedComponents,
			defaultComponent:  defaultComponent,
		},
		nil
}

type component struct {
	name      string
	modMap    map[string]struct{}
	setFilter func(s yang.Node, d datanode.DataNode,
		children []datanode.DataNode) bool
	checkMap    map[string]struct{}
	checkFilter func(s yang.Node, d datanode.DataNode,
		children []datanode.DataNode) bool
}

func (c *component) FilterSetTree(n Node, dn datanode.DataNode) []byte {
	filteredCandidate := yang.FilterTree(n, dn, c.setFilter)
	return encoding.ToRFC7951(n, filteredCandidate)
}

func (c *component) FilterCheckTree(n Node, dn datanode.DataNode) []byte {
	filteredCandidate := yang.FilterTree(n, dn, c.checkFilter)
	return encoding.ToRFC7951(n, filteredCandidate)
}

func (c *component) HasConfiguration(n Node, dn datanode.DataNode) bool {
	return string(c.FilterSetTree(n, dn)) != "{}"
}

// ComponentManager encapsulates bus operations to/from components, and service
// queries against the components' service status.
type ComponentManager interface {
	OperationsManager
	ServiceManager

	ComponentValidation(ModelSet, datanode.DataNode, commitTimeLogFn) []error
}

type compMgr struct {
	OperationsManager
	ServiceManager

	compMappings *componentMappings
}

var _ OperationsManager = (*compMgr)(nil)
var _ ServiceManager = (*compMgr)(nil)
var _ ComponentManager = (*compMgr)(nil)

func NewCompMgr(
	opsMgr OperationsManager,
	svcMgr ServiceManager,
	m yang.ModelSet,
	modelSetName string,
	compConfig []*conf.ServiceConfig,
) *compMgr {
	mappings, err := createComponentMappings(m, modelSetName, compConfig)
	if err != nil {
		fmt.Printf("Unable to create component mappings: %s\n", err)
		return nil
	}

	return &compMgr{
		OperationsManager: opsMgr,
		ServiceManager:    svcMgr,
		compMappings:      mappings,
	}
}

// listActiveModels returns the topologically sorted list of models
// that are active in the provided config.  If they have config but are
// not running, they will not be returned.
//
// Typical usage would be for getting a list of models to query for state.
func (cm *compMgr) listActiveModels(
	m ModelSet,
	config datanode.DataNode) []string {

	out := make([]string, 0)

	for _, modelName := range cm.compMappings.orderedComponents {
		comp := cm.compMappings.components[modelName]
		active, err := cm.IsActive(comp.name)
		if err != nil {
			log(err.Error())
		}
		if !active {
			continue
		}
		out = append(out, modelName)
	}
	return out
}

// listActiveOrConfiguredModels returns the topologically sorted list of models
// that are active in the provided config.  Models that have config but are
// not active are returned as they may need to be activated eg for validation.
func (cm *compMgr) listActiveOrConfiguredModels(
	m ModelSet,
	config datanode.DataNode,
) []string {

	out := make([]string, 0)
	for _, modelName := range cm.compMappings.orderedComponents {
		comp := cm.compMappings.components[modelName]

		active, err := cm.IsActive(comp.name)
		if err != nil {
			log(err.Error())
		}

		// Either the model has been activated by default or it has config.
		// Only query models in one of these two states.
		// NB: FilterTree() can impact performance, esp on low-powered devices
		//     such as SIADs.  So, only call if component isn't active.
		//     A future enhancement would be to do a single pass to extract
		//     all active namespaces as we only need to know if a service is
		//     configured or not.  Actual config is irrelevant.
		if active || comp.HasConfiguration(m, config) {
			out = append(out, modelName)
		}
	}
	return out
}

func (cm *compMgr) ComponentValidation(
	m ModelSet,
	dn datanode.DataNode,
	logFn commitTimeLogFn,
) []error {

	if err := cm.Dial(); err != nil {
		return []error{
			fmt.Errorf("Validation: Unable to initialise component comms: %s",
				err)}
	}

	var errs []error
	for _, modelName := range cm.listActiveOrConfiguredModels(m, dn) {
		startTime := time.Now()

		svc := cm.compMappings.components[modelName]
		jsonTree := svc.FilterCheckTree(m, dn)

		err := cm.CheckConfigForModel(modelName, string(jsonTree))
		if err != nil {
			errs = append(errs, err)
		}
		if logFn != nil {
			logFn(fmt.Sprintf("Check %s", modelName), startTime)
		}
	}
	return errs
}
