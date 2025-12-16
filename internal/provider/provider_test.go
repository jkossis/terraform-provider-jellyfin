// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
// The factory function is called for each Terraform CLI command to create a provider
// server that the CLI can connect to and interact with.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"jellyfin": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// Verify that required environment variables are set for acceptance tests
	if v := os.Getenv("JELLYFIN_ENDPOINT"); v == "" {
		t.Fatal("JELLYFIN_ENDPOINT must be set for acceptance tests")
	}

	if v := os.Getenv("JELLYFIN_USERNAME"); v == "" {
		t.Fatal("JELLYFIN_USERNAME must be set for acceptance tests")
	}

	if v := os.Getenv("JELLYFIN_PASSWORD"); v == "" {
		t.Fatal("JELLYFIN_PASSWORD must be set for acceptance tests")
	}
}
