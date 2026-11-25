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
			"enabled": optBool("Enable this rule.", true),
			"sequence": schema.Int64Attribute{
				MarkdownDescription: "Order in which the rule will be evaluated (lowest first). The first matching rule wins. Accepted range is 1-1000000. Defaults to `1`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
			},
			"interface": schema.StringAttribute{
				MarkdownDescription: "Interface the traffic passes, e.g. `wan`.",
				Required:            true,
			},
			"interface2": optString("Secondary interface, matches packets traveling to/from interface (1) to/from interface (2). Can be combined with direction. Interface (1) is `interface`; empty disables the match.", ""),
			"protocol": optString("Select the applicable protocol. One of `ip`, `ip4`, `ip6`, `udp`, `tcp`, `tcp_ack` (ACK packets only), `tcp_ack_not` (non-ACK packets), `icmp`, `ipv6-icmp`, `igmp`, `esp`, `ah` or `gre`.", "ip",
				stringvalidator.OneOf("ip", "ip4", "ip6", "udp", "tcp", "tcp_ack", "tcp_ack_not", "icmp", "ipv6-icmp", "igmp", "esp", "ah", "gre")),
			"max_packet_length": optInt("Specifies the maximum size of packets to match in bytes. Accepted range is 2-65535; unset matches any length."),
			"source":            optSet("Source IP or network, examples `10.0.0.0/24`, `10.0.0.1`. Use `any` to match everything. Defaults to `[\"any\"]`.", "any"),
			"source_not":        optBool("Invert source (not).", false),
			"source_port":       optString("Source port number or well known name (`imap`, `imaps`, `http`, `https`, ...), for ranges use a dash. Use `any` to match every port.", "any"),
			"destination":       optSet("Destination IP or network, examples `10.0.0.0/24`, `10.0.0.1`. Use `any` to match everything. Defaults to `[\"any\"]`.", "any"),
			"destination_not":   optBool("Invert destination (not).", false),
			"destination_port":  optString("Destination port number or well known name (`imap`, `imaps`, `http`, `https`, ...), for ranges use a dash. Use `any` to match every port.", "any"),
			"dscp": schema.SetAttribute{
				MarkdownDescription: "Match against one or multiple DSCP values. Accepted values are `be`, `ef`, `af11`, `af12`, `af13`, `af21`, `af22`, `af23`, `af31`, `af32`, `af33`, `af41`, `af42`, `af43`, `cs1`, `cs2`, `cs3`, `cs4`, `cs5`, `cs6` or `cs7`; empty matches any. Defaults to `[]`.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             emptySetDefault(),
				Validators:          []validator.Set{setvalidator.SizeAtMost(64)},
			},
			"direction": optString("Matches incoming or outgoing packets or both (default). One of `\"\"` (both), `in` or `out`.", "", stringvalidator.OneOf("", "in", "out")),
			"target": schema.StringAttribute{
				MarkdownDescription: "Target pipe or queue. Set to the UUID (`id`) of an `opnsense_trafficshaper_pipe` or `opnsense_trafficshaper_queue`; the target must have a non-empty description.",
				Required:            true,
			},
			"description": optString("Description to identify this rule.", ""),
			"id":          computedString("UUID of the rule."),
		},
	}
}

func ruleDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Traffic shaper rules (Firewall > Shaper > Rules).",

		Attributes: map[string]dschema.Attribute{
			"id":                dsID("UUID of the rule."),
			"enabled":           dsBool("Enable this rule."),
			"sequence":          dsInt("Order in which the rule will be evaluated (lowest first)."),
			"interface":         dsString("Interface."),
			"interface2":        dsString("Secondary interface, matches packets traveling to/from interface (1) to/from interface (2)."),
			"protocol":          dsString("Select the applicable protocol."),
			"max_packet_length": dsInt("Specifies the maximum size of packets to match in bytes."),
			"source":            dsSet("Source IP or network, examples `10.0.0.0/24`, `10.0.0.1`."),
			"source_not":        dsBool("Invert source (not)."),
			"source_port":       dsString("Source port number or well known name (`imap`, `imaps`, `http`, `https`, ...), for ranges use a dash."),
			"destination":       dsSet("Destination IP or network, examples `10.0.0.0/24`, `10.0.0.1`."),
			"destination_not":   dsBool("Invert destination (not)."),
			"destination_port":  dsString("Destination port number or well known name (`imap`, `imaps`, `http`, `https`, ...), for ranges use a dash."),
			"dscp":              dsSet("Match against one or multiple DSCP values."),
			"direction":         dsString("Matches incoming or outgoing packets or both (empty)."),
			"target":            dsString("Target pipe or queue (UUID of the pipe or queue)."),
			"description":       dsString("Description to identify this rule."),
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
