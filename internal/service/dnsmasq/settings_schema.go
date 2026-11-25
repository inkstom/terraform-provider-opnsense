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
			"enabled":                  optionalBool("Enable Dnsmasq. Defaults to `false`.", false),
			"interfaces":               optionalStringSet("Interface IPs used to respond to queries from clients. If no interfaces are selected, Dnsmasq will listen on all available IPv4 and IPv6 addresses by default. However, DHCP related firewall rules will only be added for explicitly selected interfaces, never for all interfaces. Example `[\"lan\", \"opt1\"]`."),
			"strict_interface_binding": optionalBool("By default we bind the wildcard address, even when listening on some interfaces. Requests that should not be handled are discarded, this has the advantage of working even when interfaces come and go and change address. This option forces binding to only the interfaces we are listening on, which is less stable in non static environments. Defaults to `false`.", false),
			"port":                     optionalInt("The port used for responding to DNS queries. It should normally be left blank unless another service needs to bind to TCP/UDP port 53. Setting this to zero (`0`) completely disables DNS function, e.g. for DHCP only alongside Unbound.", int64validator.Between(0, 65535)),
			"dnssec":                   optionalBool("Validate DNSSEC. Defaults to `false`.", false),
			"no_hosts":                 optionalBool("Do not read hostnames in `/etc/hosts`. Defaults to `false`.", false),
			"expand_hosts":             optionalBool("Append the configured domain to simple hostnames read from hosts files. Defaults to `false`.", false),
			"log_queries":              optionalBool("Log the results of DNS queries. Defaults to `false`.", false),
			"dns_forward_max":          optionalInt("Set the maximum number of concurrent DNS queries. On configurations with tight resources, this value may need to be reduced.", int64validator.AtLeast(0)),
			"cache_size":               optionalInt("Set the size of the cache. Setting the cache size to zero disables caching. Please note that huge cache size impacts performance.", int64validator.AtLeast(0)),
			"local_ttl":                optionalInt("This option allows a time-to-live (in seconds) to be given for local DNS entries, i.e. `/etc/hosts` or DHCP leases. This will reduce the load on the server at the expense of clients using stale data under some circumstances. A value of zero will disable client-side caching.", int64validator.AtLeast(0)),
			"no_ident":                 optionalBool("Do not respond to class CHAOS and type TXT in domain bind queries. Without this option being set, the cache statistics are also available in the DNS as answers to queries of class CHAOS and type TXT in domain bind. Defaults to `true`.", true),
			"strict_order":             optionalBool("If this option is set, we will query the DNS servers sequentially in the order specified (System: General Setup: DNS Servers), rather than all at once in parallel. Defaults to `false`.", false),
			"domain_needed":            optionalBool("If this option is set, we will not forward A or AAAA queries for plain names, without dots or domain parts, to upstream name servers. If the name is not known from `/etc/hosts` or DHCP then a \"not found\" answer is returned. Defaults to `false`.", false),
			"no_resolv":                optionalBool("If this option is set, DNS forwarding to system nameservers (defined in System: General Setup: DNS Servers) will be disabled. Upstream servers defined in Services: Dnsmasq DNS & DHCP: Domains will still be used. This option is recommended when Unbound forwards local domain queries to Dnsmasq, so that all queries terminate without further lookups if they are unknown. Defaults to `false`.", false),
			"no_private_reverse":       optionalBool("If this option is set, we will not forward reverse DNS lookups (PTR) for private addresses (RFC 1918) to upstream name servers. Any entries in the Domain Overrides section forwarding private `n.n.n.in-addr.arpa` names to a specific server are still forwarded. If the IP to name is not known from `/etc/hosts`, DHCP or a specific domain override then a \"not found\" answer is immediately returned. Defaults to `false`.", false),
			"add_mac": schema.StringAttribute{
				MarkdownDescription: "Add the MAC address of the requestor to DNS queries which are forwarded upstream. The MAC address will only be added if the upstream DNS Server is in the same subnet as the requestor. Since this is not standardized, it should be considered experimental. This is useful for selective DNS filtering on the upstream DNS server. One of `standard`, `base64`, `text`, or empty to disable. Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				Validators:          []validator.String{stringvalidator.OneOf("", "standard", "base64", "text")},
			},
			"add_subnet":                    optionalBool("Add the real client IPv4 and IPv6 addresses (`add-subnet=32,128`) to DNS queries which are forwarded upstream. Be careful setting this option as it can undermine privacy. This is useful for selective DNS filtering on the upstream DNS server. Defaults to `false`.", false),
			"strip_subnet":                  optionalBool("Strip the subnet received by a downstream DNS server. If \"Add subnet\" (`add_subnet`) is used and the downstream DNS server already added a subnet, Dnsmasq will not replace it without setting \"Strip subnet\". Defaults to `false`.", false),
			"register_dhcp_leases":          optionalBool("If this option is set, then machines that specify their hostname when requesting a DHCP lease will be registered, so that their name can be resolved. Applies to ISC DHCPv4 leases (legacy). Defaults to `false`.", false),
			"dhcp_domain_override":          optionalString("The domain name to use for DHCP hostname registration. If empty, the default system domain is used. Note that all DHCP leases will be assigned to the same domain. If this is undesired, static DHCP lease registration is able to provide coherent mappings. Applies to ISC DHCP (legacy)."),
			"register_dhcp_static_mappings": optionalBool("If this option is set, then DHCP static mappings will be registered, so that their name can be resolved. Applies to ISC DHCP (legacy). Defaults to `false`.", false),
			"prefer_dhcp":                   optionalBool("If this option is set, then DHCP mappings will be resolved before the manual list of names below. This only affects the name given for a reverse lookup (PTR). Applies to ISC DHCP (legacy). Defaults to `false`.", false),

			"dhcp_no_interfaces":    optionalStringSet("Do not provide DHCP, TFTP or router advertisement on the specified interfaces, but do provide DNS service. Please note that Dnsmasq continues to listen on the default DHCP ports as long as any DHCP ranges are configured; setting this option only ignores these packets on the selected interfaces."),
			"dhcp_fqdn":             optionalBool("If disabled, registers the unqualified names of DHCP clients into the DNS (e.g. `smartphone`), in which case they should be unique. If enabled, the qualified name (e.g. `smartphone.lan.internal`) is registered. This option must be enabled if you are forwarding from Unbound to Dnsmasq for specific local domains. Defaults to `true`.", true),
			"dhcp_domain":           optionalString("To ensure that all names have a domain part, there must be a default domain specified when DHCP FQDN is set. Leave empty to use the system domain."),
			"dhcp_local":            optionalBool("Sets all DHCP domains as local. This will configure this DNS server as authoritative; it will not forward queries to any upstream servers for these domains. Defaults to `true`.", true),
			"dhcp_lease_max":        optionalInt("Limits Dnsmasq to the specified maximum number of DHCP leases. This limit is to prevent DoS attacks from hosts which create thousands of leases and use lots of memory in the Dnsmasq process.", int64validator.AtLeast(0)),
			"dhcp_authoritative":    optionalBool("Should be set when Dnsmasq is definitely the only DHCP server on a network. For DHCPv4, it changes the behaviour from strict RFC compliance so that DHCP requests on unknown leases from unknown hosts are not ignored. Defaults to `false`.", false),
			"dhcp_reply_delay":      optionalInt("Delays sending DHCPOFFER and PROXYDHCP replies for at least the specified number of seconds. This can be practical for split DHCP solutions, to make sure the secondary server answers slower than the primary. Must be between `0` and `60`.", int64validator.Between(0, 60)),
			"dhcp_default_fw_rules": optionalBool("Automatically register firewall rules to allow DHCP traffic for all explicitly selected interfaces, can be disabled for more fine grained control if needed. Defaults to `true`.", true),
			"dhcp_enable_ra":        optionalBool("Setting this will enable Router Advertisements for all configured DHCPv6 ranges with the managed address bits set, and the use SLAAC bit reset. To change this default, select a combination of the possible options in the individual DHCPv6 ranges. Keep in mind that this is a global option; if there are configured DHCPv6 ranges, RAs will be sent unconditionally and cannot be deactivated selectively. Setting Router Advertisement modes in DHCPv6 ranges will have no effect without this global option enabled. Defaults to `false`.", false),
			"dhcp_host_ping":        optionalBool("By default, the DHCP server will use a ping to ensure that an address is not in use before allocating it to a host. Defaults to `true`.", true),
			"dhcp_disable_ha_sync":  optionalBool("Ignore the DHCP general settings from being updated using HA sync. Defaults to `false`.", false),
			"dhcp_log_dhcp":         optionalBool("Extra logging for DHCP, log all the options sent to DHCP clients and the tags used to determine them. Defaults to `false`.", false),
			"dhcp_log_quiet":        optionalBool("Suppress logging of the routine operation of DHCP, RA and TFTP. Errors and problems will still be logged. Defaults to `false`.", false),

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
