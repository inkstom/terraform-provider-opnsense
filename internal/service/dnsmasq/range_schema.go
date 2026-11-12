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
			"interface":     optionalString("Interface to serve, e.g. `lan` or `opt1`. Empty matches any interface."),
			"tag":           optionalString("UUID of a dnsmasq tag to set on clients that get an address from this range."),
			"start_address": schema.StringAttribute{MarkdownDescription: "First address of the range (IPv4 or IPv6).", Required: true},
			"end_address":   optionalString("Last address of the range. Leave empty for a static-only range."),
			"subnet_mask":   optionalString("Subnet mask, if it can't be derived from the interface."),
			"constructor":   optionalString("IPv6 constructor interface (prefix delegation)."),
			"mode": schema.StringAttribute{
				MarkdownDescription: "`static` to only hand out reserved addresses, or empty for a dynamic range. Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				Validators:          []validator.String{stringvalidator.OneOf("", "static")},
			},
			"prefix_length": optionalInt("IPv6 prefix length.", int64validator.Between(1, 64)),
			"lease_time":    optionalInt("Lease time in seconds."),
			"domain_type": schema.StringAttribute{
				MarkdownDescription: "Whether `domain` applies to this `range` or the whole `interface`. Defaults to `\"range\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("range"),
				Validators:          []validator.String{stringvalidator.OneOf("range", "interface")},
			},
			"domain":          optionalString("Domain to offer to clients in this range."),
			"disable_ha_sync": optionalBool("Exclude this range from HA sync. Defaults to `false`.", false),
			"ra_mode":         optionalStringSet("IPv6 router advertisement modes (`ra-only`, `slaac`, `ra-names`, `ra-stateless`, `ra-advrouter`, `off-link`)."),
			"ra_priority": schema.StringAttribute{
				MarkdownDescription: "Router advertisement priority: `high`, `low` or empty. Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				Validators:          []validator.String{stringvalidator.OneOf("", "high", "low")},
			},
			"ra_mtu":             optionalInt("MTU advertised in router advertisements."),
			"ra_interval":        optionalInt("Router advertisement interval in seconds."),
			"ra_router_lifetime": optionalInt("Router lifetime in seconds."),
			"description":        optionalString("Description."),
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
