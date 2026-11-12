package dnsmasq

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/dnsmasq"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type domainResourceModel struct {
	Sequence      types.Int64  `tfsdk:"sequence"`
	Domain        types.String `tfsdk:"domain"`
	Server        types.String `tfsdk:"server"`
	Port          types.String `tfsdk:"port"`
	SourceIp      types.String `tfsdk:"source_ip"`
	FirewallAlias types.String `tfsdk:"firewall_alias"`
	Description   types.String `tfsdk:"description"`

	Id types.String `tfsdk:"id"`
}

func domainResourceSchema() schema.Schema {
	str := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{MarkdownDescription: desc + " Defaults to `\"\"`.", Optional: true, Computed: true, Default: stringdefault.StaticString("")}
	}
	return schema.Schema{
		MarkdownDescription: "Dnsmasq domain overrides (Services > Dnsmasq DNS & DHCP > Domains): send queries for a domain to a specific DNS server.",
		Attributes: map[string]schema.Attribute{
			"sequence": schema.Int64Attribute{
				MarkdownDescription: "Order of the entry in the generated configuration. Defaults to `1`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
			},
			"domain":         schema.StringAttribute{MarkdownDescription: "Domain to override, e.g. `corp.example.com`.", Required: true},
			"server":         schema.StringAttribute{MarkdownDescription: "DNS server to send the domain's queries to.", Required: true},
			"port":           str("DNS server port; empty uses 53."),
			"source_ip":      str("Source address for queries to the server."),
			"firewall_alias": str("Firewall alias (by name) to add the domain's resolved addresses to."),
			"description":    str("Description."),
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the override.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func domainDataSourceSchema() dschema.Schema {
	s := func(d string) dschema.StringAttribute {
		return dschema.StringAttribute{MarkdownDescription: d, Computed: true}
	}
	return dschema.Schema{
		MarkdownDescription: "Dnsmasq domain overrides (Services > Dnsmasq DNS & DHCP > Domains).",
		Attributes: map[string]dschema.Attribute{
			"id":             dschema.StringAttribute{MarkdownDescription: "UUID of the override.", Required: true},
			"sequence":       dschema.Int64Attribute{MarkdownDescription: "Order.", Computed: true},
			"domain":         s("Domain."),
			"server":         s("DNS server."),
			"port":           s("DNS server port."),
			"source_ip":      s("Source address."),
			"firewall_alias": s("Firewall alias."),
			"description":    s("Description."),
		},
	}
}

func convertDomainSchemaToStruct(d *domainResourceModel) (*dnsmasq.Domain, error) {
	return &dnsmasq.Domain{
		Sequence:      tools.Int64ToString(d.Sequence.ValueInt64()),
		Domain:        d.Domain.ValueString(),
		Ip:            d.Server.ValueString(),
		Port:          d.Port.ValueString(),
		SourceIp:      d.SourceIp.ValueString(),
		FirewallAlias: api.SelectedMap(d.FirewallAlias.ValueString()),
		Description:   d.Description.ValueString(),
	}, nil
}

func convertDomainStructToSchema(d *dnsmasq.Domain) (*domainResourceModel, error) {
	return &domainResourceModel{
		Sequence:      types.Int64Value(tools.StringToInt64(d.Sequence)),
		Domain:        types.StringValue(d.Domain),
		Server:        types.StringValue(d.Ip),
		Port:          types.StringValue(d.Port),
		SourceIp:      types.StringValue(d.SourceIp),
		FirewallAlias: types.StringValue(d.FirewallAlias.String()),
		Description:   types.StringValue(d.Description),
	}, nil
}

func mergeDomainWriteOnly(prior, read *domainResourceModel) {}
