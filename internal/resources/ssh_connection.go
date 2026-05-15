package resources

import (
	"context"
	"fmt"
	"strconv"

	truenas "github.com/deevus/truenas-go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &SSHConnectionResource{}
	_ resource.ResourceWithConfigure   = &SSHConnectionResource{}
	_ resource.ResourceWithImportState = &SSHConnectionResource{}
)

// SSHConnectionResourceModel describes the resource data model.
type SSHConnectionResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Host           types.String `tfsdk:"host"`
	Port           types.Int32  `tfsdk:"port"`
	Username       types.String `tfsdk:"username"`
	PrivateKeyID   types.Int64  `tfsdk:"private_key_id"`
	RemoteHostKey  types.String `tfsdk:"remote_host_key"`
	ConnectTimeout types.Int32  `tfsdk:"connect_timeout"`
}

// SSHConnectionResource defines the resource implementation.
type SSHConnectionResource struct {
	BaseResource
}

// NewSSHConnectionResource creates a new SSHConnectionResource.
func NewSSHConnectionResource() resource.Resource {
	return &SSHConnectionResource{}
}

func (r *SSHConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_connection"
}

func (r *SSHConnectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages settings for SSH connections.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the SSH credential.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the SSH credential.",
				Required:    true,
			},
			"host": schema.StringAttribute{
				Description: "Hostname or IP address of the remote system.",
				Required:    true,
			},
			"port": schema.Int32Attribute{
				Description: "Port number to connect to on the remote system",
				Required:    true,
			},
			"username": schema.StringAttribute{
				Description: "Username on the remote system which will be used to login via SSH.",
				Required:    true,
			},
			"private_key_id": schema.Int64Attribute{
				Description: "ID of the SSH keypair to authenticate with the remote host.",
				Required:    true,
			},
			"remote_host_key": schema.StringAttribute{
				Description: "Remote system SSH key for this system to authenticate the connection.",
				Required:    true,
			},
			"connect_timeout": schema.Int32Attribute{
				Description: "Time (in seconds) before the system stops attempting to establish a connection with the remote system.",
				Optional:    true,
			},
		},
	}
}

// buildSSHConnectionOpts builds typed options from the resource model.
func buildSSHConnectionOpts(data *SSHConnectionResourceModel) truenas.CreateSSHConnectionOpts {
	opts := truenas.CreateSSHConnectionOpts{
		Name:           data.Name.ValueString(),
		Host:           data.Host.ValueString(),
		Port:           data.Port.ValueInt32(),
		Username:       data.Username.ValueString(),
		PrivateKeyID:   data.PrivateKeyID.ValueInt64(),
		RemoteHostKey:  data.RemoteHostKey.ValueString(),
		ConnectTimeout: data.ConnectTimeout.ValueInt32(),
	}

	return opts
}

func (r *SSHConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SSHConnectionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := buildSSHConnectionOpts(&data)

	credential, err := r.services.SSH.CreateSSHConnection(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create SSH Credential",
			fmt.Sprintf("Unable to create SSH credential: %s", err.Error()),
		)
		return
	}

	if credential == nil {
		resp.Diagnostics.AddError(
			"SSH Credential Not Found",
			"SSH credential was created but could not be found.",
		)
		return
	}

	mapSSHConnectionToModel(credential, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SSHConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SSHConnectionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(data.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid ID",
			fmt.Sprintf("Unable to parse ID %q: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}

	credential, err := r.services.SSH.GetSSHConnection(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read SSH Credential",
			fmt.Sprintf("Unable to query SSH credential: %s", err.Error()),
		)
		return
	}

	if credential == nil {
		// SSH credential was deleted outside Terraform
		resp.State.RemoveResource(ctx)
		return
	}

	mapSSHConnectionToModel(credential, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SSHConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state SSHConnectionResourceModel
	var plan SSHConnectionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse ID from state
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid ID",
			fmt.Sprintf("Unable to parse ID %q: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	opts := buildSSHConnectionOpts(&plan)

	credential, err := r.services.SSH.UpdateSSHConnection(ctx, id, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update SSH Credential",
			fmt.Sprintf("Unable to update SSH credential: %s", err.Error()),
		)
		return
	}

	if credential == nil {
		resp.Diagnostics.AddError(
			"SSH Credential Not Found",
			"SSH credential was updated but could not be found.",
		)
		return
	}

	// Set state from response
	mapSSHConnectionToModel(credential, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SSHConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SSHConnectionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(data.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid ID",
			fmt.Sprintf("Unable to parse ID %q: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}

	err = r.services.SSH.DeleteSSHConnection(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete SSH Credential",
			fmt.Sprintf("Unable to delete SSH credential: %s", err.Error()),
		)
		return
	}
}

// mapSSHConnectionToModel maps a typed SSHConnection to the resource model.
func mapSSHConnectionToModel(sshConnection *truenas.SSHConnection, data *SSHConnectionResourceModel) {
	data.ID = types.StringValue(strconv.FormatInt(sshConnection.ID, 10))
	data.Name = types.StringValue(sshConnection.Name)
	data.Host = types.StringValue(sshConnection.Host)
	data.Port = types.Int32Value(sshConnection.Port)
	data.Username = types.StringValue(sshConnection.Username)
	data.PrivateKeyID = types.Int64Value(sshConnection.PrivateKeyID)
	data.RemoteHostKey = types.StringValue(sshConnection.RemoteHostKey)
	data.ConnectTimeout = types.Int32Value(sshConnection.ConnectTimeout)
}
