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

// Ensure VirtfusionProvider satisfies provider.Provider interface
var _ provider.Provider = &VirtfusionProvider{}

// VirtfusionProvider implements the provider.
type VirtfusionProvider struct {
	version string
}

// ProviderConfig is shared with resources and data sources.
type ProviderConfig struct {
	Client          *http.Client
	Endpoint        string
	ApiToken        string
	OsTemplate      string
	ResourcePackage int64
	PublicIPs       int64
	PrivateIPs      int64
	HypervisorGroup int64
}

// VirtfusionProviderModel describes the provider schema.
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
				MarkdownDescription: "VirtFusion API endpoint (default: cloud.breezehost.io).",
				Optional:            true,
			},
			"api_token": schema.StringAttribute{
				MarkdownDescription: "API token for authentication.",
				Optional:            true,
				Sensitive:           true,
			},
			"os_template": schema.StringAttribute{
				MarkdownDescription: "Default OS template name (default: Ubuntu Server 22.04).",
				Optional:            true,
			},
			"resource_package": schema.Int64Attribute{
				MarkdownDescription: "Default resource package ID.",
				Optional:            true,
			},
			"public_ips": schema.Int64Attribute{
				MarkdownDescription: "Default number of public IPs (default: 1).",
				Optional:            true,
			},
			"private_ips": schema.Int64Attribute{
				MarkdownDescription: "Default number of private IPs (default: 0).",
				Optional:            true,
			},
			"hypervisor_group": schema.Int64Attribute{
				MarkdownDescription: "Default hypervisor group ID (location).",
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

	// Environment defaults
	apiToken := os.Getenv("VIRTFUSION_API_TOKEN")
	endpoint := os.Getenv("VIRTFUSION_ENDPOINT")
	osTemplate := os.Getenv("VIRTFUSION_OS_TEMPLATE")
	resourcePackage := int64(0)
	publicIPs := int64(1)
	privateIPs := int64(0)
	hypervisorGroup := int64(1)

	// Override from config
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
		resourcePackage = data.ResourcePackage.ValueInt64()
	} else if env := os.Getenv("VIRTFUSION_RESOURCE_PACKAGE"); env != "" {
		if v, err := strconv.ParseInt(env, 10, 64); err == nil {
			resourcePackage = v
		}
	}

	if !data.PublicIPs.IsNull() {
		publicIPs = data.PublicIPs.ValueInt64()
	} else if env := os.Getenv("VIRTFUSION_PUBLIC_IPS"); env != "" {
		if v, err := strconv.ParseInt(env, 10, 64); err == nil {
			publicIPs = v
		}
	}

	if !data.PrivateIPs.IsNull() {
		privateIPs = data.PrivateIPs.ValueInt64()
	} else if env := os.Getenv("VIRTFUSION_PRIVATE_IPS"); env != "" {
		if v, err := strconv.ParseInt(env, 10, 64); err == nil {
			privateIPs = v
		}
	}

	if !data.HypervisorGroup.IsNull() {
		hypervisorGroup = data.HypervisorGroup.ValueInt64()
	} else if env := os.Getenv("VIRTFUSION_HYPERVISOR_GROUP"); env != "" {
		if v, err := strconv.ParseInt(env, 10, 64); err == nil {
			hypervisorGroup = v
		}
	}

	if apiToken == "" {
		resp.Diagnostics.AddError(
			"Missing API Token",
			"No API token provided via config or VIRTFUSION_API_TOKEN env var.",
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

	// Share provider config with resources
	config := &ProviderConfig{
		Client:          client,
		Endpoint:        endpoint,
		ApiToken:        apiToken,
		OsTemplate:      osTemplate,
		ResourcePackage: resourcePackage,
		PublicIPs:       publicIPs,
		PrivateIPs:      privateIPs,
		HypervisorGroup: hypervisorGroup,
	}

	resp.DataSourceData = config
	resp.ResourceData = config
}

func (p *VirtfusionProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewVirtfusionServerResource,
		NewVirtfusionServerBuildResource,
		NewVirtfusionSSHResource,
	}
}

func (p *VirtfusionProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
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
		return &VirtfusionProvider{version: version}
	}
}
