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
				MarkdownDescription: "Interfaces participating in the bridge. Takes interface identifiers, e.g. `opt1`.",
				ElementType:         types.StringType,
				Required:            true,
				Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			},
			"description":     optStr(descrHelp, ""),
			"link_local":      optBool("By default, link-local addresses for bridges are disabled. You can enable them manually using this option. However, when a bridge interface has IPv6 addresses, IPv6 addresses on a member interface will be automatically removed before the interface is added.", false),
			"stp_enabled":     optBool("Enable spanning tree options for this bridge.", false),
			"stp_protocol":    optStrValidated("Protocol used for spanning tree. One of `rstp` or `stp`.", "rstp", stringvalidator.OneOf("rstp", "stp")),
			"stp_interfaces":  optSetDefault("Enable Spanning Tree Protocol on interface. The if_bridge(4) driver has support for the IEEE 802.1D Spanning Tree Protocol (STP). STP is used to detect and remove loops in a network topology."),
			"span_interface":  optStr("Add the interface named by interface as a span port on the bridge. Span ports transmit a copy of every frame received by the bridge. This is most useful for snooping a bridged network passively on another host connected to one of the span ports of the bridge.", ""),
			"edge_ports":      optSetDefault("Set interface as an edge port. An edge port connects directly to end stations and cannot create bridging loops in the network; this allows it to transition straight to forwarding."),
			"auto_edge_ports": optSetDefault("Allow interface to automatically detect edge status. This is the default for all interfaces added to a bridge. (This will disable the autoedge status of interfaces.)"),
			"ptp_ports":       optSetDefault("Set the interface as a point-to-point link. This is required for straight transitions to forwarding and should be enabled on a direct link to another RSTP-capable switch."),
			"auto_ptp_ports":  optSetDefault("Automatically detect the point-to-point status on interface by checking the full duplex link status. This is the default for interfaces added to the bridge. (The interfaces selected here will be removed from default autoedge status.)"),
			"static_ports":    optSetDefault("Mark an interface as a \"sticky\" interface. Dynamically learned address entries are treated as static once entered into the cache. Sticky entries are never aged out of the cache or replaced, even if the address is seen on a different interface."),
			"private_ports":   optSetDefault("Mark an interface as a \"private\" interface. A private interface does not forward any traffic to any other port that is also a private interface."),
			"max_age":         schema.Int64Attribute{MarkdownDescription: "Set the time that a Spanning Tree Protocol configuration is valid. The default is 20 seconds. The minimum is 6 seconds and the maximum is 40 seconds. Leave unset to use the default.", Optional: true},
			"forward_delay":   schema.Int64Attribute{MarkdownDescription: "Set the time that must pass before an interface begins forwarding packets when Spanning Tree is enabled. The default is 15 seconds. The minimum is 4 seconds and the maximum is 30 seconds. Leave unset to use the default.", Optional: true},
			"hold_count":      schema.Int64Attribute{MarkdownDescription: "Set the transmit hold count for Spanning Tree. This is the number of packets transmitted before being rate limited. The default is 6. The minimum is 1 and the maximum is 10. Leave unset to use the default.", Optional: true},
			"max_addresses":   schema.Int64Attribute{MarkdownDescription: "Set the size of the bridge address cache to size. The default is 2000 entries. The minimum is `1`. Leave unset to use the default.", Optional: true},
			"address_timeout": schema.Int64Attribute{MarkdownDescription: "Set the timeout of address cache entries to this number of seconds. If seconds is zero, then address cache entries will not be expired. The default is 1200 seconds. Leave unset to use the default.", Optional: true},
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
			"members":         dsSet("Interfaces participating in the bridge."),
			"description":     dsStr(descrHelp),
			"link_local":      dsBool("By default, link-local addresses for bridges are disabled. You can enable them manually using this option. However, when a bridge interface has IPv6 addresses, IPv6 addresses on a member interface will be automatically removed before the interface is added."),
			"stp_enabled":     dsBool("Enable spanning tree options for this bridge."),
			"stp_protocol":    dsStr("Protocol used for spanning tree."),
			"stp_interfaces":  dsSet("Enable Spanning Tree Protocol on interface. The if_bridge(4) driver has support for the IEEE 802.1D Spanning Tree Protocol (STP). STP is used to detect and remove loops in a network topology."),
			"span_interface":  dsStr("Add the interface named by interface as a span port on the bridge. Span ports transmit a copy of every frame received by the bridge. This is most useful for snooping a bridged network passively on another host connected to one of the span ports of the bridge."),
			"edge_ports":      dsSet("Set interface as an edge port. An edge port connects directly to end stations and cannot create bridging loops in the network; this allows it to transition straight to forwarding."),
			"auto_edge_ports": dsSet("Allow interface to automatically detect edge status. This is the default for all interfaces added to a bridge. (This will disable the autoedge status of interfaces.)"),
			"ptp_ports":       dsSet("Set the interface as a point-to-point link. This is required for straight transitions to forwarding and should be enabled on a direct link to another RSTP-capable switch."),
			"auto_ptp_ports":  dsSet("Automatically detect the point-to-point status on interface by checking the full duplex link status. This is the default for interfaces added to the bridge. (The interfaces selected here will be removed from default autoedge status.)"),
			"static_ports":    dsSet("Mark an interface as a \"sticky\" interface. Dynamically learned address entries are treated as static once entered into the cache. Sticky entries are never aged out of the cache or replaced, even if the address is seen on a different interface."),
			"private_ports":   dsSet("Mark an interface as a \"private\" interface. A private interface does not forward any traffic to any other port that is also a private interface."),
			"max_age":         dsInt("Set the time that a Spanning Tree Protocol configuration is valid. The default is 20 seconds. The minimum is 6 seconds and the maximum is 40 seconds."),
			"forward_delay":   dsInt("Set the time that must pass before an interface begins forwarding packets when Spanning Tree is enabled. The default is 15 seconds. The minimum is 4 seconds and the maximum is 30 seconds."),
			"hold_count":      dsInt("Set the transmit hold count for Spanning Tree. This is the number of packets transmitted before being rate limited. The default is 6. The minimum is 1 and the maximum is 10."),
			"max_addresses":   dsInt("Set the size of the bridge address cache to size. The default is 2000 entries."),
			"address_timeout": dsInt("Set the timeout of address cache entries to this number of seconds. If seconds is zero, then address cache entries will not be expired. The default is 1200 seconds."),
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
