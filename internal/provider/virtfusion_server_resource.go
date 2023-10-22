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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	PackageId            types.Int64  `tfsdk:"package_id"`
	UserId               types.Int64  `tfsdk:"user_id"`
	HypervisorId         types.Int64  `tfsdk:"hypervisor_id"`
	Ipv4                 types.Int64  `tfsdk:"ipv4"`
	Storage              types.Int64  `tfsdk:"storage"`
	Memory               types.Int64  `tfsdk:"memory"`
	Cores                types.Int64  `tfsdk:"cores"`
	Traffic              types.Int64  `tfsdk:"traffic"`
	InboundNetworkSpeed  types.Int64  `tfsdk:"inbound_network_speed"`
	OutboundNetworkSpeed types.Int64  `tfsdk:"outbound_network_speed"`
	StorageProfile       types.Int64  `tfsdk:"storage_profile"`
	NetworkProfile       types.Int64  `tfsdk:"network_profile"`
	Name                 types.String `tfsdk:"name"`
	Id                   types.Int64  `tfsdk:"id"`
}

func (r *VirtfusionServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_example"
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
				Default:             stringdefault.StaticString(GenerateRandomName()),
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
				MarkdownDescription: "Hypervisor ID",
				Required:            true,
			},
			"ipv4": schema.Int64Attribute{
				MarkdownDescription: "How many IPv4 addresses to allocate. Default is 1.",
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
			Id   types.Int64  `json:"id"`
			Uuid types.String `json:"uuid"`
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
	data.Id = types.Int64(responseData.Data.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	//tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	//resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
	//     return
	// }
}

func (r *VirtfusionServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
