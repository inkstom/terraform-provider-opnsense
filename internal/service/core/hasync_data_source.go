package core

import (
	"context"
	"fmt"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &hasyncDataSource{}
var _ datasource.DataSourceWithConfigure = &hasyncDataSource{}

func newHasyncDataSource() datasource.DataSource {
	return &hasyncDataSource{}
}

type hasyncDataSource struct {
	client opnsense.Client
}

type hasyncDataSourceModel struct {
	DisablePreempt   types.Bool   `tfsdk:"disable_preempt"`
	DisconnectDialup types.Bool   `tfsdk:"disconnect_dialup"`
	PfsyncInterface  types.String `tfsdk:"pfsync_interface"`
	PfsyncPeerIp     types.String `tfsdk:"pfsync_peer_ip"`
	PfsyncVersion    types.String `tfsdk:"pfsync_version"`
	PfsyncDefer      types.Bool   `tfsdk:"pfsync_defer"`
	SynchronizeToIp  types.String `tfsdk:"synchronize_to_ip"`
	VerifyPeer       types.Bool   `tfsdk:"verify_peer"`
	Username         types.String `tfsdk:"username"`
	SyncItems        types.Set    `tfsdk:"sync_items"`
	Id               types.String `tfsdk:"id"`
}

func (d *hasyncDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_core_hasync"
}

func (d *hasyncDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = hasyncDataSourceSchema()
}

func (d *hasyncDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *hasyncDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.Core().HasyncGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read HA sync settings, got error: %s", err))
		return
	}
	h := result.Hasync
	resp.Diagnostics.Append(resp.State.Set(ctx, &hasyncDataSourceModel{
		DisablePreempt:   types.BoolValue(tools.StringToBool(h.DisablePreempt)),
		DisconnectDialup: types.BoolValue(tools.StringToBool(h.DisconnectPPPs)),
		PfsyncInterface:  types.StringValue(h.PfsyncInterface.String()),
		PfsyncPeerIp:     types.StringValue(h.PfsyncPeerIp),
		PfsyncVersion:    types.StringValue(h.PfsyncVersion.String()),
		PfsyncDefer:      types.BoolValue(tools.StringToBool(h.PfsyncDefer)),
		SynchronizeToIp:  types.StringValue(h.SynchronizeToIp),
		VerifyPeer:       types.BoolValue(tools.StringToBool(h.VerifyPeer)),
		Username:         types.StringValue(h.Username),
		SyncItems:        tools.StringSliceToSet(h.SyncItems),
		Id:               types.StringValue(hasyncID),
	})...)
}
