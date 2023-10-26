package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &VirtfusionServerDataSource{}

func NewVirtfusionServerDataSource() datasource.DataSource {
	return &VirtfusionServerDataSource{}
}

// VirtfusionServerDataSource defines the data source implementation.
type VirtfusionServerDataSource struct {
	client *http.Client
}

// VirtfusionServerDataSourceModel describes the data source data model.
type VirtfusionServerDataSourceModel struct {
	Id                  types.Int64  `tfsdk:"id" json:"id"`
	OwnerId             types.Int64  `tfsdk:"owner_id" json:"ownerId"`
	HypervisorId        types.Int64  `tfsdk:"hypervisor_id" json:"hypervisorId"`
	Name                types.String `tfsdk:"name" json:"name"`
	Hostname            types.String `tfsdk:"hostname" json:"hostname"`
	CommissionStatus    types.Int64  `tfsdk:"commission_status" json:"commissionStatus"`
	Uuid                types.String `tfsdk:"uuid" json:"uuid"`
	State               types.String `tfsdk:"state" json:"state"`
	MigrateLevel        types.Int64  `tfsdk:"migrate_level" json:"migrateLevel"`
	DeleteLevel         types.Int64  `tfsdk:"delete_level" json:"deleteLevel"`
	ConfigLevel         types.Int64  `tfsdk:"config_level" json:"configLevel"`
	Rebuild             types.Bool   `tfsdk:"rebuild" json:"rebuild"`
	Suspended           types.Bool   `tfsdk:"suspended" json:"suspended"`
	Protected           types.Bool   `tfsdk:"protected" json:"protected"`
	BuildFailed         types.Bool   `tfsdk:"build_failed" json:"buildFailed"`
	PrimaryNetworkDhcp4 types.Bool   `tfsdk:"primary_network_dhcp4" json:"primaryNetworkDhcp4"`
	PrimaryNetworkDhcp6 types.Bool   `tfsdk:"primary_network_dhcp6" json:"primaryNetworkDhcp6"`
	Built               types.String `tfsdk:"built" json:"built"`
	Created             types.String `tfsdk:"created" json:"created"`
	Updated             types.String `tfsdk:"updated" json:"updated"`
	//Settings            struct {
	//	OsTemplateInstall   types.Bool   `tfsdk:"os_template_install" json:"osTemplateInstall"`
	//	OsTemplateInstallId types.Int64  `tfsdk:"os_template_install_id" json:"osTemplateInstallId"`
	//	EncryptedPassword   types.String `tfsdk:"encrypted_password" json:"encryptedPassword"`
	//	BackupPlan          types.Int64  `tfsdk:"backup_plan" json:"backupPlan"`
	//	Uefi                types.Bool   `tfsdk:"uefi" json:"uefi"`
	//	CloudInit           types.Bool   `tfsdk:"cloud_init" json:"cloudInit"`
	//	CloudInitType       types.Int64  `tfsdk:"cloud_init_type" json:"cloudInitType"`
	//	Resources           struct {
	//		Memory  types.Int64 `tfsdk:"memory" json:"memory"`
	//		Cores   types.Int64 `tfsdk:"cores" json:"cpuCores"`
	//		Storage types.Int64 `tfsdk:"storage" json:"storage"`
	//		Traffic types.Int64 `tfsdk:"traffic" json:"traffic"`
	//	} `tfsdk:"resources" json:"resources"`
	//} `tfsdk:"settings" json:"settings"`
	Network struct {
		Interfaces []struct {
			Enabled types.Bool   `tfsdk:"enabled" json:"enabled"`
			Name    types.String `tfsdk:"name" json:"name"`
			Type    types.String `tfsdk:"type" json:"type"`
			Mac     types.String `tfsdk:"mac" json:"mac"`
			Ipv4    []struct {
				Id          types.Int64  `tfsdk:"id" json:"id"`
				Enabled     types.Bool   `tfsdk:"enabled" json:"enabled"`
				Address     types.String `tfsdk:"address" json:"address"`
				Netmask     types.String `tfsdk:"netmask" json:"netmask"`
				Gateway     types.String `tfsdk:"gateway" json:"gateway"`
				ResolverOne types.String `tfsdk:"resolver_one" json:"resolverOne"`
				ResolverTwo types.String `tfsdk:"resolver_two" json:"resolverTwo"`
				Rdns        types.String `tfsdk:"rdns" json:"rdns"`
				Mac         types.String `tfsdk:"mac" json:"mac"`
			} `tfsdk:"ipv4" json:"ipv4"`
			Ipv6 []struct {
			} `tfsdk:"ipv6" json:"ipv6"`
		} `tfsdk:"interfaces" json:"interfaces"`
	} `tfsdk:"network" json:"network"`
	//Owner struct {
	//	Id            types.Int64  `tfsdk:"id" json:"id"`
	//	Name          types.String `tfsdk:"name" json:"name"`
	//	Email         types.String `tfsdk:"email" json:"email"`
	//	ExtId         types.Int64  `tfsdk:"ext_id" json:"extRelationId"`
	//	Timezone      types.String `tfsdk:"timezone" json:"timezone"`
	//	Suspended     types.Bool   `tfsdk:"suspended" json:"suspended"`
	//	Created       types.String `tfsdk:"created" json:"created"`
	//	Updated       types.String `tfsdk:"updated" json:"updated"`
	//	TwoFactorAuth types.Bool   `tfsdk:"two_factor_auth" json:"twoFactorAuth"`
	//} `tfsdk:"owner" json:"owner"`
}

func (d *VirtfusionServerDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (d *VirtfusionServerDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "The Virtfusion Server data source allows access to the details of a server." +
			" This data source can be used to read information about an existing server.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Server ID",
				Required:            true,
			},
			"owner_id": schema.Int64Attribute{
				MarkdownDescription: "Owner ID",
				Computed:            true,
			},
			"hypervisor_id": schema.Int64Attribute{
				MarkdownDescription: "Hypervisor ID",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Server name",
				Computed:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Server hostname",
				Computed:            true,
			},
			"commission_status": schema.Int64Attribute{
				MarkdownDescription: "Commission status",
				Computed:            true,
			},
			"uuid": schema.StringAttribute{
				MarkdownDescription: "Server UUID",
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Server state",
				Computed:            true,
			},
			"migrate_level": schema.Int64Attribute{
				MarkdownDescription: "Migrate level",
				Computed:            true,
			},
			"delete_level": schema.Int64Attribute{
				MarkdownDescription: "Delete level",
				Computed:            true,
			},
			"config_level": schema.Int64Attribute{
				MarkdownDescription: "Config level",
				Computed:            true,
			},
			"rebuild": schema.BoolAttribute{
				MarkdownDescription: "Rebuild",
				Computed:            true,
			},
			"suspended": schema.BoolAttribute{
				MarkdownDescription: "Suspended",
				Computed:            true,
			},
			"protected": schema.BoolAttribute{
				MarkdownDescription: "Protected",
				Computed:            true,
			},
			"build_failed": schema.BoolAttribute{
				MarkdownDescription: "Build failed",
				Computed:            true,
			},
			"primary_network_dhcp4": schema.BoolAttribute{
				MarkdownDescription: "Primary network DHCP4",
				Computed:            true,
			},
			"primary_network_dhcp6": schema.BoolAttribute{
				MarkdownDescription: "Primary network DHCP6",
				Computed:            true,
			},
			"built": schema.StringAttribute{
				MarkdownDescription: "Built",
				Computed:            true,
			},
			"created": schema.StringAttribute{
				MarkdownDescription: "Created",
				Computed:            true,
			},
			"updated": schema.StringAttribute{
				MarkdownDescription: "Updated",
				Computed:            true,
			},
			"network": schema.SingleNestedAttribute{
				MarkdownDescription: "Network interfaces",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"interfaces": schema.SingleNestedAttribute{
						MarkdownDescription: "Network interfaces",
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								MarkdownDescription: "Enabled",
								Computed:            true,
							},
							"name": schema.StringAttribute{
								MarkdownDescription: "Name",
								Computed:            true,
							},
							"type": schema.StringAttribute{
								MarkdownDescription: "Type",
								Computed:            true,
							},
							"mac": schema.StringAttribute{
								MarkdownDescription: "MAC",
								Computed:            true,
							},
							"ipv4": schema.SingleNestedAttribute{
								MarkdownDescription: "IPv4",
								Computed:            true,
								Attributes: map[string]schema.Attribute{
									"id": schema.Int64Attribute{
										MarkdownDescription: "ID",
										Computed:            true,
									},
									"enabled": schema.BoolAttribute{
										MarkdownDescription: "Enabled",
										Computed:            true,
									},
									"address": schema.StringAttribute{
										MarkdownDescription: "Address",
										Computed:            true,
									},
									"netmask": schema.StringAttribute{
										MarkdownDescription: "Netmask",
										Computed:            true,
									},
									"gateway": schema.StringAttribute{
										MarkdownDescription: "Gateway",
										Computed:            true,
									},
									"resolver_one": schema.StringAttribute{
										MarkdownDescription: "Resolver one",
										Computed:            true,
									},
									"resolver_two": schema.StringAttribute{
										MarkdownDescription: "Resolver two",
										Computed:            true,
									},
									"rdns": schema.StringAttribute{
										MarkdownDescription: "RDNS",
										Computed:            true,
									},
									"mac": schema.StringAttribute{
										MarkdownDescription: "MAC",
										Computed:            true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *VirtfusionServerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *VirtfusionServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VirtfusionServerDataSourceModel
	var responseData struct {
		Data struct {
			Id                  int64  `json:"id"`
			OwnerId             int64  `json:"owner_id"`
			HypervisorId        int64  `json:"hypervisor_id"`
			Name                string `json:"name"`
			Hostname            string `json:"hostname"`
			CommissionStatus    int64  `json:"commission_status"`
			Uuid                string `json:"uuid"`
			State               string `json:"state"`
			MigrateLevel        int64  `json:"migrate_level"`
			DeleteLevel         int64  `json:"delete_level"`
			ConfigLevel         int64  `json:"config_level"`
			Rebuild             bool   `json:"rebuild"`
			Suspended           bool   `json:"suspended"`
			Protected           bool   `json:"protected"`
			BuildFailed         bool   `json:"build_failed"`
			PrimaryNetworkDhcp4 bool   `json:"primary_network_dhcp4"`
			PrimaryNetworkDhcp6 bool   `json:"primary_network_dhcp6"`
			Built               string `json:"built"`
			Created             string `json:"created"`
			Updated             string `json:"updated"`
			//Settings            struct {
			//	OsTemplateInstall   types.Bool   `json:"os_template_install"`
			//	OsTemplateInstallId types.Int64  `json:"os_template_install_id"`
			//	EncryptedPassword   types.String `json:"encrypted_password"`
			//	BackupPlan          types.Int64  `json:"backup_plan"`
			//	Uefi                types.Bool   `json:"uefi"`
			//	CloudInit           types.Bool   `json:"cloud_init"`
			//	CloudInitType       types.Int64  `json:"cloud_init_type"`
			//	Resources           struct {
			//		Memory  types.Int64 `json:"memory"`
			//		Cores   types.Int64 `json:"cpu_cores"`
			//		Storage types.Int64 `json:"storage"`
			//		Traffic types.Int64 `json:"traffic"`
			//	} `json:"resources"`
			//} `json:"settings"`
			Network struct {
				Interfaces []struct {
					Enabled bool   `json:"enabled"`
					Name    string `json:"name"`
					Type    string `json:"type"`
					Mac     string `json:"mac"`
					Ipv4    []struct {
						Id          int64  `json:"id"`
						Enabled     bool   `json:"enabled"`
						Address     string `json:"address"`
						Netmask     string `json:"netmask"`
						Gateway     string `json:"gateway"`
						ResolverOne string `json:"resolver_one"`
						ResolverTwo string `json:"resolver_two"`
						Rdns        string `json:"rdns"`
						Mac         string `json:"mac"`
					} `json:"ipv4"`
					Ipv6 []struct {
					} `json:"ipv6"`
				} `json:"interfaces"`
			} `json:"network"`
			//Owner struct {
			//	Id            types.Int64  `json:"id"`
			//	Name          types.String `json:"name"`
			//	Email         types.String `json:"email"`
			//	ExtId         types.Int64  `json:"ext_id"`
			//	Timezone      types.String `json:"timezone"`
			//	Suspended     types.Bool   `json:"suspended"`
			//	Created       types.String `json:"created"`
			//	Updated       types.String `json:"updated"`
			//	TwoFactorAuth types.Bool   `json:"two_factor_auth"`
			//} `json:"owner"`
		} `json:"data"`
	}

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Make an API request using the configured client
	httpReq, err := http.NewRequest("GET", fmt.Sprintf("/servers/%d", data.Id.ValueInt64()), nil)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create Request",
			fmt.Sprintf("Failed to create a new HTTP request: %s", err.Error()),
		)
		return
	}

	// If the resource returns a 404, then the resource has been deleted. Return an empty state.
	httpResponse, err := d.client.Do(httpReq)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to close response body",
				fmt.Sprintf("Failed to close response body: %s", err.Error()),
			)
		}
	}(httpResponse.Body)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Execute Request",
			fmt.Sprintf("Failed to execute HTTP request: %s", err.Error()),
		)
		return
	}

	if httpResponse.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	err = json.NewDecoder(httpResponse.Body).Decode(&responseData)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to decode response body",
			fmt.Sprintf("Failed to decode response body: %s", err.Error()),
		)
		return
	}

	data.Id = types.Int64Value(responseData.Data.Id)
	data.OwnerId = types.Int64Value(responseData.Data.OwnerId)
	data.HypervisorId = types.Int64Value(responseData.Data.HypervisorId)
	data.Name = types.StringValue(responseData.Data.Name)
	data.Hostname = types.StringValue(responseData.Data.Hostname)
	data.CommissionStatus = types.Int64Value(responseData.Data.CommissionStatus)
	data.Uuid = types.StringValue(responseData.Data.Uuid)
	data.State = types.StringValue(responseData.Data.State)
	data.MigrateLevel = types.Int64Value(responseData.Data.MigrateLevel)
	data.DeleteLevel = types.Int64Value(responseData.Data.DeleteLevel)
	data.ConfigLevel = types.Int64Value(responseData.Data.ConfigLevel)
	data.Rebuild = types.BoolValue(responseData.Data.Rebuild)
	data.Suspended = types.BoolValue(responseData.Data.Suspended)
	data.Protected = types.BoolValue(responseData.Data.Protected)
	data.BuildFailed = types.BoolValue(responseData.Data.BuildFailed)
	data.PrimaryNetworkDhcp4 = types.BoolValue(responseData.Data.PrimaryNetworkDhcp4)
	data.PrimaryNetworkDhcp6 = types.BoolValue(responseData.Data.PrimaryNetworkDhcp6)
	data.Built = types.StringValue(responseData.Data.Built)
	data.Created = types.StringValue(responseData.Data.Created)
	data.Updated = types.StringValue(responseData.Data.Updated)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
