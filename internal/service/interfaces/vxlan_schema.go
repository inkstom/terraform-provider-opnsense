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
			"vni":            reqStr("VXLAN network identifier (1-16777215)."),
			"local_address":  reqStr("Source address for the tunnel."),
			"local_port":     optStr("Local UDP port; empty uses 4789.", ""),
			"remote_address": optStr("Unicast remote address.", ""),
			"remote_port":    optStr("Remote UDP port; empty uses 4789.", ""),
			"group":          optStr("Multicast group address.", ""),
			"interface":      optStr("Interface for multicast traffic, e.g. `wan`.", ""),
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
			"vni":            dsStr("VXLAN network identifier."),
			"local_address":  dsStr("Source address."),
			"local_port":     dsStr("Local UDP port."),
			"remote_address": dsStr("Unicast remote address."),
			"remote_port":    dsStr("Remote UDP port."),
			"group":          dsStr("Multicast group."),
			"interface":      dsStr("Multicast interface."),
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
