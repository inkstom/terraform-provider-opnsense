package ipsec

import (
	"context"
	"fmt"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// byIDDataSource reads one IPsec object by UUID and reuses the resource's
// struct-to-schema converter. M is the resource model; it must have an
// `id` attribute.
type byIDDataSource[M any] struct {
	client   opnsense.Client
	typeName string
	what     string
	schema   func() dschema.Schema
	read     func(ctx context.Context, c opnsense.Client, id string) (*M, error)
	setID    func(m *M, id types.String)
}

var _ datasource.DataSourceWithConfigure = &byIDDataSource[pskResourceModel]{}

func (d *byIDDataSource[M]) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + d.typeName
}

func (d *byIDDataSource[M]) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = d.schema()
}

func (d *byIDDataSource[M]) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	apiClient, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}
	d.client = opnsense.NewClient(apiClient)
}

func (d *byIDDataSource[M]) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var id types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("id"), &id)...)
	if resp.Diagnostics.HasError() {
		return
	}
	model, err := d.read(ctx, d.client, id.ValueString())
	if err == nil && model == nil {
		err = fmt.Errorf("not found")
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read %s %s, got error: %s", d.what, id.ValueString(), err))
		return
	}
	d.setID(model, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func newConnectionDataSource() datasource.DataSource {
	return &byIDDataSource[connectionResourceModel]{
		typeName: "_ipsec_connection", what: "IPsec connection", schema: connectionDataSourceSchema,
		read: func(ctx context.Context, c opnsense.Client, id string) (*connectionResourceModel, error) {
			obj, err := c.Ipsec().GetIPsecConnection(ctx, id)
			if err != nil || obj == nil {
				return nil, err
			}
			return convertConnectionStructToSchema(obj)
		},
		setID: func(m *connectionResourceModel, id types.String) { m.Id = id },
	}
}

func newAuthLocalDataSource() datasource.DataSource {
	return &byIDDataSource[authLocalResourceModel]{
		typeName: "_ipsec_auth_local", what: "IPsec local authentication", schema: authLocalDataSourceSchema,
		read: func(ctx context.Context, c opnsense.Client, id string) (*authLocalResourceModel, error) {
			obj, err := c.Ipsec().GetIPsecAuthLocal(ctx, id)
			if err != nil || obj == nil {
				return nil, err
			}
			return convertAuthLocalStructToSchema(obj)
		},
		setID: func(m *authLocalResourceModel, id types.String) { m.Id = id },
	}
}

func newAuthRemoteDataSource() datasource.DataSource {
	return &byIDDataSource[authRemoteResourceModel]{
		typeName: "_ipsec_auth_remote", what: "IPsec remote authentication", schema: authRemoteDataSourceSchema,
		read: func(ctx context.Context, c opnsense.Client, id string) (*authRemoteResourceModel, error) {
			obj, err := c.Ipsec().GetIPsecAuthRemote(ctx, id)
			if err != nil || obj == nil {
				return nil, err
			}
			return convertAuthRemoteStructToSchema(obj)
		},
		setID: func(m *authRemoteResourceModel, id types.String) { m.Id = id },
	}
}

func newChildDataSource() datasource.DataSource {
	return &byIDDataSource[childResourceModel]{
		typeName: "_ipsec_child", what: "IPsec child SA", schema: childDataSourceSchema,
		read: func(ctx context.Context, c opnsense.Client, id string) (*childResourceModel, error) {
			obj, err := c.Ipsec().GetIPsecChild(ctx, id)
			if err != nil || obj == nil {
				return nil, err
			}
			return convertChildStructToSchema(obj)
		},
		setID: func(m *childResourceModel, id types.String) { m.Id = id },
	}
}

func newPskDataSource() datasource.DataSource {
	return &byIDDataSource[pskResourceModel]{
		typeName: "_ipsec_psk", what: "IPsec pre-shared key", schema: pskDataSourceSchema,
		read: func(ctx context.Context, c opnsense.Client, id string) (*pskResourceModel, error) {
			obj, err := c.Ipsec().GetIPsecPSK(ctx, id)
			if err != nil || obj == nil {
				return nil, err
			}
			return convertPskStructToSchema(obj)
		},
		setID: func(m *pskResourceModel, id types.String) { m.Id = id },
	}
}

func newVtiDataSource() datasource.DataSource {
	return &byIDDataSource[vtiResourceModel]{
		typeName: "_ipsec_vti", what: "IPsec VTI", schema: vtiDataSourceSchema,
		read: func(ctx context.Context, c opnsense.Client, id string) (*vtiResourceModel, error) {
			obj, err := c.Ipsec().GetIPsecVTI(ctx, id)
			if err != nil || obj == nil {
				return nil, err
			}
			return convertVtiStructToSchema(obj)
		},
		setID: func(m *vtiResourceModel, id types.String) { m.Id = id },
	}
}
