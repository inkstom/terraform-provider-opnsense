package ipsec

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/ipsec"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type poolResourceModel struct {
	Enabled   types.Bool   `tfsdk:"enabled"`
	Name      types.String `tfsdk:"name"`
	Addresses types.String `tfsdk:"addresses"`
	Dns       types.Set    `tfsdk:"dns"`

	Id types.String `tfsdk:"id"`
}

func poolResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "IPsec virtual address pools (VPN > IPsec > Pools), handed to road-warrior clients. Reference a pool by `id` in `opnsense_ipsec_connection.ip_pools`.",

		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable the pool. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Pool name.",
				Required:            true,
			},
			"addresses": schema.StringAttribute{
				MarkdownDescription: "Address range to assign, as a network, e.g. `10.200.0.0/24`.",
				Required:            true,
			},
			"dns": schema.SetAttribute{
				MarkdownDescription: "DNS servers pushed to clients. Defaults to `[]`.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(tools.EmptySetValue(types.StringType)),
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the pool.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func poolDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "IPsec virtual address pools (VPN > IPsec > Pools).",

		Attributes: map[string]dschema.Attribute{
			"id":        dschema.StringAttribute{MarkdownDescription: "UUID of the pool.", Required: true},
			"enabled":   dschema.BoolAttribute{MarkdownDescription: "Whether the pool is enabled.", Computed: true},
			"name":      dschema.StringAttribute{MarkdownDescription: "Pool name.", Computed: true},
			"addresses": dschema.StringAttribute{MarkdownDescription: "Address range.", Computed: true},
			"dns":       dschema.SetAttribute{MarkdownDescription: "DNS servers.", ElementType: types.StringType, Computed: true},
		},
	}
}

func convertPoolSchemaToStruct(d *poolResourceModel) (*ipsec.IPsecPool, error) {
	return &ipsec.IPsecPool{
		Enabled:   tools.BoolToString(d.Enabled.ValueBool()),
		Name:      d.Name.ValueString(),
		Addresses: d.Addresses.ValueString(),
		Dns:       api.SelectedMapList(tools.SetToStringSlice(d.Dns)),
	}, nil
}

func convertPoolStructToSchema(d *ipsec.IPsecPool) (*poolResourceModel, error) {
	return &poolResourceModel{
		Enabled:   types.BoolValue(tools.StringToBool(d.Enabled)),
		Name:      types.StringValue(d.Name),
		Addresses: types.StringValue(d.Addresses),
		Dns:       tools.StringSliceToSet(d.Dns),
	}, nil
}

func mergePoolWriteOnly(prior, read *poolResourceModel) {}
