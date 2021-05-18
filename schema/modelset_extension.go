// Copyright (c) 2017-2021, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/danos/config/data"
	rfc7951data "github.com/danos/encoding/rfc7951/data"
	"github.com/danos/mgmterror"
	"github.com/danos/utils/exec"
	"github.com/danos/yang/data/datanode"
	yang "github.com/danos/yang/schema"
	"github.com/godbus/dbus"
)

type ConfigMultiplexerFn func([][]byte, ModelSet) (*data.Node, error)

// Needs to match configd: (*commitctx) LogCommitTime()
type commitTimeLogFn func(string, time.Time)

type ModelSet interface {
	yang.ModelSet
	ExtendedNode
	PathDescendant([]string) *TmplCompat
	OpdPathDescendant([]string) *TmplCompat
	ListActiveModels(
		compMgr ComponentManager,
		config datanode.DataNode) []string
	ListActiveOrConfiguredModels(
		compMgr ComponentManager,
		config datanode.DataNode) []string
	ServiceValidation(
		ComponentManager,
		datanode.DataNode,
		commitTimeLogFn,
	) []error
	ServiceSetRunning(
		ComponentManager,
		datanode.DataNode,
		*map[string]bool,
	) []*exec.Output
	ServiceSetRunningWithLog(
		ComponentManager,
		datanode.DataNode,
		*map[string]bool,
		commitTimeLogFn,
	) []*exec.Output
	ServiceGetRunning(
		ComponentManager,
		ConfigMultiplexerFn) (*data.Node, error)
	ServiceGetState(
		ComponentManager,
		datanode.DataNode,
		*rfc7951data.Tree,
		StateLogger) (*rfc7951data.Tree, error)
	GetModelNameForNamespace(string) (string, bool)
	GetDefaultComponentModuleMap() map[string]struct{}
}

type modelSet struct {
	yang.ModelSet
	*extensions
	*state
	compMappings *componentMappings
}

// Compile time check that the concrete type meets the interface
var _ ModelSet = (*modelSet)(nil)

type namespaceToComponent func(string) *component

// For now there is an implicit assumption that we are only dealing with the
// single 'vyatta-v1' model set.  As and when we support multiple model sets
// we should probably pass the required model set name in to this function,
// probably provided initially by the call to start yangd that provides the
// YANG directory to be parsed, as we will have a separate YANG directory
// per modelset.
const VyattaV1ModelSet = "vyatta-v1"

func (c *CompilationExtensions) ExtendModelSet(
	m yang.ModelSet,
) (yang.ModelSet, error) {

	modelToNamespaceMap, globalNSMap, defaultComponent, err :=
		getModelToNamespaceMapsForModelSet(
			m, VyattaV1ModelSet, c.ComponentConfig)
	if err != nil {
		return nil, err
	}
	var componentMap map[string]*component

	componentMap = getComponentMap(modelToNamespaceMap)

	orderedComponents, err := getOrderedComponentsList(
		VyattaV1ModelSet, defaultComponent, c.ComponentConfig)
	if err != nil {
		return nil, err
	}

	if len(componentMap) != len(orderedComponents) {
		return nil, fmt.Errorf(
			"Mismatch between number of ordered (%d) "+
				"and unordered (%d) components.",
			len(orderedComponents), len(componentMap))
	}

	compMappings := &componentMappings{
		components:        componentMap,
		nsMap:             globalNSMap,
		orderedComponents: orderedComponents,
		defaultComponent:  defaultComponent,
	}

	ext := newExtend(nil)
	return &modelSet{
			m, ext, newState(m, ext),
			compMappings},
		err
}

func checkAndInitOpsMgr(compMgr ComponentManager, operation string) error {
	if compMgr == nil || reflect.ValueOf(compMgr).IsNil() {
		return fmt.Errorf("%s: No component manager provided.", operation)
	}
	if err := compMgr.Dial(); err != nil {
		return fmt.Errorf("%s: Unable to initialise component comms: %s",
			operation, err)
	}
	return nil
}

// ListActiveModels returns the topologically sorted list of models
// that are active in the provided config.  If they have config but are
// not running, they will not be returned.
//
// Typical usage would be for getting a list of models to query for state.
func (m *modelSet) ListActiveModels(
	compMgr ComponentManager,
	config datanode.DataNode) []string {

	out := make([]string, 0)

	for _, modelName := range m.compMappings.orderedComponents {
		comp := m.compMappings.components[modelName]
		active, err := compMgr.IsActive(comp.name)
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

// ListActiveModels returns the topologically sorted list of models
// that are active in the provided config.  Models that have config but are
// not active are returned as they may need to be activated eg for validation.
func (m *modelSet) ListActiveOrConfiguredModels(
	compMgr ComponentManager,
	config datanode.DataNode,
) []string {

	out := make([]string, 0)
	for _, modelName := range m.compMappings.orderedComponents {
		comp := m.compMappings.components[modelName]

		active, err := compMgr.IsActive(comp.name)
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

func (m *modelSet) ServiceValidation(
	compMgr ComponentManager,
	dn datanode.DataNode,
	logFn commitTimeLogFn,
) []error {

	if err := checkAndInitOpsMgr(compMgr, "ServiceValidation"); err != nil {
		log(err.Error())
		return []error{err}
	}

	var errs []error
	for _, modelName := range m.ListActiveOrConfiguredModels(
		compMgr, dn) {
		startTime := time.Now()

		svc := m.compMappings.components[modelName]
		jsonTree := svc.FilterCheckTree(m, dn)

		err := compMgr.CheckConfigForModel(modelName, string(jsonTree))
		if err != nil {
			errs = append(errs, err)
		}
		if logFn != nil {
			logFn(fmt.Sprintf("Check %s", modelName), startTime)
		}
	}
	return errs
}

func (m *modelSet) GetModelNameForNamespace(ns string) (string, bool) {
	for svcName, svc := range m.compMappings.components {
		if _, ok := svc.modMap[ns]; ok {
			return svcName, true
		}
	}
	return "", false
}

func (m *modelSet) GetDefaultComponentModuleMap() map[string]struct{} {
	return m.compMappings.components[m.compMappings.defaultComponent].modMap
}

func log(output string) {
	for _, line := range strings.Split(output, "\n") {
		if len(line) == 0 {
			continue
		}
		fmt.Printf("VCI: %s\n", line)
	}
}

func (m *modelSet) ServiceSetRunning(
	compMgr ComponentManager,
	dn datanode.DataNode,
	changedNSMap *map[string]bool,
) []*exec.Output {
	return m.ServiceSetRunningWithLog(compMgr, dn, changedNSMap, nil)
}

func (m *modelSet) ServiceSetRunningWithLog(
	compMgr ComponentManager,
	dn datanode.DataNode,
	changedNSMap *map[string]bool,
	commitLogFn commitTimeLogFn,
) []*exec.Output {

	log("Set Running configuration:\n")

	var outs []*exec.Output

	if err := checkAndInitOpsMgr(compMgr, "ServiceSetRunning"); err != nil {
		ee := &exec.Output{Path: []string{""}, Output: err.Error()}
		outs = append(outs, ee)
		return outs
	}

	var changedComps map[string]bool
	if changedNSMap != nil {
		changedComps = make(map[string]bool, 0)
		for ns, _ := range *changedNSMap {
			changedComps[m.compMappings.nsMap[ns]] = true
		}
		changedComps[m.compMappings.defaultComponent] = true
	}

	for _, ordComp := range m.compMappings.orderedComponents {
		if changedComps != nil {
			if _, ok := changedComps[ordComp]; !ok {
				log(fmt.Sprintf("\t'%s' hasn't changed.\n", ordComp))
				continue
			}
		}
		startTime := time.Now()
		comp, ok := m.compMappings.components[ordComp]
		if !ok {
			log(fmt.Sprintf("Unable to set running config for '%s' component.\n",
				ordComp))
			continue
		}
		log(fmt.Sprintf("\t'%s' has changed.\n", ordComp))

		jsonTree := comp.FilterSetTree(m, dn)
		err := compMgr.SetConfigForModel(ordComp, string(jsonTree))
		if err != nil {
			fmt.Printf("Failed to run component provisioning for %s: %s\n",
				ordComp, err.Error())
			if e, ok := err.(dbus.Error); ok {
				new_out := &exec.Output{Path: []string{""},
					Output: fmt.Sprint(e)}
				outs = append(outs, new_out)
			}
		}

		if commitLogFn != nil {
			commitLogFn(fmt.Sprintf("Commit %s", ordComp), startTime)
		}
	}
	return outs
}

func (m *modelSet) ServiceGetRunning(
	compMgr ComponentManager,
	cfgMuxFn ConfigMultiplexerFn,
) (*data.Node, error) {

	if err := checkAndInitOpsMgr(compMgr, "ServiceGetRunning"); err != nil {
		return nil, err
	}

	var configs = make([][]byte, 0, len(m.compMappings.components))

	for _, comp := range m.compMappings.components {
		// Build up JSON config, then decode ...
		var jsonInput string
		err := compMgr.StoreConfigByModelInto(
			comp.name, &jsonInput)

		if err != nil {
			return nil, fmt.Errorf("Unable to get running config for %s: %s",
				comp.name, err)
		}
		configs = append(configs, []byte(jsonInput))
	}

	return cfgMuxFn(configs, m)
}

func (m *modelSet) ServiceGetState(
	compMgr ComponentManager,
	dn datanode.DataNode,
	ft *rfc7951data.Tree,
	logger StateLogger,
) (*rfc7951data.Tree, error) {

	if err := checkAndInitOpsMgr(compMgr, "ServiceGetState"); err != nil {
		return nil, err
	}

	allState := newRFC7951Merger(m, ft)

	for _, model := range m.ListActiveModels(compMgr, dn) {
		compStartTime := time.Now()
		state := rfc7951data.TreeNew()
		err := compMgr.StoreStateByModelInto(model, state)
		if err != nil {
			// No error if component doesn't implement state.
			_, ok := err.(*mgmterror.OperationNotSupportedApplicationError)
			if ok {
				logStateEvent(logger, fmt.Sprintf("%s no state fn", model))
				continue
			}
			logStateEvent(logger, fmt.Sprintf("%s store fail: %s", model, err))
			continue
		}
		allState.merge(state)
		logStateTime(logger, fmt.Sprintf("  %s", model), compStartTime)
	}

	return allState.getTree(), nil
}
