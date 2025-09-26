// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	UserId    *int64      `tfsdk:"user_id" json:"userId"`
	Name      *string     `tfsdk:"name" json:"name"`
	PublicKey *string     `tfsdk:"public_key" json:"publicKey"`
	Id        types.Int64 `tfsdk:"id" json:"id,omitempty"`
}

func (r *VirtfusionSSHResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh"
}

func (r *VirtfusionSSHResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "VirtFusion SSH Key Resource",

		Attributes: map[string]schema.Attribute{
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "User ID to associate the key with.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Friendly name for the SSH key.",
				Required:            true,
			},
			"public_key": schema.StringAttribute{
				MarkdownDescription: "Public key string (e.g. ssh-ed25519 ...).",
				Required:            true,
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "SSH key ID in VirtFusion.",
				Computed:            true,
			},
		},
	}
}

func (r *VirtfusionSSHResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*http.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *VirtfusionSSHResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VirtfusionSSHResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReqBody, err := json.Marshal(data)
	if err != nil {
		resp.Diagnostics.AddError("Marshal Error", err.Error())
		return
	}

	httpReq, err := r.client.Post("/ssh_keys", "application/json", bytes.NewBuffer(httpReqBody))
	if err != nil {
		resp.Diagnostics.AddError("HTTP Request Error", err.Error())
		return
	}
	defer httpReq.Body.Close()

	if httpReq.StatusCode == 422 {
		body, _ := io.ReadAll(httpReq.Body)
		resp.Diagnostics.AddError("Validation Error", string(body))
		return
	}
	if httpReq.StatusCode != 201 {
		resp.Diagnostics.AddError("Unexpected HTTP Status", httpReq.Status)
		return
	}

	var response struct {
		Data struct {
			Id   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(httpReq.Body).Decode(&response); err != nil {
		resp.Diagnostics.AddError("Decode Error", err.Error())
		return
	}

	data.Id = types.Int64Value(response.Data.Id)
	data.Name = &response.Data.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionSSHResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VirtfusionSSHResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequest("GET", fmt.Sprintf("/ssh_keys/%d", data.Id.ValueInt64()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Request Error", err.Error())
		return
	}

	httpResponse, err := r.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Execute Error", err.Error())
		return
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode == 404 {
		// Key no longer exists
		resp.State.RemoveResource(ctx)
		return
	}
	if httpResponse.StatusCode != 200 {
		resp.Diagnostics.AddError("Unexpected HTTP Status", httpResponse.Status)
		return
	}

	var response struct {
		Data struct {
			Id   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(httpResponse.Body).Decode(&response); err != nil {
		resp.Diagnostics.AddError("Decode Error", err.Error())
		return
	}

	data.Id = types.Int64Value(response.Data.Id)
	data.Name = &response.Data.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionSSHResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VirtfusionSSHResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// No-op for now — SSH keys in VirtFusion are immutable except for delete/create.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionSSHResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VirtfusionSSHResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequest("DELETE", fmt.Sprintf("/ssh_keys/%d", data.Id.ValueInt64()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Request Error", err.Error())
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResponse, err := r.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Execute Error", err.Error())
		return
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != 204 {
		resp.Diagnostics.AddError("Delete Failed", httpResponse.Status)
		return
	}
}

func (r *VirtfusionSSHResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
