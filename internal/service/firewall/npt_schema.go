package firewall

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/firewall"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type nptResourceModel struct {
	Enabled        types.Bool   `tfsdk:"enabled"`
	Log            types.Bool   `tfsdk:"log"`
	Sequence       types.Int64  `tfsdk:"sequence"`
	Interface      types.String `tfsdk:"interface"`
	InternalPrefix types.String `tfsdk:"internal_prefix"`
	ExternalPrefix types.String `tfsdk:"external_prefix"`
	TrackInterface types.String `tfsdk:"track_interface"`
	Categories     types.Set    `tfsdk:"categories"`
	Description    types.String `tfsdk:"description"`

	Id types.String `tfsdk:"id"`
}

func nptResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "IPv6 network prefix translation rules (Firewall > NAT > NPTv6): map an internal prefix to an external one of the same length.",

		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable the rule. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"log": schema.BoolAttribute{
				MarkdownDescription: "Log packets that match the rule. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"sequence": schema.Int64Attribute{
				MarkdownDescription: "Order of evaluation. Defaults to `100`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(100),
			},
			"interface": schema.StringAttribute{
				MarkdownDescription: "Interface the translation applies on, e.g. `wan`.",
				Required:            true,
			},
			"internal_prefix": schema.StringAttribute{
				MarkdownDescription: "Internal (source) IPv6 prefix, e.g. `fd00:1::/64`.",
				Required:            true,
			},
			"external_prefix": schema.StringAttribute{
				MarkdownDescription: "External (translated) IPv6 prefix of the same length. Leave empty with `track_interface` to follow a dynamic prefix. Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"track_interface": schema.StringAttribute{
				MarkdownDescription: "Interface whose delegated prefix is used as the external prefix; the interface must have a prefix ID. Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"categories": schema.SetAttribute{
				MarkdownDescription: "UUIDs of `opnsense_firewall_category` to tag the rule with. Defaults to `[]`.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(tools.EmptySetValue(types.StringType)),
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description. Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the rule.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func nptDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "IPv6 network prefix translation rules (Firewall > NAT > NPTv6).",

		Attributes: map[string]dschema.Attribute{
			"id":              dschema.StringAttribute{MarkdownDescription: "UUID of the rule.", Required: true},
			"enabled":         dschema.BoolAttribute{MarkdownDescription: "Whether the rule is enabled.", Computed: true},
			"log":             dschema.BoolAttribute{MarkdownDescription: "Whether matches are logged.", Computed: true},
			"sequence":        dschema.Int64Attribute{MarkdownDescription: "Order of evaluation.", Computed: true},
			"interface":       dschema.StringAttribute{MarkdownDescription: "Interface.", Computed: true},
			"internal_prefix": dschema.StringAttribute{MarkdownDescription: "Internal prefix.", Computed: true},
			"external_prefix": dschema.StringAttribute{MarkdownDescription: "External prefix.", Computed: true},
			"track_interface": dschema.StringAttribute{MarkdownDescription: "Tracked interface.", Computed: true},
			"categories":      dschema.SetAttribute{MarkdownDescription: "Category UUIDs.", ElementType: types.StringType, Computed: true},
			"description":     dschema.StringAttribute{MarkdownDescription: "Description.", Computed: true},
		},
	}
}

func convertNptSchemaToStruct(d *nptResourceModel) (*firewall.NPT, error) {
	return &firewall.NPT{
		Enabled:        tools.BoolToString(d.Enabled.ValueBool()),
		Log:            tools.BoolToString(d.Log.ValueBool()),
		Sequence:       tools.Int64ToString(d.Sequence.ValueInt64()),
		Interface:      api.SelectedMap(d.Interface.ValueString()),
		SourceNet:      d.InternalPrefix.ValueString(),
		DestinationNet: d.ExternalPrefix.ValueString(),
		TrackInterface: api.SelectedMap(d.TrackInterface.ValueString()),
		Categories:     api.SelectedMapList(tools.SetToStringSlice(d.Categories)),
		Description:    d.Description.ValueString(),
	}, nil
}

func convertNptStructToSchema(d *firewall.NPT) (*nptResourceModel, error) {
	return &nptResourceModel{
		Enabled:        types.BoolValue(tools.StringToBool(d.Enabled)),
		Log:            types.BoolValue(tools.StringToBool(d.Log)),
		Sequence:       types.Int64Value(tools.StringToInt64(d.Sequence)),
		Interface:      types.StringValue(d.Interface.String()),
		InternalPrefix: types.StringValue(d.SourceNet),
		ExternalPrefix: types.StringValue(d.DestinationNet),
		TrackInterface: types.StringValue(d.TrackInterface.String()),
		Categories:     tools.StringSliceToSet(d.Categories),
		Description:    types.StringValue(d.Description),
	}, nil
}

func mergeNptWriteOnly(prior, read *nptResourceModel) {}
