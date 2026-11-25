package interfaces

import (
	"github.com/browningluke/opnsense-go/pkg/interfaces"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type gifResourceModel struct {
	LocalAddress        types.String `tfsdk:"local_address"`
	RemoteAddress       types.String `tfsdk:"remote_address"`
	TunnelLocalAddress  types.String `tfsdk:"tunnel_local_address"`
	TunnelRemoteAddress types.String `tfsdk:"tunnel_remote_address"`
	TunnelRemoteNet     types.String `tfsdk:"tunnel_remote_net"`
	Description         types.String `tfsdk:"description"`
	Link1               types.Bool   `tfsdk:"link1"`
	Link2               types.Bool   `tfsdk:"link2"`
	NoClamp             types.Bool   `tfsdk:"no_clamp"`
	Device              types.String `tfsdk:"device"`

	Id types.String `tfsdk:"id"`
}

func gifResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "GIF (IP-in-IP) tunnels (Interfaces > Devices > GIF). Assign the device with `opnsense_interfaces_assignment` to route over it.",
		Attributes: map[string]schema.Attribute{
			"local_address":         reqStr("Local endpoint: an interface such as `wan`, or a virtual IP."),
			"remote_address":        reqStr("Peer address where encapsulated gif packets will be sent."),
			"tunnel_local_address":  reqStr("Local gif tunnel endpoint."),
			"tunnel_remote_address": reqStr("Remote gif tunnel endpoint."),
			"tunnel_remote_net":     optStr("Netmask [ipv4] or prefix [ipv6] to use for this tunnel. Given as a prefix length from `1` to `128`.", "32"),
			"description":           optStr(descrHelp, ""),
			"link1":                 optBool("Note that the ECN friendly behavior violates RFC2893. This should be used in mutual agreement with the peer. Setting this enables ECN friendly behavior (`link1` flag).", false),
			"link2":                 optBool("Ingress filtering on outer tunnel source can break tunnel operation in asymmetrically routed networks, in which case this can be disabled by marking this option. Setting this disables ingress filtering (`link2` flag).", false),
			"no_clamp":              optBool("Do not clamp the effective MTU to 1280 bytes when the outer tunnel protocol is IPv6.", false),
			"device":                computedStr("Device name assigned by OPNsense, e.g. `gif0`."),
			"id":                    computedStr("UUID of the tunnel."),
		},
	}
}

func gifDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "GIF (IP-in-IP) tunnels (Interfaces > Devices > GIF).",
		Attributes: map[string]dschema.Attribute{
			"id":                    dsID("UUID of the tunnel."),
			"local_address":         dsStr("Local endpoint."),
			"remote_address":        dsStr("Peer address where encapsulated gif packets will be sent."),
			"tunnel_local_address":  dsStr("Local gif tunnel endpoint."),
			"tunnel_remote_address": dsStr("Remote gif tunnel endpoint."),
			"tunnel_remote_net":     dsStr("Netmask [ipv4] or prefix [ipv6] to use for this tunnel, as a prefix length."),
			"description":           dsStr(descrHelp),
			"link1":                 dsBool("ECN friendly behavior (`link1` flag)."),
			"link2":                 dsBool("Ingress filtering disabled (`link2` flag)."),
			"no_clamp":              dsBool("Do not clamp the effective MTU to 1280 bytes when the outer tunnel protocol is IPv6."),
			"device":                dsStr("Device name."),
		},
	}
}

func convertGifSchemaToStruct(d *gifResourceModel) (*interfaces.Gif, error) {
	return &interfaces.Gif{
		LocalAddress:        d.LocalAddress.ValueString(),
		RemoteAddress:       d.RemoteAddress.ValueString(),
		TunnelLocalAddress:  d.TunnelLocalAddress.ValueString(),
		TunnelRemoteAddress: d.TunnelRemoteAddress.ValueString(),
		TunnelRemoteNet:     d.TunnelRemoteNet.ValueString(),
		Description:         d.Description.ValueString(),
		Link1:               tools.BoolToString(d.Link1.ValueBool()),
		Link2:               tools.BoolToString(d.Link2.ValueBool()),
		NoClamp:             tools.BoolToString(d.NoClamp.ValueBool()),
	}, nil
}

func convertGifStructToSchema(d *interfaces.Gif) (*gifResourceModel, error) {
	return &gifResourceModel{
		LocalAddress:        types.StringValue(d.LocalAddress),
		RemoteAddress:       types.StringValue(d.RemoteAddress),
		TunnelLocalAddress:  types.StringValue(d.TunnelLocalAddress),
		TunnelRemoteAddress: types.StringValue(d.TunnelRemoteAddress),
		TunnelRemoteNet:     types.StringValue(d.TunnelRemoteNet),
		Description:         types.StringValue(d.Description),
		Link1:               types.BoolValue(tools.StringToBool(d.Link1)),
		Link2:               types.BoolValue(tools.StringToBool(d.Link2)),
		NoClamp:             types.BoolValue(tools.StringToBool(d.NoClamp)),
		Device:              types.StringValue(d.Device),
	}, nil
}

func mergeGifWriteOnly(prior, read *gifResourceModel) {}
