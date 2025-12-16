// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
// The factory function is called for each Terraform CLI command to create a provider
// server that the CLI can connect to and interact with.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"jellyfin": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccProtoV6ProviderFactoriesWithEcho includes the echo provider alongside the jellyfin provider.
// It allows for testing assertions on data returned by an ephemeral resource during Open.
// The echoprovider is used to arrange tests by echoing ephemeral data into the Terraform state.
// This lets the data be referenced in test assertions with state checks.
var testAccProtoV6ProviderFactoriesWithEcho = map[string]func() (tfprotov6.ProviderServer, error){
	"jellyfin": providerserver.NewProtocol6WithError(New("test")()),
	"echo":     echoprovider.NewProviderServer(),
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
