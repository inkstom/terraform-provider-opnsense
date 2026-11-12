package dnsmasq

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/dnsmasq"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const settingsID = "dnsmasq_settings"

type settingsResourceModel struct {
	Enabled                types.Bool   `tfsdk:"enabled"`
	Interfaces             types.Set    `tfsdk:"interfaces"`
	StrictInterfaceBinding types.Bool   `tfsdk:"strict_interface_binding"`
	Port                   types.Int64  `tfsdk:"port"`
	DNSSEC                 types.Bool   `tfsdk:"dnssec"`
	NoHosts                types.Bool   `tfsdk:"no_hosts"`
	ExpandHosts            types.Bool   `tfsdk:"expand_hosts"`
	LogQueries             types.Bool   `tfsdk:"log_queries"`
	DNSForwardMax          types.Int64  `tfsdk:"dns_forward_max"`
	CacheSize              types.Int64  `tfsdk:"cache_size"`
	LocalTTL               types.Int64  `tfsdk:"local_ttl"`
	NoIdent                types.Bool   `tfsdk:"no_ident"`
	StrictOrder            types.Bool   `tfsdk:"strict_order"`
	DomainNeeded           types.Bool   `tfsdk:"domain_needed"`
	NoResolv               types.Bool   `tfsdk:"no_resolv"`
	NoPrivateReverse       types.Bool   `tfsdk:"no_private_reverse"`
	AddMac                 types.String `tfsdk:"add_mac"`
	AddSubnet              types.Bool   `tfsdk:"add_subnet"`
	StripSubnet            types.Bool   `tfsdk:"strip_subnet"`
	RegisterDHCPLeases     types.Bool   `tfsdk:"register_dhcp_leases"`
	DHCPDomainOverride     types.String `tfsdk:"dhcp_domain_override"`
	RegisterDHCPStatic     types.Bool   `tfsdk:"register_dhcp_static_mappings"`
	PreferDHCP             types.Bool   `tfsdk:"prefer_dhcp"`

	DHCPNoInterfaces   types.Set    `tfsdk:"dhcp_no_interfaces"`
	DHCPFQDN           types.Bool   `tfsdk:"dhcp_fqdn"`
	DHCPDomain         types.String `tfsdk:"dhcp_domain"`
	DHCPLocal          types.Bool   `tfsdk:"dhcp_local"`
	DHCPLeaseMax       types.Int64  `tfsdk:"dhcp_lease_max"`
	DHCPAuthoritative  types.Bool   `tfsdk:"dhcp_authoritative"`
	DHCPReplyDelay     types.Int64  `tfsdk:"dhcp_reply_delay"`
	DHCPDefaultFWRules types.Bool   `tfsdk:"dhcp_default_fw_rules"`
	DHCPEnableRA       types.Bool   `tfsdk:"dhcp_enable_ra"`
	DHCPHostPing       types.Bool   `tfsdk:"dhcp_host_ping"`
	DHCPDisableHASync  types.Bool   `tfsdk:"dhcp_disable_ha_sync"`
	DHCPLogDHCP        types.Bool   `tfsdk:"dhcp_log_dhcp"`
	DHCPLogQuiet       types.Bool   `tfsdk:"dhcp_log_quiet"`

	Id types.String `tfsdk:"id"`
}

func settingsResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Global dnsmasq DNS and DHCP settings (Services: Dnsmasq DNS & DHCP: General). " +
			"This is a singleton: import it with ID `" + settingsID + "` before managing it. " +
			"Destroying it only removes it from state; the firewall keeps its settings.",
		Attributes: map[string]schema.Attribute{
			"enabled":                  optionalBool("Enable dnsmasq. Defaults to `false`.", false),
			"interfaces":               optionalStringSet("Interfaces dnsmasq listens on, e.g. `[\"lan\", \"opt1\"]`. Empty means all."),
			"strict_interface_binding": optionalBool("Bind only to the listed interfaces. Defaults to `false`.", false),
			"port":                     optionalInt("DNS listen port. `0` disables DNS (DHCP-only, e.g. alongside Unbound).", int64validator.Between(0, 65535)),
			"dnssec":                   optionalBool("Validate DNSSEC. Defaults to `false`.", false),
			"no_hosts":                 optionalBool("Don't read /etc/hosts. Defaults to `false`.", false),
			"expand_hosts":             optionalBool("Add the domain to simple names in hosts files. Defaults to `false`.", false),
			"log_queries":              optionalBool("Log DNS queries. Defaults to `false`.", false),
			"dns_forward_max":          optionalInt("Maximum concurrent DNS queries.", int64validator.AtLeast(0)),
			"cache_size":               optionalInt("DNS cache size.", int64validator.AtLeast(0)),
			"local_ttl":                optionalInt("TTL for local answers.", int64validator.AtLeast(0)),
			"no_ident":                 optionalBool("Don't answer version/author queries. Defaults to `true`.", true),
			"strict_order":             optionalBool("Query upstream servers in order. Defaults to `false`.", false),
			"domain_needed":            optionalBool("Don't forward plain names without a domain. Defaults to `false`.", false),
			"no_resolv":                optionalBool("Don't forward to the system DNS servers. Defaults to `false`.", false),
			"no_private_reverse":       optionalBool("Don't forward private reverse lookups. Defaults to `false`.", false),
			"add_mac": schema.StringAttribute{
				MarkdownDescription: "Add the client MAC to forwarded queries: `standard`, `base64`, `text` or empty. Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				Validators:          []validator.String{stringvalidator.OneOf("", "standard", "base64", "text")},
			},
			"add_subnet":                    optionalBool("Add the client subnet to forwarded queries. Defaults to `false`.", false),
			"strip_subnet":                  optionalBool("Strip client subnet info from incoming queries. Defaults to `false`.", false),
			"register_dhcp_leases":          optionalBool("Register ISC DHCP leases (legacy). Defaults to `false`.", false),
			"dhcp_domain_override":          optionalString("Domain for registered ISC DHCP leases (legacy)."),
			"register_dhcp_static_mappings": optionalBool("Register ISC DHCP static mappings (legacy). Defaults to `false`.", false),
			"prefer_dhcp":                   optionalBool("Prefer DHCP registrations over host overrides (legacy). Defaults to `false`.", false),

			"dhcp_no_interfaces":    optionalStringSet("Interfaces where DHCP is disabled while DNS stays on."),
			"dhcp_fqdn":             optionalBool("Register clients as name.domain. Defaults to `true`.", true),
			"dhcp_domain":           optionalString("Default DHCP domain. Empty uses the system domain."),
			"dhcp_local":            optionalBool("Treat the DHCP domain as local (never forwarded). Defaults to `true`.", true),
			"dhcp_lease_max":        optionalInt("Maximum number of leases.", int64validator.AtLeast(0)),
			"dhcp_authoritative":    optionalBool("Act as the authoritative DHCP server. Defaults to `false`.", false),
			"dhcp_reply_delay":      optionalInt("Seconds to delay DHCP replies.", int64validator.Between(0, 60)),
			"dhcp_default_fw_rules": optionalBool("Add the automatic firewall rules for DHCP. Defaults to `true`.", true),
			"dhcp_enable_ra":        optionalBool("Send IPv6 router advertisements. Defaults to `false`.", false),
			"dhcp_host_ping":        optionalBool("Ping an address before leasing it. Defaults to `true`.", true),
			"dhcp_disable_ha_sync":  optionalBool("Exclude DHCP settings from HA sync. Defaults to `false`.", false),
			"dhcp_log_dhcp":         optionalBool("Log extra DHCP details. Defaults to `false`.", false),
			"dhcp_log_quiet":        optionalBool("Suppress routine DHCP logging. Defaults to `false`.", false),

			"id": schema.StringAttribute{
				MarkdownDescription: "Always `" + settingsID + "`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func convertSettingsSchemaToStruct(d *settingsResourceModel) *dnsmasq.GeneralSettings {
	b := tools.BoolToString
	return &dnsmasq.GeneralSettings{
		IsEnabled:                               b(d.Enabled.ValueBool()),
		Interface:                               tools.SetToStringSlice(d.Interfaces),
		StrictInterfaceBinding:                  b(d.StrictInterfaceBinding.ValueBool()),
		DNS_Port:                                intToAPI(d.Port),
		DNS_ExpandHosts:                         b(d.ExpandHosts.ValueBool()),
		DNS_DNSSEC:                              b(d.DNSSEC.ValueBool()),
		DNS_NoHostLookup:                        b(d.NoHosts.ValueBool()),
		DNS_LogDnsQueries:                       b(d.LogQueries.ValueBool()),
		DNS_MaxConcurrentQueries:                intToAPI(d.DNSForwardMax),
		DNS_CacheSize:                           intToAPI(d.CacheSize),
		DNS_LocalEntryTTL:                       intToAPI(d.LocalTTL),
		DNS_NoIdent:                             b(d.NoIdent.ValueBool()),
		DNS_QF_QuerySequentially:                b(d.StrictOrder.ValueBool()),
		DNS_QF_RequireDomain:                    b(d.DomainNeeded.ValueBool()),
		DNS_QF_DoNotForwardSystemDNS:            b(d.NoResolv.ValueBool()),
		DNS_QF_DoNotForwardPrivateReverseLookup: b(d.NoPrivateReverse.ValueBool()),
		DNS_QF_AddMac:                           api.SelectedMap(d.AddMac.ValueString()),
		DNS_QF_AddSubnet:                        b(d.AddSubnet.ValueBool()),
		DNS_QF_StripSubnet:                      b(d.StripSubnet.ValueBool()),
		Legacy_RegisterISCDhcp4Leases:           b(d.RegisterDHCPLeases.ValueBool()),
		Legacy_DhcpDomainOverride:               d.DHCPDomainOverride.ValueString(),
		Legacy_RegisterDhcpStaticMappings:       b(d.RegisterDHCPStatic.ValueBool()),
		Legacy_PreferDhcp:                       b(d.PreferDHCP.ValueBool()),
		DHCPSettings: dnsmasq.GeneralDHCPSettings{
			InterfaceNoDhcp:       tools.SetToStringSlice(d.DHCPNoInterfaces),
			FQDN:                  b(d.DHCPFQDN.ValueBool()),
			DefaultDomain:         d.DHCPDomain.ValueString(),
			LocalDomain:           b(d.DHCPLocal.ValueBool()),
			MaxLeases:             intToAPI(d.DHCPLeaseMax),
			Authoritative:         b(d.DHCPAuthoritative.ValueBool()),
			ReplyDelay:            intToAPI(d.DHCPReplyDelay),
			RegisterFirewallRules: b(d.DHCPDefaultFWRules.ValueBool()),
			RouterAdvertisements:  b(d.DHCPEnableRA.ValueBool()),
			HostPing:              b(d.DHCPHostPing.ValueBool()),
			DisableHASync:         b(d.DHCPDisableHASync.ValueBool()),
			LogDhcp:               b(d.DHCPLogDHCP.ValueBool()),
			LogQuiet:              b(d.DHCPLogQuiet.ValueBool()),
		},
	}
}

func convertSettingsStructToSchema(s *dnsmasq.GeneralSettings) *settingsResourceModel {
	b := tools.StringToBool
	h := s.DHCPSettings
	return &settingsResourceModel{
		Enabled:                types.BoolValue(b(s.IsEnabled)),
		Interfaces:             tools.StringSliceToSet(s.Interface),
		StrictInterfaceBinding: types.BoolValue(b(s.StrictInterfaceBinding)),
		Port:                   tools.StringToInt64Null(s.DNS_Port),
		DNSSEC:                 types.BoolValue(b(s.DNS_DNSSEC)),
		NoHosts:                types.BoolValue(b(s.DNS_NoHostLookup)),
		ExpandHosts:            types.BoolValue(b(s.DNS_ExpandHosts)),
		LogQueries:             types.BoolValue(b(s.DNS_LogDnsQueries)),
		DNSForwardMax:          tools.StringToInt64Null(s.DNS_MaxConcurrentQueries),
		CacheSize:              tools.StringToInt64Null(s.DNS_CacheSize),
		LocalTTL:               tools.StringToInt64Null(s.DNS_LocalEntryTTL),
		NoIdent:                types.BoolValue(b(s.DNS_NoIdent)),
		StrictOrder:            types.BoolValue(b(s.DNS_QF_QuerySequentially)),
		DomainNeeded:           types.BoolValue(b(s.DNS_QF_RequireDomain)),
		NoResolv:               types.BoolValue(b(s.DNS_QF_DoNotForwardSystemDNS)),
		NoPrivateReverse:       types.BoolValue(b(s.DNS_QF_DoNotForwardPrivateReverseLookup)),
		AddMac:                 types.StringValue(s.DNS_QF_AddMac.String()),
		AddSubnet:              types.BoolValue(b(s.DNS_QF_AddSubnet)),
		StripSubnet:            types.BoolValue(b(s.DNS_QF_StripSubnet)),
		RegisterDHCPLeases:     types.BoolValue(b(s.Legacy_RegisterISCDhcp4Leases)),
		DHCPDomainOverride:     types.StringValue(s.Legacy_DhcpDomainOverride),
		RegisterDHCPStatic:     types.BoolValue(b(s.Legacy_RegisterDhcpStaticMappings)),
		PreferDHCP:             types.BoolValue(b(s.Legacy_PreferDhcp)),
		DHCPNoInterfaces:       tools.StringSliceToSet(h.InterfaceNoDhcp),
		DHCPFQDN:               types.BoolValue(b(h.FQDN)),
		DHCPDomain:             types.StringValue(h.DefaultDomain),
		DHCPLocal:              types.BoolValue(b(h.LocalDomain)),
		DHCPLeaseMax:           tools.StringToInt64Null(h.MaxLeases),
		DHCPAuthoritative:      types.BoolValue(b(h.Authoritative)),
		DHCPReplyDelay:         tools.StringToInt64Null(h.ReplyDelay),
		DHCPDefaultFWRules:     types.BoolValue(b(h.RegisterFirewallRules)),
		DHCPEnableRA:           types.BoolValue(b(h.RouterAdvertisements)),
		DHCPHostPing:           types.BoolValue(b(h.HostPing)),
		DHCPDisableHASync:      types.BoolValue(b(h.DisableHASync)),
		DHCPLogDHCP:            types.BoolValue(b(h.LogDhcp)),
		DHCPLogQuiet:           types.BoolValue(b(h.LogQuiet)),
		Id:                     types.StringValue(settingsID),
	}
}
