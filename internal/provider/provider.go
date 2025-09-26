// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure VirtfusionProvider satisfies various provider interfaces.
var _ provider.Provider = &VirtfusionProvider{}

// VirtfusionProvider defines the provider implementation.
type VirtfusionProvider struct {
	// version is set at release time.
	version string
}

// VirtfusionProviderModel describes the provider data model.
type VirtfusionProviderModel struct {
	Endpoint        types.String `tfsdk:"endpoint"`
	ApiToken        types.String `tfsdk:"api_token"`
	OsTemplate      types.String `tfsdk:"os_template"`
	ResourcePackage types.Int64  `tfsdk:"resource_package"`
	PublicIPs       types.Int64  `tfsdk:"public_ips"`
	PrivateIPs      types.Int64  `tfsdk:"private_ips"`
	HypervisorGroup types.Int64  `tfsdk:"hypervisor_group"`
}

func (p *VirtfusionProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "virtfusion"
	resp.Version = p.version
}

func (p *VirtfusionProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The endpoint to use for API requests. Defaults to https://cloud.breezehost.io",
				Optional:            true,
			},
			"api_token": schema.StringAttribute{
				MarkdownDescription: "The API token to use for API requests.",
				Optional:            true,
				Sensitive:           true,
			},
			"os_template": schema.StringAttribute{
				MarkdownDescription: "The OS template to deploy. Defaults to Ubuntu Server 22.04.",
				Optional:            true,
			},
			"resource_package": schema.Int64Attribute{
				MarkdownDescription: "Resource package ID (RAM/CPU/Disk plan).",
				Optional:            true,
			},
			"public_ips": schema.Int64Attribute{
				MarkdownDescription: "Number of public IPs to assign (default: 1).",
				Optional:            true,
			},
			"private_ips": schema.Int64Attribute{
				MarkdownDescription: "Number of private IPs to assign (default: 0).",
				Optional:            true,
			},
			"hypervisor_group": schema.Int64Attribute{
				MarkdownDescription: "Hypervisor group (location) ID to deploy into.",
				Optional:            true,
			},
		},
	}
}

func (p *VirtfusionProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data VirtfusionProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Pull from env first
	apiToken := os.Getenv("VIRTFUSION_API_TOKEN")
	endpoint := os.Getenv("VIRTFUSION_ENDPOINT")
	osTemplate := os.Getenv("VIRTFUSION_OS_TEMPLATE")
	resourcePackage := os.Getenv("VIRTFUSION_RESOURCE_PACKAGE")
	publicIPs := os.Getenv("VIRTFUSION_PUBLIC_IPS")
	privateIPs := os.Getenv("VIRTFUSION_PRIVATE_IPS")
	hypervisorGroup := os.Getenv("VIRTFUSION_HYPERVISOR_GROUP")

	// Override from config block if set
	if !data.Endpoint.IsNull() {
		endpoint = data.Endpoint.ValueString()
	}
	if endpoint == "" {
		endpoint = "cloud.breezehost.io"
	}

	if !data.ApiToken.IsNull() {
		apiToken = data.ApiToken.ValueString()
	}

	if !data.OsTemplate.IsNull() {
		osTemplate = data.OsTemplate.ValueString()
	}
	if osTemplate == "" {
		osTemplate = "Ubuntu Server 22.04"
	}

	if !data.ResourcePackage.IsNull() {
		resourcePackage = strconv.FormatInt(data.ResourcePackage.ValueInt64(), 10)
	}

	if !data.PublicIPs.IsNull() {
		publicIPs = strconv.FormatInt(data.PublicIPs.ValueInt64(), 10)
	}
	if publicIPs == "" {
		publicIPs = "1"
	}

	if !data.PrivateIPs.IsNull() {
		privateIPs = strconv.FormatInt(data.PrivateIPs.ValueInt64(), 10)
	}
	if privateIPs == "" {
		privateIPs = "0"
	}

	if !data.HypervisorGroup.IsNull() {
		hypervisorGroup = strconv.FormatInt(data.HypervisorGroup.ValueInt64(), 10)
	}
	if hypervisorGroup == "" {
		hypervisorGroup = "1"
	}

	if apiToken == "" {
		resp.Diagnostics.AddError(
			"Missing API Token",
			"No API token provided via VIRTFUSION_API_TOKEN or provider config.",
		)
		return
	}

	// Build HTTP client
	customTransport := &CustomTransport{
		Transport: http.DefaultTransport,
		BaseURL:   &url.URL{Scheme: "https", Host: endpoint, Path: "/api/v1"},
		Token:     apiToken,
	}

	client := &http.Client{Transport: customTransport}

	// Pass client to resources and datasources
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *VirtfusionProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewVirtfusionServerResource,
		NewVirtfusionServerBuildResource,
		NewVirtfusionSSHResource,
	}
}

func (p *VirtfusionProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// add datasources here (e.g. templates, hypervisors)
	}
}

type CustomTransport struct {
	Transport http.RoundTripper
	BaseURL   *url.URL
	Token     string
}

func (c *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.URL.Scheme = c.BaseURL.Scheme
	req.URL.Host = c.BaseURL.Host
	req.URL.Path = path.Join(c.BaseURL.Path, req.URL.Path)
	return c.Transport.RoundTrip(req)
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &VirtfusionProvider{
			version: version,
		}
	}
}
