// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAPIKeyResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAPIKeyResourceConfig_basic("test-api-key-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("jellyfin_api_key.test", "app_name", "test-api-key-basic"),
					resource.TestCheckResourceAttrSet("jellyfin_api_key.test", "id"),
					resource.TestCheckResourceAttrSet("jellyfin_api_key.test", "access_token"),
					resource.TestCheckResourceAttrSet("jellyfin_api_key.test", "date_created"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "jellyfin_api_key.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccAPIKeyResource_multipleKeys(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create multiple keys
			{
				Config: testAccAPIKeyResourceConfig_multiple("test-api-key-1", "test-api-key-2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check first key
					resource.TestCheckResourceAttr("jellyfin_api_key.test1", "app_name", "test-api-key-1"),
					resource.TestCheckResourceAttrSet("jellyfin_api_key.test1", "id"),
					resource.TestCheckResourceAttrSet("jellyfin_api_key.test1", "access_token"),
					resource.TestCheckResourceAttrSet("jellyfin_api_key.test1", "date_created"),
					// Check second key
					resource.TestCheckResourceAttr("jellyfin_api_key.test2", "app_name", "test-api-key-2"),
					resource.TestCheckResourceAttrSet("jellyfin_api_key.test2", "id"),
					resource.TestCheckResourceAttrSet("jellyfin_api_key.test2", "access_token"),
					resource.TestCheckResourceAttrSet("jellyfin_api_key.test2", "date_created"),
				),
			},
		},
	})
}

func TestAccAPIKeyResource_appNameRequiresReplace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial key
			{
				Config: testAccAPIKeyResourceConfig_basic("test-api-key-original"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("jellyfin_api_key.test", "app_name", "test-api-key-original"),
					resource.TestCheckResourceAttrSet("jellyfin_api_key.test", "id"),
				),
			},
			// Change app_name - should require replacement (new key created)
			{
				Config: testAccAPIKeyResourceConfig_basic("test-api-key-changed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("jellyfin_api_key.test", "app_name", "test-api-key-changed"),
					resource.TestCheckResourceAttrSet("jellyfin_api_key.test", "id"),
				),
			},
		},
	})
}

func TestAccAPIKeyResource_specialCharacters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create key with special characters in name
			{
				Config: testAccAPIKeyResourceConfig_basic("Test App (v1.0) & More"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("jellyfin_api_key.test", "app_name", "Test App (v1.0) & More"),
					resource.TestCheckResourceAttrSet("jellyfin_api_key.test", "id"),
					resource.TestCheckResourceAttrSet("jellyfin_api_key.test", "access_token"),
				),
			},
		},
	})
}

func TestAccAPIKeyResource_hasValidId(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyResourceConfig_basic("test-api-key-id-check"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify that id is set (numeric Jellyfin ID)
					resource.TestCheckResourceAttrSet("jellyfin_api_key.test", "id"),
					// Verify that access_token is also set
					resource.TestCheckResourceAttrSet("jellyfin_api_key.test", "access_token"),
				),
			},
		},
	})
}

func TestAccAPIKeyResource_persistsAfterRefresh(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create key
			{
				Config: testAccAPIKeyResourceConfig_basic("test-api-key-persist"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("jellyfin_api_key.test", "access_token"),
				),
			},
			// Apply same config again - should be idempotent
			{
				Config: testAccAPIKeyResourceConfig_basic("test-api-key-persist"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("jellyfin_api_key.test", "access_token"),
				),
			},
		},
	})
}

// Test configuration functions

func testAccAPIKeyResourceConfig_basic(appName string) string {
	return fmt.Sprintf(`
resource "jellyfin_api_key" "test" {
  app_name = %[1]q
}
`, appName)
}

func testAccAPIKeyResourceConfig_multiple(appName1, appName2 string) string {
	return fmt.Sprintf(`
resource "jellyfin_api_key" "test1" {
  app_name = %[1]q
}

resource "jellyfin_api_key" "test2" {
  app_name = %[2]q
}
`, appName1, appName2)
}
