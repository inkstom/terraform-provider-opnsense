package routing

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/routing"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// gatewayResourceModel describes the resource data model.
type gatewayResourceModel struct {
	Enabled                   types.Bool   `tfsdk:"enabled"`
	Name                      types.String `tfsdk:"name"`
	Description               types.String `tfsdk:"description"`
	Interface                 types.String `tfsdk:"interface"`
	IpProtocol                types.String `tfsdk:"ip_protocol"`
	Gateway                   types.String `tfsdk:"gateway"`
	Upstream                  types.Bool   `tfsdk:"upstream"`
	FarGateway                types.Bool   `tfsdk:"far_gateway"`
	MonitorEnabled            types.Bool   `tfsdk:"monitor_enabled"`
	MonitorNoRoute            types.Bool   `tfsdk:"monitor_no_route"`
	MonitorKillStates         types.Bool   `tfsdk:"monitor_kill_states"`
	MonitorKillStatesPriority types.Bool   `tfsdk:"monitor_kill_states_priority"`
	Monitor                   types.String `tfsdk:"monitor"`
	ForceDown                 types.Bool   `tfsdk:"force_down"`
	NoSync                    types.Bool   `tfsdk:"no_sync"`
	Priority                  types.Int64  `tfsdk:"priority"`
	Weight                    types.Int64  `tfsdk:"weight"`
	LatencyLow                types.Int64  `tfsdk:"latency_low"`
	LatencyHigh               types.Int64  `tfsdk:"latency_high"`
	LossLow                   types.Int64  `tfsdk:"loss_low"`
	LossHigh                  types.Int64  `tfsdk:"loss_high"`
	Interval                  types.Int64  `tfsdk:"interval"`
	TimePeriod                types.Int64  `tfsdk:"time_period"`
	LossInterval              types.Int64  `tfsdk:"loss_interval"`
	DataLength                types.Int64  `tfsdk:"data_length"`

	Id types.String `tfsdk:"id"`
}

// optionalBool is an Optional+Computed bool with a static default.
func optionalBool(desc string, def bool) schema.BoolAttribute {
	return schema.BoolAttribute{
		MarkdownDescription: desc,
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(def),
	}
}

// optionalInt is an Optional+Computed int64 using -1 as the "unset" sentinel.
func optionalInt(desc string, min int64) schema.Int64Attribute {
	return schema.Int64Attribute{
		MarkdownDescription: desc + " Set to `-1` to use the OPNsense default. Defaults to `-1`.",
		Optional:            true,
		Computed:            true,
		Default:             int64default.StaticInt64(-1),
		Validators: []validator.Int64{
			int64validator.Any(int64validator.OneOf(-1), int64validator.AtLeast(min)),
		},
	}
}

func gatewayResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Gateways define the possible routes that can be used to access other networks, such as the internet (System > Gateways > Configuration). " +
			"A gateway cannot be deleted while it is referenced by an interface, gateway group or static route.",
		Version: 1,

		Attributes: map[string]schema.Attribute{
			"enabled": optionalBool("Enable this gateway. Inverse of the OPNsense *Disabled* checkbox. Defaults to `true`.", true),
			"name": schema.StringAttribute{
				MarkdownDescription: "Gateway name. Letters, digits, `_` and `-` only, max 32 characters. This is the value routes, gateway groups and firewall rules refer to.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(gatewayNameRegex, "must match ^[a-zA-Z0-9_-]{1,32}$"),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "You may enter a description here for your reference (not parsed).",
				Optional:            true,
			},
			"interface": schema.StringAttribute{
				MarkdownDescription: "Choose which interface this gateway applies to. Use the interface identifier, e.g. `wan`, `lan`, `opt1`.",
				Required:            true,
			},
			"ip_protocol": schema.StringAttribute{
				MarkdownDescription: "Choose the Internet Protocol this gateway uses. One of `inet` (IPv4) or `inet6` (IPv6). Defaults to `inet`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("inet"),
				Validators: []validator.String{
					stringvalidator.OneOf("inet", "inet6"),
				},
			},
			"gateway": schema.StringAttribute{
				MarkdownDescription: "IP address of the gateway (*IP Address* in OPNsense). Leave unset for a dynamic gateway (DHCP/PPP/VPN interfaces).",
				Optional:            true,
			},
			"upstream":         optionalBool("This will select the above gateway as a default gateway candidate. Shown as *Upstream Gateway* in OPNsense. Defaults to `false`.", false),
			"far_gateway":      optionalBool("This will allow the gateway to exist outside of the interface subnet. Defaults to `false`.", false),
			"monitor_enabled":  optionalBool("Monitor this gateway. When disabled, OPNsense considers this gateway as always being \"up\". Inverse of the OPNsense *Disable Gateway Monitoring* checkbox. Defaults to `false`, matching OPNsense.", false),
			"monitor_no_route": optionalBool("Do not create a dedicated host route for this monitor. Defaults to `false`.", false),
			"monitor_kill_states": optionalBool("If this gateway goes down, force clients to reconnect over a different online gateway by killing states associated with this gateway. "+
				"This option requires \"default gateway switching\" to be enabled, or this gateway assigned as part of a gateway group to trigger. Defaults to `false`.", false),
			"monitor_kill_states_priority": optionalBool("If another gateway comes up with a higher priority than this gateway, force clients to reconnect by killing states associated with this gateway. "+
				"This option requires \"default gateway switching\" to be enabled, or this gateway assigned as part of a gateway group to trigger. "+
				"The common use case for this option are metered connections over LTE that should only be used when no other gateway is online. Defaults to `false`.", false),
			"monitor": schema.StringAttribute{
				MarkdownDescription: "Enter an alternative address here to be used to monitor the link. This is used for the quality RRD graphs as well as the load balancer entries. " +
					"Use this if the gateway does not respond to ICMP echo requests (pings).",
				Optional: true,
			},
			"force_down": optionalBool("This will force this gateway to be considered \"down\". Defaults to `false`.", false),
			"no_sync":    optionalBool("Ignore this gateway from being updated using HA sync. Defaults to `false`.", false),
			"priority": schema.Int64Attribute{
				MarkdownDescription: "Choose a value between 1 and 255. Influences sort order when selecting a (default) gateway, lower means more important. " +
					"The OPNsense model also accepts `0`. Defaults to `255`.",
				Optional:   true,
				Computed:   true,
				Default:    int64default.StaticInt64(255),
				Validators: []validator.Int64{int64validator.Between(0, 255)},
			},
			"weight": schema.Int64Attribute{
				MarkdownDescription: "Weight for this gateway when used in a gateway group. Specified as an integer number between 1 and 10. Default equals 1. Defaults to `1`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
				Validators:          []validator.Int64{int64validator.Between(1, 10)},
			},
			"latency_low":   optionalInt("Low threshold for latency in milliseconds. Default is 200.", 1),
			"latency_high":  optionalInt("High threshold for latency in milliseconds. Default is 500.", 1),
			"loss_low":      optionalInt("Low threshold for packet loss in %. Default is 10. Must be between `1` and `99`.", 1),
			"loss_high":     optionalInt("High threshold for packet loss in %. Default is 20. Must be between `1` and `100`.", 1),
			"interval":      optionalInt("How often that an ICMP probe will be sent in seconds. Default is 1.", 1),
			"time_period":   optionalInt("The time period over which results are averaged. Default is 60.", 1),
			"loss_interval": optionalInt("Time interval before packets are treated as lost. Default is 4 (four times the probe interval).", 1),
			"data_length":   optionalInt("Specify the number of data bytes to be sent. Default is 1.", 0),
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the gateway.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func gatewayDataSourceSchema() dschema.Schema {
	computed := func(desc string) dschema.StringAttribute {
		return dschema.StringAttribute{MarkdownDescription: desc, Computed: true}
	}
	computedBool := func(desc string) dschema.BoolAttribute {
		return dschema.BoolAttribute{MarkdownDescription: desc, Computed: true}
	}
	computedInt := func(desc string) dschema.Int64Attribute {
		return dschema.Int64Attribute{MarkdownDescription: desc, Computed: true}
	}
	return dschema.Schema{
		MarkdownDescription: "Gateways define the possible routes that can be used to access other networks, such as the internet.",

		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				MarkdownDescription: "UUID of the gateway. Auto-generated gateways (e.g. `WAN_DHCP`) without a stored UUID can be looked up by name.",
				Required:            true,
			},
			"enabled":                      computedBool("Whether this gateway is enabled. Inverse of the OPNsense *Disabled* checkbox."),
			"name":                         computed("Gateway name."),
			"description":                  computed("You may enter a description here for your reference (not parsed)."),
			"interface":                    computed("Choose which interface this gateway applies to. The interface identifier, e.g. `wan`."),
			"ip_protocol":                  computed("Choose the Internet Protocol this gateway uses. One of `inet` (IPv4) or `inet6` (IPv6)."),
			"gateway":                      computed("IP address of the gateway (*IP Address* in OPNsense). Empty for dynamic gateways."),
			"upstream":                     computedBool("This will select the above gateway as a default gateway candidate. Shown as *Upstream Gateway* in OPNsense."),
			"far_gateway":                  computedBool("This will allow the gateway to exist outside of the interface subnet."),
			"monitor_enabled":              computedBool("Monitor this gateway. When disabled, OPNsense considers this gateway as always being \"up\". Inverse of the OPNsense *Disable Gateway Monitoring* checkbox."),
			"monitor_no_route":             computedBool("Do not create a dedicated host route for this monitor."),
			"monitor_kill_states":          computedBool("If this gateway goes down, force clients to reconnect over a different online gateway by killing states associated with this gateway."),
			"monitor_kill_states_priority": computedBool("If another gateway comes up with a higher priority than this gateway, force clients to reconnect by killing states associated with this gateway."),
			"monitor":                      computed("Enter an alternative address here to be used to monitor the link."),
			"force_down":                   computedBool("This will force this gateway to be considered \"down\"."),
			"no_sync":                      computedBool("Ignore this gateway from being updated using HA sync."),
			"priority":                     computedInt("Influences sort order when selecting a (default) gateway, lower means more important."),
			"weight":                       computedInt("Weight for this gateway when used in a gateway group."),
			"latency_low":                  computedInt("Low threshold for latency in milliseconds. `-1` when unset."),
			"latency_high":                 computedInt("High threshold for latency in milliseconds. `-1` when unset."),
			"loss_low":                     computedInt("Low threshold for packet loss in %. `-1` when unset."),
			"loss_high":                    computedInt("High threshold for packet loss in %. `-1` when unset."),
			"interval":                     computedInt("How often that an ICMP probe will be sent in seconds. `-1` when unset."),
			"time_period":                  computedInt("The time period over which results are averaged. `-1` when unset."),
			"loss_interval":                computedInt("Time interval before packets are treated as lost. `-1` when unset."),
			"data_length":                  computedInt("Specify the number of data bytes to be sent. `-1` when unset."),
		},
	}
}

func convertGatewaySchemaToStruct(d *gatewayResourceModel) (*routing.Gateway, error) {
	return &routing.Gateway{
		Disabled:                  tools.BoolToString(!d.Enabled.ValueBool()),
		Name:                      d.Name.ValueString(),
		Description:               d.Description.ValueString(),
		Interface:                 api.SelectedMap(d.Interface.ValueString()),
		IpProtocol:                api.SelectedMap(d.IpProtocol.ValueString()),
		Gateway:                   d.Gateway.ValueString(),
		DefaultGateway:            tools.BoolToString(d.Upstream.ValueBool()),
		FarGateway:                tools.BoolToString(d.FarGateway.ValueBool()),
		MonitorDisable:            tools.BoolToString(!d.MonitorEnabled.ValueBool()),
		MonitorNoRoute:            tools.BoolToString(d.MonitorNoRoute.ValueBool()),
		MonitorKillStates:         tools.BoolToString(d.MonitorKillStates.ValueBool()),
		MonitorKillStatesPriority: tools.BoolToString(d.MonitorKillStatesPriority.ValueBool()),
		Monitor:                   d.Monitor.ValueString(),
		ForceDown:                 tools.BoolToString(d.ForceDown.ValueBool()),
		NoSync:                    tools.BoolToString(d.NoSync.ValueBool()),
		Priority:                  tools.Int64ToString(d.Priority.ValueInt64()),
		Weight:                    tools.Int64ToString(d.Weight.ValueInt64()),
		LatencyLow:                tools.Int64ToStringNegative(d.LatencyLow.ValueInt64()),
		LatencyHigh:               tools.Int64ToStringNegative(d.LatencyHigh.ValueInt64()),
		LossLow:                   tools.Int64ToStringNegative(d.LossLow.ValueInt64()),
		LossHigh:                  tools.Int64ToStringNegative(d.LossHigh.ValueInt64()),
		Interval:                  tools.Int64ToStringNegative(d.Interval.ValueInt64()),
		TimePeriod:                tools.Int64ToStringNegative(d.TimePeriod.ValueInt64()),
		LossInterval:              tools.Int64ToStringNegative(d.LossInterval.ValueInt64()),
		DataLength:                tools.Int64ToStringNegative(d.DataLength.ValueInt64()),
	}, nil
}

func convertGatewayStructToSchema(d *routing.Gateway) (*gatewayResourceModel, error) {
	return &gatewayResourceModel{
		Enabled:                   types.BoolValue(!tools.StringToBool(d.Disabled)),
		Name:                      types.StringValue(d.Name),
		Description:               tools.StringOrNull(d.Description),
		Interface:                 types.StringValue(d.Interface.String()),
		IpProtocol:                types.StringValue(d.IpProtocol.String()),
		Gateway:                   tools.StringOrNull(d.Gateway),
		Upstream:                  types.BoolValue(tools.StringToBool(d.DefaultGateway)),
		FarGateway:                types.BoolValue(tools.StringToBool(d.FarGateway)),
		MonitorEnabled:            types.BoolValue(!tools.StringToBool(d.MonitorDisable)),
		MonitorNoRoute:            types.BoolValue(tools.StringToBool(d.MonitorNoRoute)),
		MonitorKillStates:         types.BoolValue(tools.StringToBool(d.MonitorKillStates)),
		MonitorKillStatesPriority: types.BoolValue(tools.StringToBool(d.MonitorKillStatesPriority)),
		Monitor:                   tools.StringOrNull(d.Monitor),
		ForceDown:                 types.BoolValue(tools.StringToBool(d.ForceDown)),
		NoSync:                    types.BoolValue(tools.StringToBool(d.NoSync)),
		Priority:                  types.Int64Value(tools.StringToInt64(d.Priority)),
		Weight:                    types.Int64Value(tools.StringToInt64(d.Weight)),
		LatencyLow:                types.Int64Value(tools.StringToInt64(d.LatencyLow)),
		LatencyHigh:               types.Int64Value(tools.StringToInt64(d.LatencyHigh)),
		LossLow:                   types.Int64Value(tools.StringToInt64(d.LossLow)),
		LossHigh:                  types.Int64Value(tools.StringToInt64(d.LossHigh)),
		Interval:                  types.Int64Value(tools.StringToInt64(d.Interval)),
		TimePeriod:                types.Int64Value(tools.StringToInt64(d.TimePeriod)),
		LossInterval:              types.Int64Value(tools.StringToInt64(d.LossInterval)),
		DataLength:                types.Int64Value(tools.StringToInt64(d.DataLength)),
	}, nil
}
