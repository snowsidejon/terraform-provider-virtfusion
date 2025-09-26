// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &VirtfusionServerBuildResource{}
var _ resource.ResourceWithImportState = &VirtfusionServerBuildResource{}

func NewVirtfusionServerBuildResource() resource.Resource {
	return &VirtfusionServerBuildResource{}
}

// VirtfusionServerBuildResource defines the resource implementation.
type VirtfusionServerBuildResource struct {
	client *http.Client
}

type VirtfusionServerBuildResourceModel struct {
	ServerId int64   `tfsdk:"server_id"`
	Name     string  `tfsdk:"name" json:"name"`
	Hostname string  `tfsdk:"hostname" json:"hostname"`
	Osid     int64   `tfsdk:"osid" json:"operatingSystemId"`
	Vnc      bool    `tfsdk:"vnc" json:"vnc"`
	Ipv6     bool    `tfsdk:"ipv6" json:"ipv6"`
	SshKeys  []int64 `tfsdk:"ssh_keys" json:"sshKeys"`
	Email    bool    `tfsdk:"email" json:"email"`
}

func (r *VirtfusionServerBuildResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_build"
}

func (r *VirtfusionServerBuildResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "VirtFusion Server Build Resource. Creates and initializes a VM with OS, hostname, and SSH keys.",

		Attributes: map[string]schema.Attribute{
			"server_id": schema.Int64Attribute{
				MarkdownDescription: "Server ID to build.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the server.",
				Required:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Hostname for the server.",
				Optional:            true,
			},
			"osid": schema.Int64Attribute{
				MarkdownDescription: "Operating system ID. If omitted, the provider will resolve from `os_template` default.",
				Optional:            true,
			},
			"vnc": schema.BoolAttribute{
				MarkdownDescription: "Enable VNC access.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"ipv6": schema.BoolAttribute{
				MarkdownDescription: "Enable IPv6 networking.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"ssh_keys": schema.ListAttribute{
				MarkdownDescription: "List of SSH key IDs to provision on the VM.",
				ElementType:         types.Int64Type,
				Optional:            true,
			},
			"email": schema.BoolAttribute{
				MarkdownDescription: "Send email on build completion.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *VirtfusionServerBuildResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VirtfusionServerBuildResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VirtfusionServerBuildResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If no osid is provided, try to resolve it from provider os_template default.
	if data.Osid == 0 {
		osid, err := r.lookupDefaultOS()
		if err != nil {
			resp.Diagnostics.AddError("OS Lookup Failed", err.Error())
			return
		}
		data.Osid = osid
	}

	createReq := VirtfusionServerBuildResourceModel{
		Name:     data.Name,
		Hostname: data.Hostname,
		Osid:     data.Osid,
		Vnc:      data.Vnc,
		Ipv6:     data.Ipv6,
		SshKeys:  data.SshKeys,
		Email:    data.Email,
	}

	httpReqBody, err := json.Marshal(createReq)
	if err != nil {
		resp.Diagnostics.AddError("JSON Marshal Error", err.Error())
		return
	}

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("/servers/%d/build", data.ServerId), bytes.NewBuffer(httpReqBody))
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
		responseBody, _ := ioutil.ReadAll(httpResponse.Body)
		resp.Diagnostics.AddError("Server Validation Error", string(responseBody))
		return
	}
	if httpResponse.StatusCode != 200 {
		resp.Diagnostics.AddError("Unexpected HTTP Status", httpResponse.Status)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionServerBuildResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VirtfusionServerBuildResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionServerBuildResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VirtfusionServerBuildResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionServerBuildResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VirtfusionServerBuildResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Nothing to delete — build is a one-time operation.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionServerBuildResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupDefaultOS resolves the provider's default os_template string into an OS ID.
// For now, this does a simple GET to /operating-systems and matches by name.
func (r *VirtfusionServerBuildResource) lookupDefaultOS() (int64, error) {
	httpReq, err := http.NewRequest("GET", "/operating-systems", nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create OS lookup request: %s", err)
	}

	httpResponse, err := r.client.Do(httpReq)
	if err != nil {
		return 0, fmt.Errorf("failed to execute OS lookup request: %s", err)
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != 200 {
		return 0, fmt.Errorf("unexpected status during OS lookup: %s", httpResponse.Status)
	}

	var response struct {
		Data []struct {
			Id   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	body, _ := io.ReadAll(httpResponse.Body)
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, fmt.Errorf("failed to parse OS lookup response: %s", err)
	}

	// Default template comes from provider env (VIRTFUSION_OS_TEMPLATE) or fallback
	defaultTemplate := os.Getenv("VIRTFUSION_OS_TEMPLATE")
	if defaultTemplate == "" {
		defaultTemplate = "Ubuntu Server 22.04"
	}

	for _, osInfo := range response.Data {
		if osInfo.Name == defaultTemplate {
			return osInfo.Id, nil
		}
	}

	return 0, fmt.Errorf("default OS template %q not found", defaultTemplate)
}
