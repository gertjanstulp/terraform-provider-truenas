package datasources

import (
	"context"
	"fmt"

	"github.com/deevus/terraform-provider-truenas/internal/services"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SSHKeyPairDataSource{}
var _ datasource.DataSourceWithConfigure = &SSHKeyPairDataSource{}

// SSHKeyPairDataSource defines the data source implementation.
type SSHKeyPairDataSource struct {
	services *services.TrueNASServices
}

// SSHKeyPairDataSourceModel describes the data source data model.
type SSHKeyPairDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	PublicKey  types.String `tfsdk:"public_key"`
	PrivateKey types.String `tfsdk:"private_key"`
}

// NewSSHKeyPairDataSource creates a new SSHKeyPairDataSource.
func NewSSHKeyPairDataSource() datasource.DataSource {
	return &SSHKeyPairDataSource{}
}

func (d *SSHKeyPairDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_keypair"
}

func (d *SSHKeyPairDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing TrueNAS ssh keypair.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the SSH keypair.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the SSH keypair to look up.",
				Required:    true,
			},
			"public_key": schema.StringAttribute{
				Description: "The public key of the SSH keypair.",
				Computed:    true,
			},
			"private_key": schema.StringAttribute{
				Description: "The private key of the SSH keypair.",
				Computed:    true,
			},
		},
	}
}

func (d *SSHKeyPairDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SSHKeyPairDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SSHKeyPairDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// List all credentials via the service
	keyPairs, err := d.services.KeychainCredential.ListSSHKeyPairs(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read SSH KeyPairs",
			fmt.Sprintf("Unable to read SSH keypairs: %s", err.Error()),
		)
		return
	}

	// Find the keypair with matching name
	searchName := data.Name.ValueString()
	found := false
	for _, keyPair := range keyPairs {
		if keyPair.Name == searchName {
			data.ID = types.StringValue(fmt.Sprintf("%d", keyPair.ID))
			data.Name = types.StringValue(keyPair.Name)
			data.PublicKey = types.StringValue(keyPair.PublicKey)
			data.PrivateKey = types.StringValue(keyPair.PrivateKey)
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError(
			"SSH keypair Not Found",
			fmt.Sprintf("SSH keypair %q was not found.", searchName),
		)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
