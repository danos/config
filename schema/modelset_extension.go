// Copyright (c) 2017-2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"fmt"
	"strings"
	"time"

	"github.com/danos/config/data"
	"github.com/danos/utils/exec"
	"github.com/danos/vci/services"
	"github.com/danos/yang/data/datanode"
	"github.com/danos/yang/data/encoding"
	yang "github.com/danos/yang/schema"
	"github.com/danos/yangd"
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
	ListActiveModels(config datanode.DataNode) []string
	ListActiveOrConfiguredModels(config datanode.DataNode) []string
	ServiceValidation(datanode.DataNode, commitTimeLogFn) []*exec.Output
	ServiceSetRunning(datanode.DataNode, *map[string]bool) []*exec.Output
	ServiceSetRunningWithLog(
		datanode.DataNode,
		*map[string]bool,
		commitTimeLogFn,
	) []*exec.Output
	ServiceGetRunning(ConfigMultiplexerFn) (*data.Node, error)
	GetModelNameForNamespace(string) (string, bool)
	GetDefaultServiceModuleMap() map[string]struct{}
}

type service struct {
	name     string
	dispatch yangd.Service
	modMap   map[string]struct{}
	nsFilter func(s yang.Node, d datanode.DataNode,
		children []datanode.DataNode) bool
}

func (s *service) FilterTree(n Node, dn datanode.DataNode) []byte {
	filteredCandidate := yang.FilterTree(n, dn, s.nsFilter)
	return encoding.ToRFC7951(n, filteredCandidate)
}

func (s *service) HasConfiguration(n Node, dn datanode.DataNode) bool {
	return string(s.FilterTree(n, dn)) != "{}"
}

func convertServiceErrors(e dbus.Error) []*exec.Output {
	if e.Name == dbus.ErrMsgNoObject.Name {
		//Ignore these until the proper error plumbing is done
		return nil
	}
	var outs []*exec.Output

	for _, e := range e.Body {
		ee := &exec.Output{Path: []string{""},
			Output: fmt.Sprint(e)}
		outs = append(outs, ee)
	}
	return outs
}

func (s *service) GetState(path []string) ([]byte, error) {
	p := strings.Join(path, "/")
	return s.dispatch.GetState(p)
}

func (s *service) ValidateCandidate(
	n Node, dn datanode.DataNode,
) []*exec.Output {

	jsonTree := s.FilterTree(n, dn)
	err := s.dispatch.ValidateCandidate(jsonTree)
	if err != nil {
		fmt.Printf("Failed to run service validation for %s: %s\n",
			s.name, err.Error())
		if e, ok := err.(dbus.Error); ok {
			return convertServiceErrors(e)
		}
	}
	return nil
}

func (s *service) SetRunning(n Node, dn datanode.DataNode) []*exec.Output {
	jsonTree := s.FilterTree(n, dn)
	err := s.dispatch.SetRunning(jsonTree)
	if err != nil {
		fmt.Printf("Failed to run service provisioning for %s: %s\n",
			s.name, err.Error())
		if e, ok := err.(dbus.Error); ok {
			return convertServiceErrors(e)
		}
	}
	return nil
}

type modelSet struct {
	yang.ModelSet
	*extensions
	*state
	dispatcher      yangd.Dispatcher
	services        map[string]*service
	nsMap           map[string]string
	orderedServices []string
	defaultService  string
}

// Compile time check that the concrete type meets the interface
var _ ModelSet = (*modelSet)(nil)

type namespaceToService func(string) *service

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

	modelToNamespaceMap, globalNSMap, defaultModel, err :=
		getModelToNamespaceMapForModelSet(
			m, VyattaV1ModelSet, c.ComponentConfig)
	if err != nil {
		return nil, err
	}
	var service_map map[string]*service

	dispatch := c.Dispatcher
	if dispatch != nil {
		service_map = getServiceMap(dispatch, modelToNamespaceMap)
	}

	orderedServices, err := getOrderedServicesList(
		VyattaV1ModelSet, defaultModel, c.ComponentConfig)
	if err != nil {
		return nil, err
	}

	if len(service_map) != len(orderedServices) {
		return nil, fmt.Errorf(
			"Mismatch between number of ordered (%d) "+
				"and unordered (%d) services.",
			len(orderedServices), len(service_map))
	}

	ext := newExtend(nil)
	return &modelSet{
			m, ext, newState(m, ext), dispatch,
			service_map, globalNSMap, orderedServices, defaultModel},
		err
}

// ListActiveModels returns the topologically sorted list of models
// that are active in the provided config.  If they have config but are
// not running, they will not be returned.
//
// Typical usage would be for getting a list of models to query for state.
func (m *modelSet) ListActiveModels(config datanode.DataNode) []string {

	out := make([]string, 0)

	svcMgr := services.NewManager()
	defer svcMgr.Close()

	for _, modelName := range m.orderedServices {
		svc := m.services[modelName]
		active, err := svcMgr.IsActive(svc.name)
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
	config datanode.DataNode,
) []string {

	out := make([]string, 0)

	svcMgr := services.NewManager()
	defer svcMgr.Close()

	for _, modelName := range m.orderedServices {
		svc := m.services[modelName]

		active, err := svcMgr.IsActive(svc.name)
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
		if active || svc.HasConfiguration(m, config) {
			out = append(out, modelName)
		}
	}
	return out
}

func (m *modelSet) ServiceValidation(
	dn datanode.DataNode,
	logFn commitTimeLogFn,
) []*exec.Output {

	if m.dispatcher == nil {
		return nil
	}

	var outs []*exec.Output
	for _, modelName := range m.ListActiveOrConfiguredModels(dn) {
		startTime := time.Now()
		svc := m.services[modelName]
		new_outs := svc.ValidateCandidate(m, dn)
		if len(new_outs) > 0 {
			outs = append(outs, new_outs...)
		}
		if logFn != nil {
			logFn(fmt.Sprintf("Check %s", modelName), startTime)
		}
	}
	return outs
}

func (m *modelSet) GetModelNameForNamespace(ns string) (string, bool) {
	for svcName, svc := range m.services {
		if _, ok := svc.modMap[ns]; ok {
			return svcName, true
		}
	}
	return "", false
}

func (m *modelSet) GetDefaultServiceModuleMap() map[string]struct{} {
	return m.services[m.defaultService].modMap
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
	dn datanode.DataNode,
	changedNSMap *map[string]bool,
) []*exec.Output {
	return m.ServiceSetRunningWithLog(dn, changedNSMap, nil)
}

func (m *modelSet) ServiceSetRunningWithLog(
	dn datanode.DataNode,
	changedNSMap *map[string]bool,
	commitLogFn commitTimeLogFn,
) []*exec.Output {

	if m.dispatcher == nil {
		return nil
	}

	log("Set Running configuration:\n")

	var changedSvcs map[string]bool
	if changedNSMap != nil {
		changedSvcs = make(map[string]bool, 0)
		for ns, _ := range *changedNSMap {
			changedSvcs[m.nsMap[ns]] = true
		}
		changedSvcs[m.defaultService] = true
	}

	var outs []*exec.Output

	for _, ordServ := range m.orderedServices {
		if changedSvcs != nil {
			if _, ok := changedSvcs[ordServ]; !ok {
				log(fmt.Sprintf("\t'%s' hasn't changed.\n",
					ordServ))
				continue
			}
		}
		startTime := time.Now()
		serv, ok := m.services[ordServ]
		if !ok {
			log(fmt.Sprintf("Unable to set running config for '%s' service.\n",
				ordServ))
			continue
		}
		log(fmt.Sprintf("\t'%s' has changed.\n", ordServ))
		new_outs := serv.SetRunning(m, dn)
		if len(new_outs) > 0 {
			outs = append(outs, new_outs...)
		}
		if commitLogFn != nil {
			commitLogFn(fmt.Sprintf("Commit %s", ordServ), startTime)
		}
	}
	return outs
}

func (m *modelSet) ServiceGetRunning(cfgMuxFn ConfigMultiplexerFn,
) (*data.Node, error) {

	if m.dispatcher == nil {
		return nil, nil
	}

	var configs = make([][]byte, 0, len(m.services))

	for _, serv := range m.services {
		// Build up JSON config, then decode ...
		jsonInput, err := serv.dispatch.GetRunning("")
		if err != nil {
			return nil, fmt.Errorf("Unable to get running config for %s",
				serv.name)
		}
		configs = append(configs, jsonInput)
	}

	return cfgMuxFn(configs, m)
}
