package interfaces

import (
	"github.com/browningluke/opnsense-go/pkg/interfaces"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type greResourceModel struct {
	LocalAddress        types.String `tfsdk:"local_address"`
	RemoteAddress       types.String `tfsdk:"remote_address"`
	TunnelLocalAddress  types.String `tfsdk:"tunnel_local_address"`
	TunnelRemoteAddress types.String `tfsdk:"tunnel_remote_address"`
	TunnelRemoteNet     types.String `tfsdk:"tunnel_remote_net"`
	Description         types.String `tfsdk:"description"`
	Device              types.String `tfsdk:"device"`

	Id types.String `tfsdk:"id"`
}

func greResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "GRE tunnels (Interfaces > Devices > GRE). Assign the device with `opnsense_interfaces_assignment` to route over it.",
		Attributes: map[string]schema.Attribute{
			"local_address":         reqStr("Local endpoint: an interface such as `wan`, or a virtual IP."),
			"remote_address":        reqStr("Peer address where encapsulated gre packets will be sent."),
			"tunnel_local_address":  reqStr("Local gre tunnel endpoint."),
			"tunnel_remote_address": reqStr("Remote gre tunnel endpoint."),
			"tunnel_remote_net":     optStr("Netmask [ipv4] or prefix [ipv6] to use for this tunnel. Given as a prefix length from `1` to `128`.", "32"),
			"description":           optStr(descrHelp, ""),
			"device":                computedStr("Device name assigned by OPNsense, e.g. `gre0`."),
			"id":                    computedStr("UUID of the tunnel."),
		},
	}
}

func greDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "GRE tunnels (Interfaces > Devices > GRE).",
		Attributes: map[string]dschema.Attribute{
			"id":                    dsID("UUID of the tunnel."),
			"local_address":         dsStr("Local endpoint."),
			"remote_address":        dsStr("Peer address where encapsulated gre packets will be sent."),
			"tunnel_local_address":  dsStr("Local gre tunnel endpoint."),
			"tunnel_remote_address": dsStr("Remote gre tunnel endpoint."),
			"tunnel_remote_net":     dsStr("Netmask [ipv4] or prefix [ipv6] to use for this tunnel, as a prefix length."),
			"description":           dsStr(descrHelp),
			"device":                dsStr("Device name."),
		},
	}
}

func convertGreSchemaToStruct(d *greResourceModel) (*interfaces.Gre, error) {
	return &interfaces.Gre{
		LocalAddress:        d.LocalAddress.ValueString(),
		RemoteAddress:       d.RemoteAddress.ValueString(),
		TunnelLocalAddress:  d.TunnelLocalAddress.ValueString(),
		TunnelRemoteAddress: d.TunnelRemoteAddress.ValueString(),
		TunnelRemoteNet:     d.TunnelRemoteNet.ValueString(),
		Description:         d.Description.ValueString(),
	}, nil
}

func convertGreStructToSchema(d *interfaces.Gre) (*greResourceModel, error) {
	return &greResourceModel{
		LocalAddress:        types.StringValue(d.LocalAddress),
		RemoteAddress:       types.StringValue(d.RemoteAddress),
		TunnelLocalAddress:  types.StringValue(d.TunnelLocalAddress),
		TunnelRemoteAddress: types.StringValue(d.TunnelRemoteAddress),
		TunnelRemoteNet:     types.StringValue(d.TunnelRemoteNet),
		Description:         types.StringValue(d.Description),
		Device:              types.StringValue(d.Device),
	}, nil
}

func mergeGreWriteOnly(prior, read *greResourceModel) {}
