// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAPIKeyDataSource_byAppName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyDataSourceConfig_byAppName("test-api-key-datasource-appname"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.jellyfin_api_key.test", "app_name", "test-api-key-datasource-appname"),
					resource.TestCheckResourceAttrSet("data.jellyfin_api_key.test", "id"),
					resource.TestCheckResourceAttrSet("data.jellyfin_api_key.test", "access_token"),
					resource.TestCheckResourceAttrSet("data.jellyfin_api_key.test", "date_created"),
				),
			},
		},
	})
}

func TestAccAPIKeyDataSource_byAccessToken(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyDataSourceConfig_byAccessToken("test-api-key-datasource-token"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.jellyfin_api_key.test", "app_name", "test-api-key-datasource-token"),
					resource.TestCheckResourceAttrPair(
						"data.jellyfin_api_key.test", "access_token",
						"jellyfin_api_key.source", "access_token",
					),
					resource.TestCheckResourceAttrPair(
						"data.jellyfin_api_key.test", "id",
						"jellyfin_api_key.source", "id",
					),
					resource.TestCheckResourceAttrSet("data.jellyfin_api_key.test", "date_created"),
				),
			},
		},
	})
}

func TestAccAPIKeyDataSource_matchesResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyDataSourceConfig_byAppName("test-api-key-datasource-match"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source attributes match the resource
					resource.TestCheckResourceAttrPair(
						"data.jellyfin_api_key.test", "id",
						"jellyfin_api_key.source", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.jellyfin_api_key.test", "app_name",
						"jellyfin_api_key.source", "app_name",
					),
					resource.TestCheckResourceAttrPair(
						"data.jellyfin_api_key.test", "access_token",
						"jellyfin_api_key.source", "access_token",
					),
					resource.TestCheckResourceAttrPair(
						"data.jellyfin_api_key.test", "date_created",
						"jellyfin_api_key.source", "date_created",
					),
				),
			},
		},
	})
}

func TestAccAPIKeyDataSource_requiresAttribute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAPIKeyDataSourceConfig_noAttribute(),
				ExpectError: regexp.MustCompile("Either 'app_name' or 'access_token' must be provided"),
			},
		},
	})
}

func TestAccAPIKeyDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAPIKeyDataSourceConfig_notFound(),
				ExpectError: regexp.MustCompile("API Key Not Found"),
			},
		},
	})
}

func TestAccAPIKeyDataSource_specialCharacters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyDataSourceConfig_byAppName("Test Data Source (v2.0) & Special"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.jellyfin_api_key.test", "app_name", "Test Data Source (v2.0) & Special"),
					resource.TestCheckResourceAttrSet("data.jellyfin_api_key.test", "id"),
				),
			},
		},
	})
}

// Test configuration functions

func testAccAPIKeyDataSourceConfig_byAppName(appName string) string {
	return fmt.Sprintf(`
resource "jellyfin_api_key" "source" {
  app_name = %[1]q
}

data "jellyfin_api_key" "test" {
  app_name = jellyfin_api_key.source.app_name
}
`, appName)
}

func testAccAPIKeyDataSourceConfig_byAccessToken(appName string) string {
	return fmt.Sprintf(`
resource "jellyfin_api_key" "source" {
  app_name = %[1]q
}

data "jellyfin_api_key" "test" {
  access_token = jellyfin_api_key.source.access_token
}
`, appName)
}

func testAccAPIKeyDataSourceConfig_noAttribute() string {
	return `
data "jellyfin_api_key" "test" {
}
`
}

func testAccAPIKeyDataSourceConfig_notFound() string {
	return `
data "jellyfin_api_key" "test" {
  app_name = "this-app-definitely-does-not-exist-12345"
}
`
}
