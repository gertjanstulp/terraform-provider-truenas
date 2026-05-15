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

func TestNewSSHConnectionDataSource(t *testing.T) {
	ds := NewSSHConnectionDataSource()
	if ds == nil {
		t.Fatal("expected non-nil data source")
	}

	// Verify it implements the required interfaces
	_ = datasource.DataSource(ds)
	_ = datasource.DataSourceWithConfigure(ds.(*SSHConnectionDataSource))
}

func TestSSHConnectionDataSource_Metadata(t *testing.T) {
	ds := NewSSHConnectionDataSource()

	req := datasource.MetadataRequest{
		ProviderTypeName: "truenas",
	}
	resp := &datasource.MetadataResponse{}

	ds.Metadata(context.Background(), req, resp)

	if resp.TypeName != "truenas_ssh_connection" {
		t.Errorf("expected TypeName 'truenas_ssh_connection', got %q", resp.TypeName)
	}
}

func TestSSHConnectionDataSource_Schema(t *testing.T) {
	ds := NewSSHConnectionDataSource()

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

	// Verify host attribute exists and is computed
	hostAttr, ok := resp.Schema.Attributes["host"]
	if !ok {
		t.Fatal("expected 'host' attribute in schema")
	}
	if !hostAttr.IsComputed() {
		t.Error("expected 'host' attribute to be computed")
	}

	// Verify port attribute exists and is computed
	portAttr, ok := resp.Schema.Attributes["port"]
	if !ok {
		t.Fatal("expected 'port' attribute in schema")
	}
	if !portAttr.IsComputed() {
		t.Error("expected 'port' attribute to be computed")
	}

	// Verify username attribute exists and is computed
	usernameAttr, ok := resp.Schema.Attributes["username"]
	if !ok {
		t.Fatal("expected 'username' attribute in schema")
	}
	if !usernameAttr.IsComputed() {
		t.Error("expected 'username' attribute to be computed")
	}

	// Verify private_key_id attribute exists and is computed
	privateKeyIDAttr, ok := resp.Schema.Attributes["private_key_id"]
	if !ok {
		t.Fatal("expected 'private_key_id' attribute in schema")
	}
	if !privateKeyIDAttr.IsComputed() {
		t.Error("expected 'private_key_id' attribute to be computed")
	}

	// Verify remote_host_key attribute exists and is computed
	remoteHostKeyAttr, ok := resp.Schema.Attributes["remote_host_key"]
	if !ok {
		t.Fatal("expected 'remote_host_key' attribute in schema")
	}
	if !remoteHostKeyAttr.IsComputed() {
		t.Error("expected 'remote_host_key' attribute to be computed")
	}

	// Verify connect_timeout attribute exists and is computed
	connectTimeoutAttr, ok := resp.Schema.Attributes["connect_timeout"]
	if !ok {
		t.Fatal("expected 'connect_timeout' attribute in schema")
	}
	if !connectTimeoutAttr.IsComputed() {
		t.Error("expected 'connect_timeout' attribute to be computed")
	}
}

func TestSSHConnectionDataSource_Configure_Success(t *testing.T) {
	ds := NewSSHConnectionDataSource().(*SSHConnectionDataSource)

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

func TestSSHConnectionDataSource_Configure_NilProviderData(t *testing.T) {
	ds := NewSSHConnectionDataSource().(*SSHConnectionDataSource)

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

func TestSSHConnectionDataSource_Configure_WrongType(t *testing.T) {
	ds := NewSSHConnectionDataSource().(*SSHConnectionDataSource)

	req := datasource.ConfigureRequest{
		ProviderData: "not a services",
	}
	resp := &datasource.ConfigureResponse{}

	ds.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error for wrong ProviderData type")
	}
}

// createTestSSHConnectionReadRequest creates a datasource.ReadRequest with the given name
func createTestSSHConnectionReadRequest(t *testing.T, name string) datasource.ReadRequest {
	t.Helper()

	// Get the schema
	ds := NewSSHConnectionDataSource()
	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(context.Background(), schemaReq, schemaResp)

	// Build config value
	configValue := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":              tftypes.String,
			"name":            tftypes.String,
			"host":            tftypes.String,
			"port":            tftypes.Number,
			"username":        tftypes.String,
			"private_key_id":  tftypes.Number,
			"remote_host_key": tftypes.String,
			"connect_timeout": tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":              tftypes.NewValue(tftypes.String, nil),
		"name":            tftypes.NewValue(tftypes.String, name),
		"host":            tftypes.NewValue(tftypes.String, nil),
		"port":            tftypes.NewValue(tftypes.Number, nil),
		"username":        tftypes.NewValue(tftypes.String, nil),
		"private_key_id":  tftypes.NewValue(tftypes.Number, nil),
		"remote_host_key": tftypes.NewValue(tftypes.String, nil),
		"connect_timeout": tftypes.NewValue(tftypes.Number, nil),
	})

	config := tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    configValue,
	}

	return datasource.ReadRequest{
		Config: config,
	}
}

func TestSSHConnectionDataSource_Read_Success(t *testing.T) {
	ds := &SSHConnectionDataSource{
		services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				ListSSHConnectionsFunc: func(ctx context.Context) ([]truenas.SSHConnection, error) {
					return []truenas.SSHConnection{
						{
							ID:             1,
							Name:           "connection",
							Host:           "host",
							Port:           123,
							Username:       "username",
							PrivateKeyID:   1,
							RemoteHostKey:  "remote host key",
							ConnectTimeout: 30,
						},
					}, nil
				},
			},
		},
	}

	req := createTestSSHConnectionReadRequest(t, "connection")

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
	var model SSHConnectionDataSourceModel
	diags := resp.State.Get(context.Background(), &model)
	if diags.HasError() {
		t.Fatalf("failed to get state: %v", diags)
	}

	if model.ID.ValueString() != "1" {
		t.Errorf("expected ID '1', got %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "connection" {
		t.Errorf("expected Name 'connection', got %q", model.Name.ValueString())
	}
	if model.Host.ValueString() != "host" {
		t.Errorf("expected Host 'host', got %q", model.Host.ValueString())
	}
	if model.Port.ValueInt32() != 123 {
		t.Errorf("expected Port '123', got %d", model.Port.ValueInt32())
	}
	if model.Username.ValueString() != "username" {
		t.Errorf("expected Username 'username', got %q", model.Username.ValueString())
	}
	if model.PrivateKeyID.ValueInt64() != 1 {
		t.Errorf("expected PrivateKeyID '1', got %d", model.PrivateKeyID.ValueInt64())
	}
	if model.RemoteHostKey.ValueString() != "remote host key" {
		t.Errorf("expected RemoteHostKey 'remote host key', got %q", model.RemoteHostKey.ValueString())
	}
	if model.ConnectTimeout.ValueInt32() != 30 {
		t.Errorf("expected ConnectTimeout '30', got %d", model.ConnectTimeout.ValueInt32())
	}
}

func TestSSHConnectionDataSource_Read_SSHConnectionNotFound(t *testing.T) {
	ds := &SSHConnectionDataSource{
		services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				ListSSHConnectionsFunc: func(ctx context.Context) ([]truenas.SSHConnection, error) {
					return []truenas.SSHConnection{}, nil
				},
			},
		},
	}

	req := createTestSSHConnectionReadRequest(t, "nonexistent")

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
		t.Fatal("expected error for ssh_connection not found")
	}
}

func TestSSHConnectionDataSource_Read_APIError(t *testing.T) {
	ds := &SSHConnectionDataSource{
		services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				ListSSHConnectionsFunc: func(ctx context.Context) ([]truenas.SSHConnection, error) {
					return nil, errors.New("connection failed")
				},
			},
		},
	}

	req := createTestSSHConnectionReadRequest(t, "tank")

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

func TestSSHConnectionDataSource_Read_ConfigError(t *testing.T) {
	ds := &SSHConnectionDataSource{
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
			"id":              tftypes.String,
			"name":            tftypes.Number, // Wrong type!
			"host":            tftypes.String,
			"port":            tftypes.Number,
			"username":        tftypes.String,
			"private_key_id":  tftypes.Number,
			"remote_host_key": tftypes.String,
			"connect_timeout": tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"id":              tftypes.NewValue(tftypes.String, nil),
		"name":            tftypes.NewValue(tftypes.Number, 123), // Wrong type!
		"host":            tftypes.NewValue(tftypes.String, nil),
		"port":            tftypes.NewValue(tftypes.Number, nil),
		"username":        tftypes.NewValue(tftypes.String, nil),
		"private_key_id":  tftypes.NewValue(tftypes.Number, nil),
		"remote_host_key": tftypes.NewValue(tftypes.String, nil),
		"connect_timeout": tftypes.NewValue(tftypes.Number, nil),
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

func TestSSHConnectionDataSource_Read_MultipleSSHConnectionsFindsMatch(t *testing.T) {
	ds := &SSHConnectionDataSource{
		services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				ListSSHConnectionsFunc: func(ctx context.Context) ([]truenas.SSHConnection, error) {
					return []truenas.SSHConnection{
						{ID: 1, Name: "connection 1", Host: "host 1", Port: 123, Username: "username 1", PrivateKeyID: 1, RemoteHostKey: "remotehostkey 1", ConnectTimeout: 10},
						{ID: 2, Name: "connection 2", Host: "host 2", Port: 456, Username: "username 2", PrivateKeyID: 2, RemoteHostKey: "remotehostkey 2", ConnectTimeout: 20},
						{ID: 3, Name: "connection 3", Host: "host 3", Port: 789, Username: "username 3", PrivateKeyID: 3, RemoteHostKey: "remotehostkey 3", ConnectTimeout: 30},
					}, nil
				},
			},
		},
	}

	req := createTestSSHConnectionReadRequest(t, "connection 2")

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

	var model SSHConnectionDataSourceModel
	diags := resp.State.Get(context.Background(), &model)
	if diags.HasError() {
		t.Fatalf("failed to get state: %v", diags)
	}

	if model.ID.ValueString() != "2" {
		t.Errorf("expected ID '2', got %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "connection 2" {
		t.Errorf("expected Name 'connection 2', got %q", model.Name.ValueString())
	}
	if model.Host.ValueString() != "host 2" {
		t.Errorf("expected Host 'host 2', got %q", model.Host.ValueString())
	}
	if model.Port.ValueInt32() != 456 {
		t.Errorf("expected Port '456', got %q", model.Port.ValueInt32())
	}
	if model.Username.ValueString() != "username 2" {
		t.Errorf("expected Username 'username 2', got %q", model.Username.ValueString())
	}
	if model.PrivateKeyID.ValueInt64() != 2 {
		t.Errorf("expected PrivateKeyID '2', got %q", model.PrivateKeyID.ValueInt64())
	}
	if model.RemoteHostKey.ValueString() != "remotehostkey 2" {
		t.Errorf("expected RemoteHostKey 'remotehostkey 2', got %q", model.RemoteHostKey.ValueString())
	}
	if model.ConnectTimeout.ValueInt32() != 20 {
		t.Errorf("expected ConnectTimeout '20', got %q", model.ConnectTimeout.ValueInt32())
	}
}

// Test that SSHConnectionDataSource implements the DataSource interface
func TestSSHConnectionDataSource_ImplementsInterfaces(t *testing.T) {
	ds := NewSSHConnectionDataSource()

	_ = datasource.DataSource(ds)
	_ = datasource.DataSourceWithConfigure(ds.(*SSHConnectionDataSource))
}
