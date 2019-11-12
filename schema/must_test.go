// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/danos/yang/xpath"
	"github.com/danos/yang/xpath/xutils"
)

//
//  Helper Functions
//
// This returns a standard checker function that can be used from NodeChecker
func checkMust(expMust, expMachine string) checkFn {
	return func(t *testing.T, actual Node) {
		actMust := actual.Musts()[0].Mach.GetExpr()

		if expMust != actMust {
			t.Errorf("Node must does not match\n"+
				"  expect = %s\n"+
				"  actual = %s",
				expMust, actMust)
		}

		actMachine := actual.Musts()[0].Mach.PrintMachine()

		if expMachine != actMachine {
			t.Errorf("Node must does not match\n"+
				"  expect = %s\n"+
				"  actual = %s",
				expMachine, actMachine)
		}
	}
}

func assertMustMatches(
	t *testing.T, st ModelSet, node, expMust, expMachine string) {

	checklist := []checkFn{
		CheckName(node),
		checkMust(expMust, expMachine),
	}
	expected := NodeChecker{node, checklist}
	actual := st.SchemaChild(node)

	expected.check(t, actual)
}

func buildSchemaWithWarnings(t *testing.T, schema_snippet string,
) (ModelSet, []xutils.Warning, error) {

	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate, schema_snippet))
	return GetConfigSchemaWithWarnings(schema_text.Bytes())
}

func buildSchemaWithWarningsAndCustomFunctions(
	t *testing.T,
	schema_snippet string,
	userFnChecker xpath.UserCustomFunctionCheckerFn,
) (ModelSet, []xutils.Warning, error) {

	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate, schema_snippet))
	return GetConfigSchemaWithWarningsAndCustomFunctions(
		userFnChecker, schema_text.Bytes())
}

func checkWarnings(
	t *testing.T,
	actWarns, expWarns []xutils.Warning,
) {
	// Basic checks on number of warnings found.
	if len(actWarns) > 0 && len(expWarns) == 0 {
		t.Fatalf("Unexpected warnings: %v\n", actWarns)
	}
	if len(actWarns) != len(expWarns) {
		t.Fatalf("Unexpected number of warnings: got %d, exp %d",
			len(actWarns), len(expWarns))
	}

	// Now check warnings match.
	for _, expWarn := range expWarns {
		matchFound := false
		for _, actWarn := range actWarns {
			if err := actWarn.MatchDebugContains(expWarn); err == nil {
				matchFound = true
				break
			}
		}
		if !matchFound {
			t.Logf("Warning not found:\n%s", expWarn)
			t.Fatalf("Warnings that were found:\n%v", actWarns)
		}
	}
}

//  Test Cases

// First set of test cases mimic behaviour of configd which will ignore errors
// in the configd:must statement (bad grammar, unknown plugin functions) and
// fall back to the parent must statement.

func TestMustValidOverride(t *testing.T) {
	schema_snippet := `
		leaf test-leaf {
		type string;
		must "original" {
			configd:must "override";
		}
	}`

	st := buildSchema(t, schema_snippet)
	assertMustMatches(t, st, "test-leaf", "override",
		"--- machine start ---\n"+
			"nameTestPush\t{urn:vyatta.com:test:configd-compile override}\n"+
			"evalLocPath\n"+
			"store\n"+
			"---- machine end ----\n")
}

func TestMustInvalidOverrideBadGrammar(t *testing.T) {
	schema_snippet := `
		leaf test-leaf {
		type string;
		must "original" {
			configd:must "bad grammar";
		}
	}`

	st := buildSchema(t, schema_snippet)
	assertMustMatches(t, st, "test-leaf", "original",
		"--- machine start ---\n"+
			"nameTestPush\t{urn:vyatta.com:test:configd-compile original}\n"+
			"evalLocPath\n"+
			"store\n"+
			"---- machine end ----\n")
}

func TestMustInvalidOverrideNonexistentFunction(t *testing.T) {
	schema_snippet := `
		leaf test-leaf {
		type string;
		must "original" {
			configd:must "non-existent()";
		}
	}`

	st := buildSchema(t, schema_snippet)
	assertMustMatches(t, st, "test-leaf", "original",
		"--- machine start ---\n"+
			"nameTestPush\t{urn:vyatta.com:test:configd-compile original}\n"+
			"evalLocPath\n"+
			"store\n"+
			"---- machine end ----\n")
}

func TestMustOverrideOnWrongNodeType(t *testing.T) {
	schema_snippet := `
		leaf test-leaf {
		type string {
			configd:must "override";
		}
	}`

	err := getSchemaBuildError(t, schema_snippet)

	expected := "invalid substatement 'configd:must'"
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("Unexpected error:\n  expect: %s\n  actual: %s",
			expected, err.Error())
	}
}

func TestTooManyMustOverrides(t *testing.T) {
	schema_snippet := `
		leaf test-leaf {
		type string;
		must "original" {
			configd:must "override1";
			configd:must "override2";
		}
	}`

	err := getSchemaBuildError(t, schema_snippet)

	expected := "only one 'configd:must' statement is allowed"
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("Unexpected error:\n  expect: %s\n  actual: %s",
			expected, err.Error())
	}
}

// Following tests explore behaviour when we are running validation via
// the yangc utility for the likes of DRAM, or validate_yang.  Here we need
// to check both must and configd:must for validity (do they compile, do they
// reference unknown functions) and for valid paths (as we don't allow refs
// to non-existent nodes).
//
// For tests, cfgd:must should include a custom function that may or may not
// exist, so for 'ok' case we are testing custom function works.
//
// 'must' and 'cfgd:must' are input must statements.
//
// 'compile' is basic compilation result.  If buildSchema() returns an error
// then the compilation failed.  Not possible to really determine which of the
// two statements was compiled successfully but we know from other tests above
// that the precedence is correct, and we will see any compile errors from
// the returned warnings.
//
// 'WARN(eval)' is a warning during path evaluation, and will contain FULL
//  path.  'WARN(compile)' is generated during compilation, and at this point
//  we only have path name, not full path.
//
// must        cfgd:must   | compile | path-eval(must) | path-eval(cfgd:must)
// ------------------------------------------------------------------------
// ok          ok            PASS(c:m) PASS              PASS
//
// ok          inv-path    	 PASS(c:m) PASS              WARN(eval)
// ok          unk-fn      	 PASS(m)   PASS              WARN(compile)
// ok          bad-grammar 	 PASS(m)   PASS              ERROR
//
// inv-path    ok          	 PASS(c:m) WARN(eval)        PASS
// unk-fn      ok          	 PASS(c:m) ERROR             PASS
// bad-grammar ok          	 PASS(c:m) ERROR             PASS
//
// inv-path    inv-path      PASS(c:m) WARN(eval)        WARN(eval)
// inv-path    unk-fn        PASS(m)   WARN(eval)        WARN(compile)
// inv-path    bad-grammar   PASS(m)   WARN(eval)        ERROR
//
// unk-fn      inv-path      PASS(c:m) FAIL              WARN(eval)
// unk-fn      unk-fn        FAIL(m)   N/A [so no report on cfgd:must BUT
// unk-fn      bad-grammar   FAIL(m)   N/A  you'll get it when must fixed!]
//
// bad-grammar inv-path      PASS(c:m) FAIL              WARN(eval)
// bad-grammar unk-fn        FAIL      N/A
// bad-grammar bad-grammar   FAIL      N/A

type result bool

const (
	pass result = true
	fail        = false
)

type yangcMustTest struct {
	name,
	must,
	cfgdMust string
	compileResult result
	compileError  string
	pathEvalWarns []xutils.Warning
}

var customFnInfo = []xpath.CustomFunctionInfo{
	{
		Name:          "custom-fn",
		FnPtr:         nil, // Not going to be called ...
		Args:          []xpath.DatumTypeChecker{xpath.TypeIsNodeset},
		RetType:       xpath.TypeIsLiteral,
		DefaultRetVal: xpath.NewLiteralDatum(""),
	},
}

const (
	ok               = "true() or node-that-exists"
	okCustomFn       = "custom-fn(.) or node-that-exists"
	invalidPath      = "non-existent-node"
	invalidPath2     = "also-non-existent"
	unknownFunction  = "unknown-function()"
	unknownFunction2 = "another-fn()"
	badGrammar       = "bad grammar"
)

const mustSchema = `
container top {
	container mustCont {
		presence "Stop must-on-np-container warning";
		must '%s' {
			configd:must '%s';
		}
		leaf node-that-exists {
			type string;
		}
	}
}`

const (
	badGrammarDebug       = "Failed to compile 'bad grammar'"
	invalidPathDebug      = "ValidatePath:"
	invalidPath2Debug     = "test:configd-compile also-non-existent"
	unknownFunctionDebug  = "Unknown function or node type: 'unknown-function'"
	unknownFunction2Debug = "Unknown function or node type: 'another-fn'"
)

var badGrammarWarning = xutils.NewWarning(
	xutils.CompilerError,
	"mustCont", // Error at compile time doesn't have access to path.
	badGrammar, "schema0:13", "", badGrammarDebug,
)

var badGrammarWarningCfgdMust = xutils.NewWarning(
	xutils.ConfigdMustCompilerError,
	"mustCont", // Error at compile time doesn't have access to path.
	badGrammar, "schema0:13", "", badGrammarDebug,
)

var invalidPathWarning = xutils.NewWarning(
	xutils.DoesntExist,
	"/top/mustCont", // Error at path eval time DOES have access to full path.
	invalidPath, "schema0:13", "", invalidPathDebug,
)

var invalidPath2Warning = xutils.NewWarning(
	xutils.DoesntExist,
	"/top/mustCont",
	invalidPath2, "schema0:13", "", invalidPath2Debug,
)

var unknownFunctionWarning = xutils.NewWarning(
	xutils.CompilerError,
	"mustCont",
	unknownFunction, "schema0:13", "", unknownFunctionDebug,
)

var unknownFunctionWarningCfgdMust = xutils.NewWarning(
	xutils.ConfigdMustCompilerError,
	"mustCont",
	unknownFunction, "schema0:13", "", unknownFunctionDebug,
)

var unknownFunction2Warning = xutils.NewWarning(
	xutils.CompilerError,
	"mustCont",
	unknownFunction2, "schema0:13", "", unknownFunction2Debug,
)

// See comment block above for full description of tests.
var yangcTests = []yangcMustTest{
	// ok          ok            PASS(c:m) PASS              PASS
	{
		name:          "All pass",
		must:          ok,
		cfgdMust:      okCustomFn,
		compileResult: pass,
		pathEvalWarns: nil,
	},
	// ok          inv-path    	 PASS(c:m) PASS              WARN(eval)
	{
		name:          "cfgd:must invalid path",
		must:          ok,
		cfgdMust:      invalidPath,
		compileResult: pass,
		pathEvalWarns: []xutils.Warning{invalidPathWarning},
	},
	// ok          unk-fn      	 PASS(m)   PASS              WARN(compile)
	{
		name:          "cfgd:must unknown function",
		must:          ok,
		cfgdMust:      unknownFunction,
		compileResult: pass,
		pathEvalWarns: []xutils.Warning{unknownFunctionWarningCfgdMust},
	},
	// ok          bad-grammar 	 PASS(m)   PASS              ERROR
	{
		name:          "cfgd:must bad grammar",
		must:          ok,
		cfgdMust:      badGrammar,
		compileResult: pass,
		pathEvalWarns: []xutils.Warning{badGrammarWarningCfgdMust},
	},
	// inv-path    ok          	 PASS(c:m) WARN(eval)        PASS
	{
		name:          "must invalid path",
		must:          invalidPath,
		cfgdMust:      ok,
		compileResult: pass,
		pathEvalWarns: []xutils.Warning{invalidPathWarning},
	},
	// unk-fn      ok          	 PASS(c:m) ERROR             PASS
	{
		name:          "must unknown function",
		must:          unknownFunction,
		cfgdMust:      ok,
		compileResult: pass,
		pathEvalWarns: []xutils.Warning{unknownFunctionWarning},
	},
	// bad-grammar ok          	 PASS(c:m) ERROR             PASS
	{
		name:          "must bad grammar",
		must:          badGrammar,
		cfgdMust:      ok,
		compileResult: pass,
		pathEvalWarns: []xutils.Warning{badGrammarWarning},
	},
	// inv-path    inv-path      PASS(c:m) WARN(eval)        WARN(eval)
	{
		name:          "must and configd:must invalid path",
		must:          invalidPath,
		cfgdMust:      invalidPath2,
		compileResult: pass,
		pathEvalWarns: []xutils.Warning{
			invalidPathWarning, invalidPath2Warning},
	},
	// inv-path    unk-fn        PASS(m)   WARN(eval)        WARN(compile)
	{
		name:          "must invalid path, cfgd:must unknown function",
		must:          invalidPath,
		cfgdMust:      unknownFunction,
		compileResult: pass,
		pathEvalWarns: []xutils.Warning{
			invalidPathWarning, unknownFunctionWarningCfgdMust},
	},
	// inv-path    bad-grammar   PASS(m)   WARN(eval)        ERROR
	{
		name:          "must invalid path, cfgd:must bad grammar",
		must:          invalidPath,
		cfgdMust:      badGrammar,
		compileResult: pass,
		pathEvalWarns: []xutils.Warning{
			invalidPathWarning, badGrammarWarningCfgdMust},
	},
	// unk-fn      inv-path      PASS(c:m) FAIL              WARN(eval)
	{
		name:          "must unknown function, cfgd:must invalid path",
		must:          unknownFunction,
		cfgdMust:      invalidPath,
		compileResult: pass,
		pathEvalWarns: []xutils.Warning{
			unknownFunctionWarning, invalidPathWarning},
	},
	// unk-fn      unk-fn        FAIL(m)   N/A
	{
		name:          "must unknown function, cfgd:must unknown function",
		must:          unknownFunction,
		cfgdMust:      unknownFunction2,
		compileResult: fail,
		compileError:  unknownFunctionDebug,
	},
	// unk-fn      bad-grammar   FAIL(m)   N/A
	{
		name:          "must unknown function, cfgd:must bad grammar",
		must:          unknownFunction,
		cfgdMust:      badGrammar,
		compileResult: fail,
		compileError:  unknownFunctionDebug,
	},
	// bad-grammar inv-path      PASS(c:m) FAIL              WARN(eval)
	{
		name:          "must bad grammar, cfgd:must invalid path",
		must:          badGrammar,
		cfgdMust:      invalidPath,
		compileResult: pass,
		pathEvalWarns: []xutils.Warning{
			badGrammarWarning, invalidPathWarning},
	},
	// bad-grammar unk-fn        FAIL      N/A
	{
		name:          "must bad grammar, cfgd:must unknown function",
		must:          badGrammar,
		cfgdMust:      unknownFunction,
		compileResult: fail,
		compileError:  badGrammarDebug,
	},
	// bad-grammar bad-grammar   FAIL      N/A
	{
		name:          "must bad grammar, cfgd:must bad grammar",
		must:          badGrammar,
		cfgdMust:      badGrammar,
		compileResult: fail,
		compileError:  badGrammarDebug,
	},
}

func TestYangcMustBehaviour(t *testing.T) {

	// Inject custom functions so they are used for both must machines and
	// path evaluation machines.
	xpath.RegisterCustomFunctions(customFnInfo)

	for _, test := range yangcTests {
		t.Run(test.name, func(t *testing.T) {
			// Build schema
			_, actWarns, err := buildSchemaWithWarnings(t,
				fmt.Sprintf(mustSchema, test.must, test.cfgdMust))

			// Check for compilation failure
			if test.compileResult == fail {
				if err == nil {
					t.Fatalf("Unexpected compilation pass.")
				}
				if test.compileError == "" {
					t.Fatalf("Must specify compile error!")
				}
				if !strings.Contains(err.Error(), test.compileError) {
					t.Fatalf("Expected compile error:\n%s\n\nGot:\n%s\n",
						test.compileError, err.Error())
				}
			}

			// Check for compilation pass
			if test.compileResult == pass && err != nil {
				t.Fatalf("Unexpected compilation fail.")
			}

			// Check warnings if compilation failed.
			checkWarnings(t, actWarns, test.pathEvalWarns)
		})
	}
}

var customFnNotAllowedWarning = xutils.NewWarning(
	xutils.ConfigdMustCompilerError,
	"mustCont",
	"custom-fn-not-allowed()", "schema0:13", "",
	"Unknown function or node type: 'custom-fn-not-allowed'",
)
var customFnAllowedWarning = xutils.NewWarning(
	xutils.ConfigdMustCompilerError,
	"mustCont",
	"custom-fn-allowed()", "schema0:16", "",
	"Unknown function or node type: 'custom-fn-allowed'",
)
var customFnAlsoAllowedWarning = xutils.NewWarning(
	xutils.ConfigdMustCompilerError,
	"mustCont",
	"custom-fn-also-allowed() then bad grammar", "schema0:19", "",
	"Unknown function or node type: 'custom-fn-also-allowed'",
)

var notAllowedBadGrammarWarning = xutils.NewWarning(
	xutils.ConfigdMustCompilerError,
	"mustCont",
	"custom-fn-also-allowed() then bad grammar", "schema0:19", "",
	"Unrecognised operator name: 'then'",
)

type customFnFilterTest struct {
	name       string
	allowedFns []string
	expWarns   []xutils.Warning
}

func TestCustomFunctionWarningFiltering(t *testing.T) {

	customFnSchema := `
	container top {
		container mustCont {
			presence "Stop must-on-np-container warning";
			must 'true()' {
				configd:must 'custom-fn-not-allowed()';
			}
			must 'true()' {
				configd:must 'custom-fn-allowed()';
			}
			must 'true()' {
				configd:must 'custom-fn-also-allowed() then bad grammar';
			}
			leaf node-that-exists {
				type string;
			}
		}
	}`

	customFnTests := []customFnFilterTest{
		{
			name: "No allowed functions",
			expWarns: []xutils.Warning{
				customFnNotAllowedWarning,
				customFnAllowedWarning,
				customFnAlsoAllowedWarning,
			},
		},
		{
			name: "Allowed functions",
			allowedFns: []string{
				"custom-fn-allowed",
				"custom-fn-not-used",
				"custom-fn-also-allowed",
			},
			expWarns: []xutils.Warning{
				customFnNotAllowedWarning,
				notAllowedBadGrammarWarning,
			},
		},
		{
			name: "Truncated function name",
			allowedFns: []string{
				"custom-fn-allow", // truncated, should NOT match
				"custom-fn-not-used",
				"custom-fn-also-allowed",
			},
			expWarns: []xutils.Warning{
				customFnNotAllowedWarning,
				customFnAllowedWarning,
				notAllowedBadGrammarWarning,
			},
		},
		/*
			{
				name: "Invalid custom func and bad grammar in configd:must",
			},
			{
				name: "Valid custom function, bad grammar",
			},
		*/
	}

	for _, test := range customFnTests {
		t.Run(test.name, func(t *testing.T) {
			_, actWarns, err :=
				buildSchemaWithWarningsAndCustomFunctions(t, customFnSchema,
					func(name string) (*xpath.Symbol, bool) {
						for _, fn := range test.allowedFns {
							if fn == name {
								return xpath.NewDummyFnSym(name), true
							}
						}
						return nil, false
					})
			if err != nil {
				t.Fatalf("Unexpected error building schema: %s\n", err)
			}

			//filteredWarns := xutils.FilterCustomXpathFunctionWarnings(
			//	actWarns, test.allowedFns)
			checkWarnings(t, actWarns, test.expWarns)
		})
	}
}
