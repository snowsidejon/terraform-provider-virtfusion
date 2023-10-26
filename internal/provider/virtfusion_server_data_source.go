package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
	//	OsTemplateInstall   *bool   `tfsdk:"os_template_install" json:"osTemplateInstall"`
	//	OsTemplateInstallId types.Int64  `tfsdk:"os_template_install_id" json:"osTemplateInstallId"`
	//	EncryptedPassword   *string `tfsdk:"encrypted_password" json:"encryptedPassword"`
	//	BackupPlan          types.Int64  `tfsdk:"backup_plan" json:"backupPlan"`
	//	Uefi                *bool   `tfsdk:"uefi" json:"uefi"`
	//	CloudInit           *bool   `tfsdk:"cloud_init" json:"cloudInit"`
	//	CloudInitType       types.Int64  `tfsdk:"cloud_init_type" json:"cloudInitType"`
	//	Resources           struct {
	//		Memory  types.Int64 `tfsdk:"memory" json:"memory"`
	//		Cores   types.Int64 `tfsdk:"cores" json:"cpuCores"`
	//		Storage types.Int64 `tfsdk:"storage" json:"storage"`
	//		Traffic types.Int64 `tfsdk:"traffic" json:"traffic"`
	//	} `tfsdk:"resources" json:"resources"`
	//} `tfsdk:"settings" json:"settings"`
	Network *NetworkStruct `tfsdk:"network" json:"network"`
	//Owner struct {
	//	Id            types.Int64  `tfsdk:"id" json:"id"`
	//	Name          *string `tfsdk:"name" json:"name"`
	//	Email         *string `tfsdk:"email" json:"email"`
	//	ExtId         types.Int64  `tfsdk:"ext_id" json:"extRelationId"`
	//	Timezone      *string `tfsdk:"timezone" json:"timezone"`
	//	Suspended     *bool   `tfsdk:"suspended" json:"suspended"`
	//	Created       *string `tfsdk:"created" json:"created"`
	//	Updated       *string `tfsdk:"updated" json:"updated"`
	//	TwoFactorAuth *bool   `tfsdk:"two_factor_auth" json:"twoFactorAuth"`
	//} `tfsdk:"owner" json:"owner"`
}

type NetworkStruct struct {
	Interfaces []struct {
		Enabled bool   `tfsdk:"enabled" json:"enabled"`
		Name    string `tfsdk:"name" json:"name"`
		Type    string `tfsdk:"type" json:"type"`
		Mac     string `tfsdk:"mac" json:"mac"`
		Ipv4    []struct {
			Id          int64  `tfsdk:"id" json:"id"`
			Enabled     bool   `tfsdk:"enabled" json:"enabled"`
			Address     string `tfsdk:"address" json:"address"`
			Netmask     string `tfsdk:"netmask" json:"netmask"`
			Gateway     string `tfsdk:"gateway" json:"gateway"`
			ResolverOne string `tfsdk:"resolver_one" json:"resolverOne"`
			ResolverTwo string `tfsdk:"resolver_two" json:"resolverTwo"`
			Rdns        string `tfsdk:"rdns" json:"rdns"`
			Mac         string `tfsdk:"mac" json:"mac"`
		} `tfsdk:"ipv4" json:"ipv4"`
		Ipv6 []struct {
		} `tfsdk:"ipv6" json:"ipv6"`
	} `tfsdk:"interfaces" json:"interfaces"`
}

type VirtfusionServerDataResponseModel struct {
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
		//	OsTemplateInstall   *bool   `json:"os_template_install"`
		//	OsTemplateInstallId *int64  `json:"os_template_install_id"`
		//	EncryptedPassword   *string `json:"encrypted_password"`
		//	BackupPlan          *int64  `json:"backup_plan"`
		//	Uefi                *bool   `json:"uefi"`
		//	CloudInit           *bool   `json:"cloud_init"`
		//	CloudInitType       *int64  `json:"cloud_init_type"`
		//	Resources           struct {
		//		Memory  *int64 `json:"memory"`
		//		Cores   *int64 `json:"cpu_cores"`
		//		Storage *int64 `json:"storage"`
		//		Traffic *int64 `json:"traffic"`
		//	} `json:"resources"`
		//} `json:"settings"`
		Network *NetworkStruct `tfsdk:"network" json:"network"`
		//Owner struct {
		//	Id            *int64  `json:"id"`
		//	Name          *string `json:"name"`
		//	Email         *string `json:"email"`
		//	ExtId         *int64  `json:"ext_id"`
		//	Timezone      *string `json:"timezone"`
		//	Suspended     *bool   `json:"suspended"`
		//	Created       *string `json:"created"`
		//	Updated       *string `json:"updated"`
		//	TwoFactorAuth *bool   `json:"two_factor_auth"`
		//} `json:"owner"`
	} `json:"data"`
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
	var responseData VirtfusionServerDataResponseModel

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
	resp.Diagnostics.AddWarning(
		"Terraform is not able to read the server data",
		fmt.Sprintf("Terraform is not able to read the server data: %s", httpResponse.Body),
	)

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

	data.Network = responseData.Data.Network

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
