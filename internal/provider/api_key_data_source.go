// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/jkossis/terraform-provider-jellyfin/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &APIKeyDataSource{}

func NewAPIKeyDataSource() datasource.DataSource {
	return &APIKeyDataSource{}
}

// APIKeyDataSource defines the data source implementation.
type APIKeyDataSource struct {
	client *client.Client
}

// APIKeyDataSourceModel describes the data source data model.
type APIKeyDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	AppName     types.String `tfsdk:"app_name"`
	AccessToken types.String `tfsdk:"access_token"`
	DateCreated types.String `tfsdk:"date_created"`
}

func (d *APIKeyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (d *APIKeyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about a Jellyfin API key.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier for this data source (same as access_token).",
			},
			"app_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the application. Either app_name or access_token must be provided.",
			},
			"access_token": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The API key token. Either app_name or access_token must be provided.",
			},
			"date_created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time when the API key was created.",
			},
		},
	}
}

func (d *APIKeyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *APIKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data APIKeyDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	hasAppName := !data.AppName.IsNull() && !data.AppName.IsUnknown()
	hasAccessToken := !data.AccessToken.IsNull() && !data.AccessToken.IsUnknown()

	if !hasAppName && !hasAccessToken {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'app_name' or 'access_token' must be provided to look up an API key.",
		)
		return
	}

	var key *client.APIKey
	var err error

	if hasAccessToken {
		key, err = d.client.GetKeyByAccessToken(ctx, data.AccessToken.ValueString())
	} else {
		key, err = d.client.FindKeyByAppName(ctx, data.AppName.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read API key: %s", err))
		return
	}

	if key == nil {
		resp.Diagnostics.AddError(
			"API Key Not Found",
			"The specified API key was not found.",
		)
		return
	}

	// Set the data using the AccessToken as the data source ID
	data.ID = types.StringValue(key.AccessToken)
	data.AppName = types.StringValue(key.AppName)
	data.AccessToken = types.StringValue(key.AccessToken)
	data.DateCreated = types.StringValue(key.DateCreated)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
