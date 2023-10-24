// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &VirtfusionPackageResource{}
var _ resource.ResourceWithImportState = &VirtfusionPackageResource{}

func NewVirtfusionPackageResource() resource.Resource {
	return &VirtfusionPackageResource{}
}

// VirtfusionPackageResource defines the resource implementation.
type VirtfusionPackageResource struct {
	client *http.Client
}

// VirtfusionPackageResourceModel describes the resource data model.
type VirtfusionPackageResourceModel struct {
	Id          types.Int64  `tfsdk:"id" json:"id"`
	Name        types.String `tfsdk:"name" json:"name"`
	CpuCores    types.Int64  `tfsdk:"cpu_cores" json:"cpu_cores"`
	CpuModel    types.String `tfsdk:"cpu_model" json:"cpu_model"`
	CpuShares   types.Int64  `tfsdk:"cpu_shares" json:"cpu_shares"`
	Topology    CpuTopologyModel
	Description types.String `tfsdk:"description" json:"description"`
	Enabled     types.Bool   `tfsdk:"enabled" json:"enabled"`
	DiskType    types.String `tfsdk:"disk_type" json:"disk_type"`
	ForceIpv6   types.Bool   `tfsdk:"force_ipv6" json:"force_ipv6"`
	MachineType types.String `tfsdk:"machine_type" json:"machine_type"`
	Memory      types.Int64  `tfsdk:"memory" json:"memory"`
	Network     NetworkModel
	PciPorts    types.Int64 `tfsdk:"pci_ports" json:"pci_ports"`
}

type Storagemodel struct {
	DiskSize      types.Int64 `tfsdk:"disk_size" json:"storage"`
	DiskType      types.Int64 `tfsdk:"disk_type" json:"disk_type"`
	ReadBytesSec  types.Int64 `tfsdk:"read_bytes_sec" json:"read_bytes_sec"`
	ReadIopsSec   types.Int64 `tfsdk:"read_iops_sec" json:"read_iops_sec"`
	WriteBytesSec types.Int64 `tfsdk:"write_bytes_sec" json:"write_bytes_sec"`
	WriteIopsSec  types.Int64 `tfsdk:"write_iops_sec" json:"write_iops_sec"`
	StorageType   types.Int64 `tfsdk:"storage_type" json:"storage_type"`
}

type NetworkModel struct {
	Profile  types.Int64 `tfsdk:"profile" json:"network_profile"`
	SpeedIn  types.Int64 `tfsdk:"speed_in" json:"network_speed_in"`
	SpeedOut types.Int64 `tfsdk:"speed_out" json:"network_speed_out"`
	Traffic  types.Int64 `tfsdk:"traffic" json:"traffic"`
}

type CpuTopologyModel struct {
	Cores   types.Int64 `tfsdk:"cores" json:"cpu_topology_cores"`
	Sockets types.Int64 `tfsdk:"sockets" json:"cpu_topology_sockets"`
	Threads types.Int64 `tfsdk:"threads" json:"cpu_topology_threads"`
	Enabled types.Bool  `tfsdk:"enabled" json:"cpu_topology_enabled"`
}

func (r *VirtfusionPackageResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_package"
}

func (r *VirtfusionPackageResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Virtfusion Package Resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the package.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the package.",
			},
			"cpu_cores": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The number of CPU cores.",
			},
			"cpu_model": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The CPU model. Omit to use the default, which is to inherit from the hyper-visor settings.",
				Computed:            true,
				Default:             stringdefault.StaticString("inherit"),
			},
			"cpu_shares": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The CPU shares. Omit to use the default, which is to inherit from the hyper-visor settings.",
				Computed:            true,
				Default:             int64default.StaticInt64(1024),
			},
		},
	}
}

func (r *VirtfusionPackageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VirtfusionPackageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VirtfusionPackageResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionPackageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VirtfusionPackageResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionPackageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VirtfusionPackageResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return

	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionPackageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VirtfusionPackageResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionPackageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
