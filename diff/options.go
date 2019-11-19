// Copyright (c) 2019, AT&T Intellectual Property.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package diff

type options struct {
	hideSecrets bool
}

type Option func(*options)

func HideSecrets(hide bool) Option {
	return func(opts *options) {
		opts.hideSecrets = hide
	}
}

func getOptions(ops ...Option) *options {
	var opts options
	for _, opt := range ops {
		opt(&opts)
	}

	return &opts
}
