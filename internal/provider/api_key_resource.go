// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/jkossis/terraform-provider-jellyfin/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &APIKeyResource{}
var _ resource.ResourceWithImportState = &APIKeyResource{}

func NewAPIKeyResource() resource.Resource {
	return &APIKeyResource{}
}

// APIKeyResource defines the resource implementation.
type APIKeyResource struct {
	client *client.Client
}

// APIKeyResourceModel describes the resource data model.
type APIKeyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	AppName     types.String `tfsdk:"app_name"`
	AccessToken types.String `tfsdk:"access_token"`
	DateCreated types.String `tfsdk:"date_created"`
}

func (r *APIKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *APIKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Jellyfin API key.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier for this resource (same as access_token).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the application using this API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"access_token": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The API key token used for authentication.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"date_created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time when the API key was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *APIKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *APIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data APIKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	appName := data.AppName.ValueString()

	tflog.Debug(ctx, "Creating API key", map[string]interface{}{
		"app_name": appName,
	})

	// Get existing keys before creation to find the new one after
	existingKeys, err := r.client.GetKeys(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list existing API keys: %s", err))
		return
	}

	existingIDs := make(map[int64]bool)
	for _, key := range existingKeys.Items {
		existingIDs[key.Id] = true
	}

	// Create the new API key
	err = r.client.CreateKey(ctx, appName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create API key: %s", err))
		return
	}

	// Find the newly created key by comparing with existing keys
	newKeys, err := r.client.GetKeys(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list API keys after creation: %s", err))
		return
	}

	var createdKey *client.APIKey
	for _, key := range newKeys.Items {
		if !existingIDs[key.Id] && key.AppName == appName {
			createdKey = &key
			break
		}
	}

	if createdKey == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to find the newly created API key")
		return
	}

	// Set the resource data using the AccessToken as the terraform resource ID
	// (Jellyfin API doesn't return a stable Id for API keys)
	data.ID = types.StringValue(createdKey.AccessToken)
	data.AccessToken = types.StringValue(createdKey.AccessToken)
	data.DateCreated = types.StringValue(createdKey.DateCreated)

	tflog.Trace(ctx, "Created API key resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data APIKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// The ID is the AccessToken
	accessToken := data.ID.ValueString()

	key, err := r.client.GetKeyByAccessToken(ctx, accessToken)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read API key: %s", err))
		return
	}

	if key == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with key information
	data.ID = types.StringValue(key.AccessToken)
	data.AppName = types.StringValue(key.AppName)
	data.AccessToken = types.StringValue(key.AccessToken)
	data.DateCreated = types.StringValue(key.DateCreated)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data APIKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Jellyfin API doesn't support updating API keys
	// The app_name has RequiresReplace, so this should only be called if there are no changes
	tflog.Trace(ctx, "Update called for API key resource (no-op)")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data APIKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// The Jellyfin API expects the AccessToken in the delete path, not the Id
	accessToken := data.AccessToken.ValueString()

	tflog.Debug(ctx, "Deleting API key", map[string]interface{}{
		"id":           data.ID.ValueString(),
		"access_token": accessToken,
	})

	err := r.client.DeleteKey(ctx, accessToken)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete API key: %s", err))
		return
	}

	tflog.Trace(ctx, "Deleted API key resource")
}

func (r *APIKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
