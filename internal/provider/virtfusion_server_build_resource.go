package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure implementation
var _ resource.Resource = &VirtfusionServerBuildResource{}

func NewVirtfusionServerBuildResource() resource.Resource {
	return &VirtfusionServerBuildResource{}
}

type VirtfusionServerBuildResource struct {
	client *http.Client
	config *ProviderConfig
}

type VirtfusionServerBuildResourceModel struct {
	ID       types.Int64   `tfsdk:"id"`
	ServerID types.Int64   `tfsdk:"server_id"`
	Name     types.String  `tfsdk:"name"`
	Hostname types.String  `tfsdk:"hostname"`
	OsID     types.Int64   `tfsdk:"osid"`
	VNC      types.Bool    `tfsdk:"vnc"`
	IPv6     types.Bool    `tfsdk:"ipv6"`
	SSHKeys  []types.Int64 `tfsdk:"ssh_keys"`
	Email    types.Bool    `tfsdk:"email"`
}

func (r *VirtfusionServerBuildResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "virtfusion_build"
}

func (r *VirtfusionServerBuildResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
			},
			"server_id": schema.Int64Attribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"hostname": schema.StringAttribute{
				Required: true,
			},
			"osid": schema.Int64Attribute{
				Optional: true,
			},
			"vnc": schema.BoolAttribute{
				Optional: true,
			},
			"ipv6": schema.BoolAttribute{
				Optional: true,
			},
			"ssh_keys": schema.ListAttribute{
				ElementType: types.Int64Type,
				Optional:    true,
			},
			"email": schema.BoolAttribute{
				Optional: true,
			},
		},
	}
}

func (r *VirtfusionServerBuildResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(*ProviderConfig)
	if !ok || config == nil {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			fmt.Sprintf("Expected *ProviderConfig, got: %T", req.ProviderData),
		)
		return
	}

	r.client = config.Client
	r.config = config
}

func (r *VirtfusionServerBuildResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VirtfusionServerBuildResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If no osid provided, try to resolve from provider default OsTemplate
	if data.OsID.IsNull() && r.config.OsTemplate != "" {
		osid, err := resolveOsTemplateToID(r.client, r.config.Endpoint, r.config.ApiToken, r.config.OsTemplate)
		if err != nil {
			resp.Diagnostics.AddError("OS Template Resolution Failed", err.Error())
			return
		}
		data.OsID = types.Int64Value(osid)
	}

	payload := map[string]interface{}{
		"server_id": data.ServerID.ValueInt64(),
		"name":      data.Name.ValueString(),
		"hostname":  data.Hostname.ValueString(),
		"osid":      data.OsID.ValueInt64(),
		"vnc":       data.VNC.ValueBool(),
		"ipv6":      data.IPv6.ValueBool(),
		"ssh_keys":  flattenInt64List(data.SSHKeys),
		"email":     data.Email.ValueBool(),
	}

	body, _ := json.Marshal(payload)
	reqURL := r.config.Endpoint + "/api/v1/build"

	httpReq, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(body))
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+r.config.ApiToken)

	httpResp, err := r.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("API request failed", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 && httpResp.StatusCode != 201 {
		resp.Diagnostics.AddError(
			"Unexpected API Response",
			fmt.Sprintf("Status: %d", httpResp.StatusCode),
		)
		return
	}

	var respData map[string]interface{}
	if err := json.NewDecoder(httpResp.Body).Decode(&respData); err != nil {
		resp.Diagnostics.AddError("Error decoding API response", err.Error())
		return
	}

	if id, ok := respData["id"].(float64); ok {
		data.ID = types.Int64Value(int64(id))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtfusionServerBuildResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VirtfusionServerBuildResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqURL := r.config.Endpoint + "/api/v1/build/" + strconv.FormatInt(data.ID.ValueInt64(), 10)
	httpReq, _ := http.NewRequest("GET", reqURL, nil)
	httpReq.Header.Set("Authorization", "Bearer "+r.config.ApiToken)

	httpResp, err := r.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("API request failed", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if httpResp.StatusCode != 200 {
		resp.Diagnostics.AddError("Unexpected API Response", fmt.Sprintf("Status: %d", httpResp.StatusCode))
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

	reqURL := r.config.Endpoint + "/api/v1/build/" + strconv.FormatInt(data.ID.ValueInt64(), 10)
	payload := map[string]interface{}{
		"name":     data.Name.ValueString(),
		"hostname": data.Hostname.ValueString(),
		"osid":     data.OsID.ValueInt64(),
		"vnc":      data.VNC.ValueBool(),
		"ipv6":     data.IPv6.ValueBool(),
		"ssh_keys": flattenInt64List(data.SSHKeys),
		"email":    data.Email.ValueBool(),
	}

	body, _ := json.Marshal(payload)
	httpReq, _ := http.NewRequest("PUT", reqURL, bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+r.config.ApiToken)

	httpResp, err := r.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("API request failed", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 {
		resp.Diagnostics.AddError("Unexpected API Response", fmt.Sprintf("Status: %d", httpResp.StatusCode))
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

	reqURL := r.config.Endpoint + "/api/v1/build/" + strconv.FormatInt(data.ID.ValueInt64(), 10)
	httpReq, _ := http.NewRequest("DELETE", reqURL, nil)
	httpReq.Header.Set("Authorization", "Bearer "+r.config.ApiToken)

	httpResp, err := r.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("API request failed", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 && httpResp.StatusCode != 204 {
		resp.Diagnostics.AddError("Unexpected API Response", fmt.Sprintf("Status: %d", httpResp.StatusCode))
		return
	}
}

// resolveOsTemplateToID resolves a template name to its numeric ID via API
func resolveOsTemplateToID(client *http.Client, endpoint, apiToken, templateName string) (int64, error) {
	reqURL := endpoint + "/api/v1/os-templates"
	httpReq, _ := http.NewRequest("GET", reqURL, nil)
	httpReq.Header.Set("Authorization", "Bearer "+apiToken)

	httpResp, err := client.Do(httpReq)
	if err != nil {
		return 0, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 {
		return 0, fmt.Errorf("unexpected status %d while fetching OS templates", httpResp.StatusCode)
	}

	var respData []map[string]interface{}
	if err := json.NewDecoder(httpResp.Body).Decode(&respData); err != nil {
		return 0, err
	}

	for _, tpl := range respData {
		if tpl["name"] == templateName {
			if id, ok := tpl["id"].(float64); ok {
				return int64(id), nil
			}
		}
	}

	return 0, fmt.Errorf("OS template %q not found", templateName)
}

// helper to convert []types.Int64 → []int64
func flattenInt64List(list []types.Int64) []int64 {
	var result []int64
	for _, v := range list {
		if !v.IsNull() {
			result = append(result, v.ValueInt64())
		}
	}
	return result
}
