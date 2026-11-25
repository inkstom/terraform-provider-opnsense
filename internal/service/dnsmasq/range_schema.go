package dnsmasq

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/dnsmasq"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type rangeResourceModel struct {
	Interface        types.String `tfsdk:"interface"`
	Tag              types.String `tfsdk:"tag"`
	StartAddress     types.String `tfsdk:"start_address"`
	EndAddress       types.String `tfsdk:"end_address"`
	SubnetMask       types.String `tfsdk:"subnet_mask"`
	Constructor      types.String `tfsdk:"constructor"`
	Mode             types.String `tfsdk:"mode"`
	PrefixLength     types.Int64  `tfsdk:"prefix_length"`
	LeaseTime        types.Int64  `tfsdk:"lease_time"`
	DomainType       types.String `tfsdk:"domain_type"`
	Domain           types.String `tfsdk:"domain"`
	DisableHASync    types.Bool   `tfsdk:"disable_ha_sync"`
	RaMode           types.Set    `tfsdk:"ra_mode"`
	RaPriority       types.String `tfsdk:"ra_priority"`
	RaMTU            types.Int64  `tfsdk:"ra_mtu"`
	RaInterval       types.Int64  `tfsdk:"ra_interval"`
	RaRouterLifetime types.Int64  `tfsdk:"ra_router_lifetime"`
	Description      types.String `tfsdk:"description"`

	Id types.String `tfsdk:"id"`
}

func optionalString(desc string) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: desc + " Defaults to `\"\"`.",
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString(""),
	}
}

func optionalBool(desc string, def bool) schema.BoolAttribute {
	return schema.BoolAttribute{
		MarkdownDescription: desc,
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(def),
	}
}

// optionalInt is an integer the API stores as a string, where "" means unset.
// Unset maps to null so imported resources don't need to spell it out.
func optionalInt(desc string, v ...validator.Int64) schema.Int64Attribute {
	return schema.Int64Attribute{
		MarkdownDescription: desc + " Leave unset for the OPNsense default.",
		Optional:            true,
		Validators:          v,
	}
}

func optionalStringSet(desc string) schema.SetAttribute {
	return schema.SetAttribute{
		MarkdownDescription: desc + " Defaults to `[]`.",
		ElementType:         types.StringType,
		Optional:            true,
		Computed:            true,
		Default:             setdefault.StaticValue(tools.EmptySetValue(types.StringType)),
	}
}

func idAttribute(what string) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "UUID of the " + what + ".",
		Computed:            true,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
}

func intToAPI(v types.Int64) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return tools.Int64ToString(v.ValueInt64())
}

func rangeResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "A dnsmasq DHCP range (Services: Dnsmasq DNS & DHCP: DHCP ranges). " +
			"For a reservation-only range, set `mode = \"static\"` and only `start_address` (the network address).",
		Attributes: map[string]schema.Attribute{
			"interface":     optionalString("Interface to serve this range. Example `lan` or `opt1`; empty matches any interface."),
			"tag":           optionalString("Optional tag to set for requests matching this range which can be used to selectively match DHCP options. UUID of an `opnsense_dnsmasq_tag`."),
			"start_address": schema.StringAttribute{MarkdownDescription: "Start of the range, e.g. `192.168.1.100` for DHCPv4, `2000::1` for DHCPv6 or when a constructor is using a suffix like `::1`. To reveal IPv6 related options, enter a IPv6 address. When using router advertisements, it is possible to use a constructor with `::` as the start address and no end address.", Required: true},
			"end_address":   optionalString("End of the range. Leave empty for a static-only range."),
			"subnet_mask":   optionalString("Leave empty to auto-calculate the subnet mask from the interface or the network class of the start address. If a DHCP relay forwards IPv4 DHCP Discovers to Dnsmasq, setting a subnet mask is required in most cases."),
			"constructor":   optionalString("Interface to use to calculate a DHCPv6 or RA range. Start address can then be specified as a suffix (e.g. `::`, `::1` or `::400`)."),
			"mode": schema.StringAttribute{
				MarkdownDescription: "Mode flags to set for this range, `static` means no addresses will be automatically assigned. One of `static`, or empty for a dynamic range. Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				Validators:          []validator.String{stringvalidator.OneOf("", "static")},
			},
			"prefix_length": optionalInt("Prefix length offered to the client. Custom values in this field will be ignored if Router Advertisements are enabled, as SLAAC will only work with a prefix length of 64. Must be between `1` and `64`.", int64validator.Between(1, 64)),
			"lease_time":    optionalInt("Defines how long the addresses (leases) given out by the server are valid (in seconds). Set `0` for infinite."),
			"domain_type": schema.StringAttribute{
				MarkdownDescription: "Choose if the domain will only match clients in this range, or all clients in any subnets on the selected interface. If you create both IPv4 and IPv6 ranges, setting this to `interface` on both ranges is recommended. One of `range` or `interface`. Defaults to `\"range\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("range"),
				Validators:          []validator.String{stringvalidator.OneOf("range", "interface")},
			},
			"domain":          optionalString("Offer this domain to DHCP clients."),
			"disable_ha_sync": optionalBool("Ignore this range from being transferred or updated by HA sync. Defaults to `false`.", false),
			"ra_mode":         optionalStringSet("Control how IPv6 clients receive their addresses. Enabling Router Advertisements in general settings will enable it for all configured DHCPv6 ranges with the managed address bits set, and the use SLAAC bit reset. To change this default, select a combination of the possible options here. `slaac`, `ra-stateless` and `ra-names` can be freely combined, all other options shall remain single selections. Accepted values: `ra-only`, `slaac`, `ra-names`, `ra-stateless`, `ra-advrouter`, `off-link`."),
			"ra_priority": schema.StringAttribute{
				MarkdownDescription: "Priority of the RA announcements. One of `high`, `low`, or empty for normal priority. Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				Validators:          []validator.String{stringvalidator.OneOf("", "high", "low")},
			},
			"ra_mtu":             optionalInt("Optional MTU to send to clients via Router Advertisements. If unsure leave empty."),
			"ra_interval":        optionalInt("Time (seconds) between Router Advertisements."),
			"ra_router_lifetime": optionalInt("The lifetime of the route may be changed or set to zero, which allows a router to advertise prefixes but not a route via itself. When using HA, setting a short timespan here is advised for faster IPv6 failover. A good combination could be 10 seconds RA interval and 30 seconds RA router lifetime. Going lower than that can pose issues in busy networks. Value in seconds."),
			"description":        optionalString("You may enter a description here for your reference (not parsed)."),
			"id":                 idAttribute("DHCP range"),
		},
	}
}

func convertRangeSchemaToStruct(d *rangeResourceModel) (*dnsmasq.Range, error) {
	return &dnsmasq.Range{
		Interface:        api.SelectedMap(d.Interface.ValueString()),
		Tag:              api.SelectedMap(d.Tag.ValueString()),
		StartAddress:     d.StartAddress.ValueString(),
		EndAddress:       d.EndAddress.ValueString(),
		SubnetMask:       d.SubnetMask.ValueString(),
		Constructor:      api.SelectedMap(d.Constructor.ValueString()),
		Mode:             api.SelectedMap(d.Mode.ValueString()),
		PrefixLength:     intToAPI(d.PrefixLength),
		LeaseTime:        intToAPI(d.LeaseTime),
		DomainType:       api.SelectedMap(d.DomainType.ValueString()),
		Domain:           d.Domain.ValueString(),
		DisableHASync:    tools.BoolToString(d.DisableHASync.ValueBool()),
		RaMode:           tools.SetToStringSlice(d.RaMode),
		RaPriority:       api.SelectedMap(d.RaPriority.ValueString()),
		RaMTU:            intToAPI(d.RaMTU),
		RaInterval:       intToAPI(d.RaInterval),
		RaRouterLifetime: intToAPI(d.RaRouterLifetime),
		Description:      d.Description.ValueString(),
	}, nil
}

func convertRangeStructToSchema(d *dnsmasq.Range) (*rangeResourceModel, error) {
	return &rangeResourceModel{
		Interface:        types.StringValue(d.Interface.String()),
		Tag:              types.StringValue(d.Tag.String()),
		StartAddress:     types.StringValue(d.StartAddress),
		EndAddress:       types.StringValue(d.EndAddress),
		SubnetMask:       types.StringValue(d.SubnetMask),
		Constructor:      types.StringValue(d.Constructor.String()),
		Mode:             types.StringValue(d.Mode.String()),
		PrefixLength:     tools.StringToInt64Null(d.PrefixLength),
		LeaseTime:        tools.StringToInt64Null(d.LeaseTime),
		DomainType:       types.StringValue(d.DomainType.String()),
		Domain:           types.StringValue(d.Domain),
		DisableHASync:    types.BoolValue(tools.StringToBool(d.DisableHASync)),
		RaMode:           tools.StringSliceToSet(d.RaMode),
		RaPriority:       types.StringValue(d.RaPriority.String()),
		RaMTU:            tools.StringToInt64Null(d.RaMTU),
		RaInterval:       tools.StringToInt64Null(d.RaInterval),
		RaRouterLifetime: tools.StringToInt64Null(d.RaRouterLifetime),
		Description:      types.StringValue(d.Description),
	}, nil
}
