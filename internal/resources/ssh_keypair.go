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
	_ resource.Resource                = &SSHKeyPairResource{}
	_ resource.ResourceWithConfigure   = &SSHKeyPairResource{}
	_ resource.ResourceWithImportState = &SSHKeyPairResource{}
)

// SSHKeyPairResourceModel describes the resource data model.
type SSHKeyPairResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	PublicKey  types.String `tfsdk:"public_key"`
	PrivateKey types.String `tfsdk:"private_key"`
}

// SSHKeyPairResource defines the resource implementation.
type SSHKeyPairResource struct {
	BaseResource
}

// NewSSHKeyPairResource creates a new SSHKeyPairResource.
func NewSSHKeyPairResource() resource.Resource {
	return &SSHKeyPairResource{}
}

func (r *SSHKeyPairResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_keypair"
}

func (r *SSHKeyPairResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages keypair for SSH connections.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the SSH keypair.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the SSH keypair.",
				Required:    true,
			},
			"public_key": schema.StringAttribute{
				Description: "The public key of the SSH keypair.",
				Required:    true,
				Sensitive:   true,
			},
			"private_key": schema.StringAttribute{
				Description: "The private key of the SSH keypair.",
				Required:    true,
				Sensitive:   true,
			},
		},
	}
}

// buildSSHKeyPairOpts builds typed options from the resource model.
func buildSSHKeyPairOpts(data *SSHKeyPairResourceModel) truenas.CreateSSHKeyPairOpts {
	opts := truenas.CreateSSHKeyPairOpts{
		Name:       data.Name.ValueString(),
		PublicKey:  data.PublicKey.ValueString(),
		PrivateKey: data.PrivateKey.ValueString(),
	}

	return opts
}

func (r *SSHKeyPairResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SSHKeyPairResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := buildSSHKeyPairOpts(&data)

	keyPair, err := r.services.KeychainCredential.CreateSSHKeyPair(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create SSH Keypair",
			fmt.Sprintf("Unable to create SSH keypair: %s", err.Error()),
		)
		return
	}

	if keyPair == nil {
		resp.Diagnostics.AddError(
			"SSH Keypair Not Found",
			"SSH keypair was created but could not be found.",
		)
		return
	}

	mapSSHKeyPairToModel(keyPair, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SSHKeyPairResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SSHKeyPairResourceModel

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

	keyPair, err := r.services.KeychainCredential.GetSSHKeyPair(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read SSH Keypair",
			fmt.Sprintf("Unable to query SSH keypair: %s", err.Error()),
		)
		return
	}

	if keyPair == nil {
		// SSH keypair was deleted outside Terraform
		resp.State.RemoveResource(ctx)
		return
	}

	mapSSHKeyPairToModel(keyPair, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SSHKeyPairResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state SSHKeyPairResourceModel
	var plan SSHKeyPairResourceModel

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

	opts := buildSSHKeyPairOpts(&plan)

	keyPair, err := r.services.KeychainCredential.UpdateSSHKeyPair(ctx, id, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update SSH Keypair",
			fmt.Sprintf("Unable to update SSH keypair: %s", err.Error()),
		)
		return
	}

	if keyPair == nil {
		resp.Diagnostics.AddError(
			"SSH Keypair Not Found",
			"SSH keypair was updated but could not be found.",
		)
		return
	}

	// Set state from response
	mapSSHKeyPairToModel(keyPair, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SSHKeyPairResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SSHKeyPairResourceModel

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

	err = r.services.KeychainCredential.DeleteSSHKeyPair(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete SSH Keypair",
			fmt.Sprintf("Unable to delete SSH keypair: %s", err.Error()),
		)
		return
	}
}

// mapSSHKeyPairToModel maps a typed SSHKeyPair to the resource model.
func mapSSHKeyPairToModel(sshKeyPair *truenas.SSHKeyPair, data *SSHKeyPairResourceModel) {
	data.ID = types.StringValue(strconv.FormatInt(sshKeyPair.ID, 10))
	data.Name = types.StringValue(sshKeyPair.Name)
	data.PublicKey = types.StringValue(sshKeyPair.PublicKey)
	data.PrivateKey = types.StringValue(sshKeyPair.PrivateKey)
}
