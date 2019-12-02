// Copyright (c) 2017,2019, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2016-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/danos/yang/compile"
)

func assertHelpMapContains(t *testing.T, helpMap map[string]string, val, help string) {
	if helpMap[val] != help {
		t.Errorf("Help for '%s' not as expected:\n  Expect - %s\n  Actual - %s",
			val, help, helpMap[val])
	}
}

func assertHelpMapDoesNotContain(t *testing.T, helpMap map[string]string, val string) {
	if _, ok := helpMap[val]; ok {
		t.Errorf("Help for '%s' incorrectly present\n  Actual - %v",
			val, helpMap[val])
	}
}

func TestPatternHelp(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`leaf patternLeaf {
             type string {
                 pattern "[a-z]+";
                 configd:pattern-help "<lower-case>";
                 configd:help "pattern help";
             }
         }`))

	ms, err := GetConfigSchema(schema_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	node := ms.Child("patternLeaf")
	helpMap := node.(ExtendedNode).HelpMap()

	assertHelpMapContains(t, helpMap, "<lower-case>", "pattern help")
}

func TestOpdPatternHelp(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`opd:option patternOption {
             type string {
                 pattern "[a-z]+";
                 opd:pattern-help "<lower-case>";
                 opd:help "pattern help";
             }
         }`))

	ms, err := GetSchema(compile.IsOpd, schema_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	node := ms.Child("patternOption")
	helpMap := node.(ExtendedNode).HelpMap()

	assertHelpMapContains(t, helpMap, "<lower-case>", "pattern help")
}

func TestLeafrefPatternHelp(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`list list {
			key "list-key";
			leaf list-key {
				type string;
			}
		}
		leaf leafref {
             type leafref {
                 path "/list";
                 configd:pattern-help "<list-entry>";
                 configd:help "leafref help";
             }
         }`))

	ms, err := GetConfigSchema(schema_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	node := ms.Child("leafref")
	helpMap := node.(ExtendedNode).HelpMap()

	assertHelpMapContains(t, helpMap, "<list-entry>", "leafref help")
}

func TestEnumerationHelp(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`leaf enumLeaf {
             type enumeration {
                 enum alpha {
                     configd:help "alpha help";
                 }
                 enum beta {
                     configd:help "beta help";
                 }
                 enum gamma;
                 configd:help "default help";
             }
         }`))

	ms, err := GetConfigSchema(schema_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	node := ms.Child("enumLeaf")
	helpMap := node.(ExtendedNode).HelpMap()

	assertHelpMapContains(t, helpMap, "alpha", "alpha help")
	assertHelpMapContains(t, helpMap, "beta", "beta help")
	assertHelpMapContains(t, helpMap, "gamma", "default help")
}

func TestOverriddenEnumerationHelp(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`typedef myEnum {
             type enumeration {
                 enum alpha {
                     configd:help "alpha help";
                 }
                 enum beta {
                     configd:help "beta help";
                 }
                 enum gamma;
                 configd:help "hidden help";
             }
         }
		leaf enumLeaf {
             type myEnum {
                 configd:help "special help";
             }
        }`))

	ms, err := GetConfigSchema(schema_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	node := ms.Child("enumLeaf")
	helpMap := node.(ExtendedNode).HelpMap()

	assertHelpMapContains(t, helpMap, "alpha", "alpha help")
	assertHelpMapContains(t, helpMap, "beta", "beta help")
	assertHelpMapContains(t, helpMap, "gamma", "special help")
}

func TestObsoleteEnumerationHelp(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate, `
		leaf enumLeaf {
			type enumeration {
				enum alpha {
					configd:help "alpha help";
				}
				enum beta {
					status obsolete;
					configd:help "beta help";
				}
				enum gamma;
				configd:help "default help";
			}
		}`))

	ms, err := GetConfigSchema(schema_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	node := ms.Child("enumLeaf")
	helpMap := node.(ExtendedNode).HelpMap()

	assertHelpMapContains(t, helpMap, "alpha", "alpha help")
	assertHelpMapDoesNotContain(t, helpMap, "beta")
	assertHelpMapContains(t, helpMap, "gamma", "default help")
}

func TestTypeInheritsParentHelp(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate, `
		leaf alpha {
			configd:help "Leaf Alpha help";
			type uint32 {
				// Inherit leaf help text
			}
		}`))

	ms, err := GetConfigSchema(schema_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	// Ensure that a type with no help text, inherits
	node := ms.Child("alpha")
	helpMap := node.(ExtendedNode).HelpMap()

	assertHelpMapContains(t, helpMap, "<0..4294967295>", "Leaf Alpha help")
}

func TestUnionHelp(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate, `
		leaf union-leaf {
			configd:help "Leaf level help";
			type union {	
				type uint8 {
					configd:help "uint8 help";
				}
				type uint16 {
					// Will get leafs help
				}
				type uint32 {
					configd:help "uint32 help";
				}
				type string {
					configd:help "string help";
				}
				type boolean {
					configd:help "boolean help";
				}
			}
		}`))

	ms, err := GetConfigSchema(schema_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	node := ms.Child("union-leaf")
	helpMap := node.(ExtendedNode).HelpMap()

	assertHelpMapContains(t, helpMap, "true", "boolean help")
	assertHelpMapContains(t, helpMap, "false", "boolean help")
	assertHelpMapContains(t, helpMap, "<text>", "string help")
	assertHelpMapContains(t, helpMap, "<0..255>", "uint8 help")
	assertHelpMapContains(t, helpMap, "<0..65535>", "Leaf level help")
	assertHelpMapContains(t, helpMap, "<0..4294967295>", "uint32 help")
}

func TestIdentityrefHelp(t *testing.T) {
	const remoteTemplate = `
module test-remote {
        namespace "urn:vyatta.com:test:remote";
        prefix test-remote;
	import test-configd-compile {prefix test;}
        organization "AT&T Inc.";
        revision 2019-11-14 {
                description "Test remote schema";
        }
        %s
}
`
	schema_remote := bytes.NewBufferString(fmt.Sprintf(
		remoteTemplate,
		`identity asteroids {
			base test:video-games;
		}
		identity donkey-kong {
			base test:video-games;
			configd:help "Avoid those barrels";
         }`))

	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`identity video-games;
		identity space-invaders {
			base video-games;
			configd:help "Shoot those aliens";
		}
		identity pac-man {
			base video-games;
			configd:help "Avoid those ghosts";
		}
		identity manic-miner {
			base video-games;
		}
		typedef games {
			type identityref {
				base video-games;
				configd:help "Classic video games";
			}

		}
		leaf games {
			type games;
         }`))

	ms, err := GetConfigSchema(schema_text.Bytes(), schema_remote.Bytes())
	if err != nil {
		t.Fatalf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	node := ms.Child("games")
	helpMap := node.(ExtendedNode).HelpMap()

	assertHelpMapContains(t, helpMap, "manic-miner", "Classic video games")
	assertHelpMapContains(t, helpMap, "pac-man", "Avoid those ghosts")
	assertHelpMapContains(t, helpMap, "space-invaders", "Shoot those aliens")
	assertHelpMapContains(t, helpMap, "test-remote:asteroids", "Classic video games")
	assertHelpMapContains(t, helpMap, "test-remote:donkey-kong", "Avoid those barrels")
	assertHelpMapDoesNotContain(t, helpMap, "video-games")
}

func TestIdentityrefUsesHelp(t *testing.T) {
	const remoteTemplate = `
module test-remote {
        namespace "urn:vyatta.com:test:remote";
        prefix test-remote;
	import test-configd-compile {prefix test;}
        organization "AT&T Inc.";
        revision 2019-11-14 {
                description "Test remote schema";
        }
        %s
}
`
	schema_remote := bytes.NewBufferString(fmt.Sprintf(
		remoteTemplate,
		`identity asteroids {
			base test:video-games;
		}
		identity donkey-kong {
			base test:video-games;
			configd:help "Avoid those barrels";
		}
		uses test:classic-video-games;
	 `))

	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`identity video-games;
		identity space-invaders {
			base video-games;
			configd:help "Shoot those aliens";
		}
		identity pac-man {
			base video-games;
			configd:help "Avoid those ghosts";
		}
		identity manic-miner {
			base video-games;
		}
		typedef games {
			type identityref {
				base video-games;
				configd:help "Classic video games";
			}

		}
		grouping classic-video-games {
			leaf games {
				type games;
			}
         }`))

	ms, err := GetConfigSchema(schema_text.Bytes(), schema_remote.Bytes())
	if err != nil {
		t.Fatalf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	node := ms.Child("games")
	helpMap := node.(ExtendedNode).HelpMap()

	// Check that identityref in uses not local to grouping definition
	// uses correct namespace
	assertHelpMapContains(t, helpMap, "test-configd-compile:manic-miner", "Classic video games")
	assertHelpMapContains(t, helpMap, "test-configd-compile:pac-man", "Avoid those ghosts")
	assertHelpMapContains(t, helpMap, "test-configd-compile:space-invaders", "Shoot those aliens")
	assertHelpMapContains(t, helpMap, "asteroids", "Classic video games")
	assertHelpMapContains(t, helpMap, "donkey-kong", "Avoid those barrels")
	// Ensure root identity is not included
	assertHelpMapDoesNotContain(t, helpMap, "video-games")
}

func TestIdentityrefUnionHelp(t *testing.T) {
	const remoteTemplate = `
module test-remote {
        namespace "urn:vyatta.com:test:remote";
        prefix test-remote;
	import test-configd-compile {prefix test;}
        organization "AT&T Inc.";
        revision 2019-11-14 {
                description "Test remote schema";
        }
        %s
}
`
	schema_remote := bytes.NewBufferString(fmt.Sprintf(
		remoteTemplate,
		`identity asteroids {
			base test:video-games;
		}
		identity donkey-kong {
			base test:video-games;
			configd:help "Avoid those barrels";
		}

		identity board-games;
		identity chess {
			base board-games;
		}
		identity risk {
			base board-games;
			configd:help "Conquer the world";
		}

		leaf games {
			type union {
				type test:games;
				type identityref {
					base board-games;
					configd:help "Board game";
				}
			}
		}
	 `))

	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`identity video-games;
		identity space-invaders {
			base video-games;
			configd:help "Shoot those aliens";
		}
		identity pac-man {
			base video-games;
			configd:help "Avoid those ghosts";
		}
		identity manic-miner {
			base video-games;
		}
		typedef games {
			type union {
				type enumeration {
					enum monopoly {
						configd:help "Avoid bankruptcy";
					}
					enum checkers {
						configd:help "Also known as draughts";
					}
				}
				type identityref {
					base video-games;
					configd:help "Classic video games";
				}
			}
		}
         `))

	ms, err := GetConfigSchema(schema_text.Bytes(), schema_remote.Bytes())
	if err != nil {
		t.Fatalf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	node := ms.Child("games")
	helpMap := node.(ExtendedNode).HelpMap()

	// Check that identityref in uses not local to grouping definition
	// uses correct namespace
	assertHelpMapContains(t, helpMap, "test-configd-compile:manic-miner", "Classic video games")
	assertHelpMapContains(t, helpMap, "test-configd-compile:pac-man", "Avoid those ghosts")
	assertHelpMapContains(t, helpMap, "test-configd-compile:space-invaders", "Shoot those aliens")
	assertHelpMapContains(t, helpMap, "asteroids", "Classic video games")
	assertHelpMapContains(t, helpMap, "donkey-kong", "Avoid those barrels")
	assertHelpMapContains(t, helpMap, "chess", "Board game")
	assertHelpMapContains(t, helpMap, "risk", "Conquer the world")
	assertHelpMapContains(t, helpMap, "monopoly", "Avoid bankruptcy")
	assertHelpMapContains(t, helpMap, "checkers", "Also known as draughts")
	// Ensure root identities are not included
	assertHelpMapDoesNotContain(t, helpMap, "video-games")
	assertHelpMapDoesNotContain(t, helpMap, "board-games")
}
