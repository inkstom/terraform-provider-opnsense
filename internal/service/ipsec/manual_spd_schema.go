package ipsec

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/ipsec"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type manualSpdResourceModel struct {
	Enabled         types.Bool   `tfsdk:"enabled"`
	Protocol        types.String `tfsdk:"protocol"`
	ReqId           types.Int64  `tfsdk:"request_id"`
	ConnectionChild types.String `tfsdk:"connection_child"`
	Source          types.String `tfsdk:"source"`
	Destination     types.String `tfsdk:"destination"`
	Description     types.String `tfsdk:"description"`

	Id types.String `tfsdk:"id"`
}

func manualSpdResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manual IPsec security policy (SPD) entries (VPN > IPsec > Manual SPD), e.g. to route extra networks into a route-based tunnel's child SA.",

		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable the entry. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "`esp` or `ah`. Defaults to `esp`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("esp"),
				Validators:          []validator.String{stringvalidator.OneOf("esp", "ah")},
			},
			"request_id": schema.Int64Attribute{
				MarkdownDescription: "Reqid to register this manual item on. Please note that an established tunnel needs to reconnect in order to add the entry as we use an updown event. Use the `request_id` set on an `opnsense_ipsec_child`. Accepts `1` to `65535`.",
				Required:            true,
			},
			"connection_child": schema.StringAttribute{
				MarkdownDescription: "Connection child to register this manual item on. Please note that an established tunnel needs to reconnect in order to add the entry as we use an updown event. UUID of an `opnsense_ipsec_child`. Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"source": schema.StringAttribute{
				MarkdownDescription: "Source network, usually the networks you would like to accept using network address translation. For example `10.0.0.0/24`.",
				Required:            true,
			},
			"destination": schema.StringAttribute{
				MarkdownDescription: "Destination network, leave empty to use the networks propagated in the child SA. Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description. Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the entry.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func manualSpdDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Manual IPsec security policy (SPD) entries (VPN > IPsec > Manual SPD).",

		Attributes: map[string]dschema.Attribute{
			"id":               dschema.StringAttribute{MarkdownDescription: "UUID of the entry.", Required: true},
			"enabled":          dschema.BoolAttribute{MarkdownDescription: "Whether the entry is enabled.", Computed: true},
			"protocol":         dschema.StringAttribute{MarkdownDescription: "Protocol.", Computed: true},
			"request_id":       dschema.Int64Attribute{MarkdownDescription: "Reqid to register this manual item on. Please note that an established tunnel needs to reconnect in order to add the entry as we use an updown event.", Computed: true},
			"connection_child": dschema.StringAttribute{MarkdownDescription: "Connection child to register this manual item on. Please note that an established tunnel needs to reconnect in order to add the entry as we use an updown event. UUID of an `opnsense_ipsec_child`.", Computed: true},
			"source":           dschema.StringAttribute{MarkdownDescription: "Source network, usually the networks you would like to accept using network address translation.", Computed: true},
			"destination":      dschema.StringAttribute{MarkdownDescription: "Destination network, leave empty to use the networks propagated in the child SA.", Computed: true},
			"description":      dschema.StringAttribute{MarkdownDescription: "Description.", Computed: true},
		},
	}
}

func convertManualSpdSchemaToStruct(d *manualSpdResourceModel) (*ipsec.IPsecManualSPD, error) {
	return &ipsec.IPsecManualSPD{
		Enabled:         tools.BoolToString(d.Enabled.ValueBool()),
		Protocol:        api.SelectedMap(d.Protocol.ValueString()),
		ReqId:           tools.Int64ToString(d.ReqId.ValueInt64()),
		ConnectionChild: api.SelectedMap(d.ConnectionChild.ValueString()),
		Source:          d.Source.ValueString(),
		Destination:     d.Destination.ValueString(),
		Description:     d.Description.ValueString(),
	}, nil
}

func convertManualSpdStructToSchema(d *ipsec.IPsecManualSPD) (*manualSpdResourceModel, error) {
	return &manualSpdResourceModel{
		Enabled:         types.BoolValue(tools.StringToBool(d.Enabled)),
		Protocol:        types.StringValue(d.Protocol.String()),
		ReqId:           types.Int64Value(tools.StringToInt64(d.ReqId)),
		ConnectionChild: types.StringValue(d.ConnectionChild.String()),
		Source:          types.StringValue(d.Source),
		Destination:     types.StringValue(d.Destination),
		Description:     types.StringValue(d.Description),
	}, nil
}

func mergeManualSpdWriteOnly(prior, read *manualSpdResourceModel) {}
