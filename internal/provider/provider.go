// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/jkossis/terraform-provider-jellyfin/internal/client"
)

// Ensure JellyfinProvider satisfies various provider interfaces.
var _ provider.Provider = &JellyfinProvider{}
var _ provider.ProviderWithFunctions = &JellyfinProvider{}

// JellyfinProvider defines the provider implementation.
type JellyfinProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// JellyfinProviderModel describes the provider data model.
type JellyfinProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (p *JellyfinProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "jellyfin"
	resp.Version = p.version
}

func (p *JellyfinProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Terraform provider for managing Jellyfin resources via the Jellyfin API. " +
			"The provider authenticates using username and password credentials.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The Jellyfin server URL (e.g., http://localhost:8096). Can also be set via the `JELLYFIN_ENDPOINT` environment variable.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The Jellyfin username for authentication. Can also be set via the `JELLYFIN_USERNAME` environment variable.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The Jellyfin password for authentication. Can also be set via the `JELLYFIN_PASSWORD` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *JellyfinProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data JellyfinProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check for environment variables if not set in config
	endpoint := data.Endpoint.ValueString()
	if endpoint == "" {
		endpoint = os.Getenv("JELLYFIN_ENDPOINT")
	}

	username := data.Username.ValueString()
	if username == "" {
		username = os.Getenv("JELLYFIN_USERNAME")
	}

	password := data.Password.ValueString()
	if password == "" {
		password = os.Getenv("JELLYFIN_PASSWORD")
	}

	// Validate required configuration
	if endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing Jellyfin Endpoint",
			"The provider cannot create the Jellyfin API client as there is a missing or empty value for the Jellyfin endpoint. "+
				"Set the endpoint value in the configuration or use the JELLYFIN_ENDPOINT environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddError(
			"Missing Jellyfin Username",
			"The provider cannot create the Jellyfin API client as there is a missing or empty value for the Jellyfin username. "+
				"Set the username value in the configuration or use the JELLYFIN_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddError(
			"Missing Jellyfin Password",
			"The provider cannot create the Jellyfin API client as there is a missing or empty value for the Jellyfin password. "+
				"Set the password value in the configuration or use the JELLYFIN_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create Jellyfin API client with authentication
	jellyfinClient, err := client.NewClientWithAuth(ctx, endpoint, username, password)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Authenticate with Jellyfin",
			"The provider failed to authenticate with the Jellyfin server. "+
				"Please verify your credentials and ensure the Jellyfin server is accessible. "+
				"Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = jellyfinClient
	resp.ResourceData = jellyfinClient
}

func (p *JellyfinProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAPIKeyResource,
	}
}

func (p *JellyfinProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAPIKeyDataSource,
	}
}

func (p *JellyfinProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &JellyfinProvider{
			version: version,
		}
	}
}
