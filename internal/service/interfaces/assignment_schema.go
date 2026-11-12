package interfaces

import (
	"regexp"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/interfaces"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	assignmentDescrRegex = regexp.MustCompile(`^[a-zA-Z_0-9]{0,255}$`)
	hexIdRegex           = regexp.MustCompile(`^0x[a-fA-F0-9]{0,30}$`)
)

// assignmentResourceModel describes the resource data model.
type assignmentResourceModel struct {
	Device           types.String `tfsdk:"device"`
	Description      types.String `tfsdk:"description"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Lock             types.Bool   `tfsdk:"lock"`
	BlockPrivate     types.Bool   `tfsdk:"block_private"`
	BlockBogons      types.Bool   `tfsdk:"block_bogons"`
	SpoofMac         types.String `tfsdk:"spoof_mac"`
	Promiscuous      types.Bool   `tfsdk:"promiscuous"`
	MTU              types.Int64  `tfsdk:"mtu"`
	MSS              types.Int64  `tfsdk:"mss"`
	GatewayInterface types.Bool   `tfsdk:"gateway_interface"`
	IPv4Type         types.String `tfsdk:"ipv4_type"`
	IPv4Address      types.String `tfsdk:"ipv4_address"`
	IPv4Gateway      types.String `tfsdk:"ipv4_gateway"`
	IPv6Type         types.String `tfsdk:"ipv6_type"`
	IPv6Address      types.String `tfsdk:"ipv6_address"`
	IPv6Gateway      types.String `tfsdk:"ipv6_gateway"`
	DhcpHostname     types.String `tfsdk:"dhcp_hostname"`
	Track6Interface  types.String `tfsdk:"track6_interface"`
	Track6PrefixId   types.String `tfsdk:"track6_prefix_id"`

	Id types.String `tfsdk:"id"`
}

func assignmentResourceSchema() schema.Schema {
	optBool := func(desc string, def bool) schema.BoolAttribute {
		return schema.BoolAttribute{MarkdownDescription: desc, Optional: true, Computed: true, Default: booldefault.StaticBool(def)}
	}
	optString := func(desc string, v ...validator.String) schema.StringAttribute {
		return schema.StringAttribute{MarkdownDescription: desc, Optional: true, Validators: v}
	}
	optSize := func(desc string) schema.Int64Attribute {
		return schema.Int64Attribute{
			MarkdownDescription: desc + " Set to `-1` to use the device default. Defaults to `-1`.",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(-1),
			Validators: []validator.Int64{
				int64validator.Any(int64validator.OneOf(-1), int64validator.Between(576, 65535)),
			},
		}
	}

	return schema.Schema{
		MarkdownDescription: "Assigns a network device (NIC, VLAN, WireGuard `wgX`, bridge, LAGG, ...) to a logical interface (`optN`) and configures it (Interfaces > Assignments). " +
			"Requires OPNsense 26.7.7 or later.\n\n" +
			"~> **Warning:** every change runs a full interface apply on the firewall. Changing the interface you reach the API through can cut off Terraform mid-apply. " +
			"Deleting an assignment also deletes firewall/NAT rules that reference only that interface.\n\n" +
			"Only static, DHCP and track/SLAAC-style addressing are modelled; settings not exposed here are left untouched on update.",
		Version: 1,

		Attributes: map[string]schema.Attribute{
			"device": schema.StringAttribute{
				MarkdownDescription: "Device to assign, e.g. `igc2`, `vlan0.10` (reference `opnsense_interfaces_vlan.device` only when you set that name explicitly; auto-generated names are not exposed), `wg0` (see `opnsense_wireguard_server.device`). A device can only be assigned once.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Interface name shown throughout the UI, e.g. `IOT`. Letters, digits and `_` only; must be unique. When unset, OPNsense shows the identifier (e.g. `OPT3`).",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(assignmentDescrRegex, "may only contain letters, digits and underscores"),
				},
			},
			"enabled": optBool("Enable the interface. Defaults to `true`.", true),
			"lock": optBool("Prevent removal from the UI. The provider clears the lock itself before destroying. "+
				"Defaults to `false` (OPNsense's UI defaults to `true`).", false),
			"block_private":     optBool("Block traffic from private networks (RFC 1918) on this interface. Defaults to `false`.", false),
			"block_bogons":      optBool("Block traffic from bogon networks on this interface. Defaults to `false`.", false),
			"promiscuous":       optBool("Put the device in promiscuous mode. Defaults to `false`.", false),
			"gateway_interface": optBool("Dynamic gateway policy: allow sending traffic to this interface without an intermediate gateway (useful for VPN interfaces such as WireGuard). Defaults to `false`.", false),
			"spoof_mac": optString("MAC address to spoof on this interface, e.g. `02:00:00:00:00:01`.",
				stringvalidator.RegexMatches(regexp.MustCompile(`^([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}$`), "must be a MAC address")),
			"mtu": optSize("Interface MTU (576-65535)."),
			"mss": optSize("TCP MSS clamping (576-65535)."),
			"ipv4_type": optString("IPv4 configuration type: `staticv4` or `dhcp`. Leave unset for none.",
				stringvalidator.OneOf("staticv4", "dhcp")),
			"ipv4_address": optString("IPv4 address in CIDR notation, e.g. `10.10.0.1/24`. Required when `ipv4_type` is `staticv4`."),
			"ipv4_gateway": optString("Name of an IPv4 upstream gateway for this interface (WAN-type interfaces only), e.g. `opnsense_routing_gateway.wan2.name`. " +
				"Note: a gateway references its interface, so setting both creates a cycle — prefer `opnsense_routing_gateway.upstream`."),
			"ipv6_type": optString("IPv6 configuration type: `staticv6`, `dhcp6`, `slaac`, `linklocal`, `idassoc6` or `track6`. Leave unset for none.",
				stringvalidator.OneOf("staticv6", "dhcp6", "slaac", "linklocal", "idassoc6", "track6")),
			"ipv6_address":     optString("IPv6 address in CIDR notation. Required when `ipv6_type` is `staticv6`."),
			"ipv6_gateway":     optString("Name of an IPv6 upstream gateway for this interface."),
			"dhcp_hostname":    optString("Hostname sent to the DHCP server when `ipv4_type` is `dhcp`."),
			"track6_interface": optString("Parent interface whose delegated prefix to use when `ipv6_type` is `idassoc6` or `track6`, e.g. `wan`."),
			"track6_prefix_id": optString("Hexadecimal prefix ID (e.g. `0x1`) selected from the delegated prefix.",
				stringvalidator.RegexMatches(hexIdRegex, "must be a hex value like 0x1")),
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Interface identifier assigned by OPNsense, e.g. `opt3`. Use this wherever an interface is expected (firewall rules, gateways, DHCP).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func assignmentDataSourceSchema() dschema.Schema {
	str := func(desc string) dschema.StringAttribute {
		return dschema.StringAttribute{MarkdownDescription: desc, Computed: true}
	}
	bl := func(desc string) dschema.BoolAttribute {
		return dschema.BoolAttribute{MarkdownDescription: desc, Computed: true}
	}
	in := func(desc string) dschema.Int64Attribute {
		return dschema.Int64Attribute{MarkdownDescription: desc, Computed: true}
	}

	return dschema.Schema{
		MarkdownDescription: "Reads an interface assignment by identifier (e.g. `wan`, `lan`, `opt3`). Requires OPNsense 26.7.7 or later.",

		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				MarkdownDescription: "Interface identifier, e.g. `wan`, `lan`, `opt3`.",
				Required:            true,
			},
			"device":            str("Assigned device."),
			"description":       str("Interface name."),
			"enabled":           bl("Whether the interface is enabled."),
			"lock":              bl("Whether the interface is locked against removal."),
			"block_private":     bl("Whether private networks are blocked."),
			"block_bogons":      bl("Whether bogon networks are blocked."),
			"promiscuous":       bl("Whether promiscuous mode is on."),
			"gateway_interface": bl("Whether the dynamic gateway policy is on."),
			"spoof_mac":         str("Spoofed MAC address."),
			"mtu":               in("MTU (`-1` when default)."),
			"mss":               in("MSS (`-1` when default)."),
			"ipv4_type":         str("IPv4 configuration type."),
			"ipv4_address":      str("IPv4 address (CIDR)."),
			"ipv4_gateway":      str("IPv4 upstream gateway."),
			"ipv6_type":         str("IPv6 configuration type."),
			"ipv6_address":      str("IPv6 address (CIDR)."),
			"ipv6_gateway":      str("IPv6 upstream gateway."),
			"dhcp_hostname":     str("DHCP client hostname."),
			"track6_interface":  str("Track/IA parent interface."),
			"track6_prefix_id":  str("Track/IA prefix ID."),
		},
	}
}

func convertAssignmentSchemaToStruct(d *assignmentResourceModel) (*interfaces.Assignment, error) {
	return &interfaces.Assignment{
		Description:      d.Description.ValueString(),
		Device:           api.SelectedMap(d.Device.ValueString()),
		Enable:           tools.BoolToString(d.Enabled.ValueBool()),
		Lock:             tools.BoolToString(d.Lock.ValueBool()),
		BlockPrivate:     tools.BoolToString(d.BlockPrivate.ValueBool()),
		BlockBogons:      tools.BoolToString(d.BlockBogons.ValueBool()),
		SpoofMac:         d.SpoofMac.ValueString(),
		Promiscuous:      tools.BoolToString(d.Promiscuous.ValueBool()),
		MTU:              tools.Int64ToStringNegative(d.MTU.ValueInt64()),
		MSS:              tools.Int64ToStringNegative(d.MSS.ValueInt64()),
		GatewayInterface: tools.BoolToString(d.GatewayInterface.ValueBool()),
		IPv4Type:         api.SelectedMap(d.IPv4Type.ValueString()),
		IPv4Address:      d.IPv4Address.ValueString(),
		Gateway:          api.SelectedMap(d.IPv4Gateway.ValueString()),
		IPv6Type:         api.SelectedMap(d.IPv6Type.ValueString()),
		IPv6Address:      d.IPv6Address.ValueString(),
		GatewayV6:        api.SelectedMap(d.IPv6Gateway.ValueString()),
		DhcpHostname:     d.DhcpHostname.ValueString(),
		Track6Interface:  api.SelectedMap(d.Track6Interface.ValueString()),
		Track6PrefixId:   d.Track6PrefixId.ValueString(),
	}, nil
}

func convertAssignmentStructToSchema(d *interfaces.Assignment) (*assignmentResourceModel, error) {
	return &assignmentResourceModel{
		Device:           types.StringValue(d.Device.String()),
		Description:      tools.StringOrNull(d.Description),
		Enabled:          types.BoolValue(tools.StringToBool(d.Enable)),
		Lock:             types.BoolValue(tools.StringToBool(d.Lock)),
		BlockPrivate:     types.BoolValue(tools.StringToBool(d.BlockPrivate)),
		BlockBogons:      types.BoolValue(tools.StringToBool(d.BlockBogons)),
		SpoofMac:         tools.StringOrNull(d.SpoofMac),
		Promiscuous:      types.BoolValue(tools.StringToBool(d.Promiscuous)),
		MTU:              types.Int64Value(tools.StringToInt64(d.MTU)),
		MSS:              types.Int64Value(tools.StringToInt64(d.MSS)),
		GatewayInterface: types.BoolValue(tools.StringToBool(d.GatewayInterface)),
		IPv4Type:         tools.StringOrNull(d.IPv4Type.String()),
		IPv4Address:      tools.StringOrNull(d.IPv4Address),
		IPv4Gateway:      tools.StringOrNull(d.Gateway.String()),
		IPv6Type:         tools.StringOrNull(d.IPv6Type.String()),
		IPv6Address:      tools.StringOrNull(d.IPv6Address),
		IPv6Gateway:      tools.StringOrNull(d.GatewayV6.String()),
		DhcpHostname:     tools.StringOrNull(d.DhcpHostname),
		Track6Interface:  tools.StringOrNull(d.Track6Interface.String()),
		Track6PrefixId:   tools.StringOrNull(d.Track6PrefixId),
	}, nil
}
