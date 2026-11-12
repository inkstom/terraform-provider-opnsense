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
	MonitorDisable            types.Bool   `tfsdk:"monitor_disable"`
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
			"enabled": optionalBool("Enable this gateway. Defaults to `true`.", true),
			"name": schema.StringAttribute{
				MarkdownDescription: "Gateway name. Letters, digits, `_` and `-` only, max 32 characters. This is the value routes, gateway groups and firewall rules refer to.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(gatewayNameRegex, "must match ^[a-zA-Z0-9_-]{1,32}$"),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional description here for your reference (not parsed).",
				Optional:            true,
			},
			"interface": schema.StringAttribute{
				MarkdownDescription: "Interface identifier this gateway is attached to, e.g. `wan`, `lan`, `opt1`.",
				Required:            true,
			},
			"ip_protocol": schema.StringAttribute{
				MarkdownDescription: "Address family, one of `inet` (IPv4) or `inet6` (IPv6). Defaults to `inet`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("inet"),
				Validators: []validator.String{
					stringvalidator.OneOf("inet", "inet6"),
				},
			},
			"gateway": schema.StringAttribute{
				MarkdownDescription: "IP address of the gateway. Leave unset for a dynamic gateway (DHCP/PPP/VPN interfaces).",
				Optional:            true,
			},
			"upstream":                     optionalBool("Upstream gateway: consider this gateway as a default gateway candidate. Defaults to `false`.", false),
			"far_gateway":                  optionalBool("Allow the gateway to exist outside of the interface subnet. Defaults to `false`.", false),
			"monitor_disable":              optionalBool("Disable gateway monitoring (consider it always online). Defaults to `true`, matching OPNsense.", true),
			"monitor_no_route":             optionalBool("Do not create a dedicated host route for the monitor IP. Defaults to `false`.", false),
			"monitor_kill_states":          optionalBool("Kill states for this gateway when it goes down. Defaults to `false`.", false),
			"monitor_kill_states_priority": optionalBool("Kill states for the higher-priority gateway when it comes back up. Defaults to `false`.", false),
			"monitor": schema.StringAttribute{
				MarkdownDescription: "Alternative address to monitor. Make sure it is reachable via this gateway.",
				Optional:            true,
			},
			"force_down": optionalBool("Mark the gateway as down so it is never considered as default gateway. Defaults to `false`.", false),
			"no_sync":    optionalBool("Do not sync this gateway to the HA peer. Defaults to `false`.", false),
			"priority": schema.Int64Attribute{
				MarkdownDescription: "Priority (0-255); a lower value means more important. Defaults to `255`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(255),
				Validators:          []validator.Int64{int64validator.Between(0, 255)},
			},
			"weight": schema.Int64Attribute{
				MarkdownDescription: "Weight (1-10) used when this gateway is part of a load-balancing group. Defaults to `1`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
				Validators:          []validator.Int64{int64validator.Between(1, 10)},
			},
			"latency_low":   optionalInt("Low latency threshold in milliseconds.", 1),
			"latency_high":  optionalInt("High latency threshold in milliseconds.", 1),
			"loss_low":      optionalInt("Low packet loss threshold in percent.", 1),
			"loss_high":     optionalInt("High packet loss threshold in percent.", 1),
			"interval":      optionalInt("Probe interval in seconds.", 1),
			"time_period":   optionalInt("Time period over which results are averaged, in seconds.", 1),
			"loss_interval": optionalInt("Time after which a probe without reply is considered lost, in seconds.", 1),
			"data_length":   optionalInt("Size of the ICMP data payload in bytes.", 0),
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
			"enabled":                      computedBool("Whether this gateway is enabled."),
			"name":                         computed("Gateway name."),
			"description":                  computed("Description."),
			"interface":                    computed("Interface identifier."),
			"ip_protocol":                  computed("Address family (`inet` or `inet6`)."),
			"gateway":                      computed("Gateway IP address; empty for dynamic gateways."),
			"upstream":                     computedBool("Whether this is an upstream (default candidate) gateway."),
			"far_gateway":                  computedBool("Whether the gateway may be outside the interface subnet."),
			"monitor_disable":              computedBool("Whether monitoring is disabled."),
			"monitor_no_route":             computedBool("Whether the monitor host route is suppressed."),
			"monitor_kill_states":          computedBool("Whether states are killed when the gateway goes down."),
			"monitor_kill_states_priority": computedBool("Whether higher-priority states are killed on recovery."),
			"monitor":                      computed("Alternative monitor address."),
			"force_down":                   computedBool("Whether the gateway is forced down."),
			"no_sync":                      computedBool("Whether HA sync is disabled."),
			"priority":                     computedInt("Priority."),
			"weight":                       computedInt("Weight."),
			"latency_low":                  computedInt("Low latency threshold (`-1` when unset)."),
			"latency_high":                 computedInt("High latency threshold (`-1` when unset)."),
			"loss_low":                     computedInt("Low packet loss threshold (`-1` when unset)."),
			"loss_high":                    computedInt("High packet loss threshold (`-1` when unset)."),
			"interval":                     computedInt("Probe interval (`-1` when unset)."),
			"time_period":                  computedInt("Averaging period (`-1` when unset)."),
			"loss_interval":                computedInt("Loss interval (`-1` when unset)."),
			"data_length":                  computedInt("ICMP payload size (`-1` when unset)."),
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
		MonitorDisable:            tools.BoolToString(d.MonitorDisable.ValueBool()),
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
		MonitorDisable:            types.BoolValue(tools.StringToBool(d.MonitorDisable)),
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
