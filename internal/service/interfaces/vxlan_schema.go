package interfaces

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/interfaces"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type vxlanResourceModel struct {
	Vni           types.String `tfsdk:"vni"`
	LocalAddress  types.String `tfsdk:"local_address"`
	LocalPort     types.String `tfsdk:"local_port"`
	RemoteAddress types.String `tfsdk:"remote_address"`
	RemotePort    types.String `tfsdk:"remote_port"`
	Group         types.String `tfsdk:"group"`
	Interface     types.String `tfsdk:"interface"`
	Device        types.String `tfsdk:"device"`

	Id types.String `tfsdk:"id"`
}

func vxlanResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "VXLAN interfaces (Interfaces > Devices > VXLAN). Set either `remote_address` (unicast) or `group` (multicast).",
		Attributes: map[string]schema.Attribute{
			"vni":            reqStr("VXLAN Network Identifier (VNI) that identifies the virtual network segment membership of the interface. Accepts `0` to `16777215`."),
			"local_address":  reqStr("The source address used in the encapsulating IPv4/IPv6 header. The address should already be assigned to an existing interface. When the interface is configured in unicast mode, the listening socket is bound to this address."),
			"local_port":     optStr("Define the port to be used. This is the local UDP port; empty uses `4789`.", ""),
			"remote_address": optStr("The interface can be configured in a unicast, or point-to-point, mode to create a tunnel between two hosts. This is the IP address of the remote end of the tunnel.", ""),
			"remote_port":    optStr("Define the port to be used. This is the remote UDP port; empty uses `4789`.", ""),
			"group":          optStr("The interface can be configured in a multicast mode to create a virtual network of hosts. This is the IP multicast group address the interface will join.", ""),
			"interface":      optStr("When the interface is configured in multicast mode, the dev interface is used to transmit IP multicast packets. Takes an interface identifier, e.g. `wan`.", ""),
			"device":         computedStr("Device name assigned by OPNsense, e.g. `vxlan1`."),
			"id":             computedStr("UUID of the VXLAN."),
		},
	}
}

func vxlanDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "VXLAN interfaces (Interfaces > Devices > VXLAN).",
		Attributes: map[string]dschema.Attribute{
			"id":             dsID("UUID of the VXLAN."),
			"vni":            dsStr("VXLAN Network Identifier (VNI) that identifies the virtual network segment membership of the interface."),
			"local_address":  dsStr("The source address used in the encapsulating IPv4/IPv6 header."),
			"local_port":     dsStr("Local UDP port."),
			"remote_address": dsStr("The IP address of the remote end of the tunnel (unicast mode)."),
			"remote_port":    dsStr("Remote UDP port."),
			"group":          dsStr("The IP multicast group address the interface will join (multicast mode)."),
			"interface":      dsStr("When the interface is configured in multicast mode, the dev interface is used to transmit IP multicast packets."),
			"device":         dsStr("Device name."),
		},
	}
}

func convertVxlanSchemaToStruct(d *vxlanResourceModel) (*interfaces.Vxlan, error) {
	return &interfaces.Vxlan{
		Vni:           d.Vni.ValueString(),
		LocalAddress:  d.LocalAddress.ValueString(),
		LocalPort:     d.LocalPort.ValueString(),
		RemoteAddress: d.RemoteAddress.ValueString(),
		RemotePort:    d.RemotePort.ValueString(),
		Group:         d.Group.ValueString(),
		Device:        api.SelectedMap(d.Interface.ValueString()),
	}, nil
}

func convertVxlanStructToSchema(d *interfaces.Vxlan) (*vxlanResourceModel, error) {
	return &vxlanResourceModel{
		Vni:           types.StringValue(d.Vni),
		LocalAddress:  types.StringValue(d.LocalAddress),
		LocalPort:     types.StringValue(d.LocalPort),
		RemoteAddress: types.StringValue(d.RemoteAddress),
		RemotePort:    types.StringValue(d.RemotePort),
		Group:         types.StringValue(d.Group),
		Interface:     types.StringValue(d.Device.String()),
		Device:        types.StringValue("vxlan" + d.DeviceId),
	}, nil
}

func mergeVxlanWriteOnly(prior, read *vxlanResourceModel) {}
