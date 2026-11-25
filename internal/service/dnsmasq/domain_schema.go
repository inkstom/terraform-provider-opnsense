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
				MarkdownDescription: "Sort with a sequence number, e.g., for strict processing order when using the \"Query DNS servers sequentially\" option in general settings. Must be between `1` and `99999`. Defaults to `1`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
			},
			"domain":         schema.StringAttribute{MarkdownDescription: "Domain to override (NOTE: this does not have to be a valid TLD!). Example: `corp.example.com`.", Required: true},
			"server":         schema.StringAttribute{MarkdownDescription: "IP address of the authoritative DNS server for this domain, leave empty to prevent lookups for this domain. Set to `\"\"` to prevent lookups.", Required: true},
			"port":           str("Specify a non standard port number here, leave blank for default (`53`)."),
			"source_ip":      str("Source IP address for queries to the DNS server for the override domain. Best to leave empty."),
			"firewall_alias": str("Choose an \"external (advanced)\" type alias from \"Firewall - Aliases\". Whenever a client successfully resolves the domain, the resolved IP addresses will be automatically added to the chosen alias. Adding a domain will also add all IP addresses of resolved subdomains. Please note that DNS record TTL is not evaluated; once an IP address is added, it will stay permanently, or until manually flushed in \"Firewall - Diagnostics - Aliases\", or until removed automatically when setting an expiration on the alias. UUID of an `opnsense_firewall_alias` of an external type."),
			"description":    str("You may enter a description here for your reference (not parsed)."),
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
			"sequence":       dschema.Int64Attribute{MarkdownDescription: "Sort with a sequence number, e.g., for strict processing order when using the \"Query DNS servers sequentially\" option in general settings.", Computed: true},
			"domain":         s("Domain to override (NOTE: this does not have to be a valid TLD!)."),
			"server":         s("IP address of the authoritative DNS server for this domain, leave empty to prevent lookups for this domain."),
			"port":           s("Specify a non standard port number here, leave blank for default."),
			"source_ip":      s("Source IP address for queries to the DNS server for the override domain."),
			"firewall_alias": s("UUID of the external type firewall alias that resolved IP addresses for the domain are added to."),
			"description":    s("You may enter a description here for your reference (not parsed)."),
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
