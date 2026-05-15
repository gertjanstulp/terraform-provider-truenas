package datasources

import (
	"context"
	"errors"
	"testing"

	"github.com/deevus/terraform-provider-truenas/internal/services"
	truenas "github.com/deevus/truenas-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestNewSSHKeyPairDataSource(t *testing.T) {
	ds := NewSSHKeyPairDataSource()
	if ds == nil {
		t.Fatal("expected non-nil data source")
	}

	// Verify it implements the required interfaces
	_ = datasource.DataSource(ds)
	_ = datasource.DataSourceWithConfigure(ds.(*SSHKeyPairDataSource))
}

func TestSSHKeyPairDataSource_Metadata(t *testing.T) {
	ds := NewSSHKeyPairDataSource()

	req := datasource.MetadataRequest{
		ProviderTypeName: "truenas",
	}
	resp := &datasource.MetadataResponse{}

	ds.Metadata(context.Background(), req, resp)

	if resp.TypeName != "truenas_ssh_keypair" {
		t.Errorf("expected TypeName 'truenas_ssh_keypair', got %q", resp.TypeName)
	}
}

func TestSSHKeyPairDataSource_Schema(t *testing.T) {
	ds := NewSSHKeyPairDataSource()

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	ds.Schema(context.Background(), req, resp)

	// Verify schema has description
	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema description")
	}

	// Verify name attribute exists and is required
	nameAttr, ok := resp.Schema.Attributes["name"]
	if !ok {
		t.Fatal("expected 'name' attribute in schema")
	}
	if !nameAttr.IsRequired() {
		t.Error("expected 'name' attribute to be required")
	}

	// Verify id attribute exists and is computed
	idAttr, ok := resp.Schema.Attributes["id"]
	if !ok {
		t.Fatal("expected 'id' attribute in schema")
	}
	if !idAttr.IsComputed() {
		t.Error("expected 'id' attribute to be computed")
	}

	// Verify public_key attribute exists and is computed
	publicKeyAttr, ok := resp.Schema.Attributes["public_key"]
	if !ok {
		t.Fatal("expected 'public_key' attribute in schema")
	}
	if !publicKeyAttr.IsComputed() {
		t.Error("expected 'public_key' attribute to be computed")
	}

	// Verify private_key attribute exists and is computed
	privateKeyAttr, ok := resp.Schema.Attributes["private_key"]
	if !ok {
		t.Fatal("expected 'private_key' attribute in schema")
	}
	if !privateKeyAttr.IsComputed() {
		t.Error("expected 'private_key' attribute to be computed")
	}
}

func TestSSHKeyPairDataSource_Configure_Success(t *testing.T) {
	ds := NewSSHKeyPairDataSource().(*SSHKeyPairDataSource)

	svc := &services.TrueNASServices{}

	req := datasource.ConfigureRequest{
		ProviderData: svc,
	}
	resp := &datasource.ConfigureResponse{}

	ds.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}
}

func TestSSHKeyPairDataSource_Configure_NilProviderData(t *testing.T) {
	ds := NewSSHKeyPairDataSource().(*SSHKeyPairDataSource)

	req := datasource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &datasource.ConfigureResponse{}

	ds.Configure(context.Background(), req, resp)

	// Should not error - nil ProviderData is valid during schema validation
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}
}

func TestSSHKeyPairDataSource_Configure_WrongType(t *testing.T) {
	ds := NewSSHKeyPairDataSource().(*SSHKeyPairDataSource)

	req := datasource.ConfigureRequest{
		ProviderData: "not a services",
	}
	resp := &datasource.ConfigureResponse{}

	ds.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error for wrong ProviderData type")
	}
}

// createTestSSHKeyPairReadRequest creates a datasource.ReadRequest with the given name
func createTestSSHKeyPairReadRequest(t *testing.T, name string) datasource.ReadRequest {
	t.Helper()

	// Get the schema
	ds := NewSSHKeyPairDataSource()
	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(context.Background(), schemaReq, schemaResp)

	// Build config value
	configValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":          tftypes.String,
			"name":        tftypes.String,
			"public_key":  tftypes.String,
			"private_key": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.String, name),
		"public_key":  tftypes.NewValue(tftypes.String, nil),
		"private_key": tftypes.NewValue(tftypes.String, nil),
	})

	config := tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    configValue,
	}

	return datasource.ReadRequest{
		Config: config,
	}
}

func TestSSHKeyPairDataSource_Read_Success(t *testing.T) {
	ds := &SSHKeyPairDataSource{
		services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				ListSSHKeyPairsFunc: func(ctx context.Context) ([]truenas.SSHKeyPair, error) {
					return []truenas.SSHKeyPair{
						{
							ID:         1,
							Name:       "keypair",
							PublicKey:  "publickey",
							PrivateKey: "privatekey",
						},
					}, nil
				},
			},
		},
	}

	req := createTestSSHKeyPairReadRequest(t, "keypair")

	// Get the schema for the state
	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(context.Background(), schemaReq, schemaResp)

	resp := &datasource.ReadResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
		},
	}

	ds.Read(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}

	// Verify the state was set correctly
	var model SSHKeyPairDataSourceModel
	diags := resp.State.Get(context.Background(), &model)
	if diags.HasError() {
		t.Fatalf("failed to get state: %v", diags)
	}

	if model.ID.ValueString() != "1" {
		t.Errorf("expected ID '1', got %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "keypair" {
		t.Errorf("expected Name 'keypair', got %q", model.Name.ValueString())
	}
	if model.PublicKey.ValueString() != "publickey" {
		t.Errorf("expected PublicKey 'publickey', got %q", model.PublicKey.ValueString())
	}
	if model.PrivateKey.ValueString() != "privatekey" {
		t.Errorf("expected PrivateKey 'privatekey', got %q", model.PrivateKey.ValueString())
	}
}

func TestSSHKeyPairDataSource_Read_SSHKeyPairNotFound(t *testing.T) {
	ds := &SSHKeyPairDataSource{
		services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				ListSSHKeyPairsFunc: func(ctx context.Context) ([]truenas.SSHKeyPair, error) {
					return []truenas.SSHKeyPair{}, nil
				},
			},
		},
	}

	req := createTestSSHKeyPairReadRequest(t, "nonexistent")

	// Get the schema for the state
	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(context.Background(), schemaReq, schemaResp)

	resp := &datasource.ReadResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
		},
	}

	ds.Read(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error for ssh_keypair not found")
	}
}

func TestSSHKeyPairDataSource_Read_APIError(t *testing.T) {
	ds := &SSHKeyPairDataSource{
		services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				ListSSHKeyPairsFunc: func(ctx context.Context) ([]truenas.SSHKeyPair, error) {
					return nil, errors.New("connection failed")
				},
			},
		},
	}

	req := createTestSSHKeyPairReadRequest(t, "tank")

	// Get the schema for the state
	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(context.Background(), schemaReq, schemaResp)

	resp := &datasource.ReadResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
		},
	}

	ds.Read(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error for API error")
	}
}

func TestSSHKeyPairDataSource_Read_ConfigError(t *testing.T) {
	ds := &SSHKeyPairDataSource{
		services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{},
		},
	}

	// Get the schema
	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(context.Background(), schemaReq, schemaResp)

	// Create an invalid config value with wrong type for name
	configValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":          tftypes.String,
			"name":        tftypes.Number, // Wrong type!
			"public_key":  tftypes.String,
			"private_key": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.Number, 123), // Wrong type!
		"public_key":  tftypes.NewValue(tftypes.String, nil),
		"private_key": tftypes.NewValue(tftypes.String, nil),
	})

	config := tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    configValue,
	}

	req := datasource.ReadRequest{
		Config: config,
	}

	resp := &datasource.ReadResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
		},
	}

	ds.Read(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error for config parse error")
	}
}

func TestSSHKeyPairDataSource_Read_MultipleSSHKeyPairsFindsMatch(t *testing.T) {
	ds := &SSHKeyPairDataSource{
		services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				ListSSHKeyPairsFunc: func(ctx context.Context) ([]truenas.SSHKeyPair, error) {
					return []truenas.SSHKeyPair{
						{ID: 1, Name: "keypair 1", PublicKey: "publickey 1", PrivateKey: "privatekey 1"},
						{ID: 2, Name: "keypair 2", PublicKey: "publickey 2", PrivateKey: "privatekey 2"},
						{ID: 3, Name: "keypair 3", PublicKey: "publickey 3", PrivateKey: "privatekey 3"},
					}, nil
				},
			},
		},
	}

	req := createTestSSHKeyPairReadRequest(t, "keypair 2")

	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(context.Background(), schemaReq, schemaResp)

	resp := &datasource.ReadResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
		},
	}

	ds.Read(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}

	var model SSHKeyPairDataSourceModel
	diags := resp.State.Get(context.Background(), &model)
	if diags.HasError() {
		t.Fatalf("failed to get state: %v", diags)
	}

	if model.ID.ValueString() != "2" {
		t.Errorf("expected ID '2', got %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "keypair 2" {
		t.Errorf("expected Name 'keypair 2', got %q", model.Name.ValueString())
	}
	if model.PublicKey.ValueString() != "publickey 2" {
		t.Errorf("expected PublicKey 'publickey 2', got %q", model.PublicKey.ValueString())
	}
	if model.PrivateKey.ValueString() != "privatekey 2" {
		t.Errorf("expected PrivateKey 'privatekey 2', got %q", model.PrivateKey.ValueString())
	}
}

// Test that SSHKeyPairDataSource implements the DataSource interface
func TestSSHKeyPairDataSource_ImplementsInterfaces(t *testing.T) {
	ds := NewSSHKeyPairDataSource()

	_ = datasource.DataSource(ds)
	_ = datasource.DataSourceWithConfigure(ds.(*SSHKeyPairDataSource))
}
