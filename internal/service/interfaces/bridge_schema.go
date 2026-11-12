package interfaces

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/interfaces"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type bridgeResourceModel struct {
	MemberSet      types.Set    `tfsdk:"members"`
	Description    types.String `tfsdk:"description"`
	LinkLocal      types.Bool   `tfsdk:"link_local"`
	StpEnabled     types.Bool   `tfsdk:"stp_enabled"`
	StpProtocol    types.String `tfsdk:"stp_protocol"`
	StpInterfaces  types.Set    `tfsdk:"stp_interfaces"`
	SpanInterface  types.String `tfsdk:"span_interface"`
	EdgePorts      types.Set    `tfsdk:"edge_ports"`
	AutoEdgePorts  types.Set    `tfsdk:"auto_edge_ports"`
	PtpPorts       types.Set    `tfsdk:"ptp_ports"`
	AutoPtpPorts   types.Set    `tfsdk:"auto_ptp_ports"`
	StaticPorts    types.Set    `tfsdk:"static_ports"`
	PrivatePorts   types.Set    `tfsdk:"private_ports"`
	MaxAge         types.Int64  `tfsdk:"max_age"`
	ForwardDelay   types.Int64  `tfsdk:"forward_delay"`
	HoldCount      types.Int64  `tfsdk:"hold_count"`
	MaxAddresses   types.Int64  `tfsdk:"max_addresses"`
	AddressTimeout types.Int64  `tfsdk:"address_timeout"`
	Device         types.String `tfsdk:"device"`

	Id types.String `tfsdk:"id"`
}

func optSetDefault(desc string) schema.SetAttribute {
	return schema.SetAttribute{
		MarkdownDescription: desc + " Defaults to `[]`.",
		ElementType:         types.StringType,
		Optional:            true,
		Computed:            true,
		Default:             setdefault.StaticValue(tools.EmptySetValue(types.StringType)),
	}
}

func intOrEmpty(v types.Int64) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return tools.Int64ToString(v.ValueInt64())
}

func dsSet(desc string) dschema.SetAttribute {
	return dschema.SetAttribute{MarkdownDescription: desc, ElementType: types.StringType, Computed: true}
}

func dsInt(desc string) dschema.Int64Attribute {
	return dschema.Int64Attribute{MarkdownDescription: desc, Computed: true}
}

func bridgeResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Bridge interfaces (Interfaces > Devices > Bridge). Members and the per-port STP sets take assigned interface identifiers (e.g. `opt1`), " +
			"such as `opnsense_interfaces_assignment.id`. Assign the bridge device itself to give it an address.",

		Attributes: map[string]schema.Attribute{
			"members": schema.SetAttribute{
				MarkdownDescription: "Member interfaces, by identifier.",
				ElementType:         types.StringType,
				Required:            true,
				Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			},
			"description":     optStr("Description.", ""),
			"link_local":      optBool("Enable IPv6 link-local addresses on the bridge.", false),
			"stp_enabled":     optBool("Enable spanning tree.", false),
			"stp_protocol":    optStrValidated("Spanning tree protocol: `rstp` or `stp`.", "rstp", stringvalidator.OneOf("rstp", "stp")),
			"stp_interfaces":  optSetDefault("Members that take part in spanning tree."),
			"span_interface":  optStr("Interface that receives a copy of all bridge traffic (SPAN port).", ""),
			"edge_ports":      optSetDefault("Members that are edge ports (no bridges behind them; skip the STP learning delay)."),
			"auto_edge_ports": optSetDefault("Members whose edge status is detected automatically."),
			"ptp_ports":       optSetDefault("Members on a point-to-point link (RSTP fast transitions)."),
			"auto_ptp_ports":  optSetDefault("Members whose point-to-point status is detected automatically."),
			"static_ports":    optSetDefault("Members whose learned addresses never age out."),
			"private_ports":   optSetDefault("Private members: they can't send traffic to each other through the bridge."),
			"max_age":         schema.Int64Attribute{MarkdownDescription: "STP max age in seconds (6-40); unset uses 20.", Optional: true},
			"forward_delay":   schema.Int64Attribute{MarkdownDescription: "STP forward delay in seconds (4-30); unset uses 15.", Optional: true},
			"hold_count":      schema.Int64Attribute{MarkdownDescription: "STP transmit hold count (1-10); unset uses 6.", Optional: true},
			"max_addresses":   schema.Int64Attribute{MarkdownDescription: "Size of the address cache; unset uses 2000.", Optional: true},
			"address_timeout": schema.Int64Attribute{MarkdownDescription: "Seconds before a learned address expires; unset uses 1200.", Optional: true},
			"device":          computedStr("Device name assigned by OPNsense, e.g. `bridge0`."),
			"id":              computedStr("UUID of the bridge."),
		},
	}
}

func bridgeDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Bridge interfaces (Interfaces > Devices > Bridge).",

		Attributes: map[string]dschema.Attribute{
			"id":              dsID("UUID of the bridge."),
			"members":         dsSet("Member interfaces."),
			"description":     dsStr("Description."),
			"link_local":      dsBool("Whether IPv6 link-local addresses are enabled."),
			"stp_enabled":     dsBool("Whether spanning tree is enabled."),
			"stp_protocol":    dsStr("Spanning tree protocol."),
			"stp_interfaces":  dsSet("Members in spanning tree."),
			"span_interface":  dsStr("SPAN port."),
			"edge_ports":      dsSet("Members that are edge ports"),
			"auto_edge_ports": dsSet("Members whose edge status is detected automatically."),
			"ptp_ports":       dsSet("Members on a point-to-point link"),
			"auto_ptp_ports":  dsSet("Members whose point-to-point status is detected automatically."),
			"static_ports":    dsSet("Members whose learned addresses never age out."),
			"private_ports":   dsSet("Private members: they can't send traffic to each other through the bridge."),
			"max_age":         dsInt("STP max age in seconds."),
			"forward_delay":   dsInt("STP forward delay in seconds."),
			"hold_count":      dsInt("STP transmit hold count."),
			"max_addresses":   dsInt("Size of the address cache."),
			"address_timeout": dsInt("Seconds before a learned address expires."),
			"device":          dsStr("Device name."),
		},
	}
}

func convertBridgeSchemaToStruct(d *bridgeResourceModel) (*interfaces.Bridge, error) {
	return &interfaces.Bridge{
		Members:        api.SelectedMapList(tools.SetToStringSlice(d.MemberSet)),
		Description:    d.Description.ValueString(),
		LinkLocal:      tools.BoolToString(d.LinkLocal.ValueBool()),
		EnableStp:      tools.BoolToString(d.StpEnabled.ValueBool()),
		StpProto:       api.SelectedMap(d.StpProtocol.ValueString()),
		StpInterfaces:  api.SelectedMapList(tools.SetToStringSlice(d.StpInterfaces)),
		SpanInterface:  api.SelectedMap(d.SpanInterface.ValueString()),
		EdgePorts:      api.SelectedMapList(tools.SetToStringSlice(d.EdgePorts)),
		AutoEdgePorts:  api.SelectedMapList(tools.SetToStringSlice(d.AutoEdgePorts)),
		PtpPorts:       api.SelectedMapList(tools.SetToStringSlice(d.PtpPorts)),
		AutoPtpPorts:   api.SelectedMapList(tools.SetToStringSlice(d.AutoPtpPorts)),
		StaticPorts:    api.SelectedMapList(tools.SetToStringSlice(d.StaticPorts)),
		PrivatePorts:   api.SelectedMapList(tools.SetToStringSlice(d.PrivatePorts)),
		MaxAge:         intOrEmpty(d.MaxAge),
		ForwardDelay:   intOrEmpty(d.ForwardDelay),
		HoldCount:      intOrEmpty(d.HoldCount),
		MaxAddresses:   intOrEmpty(d.MaxAddresses),
		AddressTimeout: intOrEmpty(d.AddressTimeout),
	}, nil
}

func convertBridgeStructToSchema(d *interfaces.Bridge) (*bridgeResourceModel, error) {
	return &bridgeResourceModel{
		MemberSet:      tools.StringSliceToSet(d.Members),
		Description:    types.StringValue(d.Description),
		LinkLocal:      types.BoolValue(tools.StringToBool(d.LinkLocal)),
		StpEnabled:     types.BoolValue(tools.StringToBool(d.EnableStp)),
		StpProtocol:    types.StringValue(d.StpProto.String()),
		StpInterfaces:  tools.StringSliceToSet(d.StpInterfaces),
		SpanInterface:  types.StringValue(d.SpanInterface.String()),
		EdgePorts:      tools.StringSliceToSet(d.EdgePorts),
		AutoEdgePorts:  tools.StringSliceToSet(d.AutoEdgePorts),
		PtpPorts:       tools.StringSliceToSet(d.PtpPorts),
		AutoPtpPorts:   tools.StringSliceToSet(d.AutoPtpPorts),
		StaticPorts:    tools.StringSliceToSet(d.StaticPorts),
		PrivatePorts:   tools.StringSliceToSet(d.PrivatePorts),
		MaxAge:         tools.StringToInt64Null(d.MaxAge),
		ForwardDelay:   tools.StringToInt64Null(d.ForwardDelay),
		HoldCount:      tools.StringToInt64Null(d.HoldCount),
		MaxAddresses:   tools.StringToInt64Null(d.MaxAddresses),
		AddressTimeout: tools.StringToInt64Null(d.AddressTimeout),
		Device:         types.StringValue(d.Device),
	}, nil
}

func mergeBridgeWriteOnly(prior, read *bridgeResourceModel) {}
