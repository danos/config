// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/danos/config/platform"
)

func showUsageAndExit() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "    %s <platform-dir> <output_file>\n", os.Args[0])
	os.Exit(1)
}

func main() {

	if len(os.Args) != 3 {
		showUsageAndExit()
	}

	platforms := platform.NewPlatform().PlatformBaseDir(os.Args[1]).
		LoadDefinitions()

	json_output, err := json.Marshal(platforms)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal platform data as JSON: %s",
			err)
		os.Exit(1)
	}

	err = ioutil.WriteFile(os.Args[2], json_output, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save platform data: %s",
			err)
		os.Exit(1)
	}

	os.Exit(0)
}
