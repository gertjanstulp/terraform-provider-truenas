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

func TestNewSSHConnectionResource(t *testing.T) {
	r := NewSSHConnectionResource()
	if r == nil {
		t.Fatal("NewSSHConnectionResource returned nil")
	}

	_, ok := r.(*SSHConnectionResource)
	if !ok {
		t.Fatalf("expected *SSHConnectionResource, got %T", r)
	}

	// Verify interface implementations
	_ = resource.Resource(r)
	_ = resource.ResourceWithConfigure(r.(*SSHConnectionResource))
	_ = resource.ResourceWithImportState(r.(*SSHConnectionResource))
}

func TestSSHConnectionResource_Metadata(t *testing.T) {
	r := NewSSHConnectionResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "truenas",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "truenas_ssh_connection" {
		t.Errorf("expected TypeName 'truenas_ssh_connection', got %q", resp.TypeName)
	}
}

func TestSSHConnectionResource_Configure_Success(t *testing.T) {
	r := NewSSHConnectionResource().(*SSHConnectionResource)

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

func TestSSHConnectionResource_Configure_NilProviderData(t *testing.T) {
	r := NewSSHConnectionResource().(*SSHConnectionResource)

	req := resource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}
}

func TestSSHConnectionResource_Configure_WrongType(t *testing.T) {
	r := NewSSHConnectionResource().(*SSHConnectionResource)

	req := resource.ConfigureRequest{
		ProviderData: "not a client",
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error for wrong ProviderData type")
	}
}

func TestSSHConnectionResource_Schema(t *testing.T) {
	r := NewSSHConnectionResource()

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
	if attrs["host"] == nil {
		t.Error("expected 'host' attribute")
	}
	if attrs["port"] == nil {
		t.Error("expected 'port' attribute")
	}
	if attrs["username"] == nil {
		t.Error("expected 'username' attribute")
	}
	if attrs["private_key_id"] == nil {
		t.Error("expected 'private_key_id' attribute")
	}
	if attrs["remote_host_key"] == nil {
		t.Error("expected 'remote_host_key' attribute")
	}
	if attrs["connect_timeout"] == nil {
		t.Error("expected 'connect_timeout' attribute")
	}
}

// Test helpers

func getSSHConnectionResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := NewSSHConnectionResource()
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), schemaReq, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("failed to get schema: %v", schemaResp.Diagnostics)
	}
	return *schemaResp
}

// sshConnectionModelParams holds parameters for creating test model values.
type sshConnectionModelParams struct {
	ID             interface{}
	Name           interface{}
	Host           interface{}
	Port           interface{}
	Username       interface{}
	PrivateKeyID   interface{}
	RemoteHostKey  interface{}
	ConnectTimeout interface{}
}

func createSSHConnectionModelValue(p sshConnectionModelParams) tftypes.Value {

	// Build the values map
	values := map[string]tftypes.Value{
		"id":              tftypes.NewValue(tftypes.String, p.ID),
		"name":            tftypes.NewValue(tftypes.String, p.Name),
		"host":            tftypes.NewValue(tftypes.String, p.Host),
		"port":            tftypes.NewValue(tftypes.Number, p.Port),
		"username":        tftypes.NewValue(tftypes.String, p.Username),
		"private_key_id":  tftypes.NewValue(tftypes.Number, p.PrivateKeyID),
		"remote_host_key": tftypes.NewValue(tftypes.String, p.RemoteHostKey),
		"connect_timeout": tftypes.NewValue(tftypes.Number, p.ConnectTimeout),
	}

	// Create object type matching the schema
	objectType := tftypes.Object{
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
	}

	return tftypes.NewValue(objectType, values)
}

// testSSHConnection returns a standard test task for S3.
func testSSHConnection(id int64, name string) *truenas.SSHConnection {
	return &truenas.SSHConnection{
		ID:             id,
		Name:           name,
		Host:           "host",
		Port:           123,
		Username:       "username",
		PrivateKeyID:   1,
		RemoteHostKey:  "remotehostkey",
		ConnectTimeout: 30,
	}
}

func TestSSHConnectionResource_Create_Success(t *testing.T) {
	var capturedOpts truenas.CreateSSHConnectionOpts

	r := &SSHConnectionResource{
		BaseResource: BaseResource{services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				CreateSSHConnectionFunc: func(ctx context.Context, opts truenas.CreateSSHConnectionOpts) (*truenas.SSHConnection, error) {
					capturedOpts = opts
					return testSSHConnection(10, "connection"), nil
				},
			},
		}},
	}

	schemaResp := getSSHConnectionResourceSchema(t)
	planValue := createSSHConnectionModelValue(sshConnectionModelParams{
		Name:           "connection",
		Host:           "host",
		Port:           123,
		Username:       "username",
		PrivateKeyID:   1,
		RemoteHostKey:  "remotehostkey",
		ConnectTimeout: 30,
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
	if capturedOpts.Name != "connection" {
		t.Errorf("expected Name 'connection', got %q", capturedOpts.Name)
	}
	if capturedOpts.Host != "host" {
		t.Errorf("expected Host 'host', got %q", capturedOpts.Host)
	}
	if capturedOpts.Port != 123 {
		t.Errorf("expected Port '123', got %q", capturedOpts.Port)
	}
	if capturedOpts.Username != "username" {
		t.Errorf("expected Username 'username', got %q", capturedOpts.Username)
	}
	if capturedOpts.PrivateKeyID != 1 {
		t.Errorf("expected PrivateKeyID '1', got %q", capturedOpts.PrivateKeyID)
	}
	if capturedOpts.RemoteHostKey != "remotehostkey" {
		t.Errorf("expected RemoteHostKey 'remotehostkey', got %q", capturedOpts.RemoteHostKey)
	}
	if capturedOpts.ConnectTimeout != 30 {
		t.Errorf("expected ConnectTimeout '30', got %q", capturedOpts.ConnectTimeout)
	}

	// Verify state was set
	var resultData SSHConnectionResourceModel
	resp.State.Get(context.Background(), &resultData)
	if resultData.ID.ValueString() != "10" {
		t.Errorf("expected ID '10', got %q", resultData.ID.ValueString())
	}
}

func TestSSHConnectionResource_Create_APIError(t *testing.T) {
	r := &SSHConnectionResource{
		BaseResource: BaseResource{services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				CreateSSHConnectionFunc: func(ctx context.Context, opts truenas.CreateSSHConnectionOpts) (*truenas.SSHConnection, error) {
					return nil, errors.New("connection refused")
				},
			},
		}},
	}

	schemaResp := getSSHConnectionResourceSchema(t)
	planValue := createSSHConnectionModelValue(sshConnectionModelParams{
		Name:           "connection",
		Host:           "host",
		Port:           123,
		Username:       "username",
		PrivateKeyID:   1,
		RemoteHostKey:  "remotehostkey",
		ConnectTimeout: 30,
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

func TestSSHConnectionResource_Read_Success(t *testing.T) {
	r := &SSHConnectionResource{
		BaseResource: BaseResource{services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				GetSSHConnectionFunc: func(ctx context.Context, id int64) (*truenas.SSHConnection, error) {
					return testSSHConnection(10, "connection"), nil
				},
			},
		}},
	}

	schemaResp := getSSHConnectionResourceSchema(t)
	stateValue := createSSHConnectionModelValue(sshConnectionModelParams{
		ID:             "10",
		Name:           "connection",
		Host:           "host",
		Port:           123,
		Username:       "username",
		PrivateKeyID:   1,
		RemoteHostKey:  "remotehostkey",
		ConnectTimeout: 30,
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
	var resultData SSHConnectionResourceModel
	resp.State.Get(context.Background(), &resultData)
	if resultData.Name.ValueString() != "connection" {
		t.Errorf("expected name 'connection', got %q", resultData.Name.ValueString())
	}
}

func TestSSHConnectionResource_Read_NotFound(t *testing.T) {
	r := &SSHConnectionResource{
		BaseResource: BaseResource{services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				GetSSHConnectionFunc: func(ctx context.Context, id int64) (*truenas.SSHConnection, error) {
					return nil, nil
				},
			},
		}},
	}

	schemaResp := getSSHConnectionResourceSchema(t)
	stateValue := createSSHConnectionModelValue(sshConnectionModelParams{
		ID:             "10",
		Name:           "connection",
		Host:           "host",
		Port:           123,
		Username:       "username",
		PrivateKeyID:   1,
		RemoteHostKey:  "remotehostkey",
		ConnectTimeout: 30,
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

func TestSSHConnectionResource_Read_APIError(t *testing.T) {
	r := &SSHConnectionResource{
		BaseResource: BaseResource{services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				GetSSHConnectionFunc: func(ctx context.Context, id int64) (*truenas.SSHConnection, error) {
					return nil, errors.New("connection refused")
				},
			},
		}},
	}

	schemaResp := getSSHConnectionResourceSchema(t)
	stateValue := createSSHConnectionModelValue(sshConnectionModelParams{
		ID:             "10",
		Name:           "connection",
		Host:           "host",
		Port:           123,
		Username:       "username",
		PrivateKeyID:   1,
		RemoteHostKey:  "remotehostkey",
		ConnectTimeout: 30,
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

func TestSSHConnectionResource_Update_Success(t *testing.T) {
	var capturedID int64
	var capturedOpts truenas.UpdateSSHConnectionOpts

	r := &SSHConnectionResource{
		BaseResource: BaseResource{services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				UpdateSSHConnectionFunc: func(ctx context.Context, id int64, opts truenas.UpdateSSHConnectionOpts) (*truenas.SSHConnection, error) {
					capturedID = id
					capturedOpts = opts
					return &truenas.SSHConnection{
						ID:             10,
						Name:           "Updated connection",
						Host:           "host",
						Port:           123,
						Username:       "username",
						PrivateKeyID:   1,
						RemoteHostKey:  "remotehostkey",
						ConnectTimeout: 30,
					}, nil
				},
			},
		}},
	}

	schemaResp := getSSHConnectionResourceSchema(t)

	// Current state
	stateValue := createSSHConnectionModelValue(sshConnectionModelParams{
		ID:             "10",
		Name:           "connection",
		Host:           "host",
		Port:           123,
		Username:       "username",
		PrivateKeyID:   1,
		RemoteHostKey:  "remotehostkey",
		ConnectTimeout: 30,
	})

	// Updated plan
	planValue := createSSHConnectionModelValue(sshConnectionModelParams{
		ID:             "10",
		Name:           "Updated connection",
		Host:           "host",
		Port:           123,
		Username:       "username",
		PrivateKeyID:   1,
		RemoteHostKey:  "remotehostkey",
		ConnectTimeout: 30,
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

	if capturedOpts.Name != "Updated connection" {
		t.Errorf("expected name 'Updated connection', got %q", capturedOpts.Name)
	}

	// Verify state was set
	var resultData SSHConnectionResourceModel
	resp.State.Get(context.Background(), &resultData)
	if resultData.Name.ValueString() != "Updated connection" {
		t.Errorf("expected name 'Updated connection', got %q", resultData.Name.ValueString())
	}
}

func TestSSHConnectionResource_Update_APIError(t *testing.T) {
	r := &SSHConnectionResource{
		BaseResource: BaseResource{services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				UpdateSSHConnectionFunc: func(ctx context.Context, id int64, opts truenas.UpdateSSHConnectionOpts) (*truenas.SSHConnection, error) {
					return nil, errors.New("connection refused")
				},
			},
		}},
	}

	schemaResp := getSSHConnectionResourceSchema(t)
	stateValue := createSSHConnectionModelValue(sshConnectionModelParams{
		ID:             "10",
		Name:           "connection",
		Host:           "host",
		Port:           123,
		Username:       "username",
		PrivateKeyID:   1,
		RemoteHostKey:  "remotehostkey",
		ConnectTimeout: 30,
	})

	planValue := createSSHConnectionModelValue(sshConnectionModelParams{
		ID:             "10",
		Name:           "connection",
		Host:           "host",
		Port:           123,
		Username:       "username",
		PrivateKeyID:   1,
		RemoteHostKey:  "remotehostkey",
		ConnectTimeout: 30,
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

func TestSSHConnectionResource_Delete_Success(t *testing.T) {
	var capturedID int64

	r := &SSHConnectionResource{
		BaseResource: BaseResource{services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				DeleteSSHConnectionFunc: func(ctx context.Context, id int64) error {
					capturedID = id
					return nil
				},
			},
		}},
	}

	schemaResp := getSSHConnectionResourceSchema(t)
	stateValue := createSSHConnectionModelValue(sshConnectionModelParams{
		ID:             "10",
		Name:           "connection",
		Host:           "host",
		Port:           123,
		Username:       "username",
		PrivateKeyID:   1,
		RemoteHostKey:  "remotehostkey",
		ConnectTimeout: 30,
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

func TestSSHConnectionResource_Delete_APIError(t *testing.T) {
	r := &SSHConnectionResource{
		BaseResource: BaseResource{services: &services.TrueNASServices{
			SSH: &truenas.MockSSHService{
				DeleteSSHConnectionFunc: func(ctx context.Context, id int64) error {
					return errors.New("task in use by active job")
				},
			},
		}},
	}

	schemaResp := getSSHConnectionResourceSchema(t)
	stateValue := createSSHConnectionModelValue(sshConnectionModelParams{
		ID:             "10",
		Name:           "connection",
		Host:           "host",
		Port:           123,
		Username:       "username",
		PrivateKeyID:   1,
		RemoteHostKey:  "remotehostkey",
		ConnectTimeout: 30,
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
