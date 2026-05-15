package datasources

import (
	"context"
	"fmt"

	"github.com/deevus/terraform-provider-truenas/internal/services"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SSHConnectionDataSource{}
var _ datasource.DataSourceWithConfigure = &SSHConnectionDataSource{}

// SSHConnectionDataSource defines the data source implementation.
type SSHConnectionDataSource struct {
	services *services.TrueNASServices
}

// SSHConnectionDataSourceModel describes the data source data model.
type SSHConnectionDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Host           types.String `tfsdk:"host"`
	Port           types.Int32  `tfsdk:"port"`
	Username       types.String `tfsdk:"username"`
	PrivateKeyID   types.Int64  `tfsdk:"private_key_id"`
	RemoteHostKey  types.String `tfsdk:"remote_host_key"`
	ConnectTimeout types.Int32  `tfsdk:"connect_timeout"`
}

// NewSSHConnectionDataSource creates a new SSHConnectionDataSource.
func NewSSHConnectionDataSource() datasource.DataSource {
	return &SSHConnectionDataSource{}
}

func (d *SSHConnectionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_connection"
}

func (d *SSHConnectionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing TrueNAS ssh connection.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the SSH connection.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the SSH connection to look up.",
				Required:    true,
			},
			"host": schema.StringAttribute{
				Description: "Hostname or IP address of the remote system.",
				Computed:    true,
			},
			"port": schema.Int32Attribute{
				Description: "Port number to connect to on the remote system",
				Computed:    true,
			},
			"username": schema.StringAttribute{
				Description: "Username on the remote system which will be used to login via SSH.",
				Computed:    true,
			},
			"private_key_id": schema.Int64Attribute{
				Description: "ID of the SSH keypair to authenticate with the remote host.",
				Computed:    true,
			},
			"remote_host_key": schema.StringAttribute{
				Description: "Remote system SSH key for this system to authenticate the connection.",
				Computed:    true,
			},
			"connect_timeout": schema.Int32Attribute{
				Description: "Time (in seconds) before the system stops attempting to establish a connection with the remote system.",
				Computed:    true,
			},
		},
	}
}

func (d *SSHConnectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured
	if req.ProviderData == nil {
		return
	}

	s, ok := req.ProviderData.(*services.TrueNASServices)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *services.TrueNASServices, got: %T.", req.ProviderData),
		)
		return
	}

	d.services = s
}

func (d *SSHConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SSHConnectionDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// List all connections via the service
	connections, err := d.services.SSH.ListSSHConnections(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read SSH Connections",
			fmt.Sprintf("Unable to read SSH connections: %s", err.Error()),
		)
		return
	}

	// Find the connection with matching name
	searchName := data.Name.ValueString()
	found := false
	for _, connection := range connections {
		if connection.Name == searchName {
			data.ID = types.StringValue(fmt.Sprintf("%d", connection.ID))
			data.Name = types.StringValue(connection.Name)
			data.Host = types.StringValue(connection.Host)
			data.Port = types.Int32Value(connection.Port)
			data.Username = types.StringValue(connection.Username)
			data.PrivateKeyID = types.Int64Value(connection.PrivateKeyID)
			data.RemoteHostKey = types.StringValue(connection.RemoteHostKey)
			data.ConnectTimeout = types.Int32Value(connection.ConnectTimeout)
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError(
			"SSH Connection Not Found",
			fmt.Sprintf("SSH keypaconnectionir %q was not found.", searchName),
		)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
