package resources

import (
	"context"
	"errors"
	"testing"

	"github.com/deevus/terraform-provider-truenas/internal/services"
	truenas "github.com/deevus/truenas-go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestNewSSHKeyPairResource(t *testing.T) {
	r := NewSSHKeyPairResource()
	if r == nil {
		t.Fatal("NewSSHKeyPairResource returned nil")
	}

	_, ok := r.(*SSHKeyPairResource)
	if !ok {
		t.Fatalf("expected *SSHKeyPairResource, got %T", r)
	}

	// Verify interface implementations
	_ = resource.Resource(r)
	_ = resource.ResourceWithConfigure(r.(*SSHKeyPairResource))
	_ = resource.ResourceWithImportState(r.(*SSHKeyPairResource))
}

func TestSSHKeyPairResource_Metadata(t *testing.T) {
	r := NewSSHKeyPairResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "truenas",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "truenas_ssh_keypair" {
		t.Errorf("expected TypeName 'truenas_ssh_keypair', got %q", resp.TypeName)
	}
}

func TestSSHKeyPairResource_Configure_Success(t *testing.T) {
	r := NewSSHKeyPairResource().(*SSHKeyPairResource)

	svc := &services.TrueNASServices{}

	req := resource.ConfigureRequest{
		ProviderData: svc,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}
}

func TestSSHKeyPairResource_Configure_NilProviderData(t *testing.T) {
	r := NewSSHKeyPairResource().(*SSHKeyPairResource)

	req := resource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}
}

func TestSSHKeyPairResource_Configure_WrongType(t *testing.T) {
	r := NewSSHKeyPairResource().(*SSHKeyPairResource)

	req := resource.ConfigureRequest{
		ProviderData: "not a client",
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error for wrong ProviderData type")
	}
}

func TestSSHKeyPairResource_Schema(t *testing.T) {
	r := NewSSHKeyPairResource()

	ctx := context.Background()
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}

	r.Schema(ctx, schemaReq, schemaResp)

	if schemaResp.Schema.Description == "" {
		t.Error("expected non-empty schema description")
	}

	// Verify required attributes exist
	attrs := schemaResp.Schema.Attributes
	if attrs["id"] == nil {
		t.Error("expected 'id' attribute")
	}
	if attrs["name"] == nil {
		t.Error("expected 'name' attribute")
	}
	if attrs["public_key"] == nil {
		t.Error("expected 'public_key' attribute")
	}
	if attrs["private_key"] == nil {
		t.Error("expected 'private_key' attribute")
	}
}

// Test helpers

func getSSHKeyPairResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := NewSSHKeyPairResource()
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), schemaReq, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("failed to get schema: %v", schemaResp.Diagnostics)
	}
	return *schemaResp
}

// sshKeyPairModelParams holds parameters for creating test model values.
type sshKeyPairModelParams struct {
	ID         interface{}
	Name       interface{}
	PublicKey  interface{}
	PrivateKey interface{}
}

func createSSHKeyPairModelValue(p sshKeyPairModelParams) tftypes.Value {

	// Build the values map
	values := map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, p.ID),
		"name":        tftypes.NewValue(tftypes.String, p.Name),
		"public_key":  tftypes.NewValue(tftypes.String, p.PublicKey),
		"private_key": tftypes.NewValue(tftypes.String, p.PrivateKey),
	}

	// Create object type matching the schema
	objectType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":          tftypes.String,
			"name":        tftypes.String,
			"public_key":  tftypes.String,
			"private_key": tftypes.String,
		},
	}

	return tftypes.NewValue(objectType, values)
}

// testSSHKeyPair returns a standard test task for S3.
func testSSHKeyPair(id int64, name string) *truenas.SSHKeyPair {
	return &truenas.SSHKeyPair{
		ID:         id,
		Name:       name,
		PublicKey:  "publickey",
		PrivateKey: "privatekey",
	}
}

func TestSSHKeyPairResource_Create_Success(t *testing.T) {
	var capturedOpts truenas.CreateSSHKeyPairOpts

	r := &SSHKeyPairResource{
		BaseResource: BaseResource{services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				CreateSSHKeyPairFunc: func(ctx context.Context, opts truenas.CreateSSHKeyPairOpts) (*truenas.SSHKeyPair, error) {
					capturedOpts = opts
					return testSSHKeyPair(10, "keypair"), nil
				},
			},
		}},
	}

	schemaResp := getSSHKeyPairResourceSchema(t)
	planValue := createSSHKeyPairModelValue(sshKeyPairModelParams{
		Name:       "keypair",
		PublicKey:  "publickey",
		PrivateKey: "privatekey",
	})

	req := resource.CreateRequest{
		Plan: tfsdk.Plan{
			Schema: schemaResp.Schema,
			Raw:    planValue,
		},
	}

	resp := &resource.CreateResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
		},
	}

	r.Create(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}

	// Verify opts sent to service
	if capturedOpts.Name != "keypair" {
		t.Errorf("expected Name 'keypair', got %q", capturedOpts.Name)
	}
	if capturedOpts.PublicKey != "publickey" {
		t.Errorf("expected PublicKey 'publickey', got %q", capturedOpts.PublicKey)
	}
	if capturedOpts.PrivateKey != "privatekey" {
		t.Errorf("expected PrivateKey 'privatekey', got %q", capturedOpts.PrivateKey)
	}

	// Verify state was set
	var resultData SSHKeyPairResourceModel
	resp.State.Get(context.Background(), &resultData)
	if resultData.ID.ValueString() != "10" {
		t.Errorf("expected ID '10', got %q", resultData.ID.ValueString())
	}
}

func TestSSHKeyPairResource_Create_APIError(t *testing.T) {
	r := &SSHKeyPairResource{
		BaseResource: BaseResource{services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				CreateSSHKeyPairFunc: func(ctx context.Context, opts truenas.CreateSSHKeyPairOpts) (*truenas.SSHKeyPair, error) {
					return nil, errors.New("connection refused")
				},
			},
		}},
	}

	schemaResp := getSSHKeyPairResourceSchema(t)
	planValue := createSSHKeyPairModelValue(sshKeyPairModelParams{
		Name:       "keypair",
		PublicKey:  "publickey",
		PrivateKey: "privatekey",
	})

	req := resource.CreateRequest{
		Plan: tfsdk.Plan{
			Schema: schemaResp.Schema,
			Raw:    planValue,
		},
	}

	resp := &resource.CreateResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
		},
	}

	r.Create(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error for API error")
	}
}

func TestSSHKeyPairResource_Read_Success(t *testing.T) {
	r := &SSHKeyPairResource{
		BaseResource: BaseResource{services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				GetSSHKeyPairFunc: func(ctx context.Context, id int64) (*truenas.SSHKeyPair, error) {
					return testSSHKeyPair(10, "keypair"), nil
				},
			},
		}},
	}

	schemaResp := getSSHKeyPairResourceSchema(t)
	stateValue := createSSHKeyPairModelValue(sshKeyPairModelParams{
		ID:         "10",
		Name:       "keypair",
		PublicKey:  "publickey",
		PrivateKey: "privatekey",
	})

	req := resource.ReadRequest{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}

	resp := &resource.ReadResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
		},
	}

	r.Read(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}

	// Verify state was updated
	var resultData SSHKeyPairResourceModel
	resp.State.Get(context.Background(), &resultData)
	if resultData.Name.ValueString() != "keypair" {
		t.Errorf("expected name 'keypair', got %q", resultData.Name.ValueString())
	}
}

func TestSSHKeyPairResource_Read_NotFound(t *testing.T) {
	r := &SSHKeyPairResource{
		BaseResource: BaseResource{services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				GetSSHKeyPairFunc: func(ctx context.Context, id int64) (*truenas.SSHKeyPair, error) {
					return nil, nil
				},
			},
		}},
	}

	schemaResp := getSSHKeyPairResourceSchema(t)
	stateValue := createSSHKeyPairModelValue(sshKeyPairModelParams{
		ID:         "10",
		Name:       "keypair",
		PublicKey:  "publickey",
		PrivateKey: "privatekey",
	})

	req := resource.ReadRequest{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}

	resp := &resource.ReadResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
		},
	}

	r.Read(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}

	// State should be removed (resource not found)
	if !resp.State.Raw.IsNull() {
		t.Error("expected state to be removed when resource not found")
	}
}

func TestSSHKeyPairResource_Read_APIError(t *testing.T) {
	r := &SSHKeyPairResource{
		BaseResource: BaseResource{services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				GetSSHKeyPairFunc: func(ctx context.Context, id int64) (*truenas.SSHKeyPair, error) {
					return nil, errors.New("connection refused")
				},
			},
		}},
	}

	schemaResp := getSSHKeyPairResourceSchema(t)
	stateValue := createSSHKeyPairModelValue(sshKeyPairModelParams{
		ID:         "10",
		Name:       "keypair",
		PublicKey:  "publickey",
		PrivateKey: "privatekey",
	})

	req := resource.ReadRequest{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}

	resp := &resource.ReadResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
		},
	}

	r.Read(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error for API error")
	}
}

func TestSSHKeyPairResource_Update_Success(t *testing.T) {
	var capturedID int64
	var capturedOpts truenas.UpdateSSHKeyPairOpts

	r := &SSHKeyPairResource{
		BaseResource: BaseResource{services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				UpdateSSHKeyPairFunc: func(ctx context.Context, id int64, opts truenas.UpdateSSHKeyPairOpts) (*truenas.SSHKeyPair, error) {
					capturedID = id
					capturedOpts = opts
					return &truenas.SSHKeyPair{
						ID:         10,
						Name:       "Updated keypair",
						PublicKey:  "publickey",
						PrivateKey: "privatekey",
					}, nil
				},
			},
		}},
	}

	schemaResp := getSSHKeyPairResourceSchema(t)

	// Current state
	stateValue := createSSHKeyPairModelValue(sshKeyPairModelParams{
		ID:         "10",
		Name:       "keypair",
		PublicKey:  "publickey",
		PrivateKey: "privatekey",
	})

	// Updated plan
	planValue := createSSHKeyPairModelValue(sshKeyPairModelParams{
		ID:         "10",
		Name:       "Updated keypair",
		PublicKey:  "publickey",
		PrivateKey: "privatekey",
	})

	req := resource.UpdateRequest{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
		Plan: tfsdk.Plan{
			Schema: schemaResp.Schema,
			Raw:    planValue,
		},
	}

	resp := &resource.UpdateResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
		},
	}

	r.Update(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}

	if capturedID != 10 {
		t.Errorf("expected ID 10, got %d", capturedID)
	}

	if capturedOpts.Name != "Updated keypair" {
		t.Errorf("expected name 'Updated keypair', got %q", capturedOpts.Name)
	}

	// Verify state was set
	var resultData SSHKeyPairResourceModel
	resp.State.Get(context.Background(), &resultData)
	if resultData.Name.ValueString() != "Updated keypair" {
		t.Errorf("expected name 'Updated keypair', got %q", resultData.Name.ValueString())
	}
}

func TestSSHKeyPairResource_Update_APIError(t *testing.T) {
	r := &SSHKeyPairResource{
		BaseResource: BaseResource{services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				UpdateSSHKeyPairFunc: func(ctx context.Context, id int64, opts truenas.UpdateSSHKeyPairOpts) (*truenas.SSHKeyPair, error) {
					return nil, errors.New("connection refused")
				},
			},
		}},
	}

	schemaResp := getSSHKeyPairResourceSchema(t)
	stateValue := createSSHKeyPairModelValue(sshKeyPairModelParams{
		ID:         "10",
		Name:       "keypair",
		PublicKey:  "publickey",
		PrivateKey: "privatekey",
	})

	planValue := createSSHKeyPairModelValue(sshKeyPairModelParams{
		ID:         "10",
		Name:       "Updated keypair",
		PublicKey:  "publickey",
		PrivateKey: "privatekey",
	})

	req := resource.UpdateRequest{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
		Plan: tfsdk.Plan{
			Schema: schemaResp.Schema,
			Raw:    planValue,
		},
	}

	resp := &resource.UpdateResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
		},
	}

	r.Update(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error for API error")
	}
}

func TestSSHKeyPairResource_Delete_Success(t *testing.T) {
	var capturedID int64

	r := &SSHKeyPairResource{
		BaseResource: BaseResource{services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				DeleteSSHKeyPairFunc: func(ctx context.Context, id int64) error {
					capturedID = id
					return nil
				},
			},
		}},
	}

	schemaResp := getSSHKeyPairResourceSchema(t)
	stateValue := createSSHKeyPairModelValue(sshKeyPairModelParams{
		ID:         "10",
		Name:       "keypair",
		PublicKey:  "publickey",
		PrivateKey: "privatekey",
	})

	req := resource.DeleteRequest{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}

	resp := &resource.DeleteResponse{}

	r.Delete(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}

	if capturedID != 10 {
		t.Errorf("expected ID 10, got %d", capturedID)
	}
}

func TestSSHKeyPairResource_Delete_APIError(t *testing.T) {
	r := &SSHKeyPairResource{
		BaseResource: BaseResource{services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				DeleteSSHKeyPairFunc: func(ctx context.Context, id int64) error {
					return errors.New("task in use by active job")
				},
			},
		}},
	}

	schemaResp := getSSHKeyPairResourceSchema(t)
	stateValue := createSSHKeyPairModelValue(sshKeyPairModelParams{
		ID:         "10",
		Name:       "keypair",
		PublicKey:  "publickey",
		PrivateKey: "privatekey",
	})

	req := resource.DeleteRequest{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    stateValue,
		},
	}

	resp := &resource.DeleteResponse{}

	r.Delete(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error for API error")
	}
}
