// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	PackageId            *int64       `tfsdk:"package_id" json:"packageId,omitempty"`
	UserId               *int64       `tfsdk:"user_id" json:"userId,omitempty"`
	HypervisorId         *int64       `tfsdk:"hypervisor_id" json:"hypervisorId,omitempty"`
	Ipv4                 *int64       `tfsdk:"ipv4" json:"ipv4,omitempty"`
	Storage              *int64       `tfsdk:"storage" json:"storage,omitempty"`
	Memory               *int64       `tfsdk:"memory" json:"memory,omitempty"`
	Cores                *int64       `tfsdk:"cores" json:"cpuCores,omitempty"`
	Traffic              *int64       `tfsdk:"traffic" json:"traffic,omitempty"`
	InboundNetworkSpeed  *int64       `tfsdk:"inbound_network_speed" json:"networkSpeedInbound,omitempty"`
	OutboundNetworkSpeed *int64       `tfsdk:"outbound_network_speed" json:"networkSpeedOutbound,omitempty"`
	StorageProfile       *int64       `tfsdk:"storage_profile" json:"storageProfile,omitempty"`
	NetworkProfile       *int64       `tfsdk:"network_profile" json:"networkProfile,omitempty"`
	Name                 types.String `tfsdk:"name" json:"name,omitempty"`
	Id                   types.Int64  `tfsdk:"id" json:"id"`
	Build                struct {
		Name     types.String  `tfsdk:"name" json:"name"`
		Hostname types.String  `tfsdk:"hostname" json:"hostname"`
		Osid     types.Int64   `tfsdk:"osid" json:"operatingSystemId"`
		Vnc      types.Bool    `tfsdk:"vnc" json:"vnc"`
		Ipv6     types.Bool    `tfsdk:"ipv6" json:"ipv6"`
		SshKeys  []types.Int64 `tfsdk:"ssh_keys" json:"sshKeys"`
		Email    types.Bool    `tfsdk:"email" json:"email"`
	}
}

func (r *VirtfusionServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func GenerateRandomName() string {
	// Generate a random name for the server if one is not provided
	// It'll be in a format like "tf-<uuid>"
	return "tf-" + uuid.New().String()[:8]
}

func (r *VirtfusionServerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Virtfusion Server Resource",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Server name. If omitted, a random UUID will be generated.",
				Optional:            true,
				Computed:            true,
				//Default:  stringdefault.StaticString(GenerateRandomName()),
			},
			"package_id": schema.Int64Attribute{
				MarkdownDescription: "Package ID",
				Required:            true,
			},
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "User ID",
				Required:            true,
			},
			"hypervisor_id": schema.Int64Attribute{
				MarkdownDescription: "Hypervisor Group ID",
				Required:            true,
			},
			"ipv4": schema.Int64Attribute{
				MarkdownDescription: "IPv4 Addresses to assign. Omit to use the default of 1 IPv4.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
			},
			"storage": schema.Int64Attribute{
				MarkdownDescription: "Primary storage size in GB. Omit to use the default storage size from the package.",
				Optional:            true,
			},
			"memory": schema.Int64Attribute{
				MarkdownDescription: "How much memory to allocate in MB. Omit to use the default memory size from the package.",
				Optional:            true,
			},
			"cores": schema.Int64Attribute{
				MarkdownDescription: "How many cores to allocate. Omit to use the default core count from the package.",
				Optional:            true,
			},
			"traffic": schema.Int64Attribute{
				MarkdownDescription: "How much traffic to allocate in GB. Omit to use the default traffic size from the package. 0=Unlimited",
				Optional:            true,
			},
			"inbound_network_speed": schema.Int64Attribute{
				MarkdownDescription: "Inbound network speed in kB/s. Omit to use the default inbound network speed from the package.",
				Optional:            true,
			},
			"outbound_network_speed": schema.Int64Attribute{
				MarkdownDescription: "Outbound network speed in kB/s. Omit to use the default outbound network speed from the package.",
				Optional:            true,
			},
			"storage_profile": schema.Int64Attribute{
				MarkdownDescription: "Storage profile ID. Omit to use the default storage profile from the package.",
				Optional:            true,
			},
			"network_profile": schema.Int64Attribute{
				MarkdownDescription: "Network profile ID. Omit to use the default network profile from the package.",
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

func (r *VirtfusionServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VirtfusionServerResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq := VirtfusionServerResourceModel{
		PackageId:            data.PackageId,
		UserId:               data.UserId,
		HypervisorId:         data.HypervisorId,
		Ipv4:                 data.Ipv4,
		Storage:              data.Storage,
		Traffic:              data.Traffic,
		Memory:               data.Memory,
		Cores:                data.Cores,
		InboundNetworkSpeed:  data.InboundNetworkSpeed,
		OutboundNetworkSpeed: data.OutboundNetworkSpeed,
		StorageProfile:       data.StorageProfile,
		NetworkProfile:       data.NetworkProfile,
	}

	httpReqBody, err := json.Marshal(createReq)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while creating the resource create request. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	httpReq, err := http.NewRequest("POST", "/servers", bytes.NewBuffer(httpReqBody))
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

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Close Request",
				fmt.Sprintf("Failed to close HTTP request: %s", err.Error()),
			)
			return
		}
	}(httpResponse.Body)

	if httpResponse.StatusCode == 422 {
		responseBody, err := ioutil.ReadAll(httpResponse.Body)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Read Response",
				fmt.Sprintf("Failed to read HTTP response body: %s", err.Error()),
			)
			return
		}

		var errorResponse map[string]interface{}
		err = json.Unmarshal(responseBody, &errorResponse)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Parse Error Response",
				fmt.Sprintf("Failed to parse HTTP response body: %s", err.Error()),
			)
			return
		}

		if errors, exists := errorResponse["errors"]; exists {
			resp.Diagnostics.AddError(
				"Server Returned Errors",
				fmt.Sprintf("Errors from server: %v", errors),
			)
		}

		return
	}

	if httpResponse.StatusCode != 201 {
		resp.Diagnostics.AddError(
			"Failed to Create Resource",
			fmt.Sprintf("Failed to create resource: %s", httpResponse.Status),
		)
		return
	}

	responseBody, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read Response",
			fmt.Sprintf("Failed to read HTTP response body: %s", err.Error()),
		)
		return
	}

	type ResponseData struct {
		Data struct {
			Id   int64  `json:"id"`
			Uuid string `json:"uuid"`
			Name string `json:"name"`
		} `json:"data"`
	}

	var responseData ResponseData

	// Unmarshal the JSON response
	err = json.Unmarshal(responseBody, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Parse Response",
			fmt.Sprintf("Failed to parse HTTP response body: %s", err.Error()),
		)
		return
	}

	// Update the Terraform state with the server ID
	data.Id = types.Int64Value(responseData.Data.Id)
	data.Name = types.StringValue(responseData.Data.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VirtfusionServerResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VirtfusionServerResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VirtfusionServerResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequest("DELETE", fmt.Sprintf("/servers/%d?delay=0", data.Id.ValueInt64()), bytes.NewBuffer([]byte{}))
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
