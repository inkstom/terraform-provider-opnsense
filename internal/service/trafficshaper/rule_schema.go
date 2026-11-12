package trafficshaper

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/trafficshaper"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ruleResourceModel struct {
	Enabled         types.Bool   `tfsdk:"enabled"`
	Sequence        types.Int64  `tfsdk:"sequence"`
	Interface       types.String `tfsdk:"interface"`
	Interface2      types.String `tfsdk:"interface2"`
	Protocol        types.String `tfsdk:"protocol"`
	MaxPacketLength types.Int64  `tfsdk:"max_packet_length"`
	Source          types.Set    `tfsdk:"source"`
	SourceNot       types.Bool   `tfsdk:"source_not"`
	SourcePort      types.String `tfsdk:"source_port"`
	Destination     types.Set    `tfsdk:"destination"`
	DestinationNot  types.Bool   `tfsdk:"destination_not"`
	DestinationPort types.String `tfsdk:"destination_port"`
	Dscp            types.Set    `tfsdk:"dscp"`
	Direction       types.String `tfsdk:"direction"`
	Target          types.String `tfsdk:"target"`
	Description     types.String `tfsdk:"description"`

	Id types.String `tfsdk:"id"`
}

func ruleResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Traffic shaper rules (Firewall > Shaper > Rules): match traffic and send it to a pipe or queue.",

		Attributes: map[string]schema.Attribute{
			"enabled": optBool("Enable the rule.", true),
			"sequence": schema.Int64Attribute{
				MarkdownDescription: "Order of evaluation; the first matching rule wins. Defaults to `1`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
			},
			"interface": schema.StringAttribute{
				MarkdownDescription: "Interface the traffic passes, e.g. `wan`.",
				Required:            true,
			},
			"interface2": optString("Second interface, to match only traffic between `interface` and this one.", ""),
			"protocol": optString("Protocol: `ip`, `ip4`, `ip6`, `udp`, `tcp`, `tcp_ack`, `tcp_ack_not`, `icmp`, `ipv6-icmp`, `igmp`, `esp`, `ah` or `gre`.", "ip",
				stringvalidator.OneOf("ip", "ip4", "ip6", "udp", "tcp", "tcp_ack", "tcp_ack_not", "icmp", "ipv6-icmp", "igmp", "esp", "ah", "gre")),
			"max_packet_length": optInt("Only match packets up to this many bytes (e.g. small ACKs); unset matches any length."),
			"source":            optSet("Source networks or addresses, or `any`. Defaults to `[\"any\"]`.", "any"),
			"source_not":        optBool("Invert the source match.", false),
			"source_port":       optString("Source port, port range or service name, or `any`.", "any"),
			"destination":       optSet("Destination networks or addresses, or `any`. Defaults to `[\"any\"]`.", "any"),
			"destination_not":   optBool("Invert the destination match.", false),
			"destination_port":  optString("Destination port, port range or service name, or `any`.", "any"),
			"dscp": schema.SetAttribute{
				MarkdownDescription: "DSCP classes to match, e.g. `ef`, `af11`, `cs1`; empty matches any. Defaults to `[]`.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             emptySetDefault(),
				Validators:          []validator.Set{setvalidator.SizeAtMost(64)},
			},
			"direction": optString("Direction: `\"\"` (both), `in` or `out`.", "", stringvalidator.OneOf("", "in", "out")),
			"target": schema.StringAttribute{
				MarkdownDescription: "UUID of the `opnsense_trafficshaper_pipe` or `opnsense_trafficshaper_queue` to send matching traffic to.",
				Required:            true,
			},
			"description": optString("Description.", ""),
			"id":          computedString("UUID of the rule."),
		},
	}
}

func ruleDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Traffic shaper rules (Firewall > Shaper > Rules).",

		Attributes: map[string]dschema.Attribute{
			"id":                dsID("UUID of the rule."),
			"enabled":           dsBool("Whether the rule is enabled."),
			"sequence":          dsInt("Order of evaluation."),
			"interface":         dsString("Interface."),
			"interface2":        dsString("Second interface."),
			"protocol":          dsString("Protocol."),
			"max_packet_length": dsInt("Maximum packet length matched."),
			"source":            dsSet("Source networks."),
			"source_not":        dsBool("Whether the source match is inverted."),
			"source_port":       dsString("Source port."),
			"destination":       dsSet("Destination networks."),
			"destination_not":   dsBool("Whether the destination match is inverted."),
			"destination_port":  dsString("Destination port."),
			"dscp":              dsSet("DSCP classes matched."),
			"direction":         dsString("Direction."),
			"target":            dsString("UUID of the pipe or queue."),
			"description":       dsString("Description."),
		},
	}
}

func convertRuleSchemaToStruct(d *ruleResourceModel) (*trafficshaper.Rule, error) {
	return &trafficshaper.Rule{
		Enabled:         tools.BoolToString(d.Enabled.ValueBool()),
		Sequence:        tools.Int64ToString(d.Sequence.ValueInt64()),
		Interface:       api.SelectedMap(d.Interface.ValueString()),
		Interface2:      api.SelectedMap(d.Interface2.ValueString()),
		Protocol:        api.SelectedMap(d.Protocol.ValueString()),
		IpLength:        intOrEmpty(d.MaxPacketLength),
		Source:          api.SelectedMapList(tools.SetToStringSlice(d.Source)),
		SourceNot:       tools.BoolToString(d.SourceNot.ValueBool()),
		SourcePort:      d.SourcePort.ValueString(),
		Destination:     api.SelectedMapList(tools.SetToStringSlice(d.Destination)),
		DestinationNot:  tools.BoolToString(d.DestinationNot.ValueBool()),
		DestinationPort: d.DestinationPort.ValueString(),
		Dscp:            api.SelectedMapList(tools.SetToStringSlice(d.Dscp)),
		Direction:       api.SelectedMap(d.Direction.ValueString()),
		Target:          api.SelectedMap(d.Target.ValueString()),
		Description:     d.Description.ValueString(),
	}, nil
}

func convertRuleStructToSchema(d *trafficshaper.Rule) (*ruleResourceModel, error) {
	return &ruleResourceModel{
		Enabled:         types.BoolValue(tools.StringToBool(d.Enabled)),
		Sequence:        types.Int64Value(tools.StringToInt64(d.Sequence)),
		Interface:       types.StringValue(d.Interface.String()),
		Interface2:      types.StringValue(d.Interface2.String()),
		Protocol:        types.StringValue(d.Protocol.String()),
		MaxPacketLength: tools.StringToInt64Null(d.IpLength),
		Source:          tools.StringSliceToSet(d.Source),
		SourceNot:       types.BoolValue(tools.StringToBool(d.SourceNot)),
		SourcePort:      types.StringValue(d.SourcePort),
		Destination:     tools.StringSliceToSet(d.Destination),
		DestinationNot:  types.BoolValue(tools.StringToBool(d.DestinationNot)),
		DestinationPort: types.StringValue(d.DestinationPort),
		Dscp:            tools.StringSliceToSet(d.Dscp),
		Direction:       types.StringValue(d.Direction.String()),
		Target:          types.StringValue(d.Target.String()),
		Description:     types.StringValue(d.Description),
	}, nil
}

func mergeRuleWriteOnly(prior, read *ruleResourceModel) {}
