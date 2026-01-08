// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

// Package main runs the provider's gRPC server.
package main

import (
	"context"
	"fmt"
	"os"

	xyz "github.com/JDetmar/pulumi-webflow/provider"
)

// Serve the provider against Pulumi's Provider protocol.
func main() {
	err := xyz.Provider().Run(context.Background(), xyz.Name, xyz.Version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(1)
	}
}
