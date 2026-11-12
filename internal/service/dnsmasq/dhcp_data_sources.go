package dnsmasq

import (
	"context"
	"fmt"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Data sources for DHCP ranges, DHCP options and the general settings. Their
// schemas are derived from the resource schemas so the two cannot drift.

type dhcpDataSourceBase struct {
	client opnsense.Client
}

func (d *dhcpDataSourceBase) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

var idPath = path.Root("id")

// Range

var _ datasource.DataSourceWithConfigure = &rangeDataSource{}

func newRangeDataSource() datasource.DataSource { return &rangeDataSource{} }

type rangeDataSource struct{ dhcpDataSourceBase }

func (d *rangeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dnsmasq_range"
}

func (d *rangeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = tools.DataSourceSchemaFromResource(rangeResourceSchema(), "id")
}

func (d *rangeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var id types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, idPath, &id)...)
	if resp.Diagnostics.HasError() {
		return
	}
	obj, err := d.client.Dnsmasq().GetRange(ctx, id.ValueString())
	if err == nil && obj == nil {
		err = fmt.Errorf("not found")
	}
	var model *rangeResourceModel
	if err == nil {
		model, err = convertRangeStructToSchema(obj)
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read dnsmasq DHCP range %s, got error: %s", id.ValueString(), err))
		return
	}
	model.Id = id
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

// Option

var _ datasource.DataSourceWithConfigure = &optionDataSource{}

func newOptionDataSource() datasource.DataSource { return &optionDataSource{} }

type optionDataSource struct{ dhcpDataSourceBase }

func (d *optionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dnsmasq_option"
}

func (d *optionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = tools.DataSourceSchemaFromResource(optionResourceSchema(), "id")
}

func (d *optionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var id types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, idPath, &id)...)
	if resp.Diagnostics.HasError() {
		return
	}
	obj, err := d.client.Dnsmasq().GetOption(ctx, id.ValueString())
	if err == nil && obj == nil {
		err = fmt.Errorf("not found")
	}
	var model *optionResourceModel
	if err == nil {
		model, err = convertOptionStructToSchema(obj)
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read dnsmasq DHCP option %s, got error: %s", id.ValueString(), err))
		return
	}
	model.Id = id
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

// Settings

var _ datasource.DataSourceWithConfigure = &settingsDataSource{}

func newSettingsDataSource() datasource.DataSource { return &settingsDataSource{} }

type settingsDataSource struct{ dhcpDataSourceBase }

func (d *settingsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dnsmasq_settings"
}

func (d *settingsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = tools.DataSourceSchemaFromResource(settingsResourceSchema())
}

func (d *settingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.Dnsmasq().GeneralSettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read dnsmasq settings, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, convertSettingsStructToSchema(&result.Dnsmasq))...)
}
