package unbound

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/unbound"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dnsblResourceModel struct {
	Enabled     types.Bool   `tfsdk:"enabled"`
	Type        types.Set    `tfsdk:"type"`
	Lists       types.Set    `tfsdk:"lists"`
	Allowlists  types.Set    `tfsdk:"allowlists"`
	Blocklists  types.Set    `tfsdk:"blocklists"`
	Wildcards   types.Set    `tfsdk:"wildcards"`
	SourceNets  types.Set    `tfsdk:"source_nets"`
	Address     types.String `tfsdk:"address"`
	NXDomain    types.Bool   `tfsdk:"nxdomain"`
	CacheTTL    types.Int64  `tfsdk:"cache_ttl"`
	Description types.String `tfsdk:"description"`

	Id types.String `tfsdk:"id"`
}

func dnsblStringSet(desc string) schema.SetAttribute {
	return schema.SetAttribute{
		MarkdownDescription: desc + " Defaults to `[]`.",
		ElementType:         types.StringType,
		Optional:            true,
		Computed:            true,
		Default:             setdefault.StaticValue(tools.EmptySetValue(types.StringType)),
	}
}

func dnsblResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Unbound DNS blocklist policies (Services > Unbound DNS > Blocklist). Requires OPNsense 26.1 or later. " +
			"Each change is applied with `unbound/service/dnsbl`, which downloads the selected lists.",

		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable the usage of DNS blocklists. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"type":        dnsblStringSet("Select which kind of DNSBL you want to use. Predefined list identifiers, e.g. `hgz002` (HaGeZi Multi NORMAL), `oisd0` or `sb`."),
			"lists":       dnsblStringSet("List of domains from where blocklist will be downloaded. Each entry is a blocklist URL."),
			"allowlists":  dnsblStringSet("List of domains to allow. You can use regular expressions. This allow list only applies to blocklist matches on items in this policy."),
			"blocklists":  dnsblStringSet("List of domains to blocklist. Only exact matches are supported."),
			"wildcards":   dnsblStringSet("List of wildcard domains to blocklist. All subdomains of the given domain will be blocked. Blocking first-level domains is not supported."),
			"source_nets": dnsblStringSet("Source networks to apply policy on. Examples are `192.168.1.0/24` or `192.168.1.1`. Leave empty to apply on everything. All specified networks should use the same protocol family and have equal sizes to avoid priority issues."),
			"address": schema.StringAttribute{
				MarkdownDescription: "Destination ip address for entries in the blocklist (leave empty to use default: `0.0.0.0`). Not used when \"Return NXDOMAIN\" (`nxdomain`) is checked. IPv4 only. Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"nxdomain": schema.BoolAttribute{
				MarkdownDescription: "Use the DNS response code NXDOMAIN instead of a destination address. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"cache_ttl": schema.Int64Attribute{
				MarkdownDescription: "TTL for the blocklists cache. Remote blocklists don't usually update more often than once a day. Therefore, when blocklists are downloaded, they are cached locally to prevent unnecessary fetches over the internet. You can change this behavior here if you know the remote files rotate faster than this. In seconds. Defaults to `72000`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(72000),
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description, 1 to 255 characters. OPNsense requires it to keep track of which policies apply to which source networks.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the policy.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func convertDnsblSchemaToStruct(d *dnsblResourceModel) (*unbound.Blocklist, error) {
	return &unbound.Blocklist{
		Enabled:     tools.BoolToString(d.Enabled.ValueBool()),
		Type:        api.SelectedMapList(tools.SetToStringSlice(d.Type)),
		Lists:       api.SelectedMapList(tools.SetToStringSlice(d.Lists)),
		Allowlists:  api.SelectedMapList(tools.SetToStringSlice(d.Allowlists)),
		Blocklists:  api.SelectedMapList(tools.SetToStringSlice(d.Blocklists)),
		Wildcards:   api.SelectedMapList(tools.SetToStringSlice(d.Wildcards)),
		SourceNets:  api.SelectedMapList(tools.SetToStringSlice(d.SourceNets)),
		Address:     d.Address.ValueString(),
		NXDomain:    tools.BoolToString(d.NXDomain.ValueBool()),
		CacheTTL:    tools.Int64ToString(d.CacheTTL.ValueInt64()),
		Description: d.Description.ValueString(),
	}, nil
}

func convertDnsblStructToSchema(d *unbound.Blocklist) (*dnsblResourceModel, error) {
	return &dnsblResourceModel{
		Enabled:     types.BoolValue(tools.StringToBool(d.Enabled)),
		Type:        tools.StringSliceToSet(d.Type),
		Lists:       tools.StringSliceToSet(d.Lists),
		Allowlists:  tools.StringSliceToSet(d.Allowlists),
		Blocklists:  tools.StringSliceToSet(d.Blocklists),
		Wildcards:   tools.StringSliceToSet(d.Wildcards),
		SourceNets:  tools.StringSliceToSet(d.SourceNets),
		Address:     types.StringValue(d.Address),
		NXDomain:    types.BoolValue(tools.StringToBool(d.NXDomain)),
		CacheTTL:    types.Int64Value(tools.StringToInt64(d.CacheTTL)),
		Description: types.StringValue(d.Description),
	}, nil
}

func mergeDnsblWriteOnly(prior, read *dnsblResourceModel) {}
