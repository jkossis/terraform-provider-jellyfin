// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/jkossis/terraform-provider-jellyfin/internal/client"
)

func TestAPIKeyDataSource_Metadata(t *testing.T) {
	ds := &APIKeyDataSource{}
	req := datasource.MetadataRequest{
		ProviderTypeName: "jellyfin",
	}
	resp := &datasource.MetadataResponse{}

	ds.Metadata(context.Background(), req, resp)

	expected := "jellyfin_api_key"
	if resp.TypeName != expected {
		t.Errorf("Expected TypeName %q, got %q", expected, resp.TypeName)
	}
}

func TestAPIKeyDataSource_Schema(t *testing.T) {
	ds := &APIKeyDataSource{}
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	ds.Schema(context.Background(), req, resp)

	// Check that schema has expected attributes
	if resp.Schema.Attributes == nil {
		t.Fatal("Expected schema attributes to be defined")
	}

	// Check id attribute
	idAttr, ok := resp.Schema.Attributes["id"]
	if !ok {
		t.Error("Expected 'id' attribute in schema")
	} else {
		if !idAttr.IsComputed() {
			t.Error("Expected 'id' attribute to be computed")
		}
	}

	// Check app_name attribute
	appNameAttr, ok := resp.Schema.Attributes["app_name"]
	if !ok {
		t.Error("Expected 'app_name' attribute in schema")
	} else {
		if !appNameAttr.IsOptional() {
			t.Error("Expected 'app_name' attribute to be optional")
		}
		if !appNameAttr.IsComputed() {
			t.Error("Expected 'app_name' attribute to be computed")
		}
	}

	// Check access_token attribute
	accessTokenAttr, ok := resp.Schema.Attributes["access_token"]
	if !ok {
		t.Error("Expected 'access_token' attribute in schema")
	} else {
		if !accessTokenAttr.IsOptional() {
			t.Error("Expected 'access_token' attribute to be optional")
		}
		if !accessTokenAttr.IsComputed() {
			t.Error("Expected 'access_token' attribute to be computed")
		}
		if !accessTokenAttr.IsSensitive() {
			t.Error("Expected 'access_token' attribute to be sensitive")
		}
	}

	// Check date_created attribute
	dateCreatedAttr, ok := resp.Schema.Attributes["date_created"]
	if !ok {
		t.Error("Expected 'date_created' attribute in schema")
	} else {
		if !dateCreatedAttr.IsComputed() {
			t.Error("Expected 'date_created' attribute to be computed")
		}
	}

	// Check schema has a description
	if resp.Schema.MarkdownDescription == "" {
		t.Error("Expected schema to have a markdown description")
	}
}

func TestAPIKeyDataSource_Configure_nilProviderData(t *testing.T) {
	ds := &APIKeyDataSource{}
	req := datasource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &datasource.ConfigureResponse{}

	ds.Configure(context.Background(), req, resp)

	// Should not error when provider data is nil (early return)
	if resp.Diagnostics.HasError() {
		t.Errorf("Unexpected error: %v", resp.Diagnostics.Errors())
	}
}

func TestAPIKeyDataSource_Configure_wrongType(t *testing.T) {
	ds := &APIKeyDataSource{}
	req := datasource.ConfigureRequest{
		ProviderData: "wrong type", // Should be *client.Client
	}
	resp := &datasource.ConfigureResponse{}

	ds.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("Expected error when provider data is wrong type")
	}
}

func TestAPIKeyDataSource_Configure_success(t *testing.T) {
	ds := &APIKeyDataSource{}
	c := client.NewClient("http://localhost:8096", "test-key")
	req := datasource.ConfigureRequest{
		ProviderData: c,
	}
	resp := &datasource.ConfigureResponse{}

	ds.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Unexpected error: %v", resp.Diagnostics.Errors())
	}

	if ds.client != c {
		t.Error("Expected client to be set")
	}
}

func TestNewAPIKeyDataSource(t *testing.T) {
	ds := NewAPIKeyDataSource()
	if ds == nil {
		t.Error("Expected data source to be instantiated")
	}

	_, ok := ds.(*APIKeyDataSource)
	if !ok {
		t.Error("Expected data source to be *APIKeyDataSource")
	}
}
