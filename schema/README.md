# configd-schema repository

## Overview

The configd/schema package extends standard YANG with configd extensions.
Each YANG node type can be extended with these extensions, and additionally
there are 'tree', 'model' and 'modelset' nodes that can be extended.

It also provides the point of integration between core configd / sessiond /
yangd functionality and VCI components.

## Configd extensions

Configd extensions represent proprietary extensions to standard YANG, eg
begin / create / update / delete / end action scripts, and allowed,
normalize, syntax, validate, secret extensions.

The mechanism for supporting extensions is that when we call the YANG
compiler (compile.CompileDir()), we pass in a CompilationExtensions object
that conforms to the Extensions interface provided by the compiler.  The
provided functions return decorated nodes (ExtendedNode) that conform to
the schema.Node interface, with additional state and extension information.

### CompilationExtensions

In addition to conforming to the compile.Extensions interface, the
CompilationExtensions object contains a Dispatcher and a slice of
ComponentConfigs.  These are provided by the caller (which in our case
is the 'configd' application during startup).

#### Dispatcher

This is provided by golang-brocade-vyatta-yangd/dbus, and provides wrappers
around Set/Get and Validate calls over DBUS to VCI components.

Bus address is 'net.vyatta.vci.config.[read|write].<method>'.

#### ComponentConfig

This is a slice of VCI component configurations as parsed from their
DotComponent files.  This provides mappings between components, models,
YANG modules and model sets.

As the DotComponent files may come from third party applications, we need to
validate their content.  Different parts of this are done in different
places as appropriate.

### ModelSet Extension

The ModelSet extension holds a slice of 'services' which represent each
component's configuration, along with a dispatcher which provides
communication over the DBUS to the components generically.

