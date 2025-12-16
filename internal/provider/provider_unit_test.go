// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func TestJellyfinProvider_Metadata(t *testing.T) {
	p := &JellyfinProvider{version: "test"}
	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}

	p.Metadata(context.Background(), req, resp)

	if resp.TypeName != "jellyfin" {
		t.Errorf("Expected TypeName 'jellyfin', got %s", resp.TypeName)
	}

	if resp.Version != "test" {
		t.Errorf("Expected Version 'test', got %s", resp.Version)
	}
}

func TestJellyfinProvider_Schema(t *testing.T) {
	p := &JellyfinProvider{}
	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	p.Schema(context.Background(), req, resp)

	// Check that schema has expected attributes
	if resp.Schema.Attributes == nil {
		t.Fatal("Expected schema attributes to be defined")
	}

	// Check endpoint attribute
	endpointAttr, ok := resp.Schema.Attributes["endpoint"]
	if !ok {
		t.Error("Expected 'endpoint' attribute in schema")
	} else {
		if !endpointAttr.IsOptional() {
			t.Error("Expected 'endpoint' attribute to be optional")
		}
	}

	// Check username attribute
	usernameAttr, ok := resp.Schema.Attributes["username"]
	if !ok {
		t.Error("Expected 'username' attribute in schema")
	} else {
		if !usernameAttr.IsOptional() {
			t.Error("Expected 'username' attribute to be optional")
		}
	}

	// Check password attribute
	passwordAttr, ok := resp.Schema.Attributes["password"]
	if !ok {
		t.Error("Expected 'password' attribute in schema")
	} else {
		if !passwordAttr.IsOptional() {
			t.Error("Expected 'password' attribute to be optional")
		}
		if !passwordAttr.IsSensitive() {
			t.Error("Expected 'password' attribute to be sensitive")
		}
	}

	// Check schema has a description
	if resp.Schema.MarkdownDescription == "" {
		t.Error("Expected schema to have a markdown description")
	}
}

func TestJellyfinProvider_Resources(t *testing.T) {
	p := &JellyfinProvider{}
	resources := p.Resources(context.Background())

	if len(resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(resources))
	}

	// Verify the resource can be instantiated
	r := resources[0]()
	if r == nil {
		t.Error("Expected resource to be instantiated")
	}
}

func TestJellyfinProvider_DataSources(t *testing.T) {
	p := &JellyfinProvider{}
	dataSources := p.DataSources(context.Background())

	if len(dataSources) != 1 {
		t.Errorf("Expected 1 data source, got %d", len(dataSources))
	}

	// Verify the data source can be instantiated
	ds := dataSources[0]()
	if ds == nil {
		t.Error("Expected data source to be instantiated")
	}
}

func TestJellyfinProvider_Functions(t *testing.T) {
	p := &JellyfinProvider{}
	functions := p.Functions(context.Background())

	if len(functions) != 0 {
		t.Errorf("Expected 0 functions, got %d", len(functions))
	}
}

func TestNew(t *testing.T) {
	factory := New("1.0.0")
	if factory == nil {
		t.Fatal("Expected factory function to be returned")
	}

	p := factory()
	if p == nil {
		t.Fatal("Expected provider to be instantiated")
	}

	// Verify the version is set correctly
	jp, ok := p.(*JellyfinProvider)
	if !ok {
		t.Fatal("Expected provider to be *JellyfinProvider")
	}

	if jp.version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %s", jp.version)
	}
}

func TestNew_differentVersions(t *testing.T) {
	testCases := []struct {
		version string
	}{
		{"dev"},
		{"test"},
		{"0.1.0"},
		{"1.2.3"},
	}

	for _, tc := range testCases {
		t.Run(tc.version, func(t *testing.T) {
			factory := New(tc.version)
			p := factory().(*JellyfinProvider)
			if p.version != tc.version {
				t.Errorf("Expected version %q, got %q", tc.version, p.version)
			}
		})
	}
}
