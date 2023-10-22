// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Virtfusion Server Build Resource",

		Attributes: map[string]schema.Attribute{
			"server_id": schema.Int64Attribute{
				MarkdownDescription: "Server ID",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Server Name",
				Required:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Server Hostname",
				Optional:            true,
			},
			"osid": schema.Int64Attribute{
				MarkdownDescription: "Server Operating System ID",
				Required:            true,
			},
			"vnc": schema.BoolAttribute{
				MarkdownDescription: "Server VNC",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"ipv6": schema.BoolAttribute{
				MarkdownDescription: "Server IPv6",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"ssh_keys": schema.ListAttribute{
				MarkdownDescription: "Server SSH Keys IDs",
				ElementType:         types.Int64Type,
				Optional:            true,
			},
			"email": schema.BoolAttribute{
				MarkdownDescription: "Server Email",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *VirtfusionServerBuildResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VirtfusionServerBuildResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VirtfusionServerBuildResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
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
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while creating the resource create request. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)

		return
	}

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("/servers/%d/build", data.ServerId), bytes.NewBuffer(httpReqBody))

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

	if httpResponse.StatusCode != 200 {
		resp.Diagnostics.AddError(
			"Failed to Create Resource",
			fmt.Sprintf("Failed to create resource: %s", httpResponse.Status),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionServerBuildResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VirtfusionServerBuildResourceModel

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

func (r *VirtfusionServerBuildResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VirtfusionServerBuildResourceModel

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

func (r *VirtfusionServerBuildResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VirtfusionServerBuildResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionServerBuildResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
