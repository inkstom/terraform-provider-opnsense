package dyndns

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/dyndns"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type accountResourceModel struct {
	Enabled        types.Bool   `tfsdk:"enabled"`
	Service        types.String `tfsdk:"service"`
	Protocol       types.String `tfsdk:"protocol"`
	Server         types.String `tfsdk:"server"`
	Username       types.String `tfsdk:"username"`
	Password       types.String `tfsdk:"password"`
	ResourceId     types.String `tfsdk:"resource_id"`
	Hostnames      types.Set    `tfsdk:"hostnames"`
	Wildcard       types.Bool   `tfsdk:"wildcard"`
	Zone           types.String `tfsdk:"zone"`
	Checkip        types.String `tfsdk:"checkip"`
	DynIpv6Host    types.String `tfsdk:"dynipv6host"`
	CheckipTimeout types.Int64  `tfsdk:"checkip_timeout"`
	ForceSsl       types.Bool   `tfsdk:"force_ssl"`
	Ttl            types.Int64  `tfsdk:"ttl"`
	Interface      types.String `tfsdk:"interface"`
	Description    types.String `tfsdk:"description"`

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

func accountResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Dynamic DNS accounts (Services > Dynamic DNS). Requires the `os-ddclient` plugin. " +
			"The `service` and `checkip` values offered depend on the backend chosen in the plugin's general settings (`native` or `ddclient`).",

		Attributes: map[string]schema.Attribute{
			"enabled": optionalBool("Enable this account. Defaults to `true`.", true),
			"service": schema.StringAttribute{
				MarkdownDescription: "Provider key, e.g. `cloudflare`, `duckdns`, `freemyip` or `custom`.",
				Required:            true,
			},
			"protocol": optionalString("Protocol for the `custom` service: `dyndns1`, `dyndns2`, `get`, `post` or `put`."),
			"server":   optionalString("Server (or URL) for the `custom` service."),
			"username": optionalString("Username."),
			"password": schema.StringAttribute{
				MarkdownDescription: "Password or API token. OPNsense never returns it, so Terraform keeps the value it last wrote, " +
					"and an empty value leaves the stored password unchanged. After an import, the first apply writes it again. Defaults to `\"\"`.",
				Optional:  true,
				Computed:  true,
				Sensitive: true,
				Default:   stringdefault.StaticString(""),
			},
			"resource_id": optionalString("Resource ID, for services that need one."),
			"hostnames": schema.SetAttribute{
				MarkdownDescription: "Hostnames to update.",
				ElementType:         types.StringType,
				Required:            true,
				Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			},
			"wildcard": optionalBool("Also update the wildcard record. Defaults to `false`.", false),
			"zone":     optionalString("DNS zone, for services that need one."),
			"checkip": schema.StringAttribute{
				MarkdownDescription: "How to find the current IP address, e.g. `web_icanhazip`, `web_cloudflare` or `if` (read it from `interface`).",
				Required:            true,
			},
			"dynipv6host": optionalString("Partial IPv6 host address appended to the prefix, e.g. `::1000`."),
			"checkip_timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout in seconds for the IP check, 10 to 60. Defaults to `10`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(10),
				Validators:          []validator.Int64{int64validator.Between(10, 60)},
			},
			"force_ssl": optionalBool("Use HTTPS for updates. Defaults to `true`.", true),
			"ttl": schema.Int64Attribute{
				MarkdownDescription: "Record TTL in seconds, for services that set one. Defaults to `300`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(300),
				Validators:          []validator.Int64{int64validator.Between(1, 604800)},
			},
			"interface":   optionalString("Interface to read the address from (required when `checkip` is `if` or `if6`), e.g. `wan`."),
			"description": optionalString("Description."),
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the account.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func convertAccountSchemaToStruct(d *accountResourceModel) (*dyndns.Account, error) {
	return &dyndns.Account{
		Enabled:        tools.BoolToString(d.Enabled.ValueBool()),
		Service:        api.SelectedMap(d.Service.ValueString()),
		Protocol:       api.SelectedMap(d.Protocol.ValueString()),
		Server:         d.Server.ValueString(),
		Username:       d.Username.ValueString(),
		Password:       d.Password.ValueString(),
		ResourceId:     d.ResourceId.ValueString(),
		Hostnames:      api.SelectedMapList(tools.SetToStringSlice(d.Hostnames)),
		Wildcard:       tools.BoolToString(d.Wildcard.ValueBool()),
		Zone:           d.Zone.ValueString(),
		Checkip:        api.SelectedMap(d.Checkip.ValueString()),
		DynIpv6Host:    d.DynIpv6Host.ValueString(),
		CheckipTimeout: tools.Int64ToString(d.CheckipTimeout.ValueInt64()),
		ForceSsl:       tools.BoolToString(d.ForceSsl.ValueBool()),
		Ttl:            tools.Int64ToString(d.Ttl.ValueInt64()),
		Interface:      api.SelectedMap(d.Interface.ValueString()),
		Description:    d.Description.ValueString(),
	}, nil
}

func convertAccountStructToSchema(d *dyndns.Account) (*accountResourceModel, error) {
	return &accountResourceModel{
		Enabled:        types.BoolValue(tools.StringToBool(d.Enabled)),
		Service:        types.StringValue(d.Service.String()),
		Protocol:       types.StringValue(d.Protocol.String()),
		Server:         types.StringValue(d.Server),
		Username:       types.StringValue(d.Username),
		Password:       types.StringValue(d.Password),
		ResourceId:     types.StringValue(d.ResourceId),
		Hostnames:      tools.StringSliceToSet(d.Hostnames),
		Wildcard:       types.BoolValue(tools.StringToBool(d.Wildcard)),
		Zone:           types.StringValue(d.Zone),
		Checkip:        types.StringValue(d.Checkip.String()),
		DynIpv6Host:    types.StringValue(d.DynIpv6Host),
		CheckipTimeout: types.Int64Value(tools.StringToInt64(d.CheckipTimeout)),
		ForceSsl:       types.BoolValue(tools.StringToBool(d.ForceSsl)),
		Ttl:            types.Int64Value(tools.StringToInt64(d.Ttl)),
		Interface:      types.StringValue(d.Interface.String()),
		Description:    types.StringValue(d.Description),
	}, nil
}

// mergeAccountWriteOnly keeps the password from prior state: GET never returns
// it. On import the prior value is null, so the read "" stands.
func mergeAccountWriteOnly(prior, read *accountResourceModel) {
	if prior != nil && !prior.Password.IsNull() {
		read.Password = prior.Password
	}
}
