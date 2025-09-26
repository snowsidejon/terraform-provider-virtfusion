// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &VirtfusionServerResource{}
var _ resource.ResourceWithImportState = &VirtfusionServerResource{}

func NewVirtfusionServerResource() resource.Resource {
	return &VirtfusionServerResource{}
}

// VirtfusionServerResource defines the resource implementation.
type VirtfusionServerResource struct {
	client *http.Client
}

// VirtfusionServerResourceModel describes the resource data model.
type VirtfusionServerResourceModel struct {
	PackageId         *int64      `tfsdk:"package_id" json:"packageId,omitempty"`
	UserId            *int64      `tfsdk:"user_id" json:"userId,omitempty"`
	HypervisorId      *int64      `tfsdk:"hypervisor_id" json:"hypervisorId,omitempty"`
	HypervisorGroupId *int64      `tfsdk:"hypervisor_group_id" json:"hypervisorGroupId,omitempty"`
	Ipv4              *int64      `tfsdk:"ipv4" json:"ipv4,omitempty"`
	Storage           *int64      `tfsdk:"storage" json:"storage,omitempty"`
	Memory            *int64      `tfsdk:"memory" json:"memory,omitempty"`
	Cores             *int64      `tfsdk:"cores" json:"cpuCores,omitempty"`
	Traffic           *int64      `tfsdk:"traffic" json:"traffic,omitempty"`
	InboundNet        *int64      `tfsdk:"inbound_network_speed" json:"networkSpeedInbound,omitempty"`
	OutboundNet       *int64      `tfsdk:"outbound_network_speed" json:"networkSpeedOutbound,omitempty"`
	StorageProfile    *int64      `tfsdk:"storage_profile" json:"storageProfile,omitempty"`
	NetworkProfile    *int64      `tfsdk:"network_profile" json:"networkProfile,omitempty"`
	Id                types.Int64 `tfsdk:"id" json:"id"`
}

func (r *VirtfusionServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *VirtfusionServerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "VirtFusion Server Resource",

		Attributes: map[string]schema.Attribute{
			"package_id": schema.Int64Attribute{
				MarkdownDescription: "Package ID. Defaults from provider-level `resource_package` if omitted.",
				Optional:            true,
			},
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "User ID.",
				Required:            true,
			},
			"hypervisor_id": schema.Int64Attribute{
				MarkdownDescription: "Specific Hypervisor ID to place the server on. Conflicts with hypervisor_group_id.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRoot("hypervisor_group_id")),
				},
			},
			"hypervisor_group_id": schema.Int64Attribute{
				MarkdownDescription: "Hypervisor Group (location) ID. Defaults from provider-level `hypervisor_group` if omitted.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRoot("hypervisor_id")),
				},
			},
			"ipv4": schema.Int64Attribute{
				MarkdownDescription: "IPv4 addresses to assign. Defaults from provider-level `public_ips` if omitted.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
			},
			"storage": schema.Int64Attribute{
				MarkdownDescription: "Primary storage size in GB. Omit for package default.",
				Optional:            true,
			},
			"memory": schema.Int64Attribute{
				MarkdownDescription: "Memory in MB. Omit for package default.",
				Optional:            true,
			},
			"cores": schema.Int64Attribute{
				MarkdownDescription: "CPU core count. Omit for package default.",
				Optional:            true,
			},
			"traffic": schema.Int64Attribute{
				MarkdownDescription: "Traffic in GB. 0 = unlimited.",
				Optional:            true,
			},
			"inbound_network_speed": schema.Int64Attribute{
				MarkdownDescription: "Inbound network speed (kB/s).",
				Optional:            true,
			},
			"outbound_network_speed": schema.Int64Attribute{
				MarkdownDescription: "Outbound network speed (kB/s).",
				Optional:            true,
			},
			"storage_profile": schema.Int64Attribute{
				MarkdownDescription: "Storage profile ID.",
				Optional:            true,
			},
			"network_profile": schema.Int64Attribute{
				MarkdownDescription: "Network profile ID.",
				Optional:            true,
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "Server ID",
				Computed:            true,
			},
		},
	}
}

func (r *VirtfusionServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VirtfusionServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VirtfusionServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build request dynamically
	createReq := map[string]interface{}{
		"userId": data.UserId,
	}
	if data.PackageId != nil {
		createReq["packageId"] = data.PackageId
	}
	if data.HypervisorId != nil {
		createReq["hypervisorId"] = data.HypervisorId
	}
	if data.HypervisorGroupId != nil {
		createReq["hypervisorGroupId"] = data.HypervisorGroupId
	}
	if data.Ipv4 != nil {
		createReq["ipv4"] = data.Ipv4
	}
	if data.Storage != nil {
		createReq["storage"] = data.Storage
	}
	if data.Traffic != nil {
		createReq["traffic"] = data.Traffic
	}
	if data.Memory != nil {
		createReq["memory"] = data.Memory
	}
	if data.Cores != nil {
		createReq["cpuCores"] = data.Cores
	}
	if data.InboundNet != nil {
		createReq["networkSpeedInbound"] = data.InboundNet
	}
	if data.OutboundNet != nil {
		createReq["networkSpeedOutbound"] = data.OutboundNet
	}
	if data.StorageProfile != nil {
		createReq["storageProfile"] = data.StorageProfile
	}
	if data.NetworkProfile != nil {
		createReq["networkProfile"] = data.NetworkProfile
	}

	httpReqBody, err := json.Marshal(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Marshal Error", err.Error())
		return
	}

	httpReq, err := http.NewRequest("POST", "/servers", bytes.NewBuffer(httpReqBody))
	if err != nil {
		resp.Diagnostics.AddError("HTTP Request Error", err.Error())
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResponse, err := r.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("HTTP Execute Error", err.Error())
		return
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode == 422 {
		body, _ := ioutil.ReadAll(httpResponse.Body)
		resp.Diagnostics.AddError("Server Validation Error", string(body))
		return
	}
	if httpResponse.StatusCode != 201 {
		resp.Diagnostics.AddError("Unexpected HTTP Status", httpResponse.Status)
		return
	}

	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		resp.Diagnostics.AddError("Read Body Error", err.Error())
		return
	}

	var response struct {
		Data struct {
			Id int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		resp.Diagnostics.AddError("Unmarshal Error", err.Error())
		return
	}

	data.Id = types.Int64Value(response.Data.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VirtfusionServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VirtfusionServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VirtfusionServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	httpReq, err := http.NewRequest("DELETE", fmt.Sprintf("/servers/%d?delay=0", data.Id.ValueInt64()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Delete Request Error", err.Error())
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpResponse, err := r.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Delete Execute Error", err.Error())
		return
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode != 204 {
		resp.Diagnostics.AddError("Delete Failed", httpResponse.Status)
		return
	}
}

func (r *VirtfusionServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
