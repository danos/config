// Copyright (c) 2017-2020, AT&T Intellectual Property. All rights reserved.
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

	"github.com/danos/utils/tsort"
	"github.com/danos/vci/conf"
	"github.com/danos/yang/data/datanode"
	yang "github.com/danos/yang/schema"
)

type modelToNamespaceMap map[string]*nsMaps

type nsMaps struct {
	setMap   map[string]struct{}
	checkMap map[string]struct{}
}

// Returns map of models to namespaces, plus the modelName for the 'default'
// component, if any.  Error returned if >1 default component, or if default
// component has any explicitly allocated namespaces.
func getModelToNamespaceMapsForModelSet(
	m yang.ModelSet,
	modelSetName string,
	comps []*conf.ServiceConfig,
) (
	modelToNamespaceMap,
	map[string]string,
	string,
	error) {

	modelToNSMaps, globalNSMap, defaultCompModelName, errs :=
		createNSMapForModelSetFromComponents(m, modelSetName, comps)

	if errs.Len() > 0 {
		return nil, nil, "", fmt.Errorf(
			"Problems found when validating components:\n\n%s\n",
			errs.String())
	}
	return modelToNSMaps, globalNSMap, defaultCompModelName, nil
}

// createNSMapForModelSetFromComponents - create the 2 required maps
//
// This function generates the model-to-namespace map that is used to
// map incoming service requests (state, config, notifications, RPCs) to
// the relevant service / component.  It can be quite hard to visualise how
// this map might look, so the following worked example should help.
//
// Let's say we have 3 components - CompA, CompB, and CompC - with the following
// models, modules, submodules and model set support:
//
// Comp A
//   Model mA1
//     Modules  mod-a, mod-a1, mod-x, submod-x1
//     ModelSet vyatta-v1
//   Model mA2
//     Modules  mod-a, mod-a2
//     ModelSet open-v1
//
// Comp B
//   Model mB
//     Modules  mod-b1, mod-b2
//     ModelSet vyatta-v1, open-v1
//
// Comp C
//   Model mC
//     Modules  mod-c, submod-x2 (NB: submodule in different comp to module)
//     ModelSet vyatta-v1
//
// Noting that the map contains namespace not module name, and thus we replace
// 'mod-a' with 'ns-a' etc, we get:
//
// NSMap for 'vyatta-v1' modelset:
//   [mA1] {[ns-a], [ns-a1], [ns-x], [submod-x1@ns-x]}
//   [mB]  {[ns-b1], [ns-b2]}
//   [mC]  {[ns-c], [submod-x2@ns-x]}
//
// NSMap for 'open-v1' modelset:
//   [mA2] {[ns-a], [ns-a2]}
//   [mB]  {[ns-b1], [ns-b2]}
//
// Note the following assumptions:
//
// - Zero or One Component may mark themselves as the default component.
//   Any model supported by this component must have NO modules or submodules
//   explicitly allocated.
//
// - Each Model contains zero or more modules and submodules
//
// - Each Model supports one or more model sets (eg vyatta-v1)
//
// - Modules may belong to multiple models, but only one model per model set.
//
// - Each component may only provide one model per model set.
//
func createNSMapForModelSetFromComponents(
	m yang.ModelSet,
	modelSetName string,
	comps []*conf.ServiceConfig,
) (
	modelToNamespaceMap,
	map[string]string,
	string,
	bytes.Buffer,
) {
	modelToNSMap := make(modelToNamespaceMap)
	globalNSMap := make(map[string]string)
	modelMap := make(map[string]bool)
	var defaultCompModelName string
	var errs bytes.Buffer

	for _, comp := range comps {
		model := comp.ModelByModelSet[modelSetName]

		if model == nil || modelDefinedTwice(
			modelMap, model.Name, modelSetName, &errs) {
			continue
		}
		modelMap[model.Name] = true

		maps := updateGlobalAndCreateCompNSMapsFromModelModules(
			comp, model, m, &globalNSMap, &errs)

		if comp.DefaultComp {
			if err := validateDefaultComponent(defaultCompModelName,
				maps.setMap, model.Name); err != nil {
				errs.WriteString(err.Error())
				return nil, nil, "", errs
			}
			defaultCompModelName = model.Name
		}
		modelToNSMap[model.Name] = maps
	}

	if defaultCompModelName != "" {
		defMap := createDefaultNamespaceList(m, globalNSMap)
		modelToNSMap[defaultCompModelName] = &nsMaps{
			setMap: defMap, checkMap: defMap}
	}

	return modelToNSMap, globalNSMap, defaultCompModelName, errs
}

// verifyModuleValidAndReturnNamespace - validate module and return namespace
//
// Check give <modOrSubmodName> exists as valid module or submodule in the
// modelset, and if so, return the namespace for it.  If not, return empty
// string.
func verifyModuleValidAndReturnNamespace(
	m yang.ModelSet,
	modOrSubmodName string,
	compName string,
) string {
	namespace := ""
	for _, module := range m.Modules() {
		if module.Identifier() == modOrSubmodName {
			namespace = module.(Model).Namespace()
			break
		}
	}
	if namespace == "" {
		for _, submod := range m.Submodules() {
			if submod.Identifier() == modOrSubmodName {
				namespace = createSubmodNS(
					submod.Identifier(), submod.Namespace())
				break
			}
		}
		if namespace == "" {
			fmt.Printf("%s:\n\t%s (sub)module not present in image.\n",
				compName, modOrSubmodName)
		}
	}

	return namespace
}

func updateGlobalAndCreateCompNSMapsFromModelModules(
	comp *conf.ServiceConfig,
	model *conf.Model,
	m yang.ModelSet,
	globalNSMap *map[string]string,
	errs *bytes.Buffer,
) *nsMaps {

	compSetMap := make(map[string]struct{})
	compCheckMap := make(map[string]struct{})
	retMaps := &nsMaps{
		setMap:   compSetMap,
		checkMap: compCheckMap,
	}

	// Create complete namespace map for Set() call, and baseline map for
	// Check() calls.
	//
	// model.Modules contains module and submodule names
	for _, modOrSubmodName := range model.Modules {

		namespace := verifyModuleValidAndReturnNamespace(
			m, modOrSubmodName, comp.Name)
		if namespace == "" {
			continue
		}

		if modelsShareNamespace(
			globalNSMap, namespace, model.Name, errs) {
			continue
		}

		(*globalNSMap)[namespace] = model.Name
		compSetMap[namespace] = struct{}{}
		compCheckMap[namespace] = struct{}{}
	}

	// Now we add the extra modules needed by the Check() call to allow
	// validation code to access candidate configuration owned by other
	// components.
	for _, modOrSubmodName := range model.ImportsForCheck {
		namespace := verifyModuleValidAndReturnNamespace(
			m, modOrSubmodName, comp.Name)
		if namespace == "" {
			continue
		}

		compCheckMap[namespace] = struct{}{}
	}

	return retMaps
}

// createSubmodNs - return string that provides unique identifier
//
// Format is arbitrary - using '@' which won't (can't?) occur in the module
// namespace allows for easy splitting later - use of ':' for example might
// be problematic.
func createSubmodNS(submod, moduleNS string) string {
	return submod + "@" + moduleNS
}

func validateDefaultComponent(
	defaultCompModelName string,
	compNSMap map[string]struct{},
	modelName string,
) error {
	if len(compNSMap) != 0 {
		return fmt.Errorf(
			"Default component (%s) cannot have assigned namespaces",
			modelName)
	}
	if defaultCompModelName != "" {
		return fmt.Errorf(
			"Can't have 2 default components: '%s' and '%s'",
			defaultCompModelName, modelName)
	}
	return nil
}

func modelDefinedTwice(
	modelMap map[string]bool,
	modelName,
	modelSetName string,
	errs *bytes.Buffer,
) bool {
	if _, ok := modelMap[modelName]; ok {
		errs.WriteString(fmt.Sprintf(
			"Model '%s' defined twice for model set '%s'.\n",
			modelName, modelSetName))
		return true
	}
	return false
}

func modelsShareNamespace(
	nsMap *map[string]string,
	namespace string,
	modelName string,
	errs *bytes.Buffer,
) bool {

	if previousModel, ok := (*nsMap)[namespace]; ok {
		errs.WriteString(fmt.Sprintf(
			"Models '%s' and '%s' cannot share '%s' namespace.\n",
			modelName, previousModel, namespace))
		return true
	}
	return false
}

func createDefaultNamespaceList(
	m yang.ModelSet,
	nsMap map[string]string,
) map[string]struct{} {

	defaultNamespaceList := make(map[string]struct{})
	for _, module := range m.Modules() {
		if _, ok := nsMap[module.(Model).Namespace()]; !ok {
			defaultNamespaceList[module.(Model).Namespace()] = struct{}{}
		}
	}
	for _, submod := range m.Submodules() {
		submodNS := createSubmodNS(submod.Identifier(), submod.Namespace())
		if _, ok := nsMap[submodNS]; !ok {
			defaultNamespaceList[submodNS] = struct{}{}
		}
	}
	return defaultNamespaceList
}

// For a specific modelset, provide a map of service objects, one per model,
// that contain:
//
// - a map of namespaces the service covers
// - a function to check if a given YANG node belongs to this service.
//
func getServiceMap(
	modelToNSMap modelToNamespaceMap,
) map[string]*service {

	service_map := make(map[string]*service, len(modelToNSMap))

	for name, modMap := range modelToNSMap {
		setMap := modMap.setMap // Avoid 'closure pitfall'
		setFilter := func(s yang.Node, d datanode.DataNode,
			children []datanode.DataNode) bool {
			if len(children) != 0 {
				return true
			}
			filter := s.Namespace()
			if s.Submodule() != "" {
				filter = createSubmodNS(s.Submodule(), filter)
			}
			_, ok := setMap[filter]
			return ok
		}
		checkMap := modMap.checkMap // Avoid 'closure pitfall'
		checkFilter := func(s yang.Node, d datanode.DataNode,
			children []datanode.DataNode) bool {
			if len(children) != 0 {
				return true
			}
			filter := s.Namespace()
			if s.Submodule() != "" {
				filter = createSubmodNS(s.Submodule(), filter)
			}
			_, ok := checkMap[filter]
			return ok
		}
		service_map[name] = &service{
			name:        name,
			modMap:      modMap.setMap,
			setFilter:   setFilter,
			checkMap:    modMap.checkMap,
			checkFilter: checkFilter,
		}
	}
	return service_map
}

func removeServiceSuffix(name string) string {
	if strings.HasSuffix(name, ".service") {
		return name[:len(name)-len(".service")]
	}
	return name
}

func getOrderedServicesList(
	modelSetName string,
	defaultService string,
	comps []*conf.ServiceConfig,
) ([]string, error) {

	orderedList := make([]string, 0)
	compNameToModelName := make(map[string]string, 0)

	g := tsort.New()

	for _, comp := range comps {
		model := comp.ModelByModelSet[modelSetName]
		if model == nil {
			continue
		}
		compNameToModelName[comp.Name] = model.Name
		g.AddVertex(comp.Name)
		for _, before := range comp.Before {
			if before == "" {
				continue
			}
			g.AddEdge(removeServiceSuffix(before), comp.Name)
		}
		for _, after := range comp.After {
			if after == "" {
				continue
			}
			g.AddEdge(comp.Name, removeServiceSuffix(after))
		}
	}

	orderedServices, err := g.Sort()
	if err != nil {
		return nil, fmt.Errorf("Unable to order services.")
	}

	for _, name := range orderedServices {
		if modelName, ok := compNameToModelName[name]; ok {
			orderedList = append(orderedList, modelName)
		}
	}

	return orderedList, nil
}
