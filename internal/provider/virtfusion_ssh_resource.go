// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"io"
	"net/http"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &VirtfusionSSHResource{}
var _ resource.ResourceWithImportState = &VirtfusionSSHResource{}

func NewVirtfusionSSHResource() resource.Resource {
	return &VirtfusionSSHResource{}
}

// VirtfusionSSHResource defines the resource implementation.
type VirtfusionSSHResource struct {
	client *http.Client
}

// VirtfusionSSHResourceModel describes the resource data model.
type VirtfusionSSHResourceModel struct {
	UserId        *int64       `tfsdk:"user_id" json:"userId"`
	Name          *string      `tfsdk:"name" json:"name"`
	PublicKey     *string      `tfsdk:"public_key" json:"publicKey"`
	PublicKeyHash types.String `tfsdk:"public_key_hash"`
	Id            types.Int64  `tfsdk:"id" json:"id,omitempty"`
}

func (r *VirtfusionSSHResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh"
}

func (r *VirtfusionSSHResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Virtfusion SSH Resource",

		Attributes: map[string]schema.Attribute{
			"user_id": schema.Int64Attribute{
				Description: "User ID",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Key Name",
				Required:    true,
			},
			"public_key": schema.StringAttribute{
				Description: "Public Key",
				Required:    true,
			},
			"public_key_hash": schema.StringAttribute{
				Description: "Public Key Hash",
				Computed:    true,
				Default:     nil,
			},
			"id": schema.Int64Attribute{
				Description: "SSH Key ID",
				Computed:    true,
			},
		},
	}
}

func (r *VirtfusionSSHResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *VirtfusionSSHResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VirtfusionSSHResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq := VirtfusionSSHResourceModel{
		UserId:    data.UserId,
		Name:      data.Name,
		PublicKey: data.PublicKey,
	}

	// Convert the model to JSON
	jsonReq, err := json.Marshal(createReq)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to marshal request body",
			fmt.Sprintf("Failed to marshal request body: %s", err.Error()),
		)
		return
	}

	httpReq, err := r.client.Post("/ssh_keys", "application/json", bytes.NewBuffer(jsonReq))

	if err != nil {
		resp.Diagnostics.AddError(
			"Request failed",
			fmt.Sprintf("Request failed: %s", err.Error()),
		)
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to close response body",
				fmt.Sprintf("Failed to close response body: %s", err.Error()),
			)
		}
	}(httpReq.Body)

	if httpReq.StatusCode != 201 {

		if httpReq.StatusCode == 422 {
			responseBody, _ := io.ReadAll(httpReq.Body)
			var errorResponse map[string]interface{}
			err = json.Unmarshal(responseBody, &errorResponse)
			if errors, exists := errorResponse["errors"]; exists {
				resp.Diagnostics.AddError(
					"Failed to create SSH key",
					fmt.Sprintf("Errors from server: %v", errors),
				)

				return
			}
		}

		resp.Diagnostics.AddError(
			"Invalid Request",
			fmt.Sprintf("Failed to create SSH key: %s", httpReq.Status),
		)
		return
	}

	// Read the response body into the model. The response is expected to be a JSON object with the body of the created
	// ssh key within the `data` field. The `data` field is a JSON object with the ssh key data.
	responseBody, err := io.ReadAll(httpReq.Body)

	type ResponseData struct {
		Data struct {
			Id        int64  `json:"id"`
			Name      string `json:"name"`
			Type      string `json:"type"`
			CreatedAt string `json:"createdAt"`
		} `json:"data"`
	}

	var responseData ResponseData

	// Unmarshal the response body into the model
	err = json.Unmarshal(responseBody, &responseData)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to unmarshal response body",
			fmt.Sprintf("Failed to unmarshal response body: %s", err.Error()),
		)
		return
	}

	data.Id = types.Int64Value(responseData.Data.Id)
	data.Name = &responseData.Data.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionSSHResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VirtfusionSSHResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequest("GET", fmt.Sprintf("/ssh_keys/%d", data.Id.ValueInt64()), nil)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create Request",
			fmt.Sprintf("Failed to create a new HTTP request: %s", err.Error()),
		)
		return
	}

	// If the resource returns a 404, then the resource has been deleted. Return an empty state.
	httpResponse, err := r.client.Do(httpReq)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to close response body",
				fmt.Sprintf("Failed to close response body: %s", err.Error()),
			)
		}
	}(httpResponse.Body)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Execute Request",
			fmt.Sprintf("Failed to execute HTTP request: %s", err.Error()),
		)
		return
	}

	if httpResponse.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	var responseData struct {
		Data struct {
			Id            int64  `json:"id"`
			Name          string `json:"name"`
			Type          string `json:"type"`
			Enabled       bool   `json:"enabled"`
			CreatedAt     string `json:"created"`
			UpdatedAt     string `json:"updated"`
			PublicKeyHash string `json:"publicKey"`
		} `json:"data"`
	}

	err = json.NewDecoder(httpResponse.Body).Decode(&responseData)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to decode response body",
			fmt.Sprintf("Failed to decode response body: %s", err.Error()),
		)
		return
	}

	data.Name = &responseData.Data.Name
	data.PublicKeyHash = types.StringValue(responseData.Data.PublicKeyHash)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionSSHResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VirtfusionSSHResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return

	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionSSHResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VirtfusionSSHResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequest("DELETE", fmt.Sprintf("/ssh_keys/%d", data.Id.ValueInt64()), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create Request",
			fmt.Sprintf("Failed to create a new HTTP request: %s", err.Error()),
		)
		return
	}

	// Add any additional headers (Content-Type, etc.)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResponse, err := r.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Execute Request",
			fmt.Sprintf("Failed to execute HTTP request: %s", err.Error()),
		)
		return
	}

	if httpResponse.StatusCode != 204 {
		resp.Diagnostics.AddError(
			"Failed to Delete Resource",
			fmt.Sprintf("Failed to delete resource: %s", httpResponse.Status),
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Request failed",
			fmt.Sprintf("Request failed: %s", err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionSSHResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
