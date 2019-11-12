// Copyright (c) 2018-2019, AT&T Intellectual Property.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package union

type unionOptions struct {
	auth             Auther
	includeDefaults  bool
	hideSecrets      bool
	forceShowSecrets bool
}

type UnionOption func(*unionOptions)

func Authorizer(auth Auther) UnionOption {
	return func(opts *unionOptions) {
		opts.auth = auth
	}
}

func IncludeDefaults(opts *unionOptions) {
	opts.includeDefaults = true
}

func HideSecrets(opts *unionOptions) {
	opts.hideSecrets = true
}

// ForceShowSecrets forces secrets to not be filtered, even if the usual secret
// filtering logic would suggest they should be filtered.
func ForceShowSecrets(opts *unionOptions) {
	opts.forceShowSecrets = true
}

func (opts *unionOptions) shouldHideSecrets(path []string) bool {
	// The interaction between the Authorizer, HideSecrets, and ForceShowSecrets
	// can be a bit confusing, so here's a table that should help:
	//
	// | HideSecrets | Auther.AuthReadSecrets | ForceShowSecrets | Shown?  |
	// |-------------|------------------------|------------------|---------|
	// | Show        | Hide                   | -                | Hide    |
	// | Hide        | Hide                   | -                | Hide    |
	// | Show        | -                      | -                | Show    |
	// | Hide        | -                      | -                | Hide    |
	// | Show        | Hide                   | Show             | Show    |
	// | Hide        | Hide                   | Show             | Show    |
	// | Show        | -                      | Show             | Show    |
	// | Hide        | -                      | Show             | Show    |
	hideSecrets := opts.hideSecrets || !authorize(opts.auth, path, "secrets")
	if opts.forceShowSecrets {
		hideSecrets = false
	}
	return hideSecrets
}
